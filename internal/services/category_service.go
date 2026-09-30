package services

import (
	"HwWach/internal/dto"
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CategoryService interface {
	Search(ctx context.Context, query string, userUUID uuid.UUID) ([]*models.Category, error)
	Sync(ctx context.Context, userUUID uuid.UUID, since *time.Time) (*dto.CategorySyncResponse, error)
	ResolveOrCreateCategory(ctx context.Context, categoryUUID *uuid.UUID, categoryText *string, userUUID uuid.UUID) (*models.Category, error)
	ListPending(ctx context.Context) ([]*models.Category, error)
	Moderate(ctx context.Context, id uuid.UUID, req *dto.ModerateCategoryRequest) (*models.Category, error)
}

type categoryService struct {
	categoryRepo repository.CategoryRepo
	assetRepo    repository.AssetRepo
}

func NewCategoryService(categoryRepo repository.CategoryRepo, assetRepo repository.AssetRepo) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
		assetRepo:    assetRepo,
	}
}

func (s *categoryService) Search(ctx context.Context, query string, userUUID uuid.UUID) ([]*models.Category, error) {
	return s.categoryRepo.Search(ctx, query, userUUID, 20)
}

func (s *categoryService) Sync(ctx context.Context, userUUID uuid.UUID, since *time.Time) (*dto.CategorySyncResponse, error) {
	categories, deletedUUIDs, err := s.categoryRepo.Sync(ctx, userUUID, since)
	if err != nil {
		return nil, err
	}

	categoryDTOs := make([]dto.CategoryResponse, 0, len(categories))
	for _, c := range categories {
		categoryDTOs = append(categoryDTOs, categoryToResponse(c))
	}

	deletedIDStrs := make([]string, 0, len(deletedUUIDs))
	for _, id := range deletedUUIDs {
		deletedIDStrs = append(deletedIDStrs, id.String())
	}

	return &dto.CategorySyncResponse{
		Categories: categoryDTOs,
		DeletedIDs: deletedIDStrs,
		SyncedAt:   time.Now().Format(time.RFC3339),
	}, nil
}

func (s *categoryService) ResolveOrCreateCategory(ctx context.Context, categoryUUID *uuid.UUID, categoryText *string, userUUID uuid.UUID) (*models.Category, error) {
	// 1. Если передан UUID категории
	if categoryUUID != nil && *categoryUUID != uuid.Nil {
		cat, err := s.categoryRepo.GetByUUID(ctx, *categoryUUID)
		if err == nil && cat != nil {
			// Если категория была объединена — резолвим целевую категорию
			if cat.Status == models.CategoryStatusMerged && cat.MergedIntoUUID != nil {
				targetCat, targetErr := s.categoryRepo.GetByUUID(ctx, *cat.MergedIntoUUID)
				if targetErr == nil && targetCat != nil {
					cat = targetCat
				}
			}

			// Если категория активна или принадлежит пользователю (L3)
			if !cat.DeletedAt.Valid && cat.Status != models.CategoryStatusRejected {
				_ = s.categoryRepo.IncrementUsage(ctx, cat.UUID)
				return cat, nil
			}
		}
		// Если по UUID не нашли или удалена, но есть categoryText — попробуем текстовый поиск/создание
		if categoryText == nil || strings.TrimSpace(*categoryText) == "" {
			return nil, errors.New("specified category was not found or is no longer available")
		}
	}

	// 2. Резолвинг по тексту категории
	if categoryText == nil || strings.TrimSpace(*categoryText) == "" {
		return nil, errors.New("category is required")
	}

	cleanName := strings.TrimSpace(*categoryText)
	normName := strings.ToLower(cleanName)

	// Проверяем, существует ли уже категория с таким именем (L1, L2 или L3 пользователя)
	existing, err := s.categoryRepo.FindExactMatch(ctx, normName, userUUID)
	if err == nil && existing != nil {
		if existing.Status == models.CategoryStatusMerged && existing.MergedIntoUUID != nil {
			target, targetErr := s.categoryRepo.GetByUUID(ctx, *existing.MergedIntoUUID)
			if targetErr == nil && target != nil {
				existing = target
			}
		}
		_ = s.categoryRepo.IncrementUsage(ctx, existing.UUID)
		return existing, nil
	}

	// Создаём новую категорию L3 (пользовательскую)
	newUUID, err := uuid.NewV7()
	if err != nil {
		newUUID = uuid.New()
	}

	newCat := &models.Category{
		UUID:           newUUID,
		Name:           cleanName,
		NormalizedName: normName,
		Level:          3,
		Status:         models.CategoryStatusPending,
		CreatedBy:      &userUUID,
		UsageCount:     1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.categoryRepo.Create(ctx, newCat); err != nil {
		// При возможном конфликте уникальности пробуем получить ещё раз
		retry, retryErr := s.categoryRepo.FindExactMatch(ctx, normName, userUUID)
		if retryErr == nil && retry != nil {
			_ = s.categoryRepo.IncrementUsage(ctx, retry.UUID)
			return retry, nil
		}
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return newCat, nil
}

func (s *categoryService) ListPending(ctx context.Context) ([]*models.Category, error) {
	return s.categoryRepo.ListPending(ctx)
}

func (s *categoryService) Moderate(ctx context.Context, id uuid.UUID, req *dto.ModerateCategoryRequest) (*models.Category, error) {
	cat, err := s.categoryRepo.GetByUUID(ctx, id)
	if err != nil {
		return nil, errors.New("category not found")
	}

	if cat.Level != 3 || cat.Status != models.CategoryStatusPending {
		return nil, errors.New("only pending user categories (level 3) can be moderated")
	}

	switch req.Action {
	case "approve":
		cat.Status = models.CategoryStatusApproved
		if req.Level != nil && (*req.Level == 1 || *req.Level == 2) {
			cat.Level = *req.Level
		} else {
			cat.Level = 2 // По умолчанию переводим в Local (L2)
		}
		if req.AdminComment != nil {
			cat.AdminComment = *req.AdminComment
		}
		cat.UpdatedAt = time.Now()
		if err := s.categoryRepo.Update(ctx, cat); err != nil {
			return nil, err
		}

	case "merge":
		if req.MergedIntoUUID == nil || *req.MergedIntoUUID == "" {
			return nil, errors.New("merged_into_uuid is required for merge action")
		}
		targetUUID, err := uuid.Parse(*req.MergedIntoUUID)
		if err != nil {
			return nil, errors.New("invalid merged_into_uuid")
		}
		targetCat, err := s.categoryRepo.GetByUUID(ctx, targetUUID)
		if err != nil || targetCat.DeletedAt.Valid {
			return nil, errors.New("target category not found or deleted")
		}

		cat.Status = models.CategoryStatusMerged
		cat.MergedIntoUUID = &targetUUID
		if req.AdminComment != nil {
			cat.AdminComment = *req.AdminComment
		}
		cat.UpdatedAt = time.Now()
		if err := s.categoryRepo.Update(ctx, cat); err != nil {
			return nil, err
		}

		// Перенаправляем все ассеты на целевую категорию
		_ = s.assetRepo.ReassignCategory(ctx, cat.UUID, targetCat.UUID, targetCat.Name)

	case "reject":
		cat.Status = models.CategoryStatusRejected
		if req.AdminComment != nil {
			cat.AdminComment = *req.AdminComment
		}
		cat.UpdatedAt = time.Now()
		if err := s.categoryRepo.Update(ctx, cat); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown moderation action: %s", req.Action)
	}

	return cat, nil
}

func categoryToResponse(c *models.Category) dto.CategoryResponse {
	resp := dto.CategoryResponse{
		UUID:         c.UUID.String(),
		Name:         c.Name,
		Level:        c.Level,
		Status:       string(c.Status),
		UsageCount:   c.UsageCount,
		AdminComment: c.AdminComment,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.Format(time.RFC3339),
	}
	if c.CreatedBy != nil {
		createdByStr := c.CreatedBy.String()
		resp.CreatedBy = &createdByStr
	}
	return resp
}
