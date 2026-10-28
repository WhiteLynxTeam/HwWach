package services

import (
	"HwWach/internal/authz"
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

// Категории (3.5): L1 общие для всех организаций — правит только админ организации платформы
// (PLATFORM_ORG_ID); L2 — своей организации; L3 — пользовательские, на модерацию админу организации
type CategoryService interface {
	Search(ctx context.Context, p *authz.Principal, query string) ([]*models.Category, error)
	Sync(ctx context.Context, p *authz.Principal, since *time.Time) (*dto.CategorySyncResponse, error)
	ResolveOrCreateCategory(ctx context.Context, p *authz.Principal, categoryUUID *uuid.UUID, categoryText *string) (*models.Category, error)
	ListPending(ctx context.Context, p *authz.Principal) ([]*models.Category, error)
	Moderate(ctx context.Context, p *authz.Principal, id uuid.UUID, req *dto.ModerateCategoryRequest) (*models.Category, error)

	// Admin methods
	ListAdmin(ctx context.Context, p *authz.Principal, level *int16, status *string, search string) ([]*dto.AdminCategoryResponse, error)
	GetAffectedAssets(ctx context.Context, p *authz.Principal, id uuid.UUID) (*dto.CategoryAffectedResponse, error)
	AdminCreate(ctx context.Context, p *authz.Principal, req *dto.AdminCreateCategoryRequest) (*dto.AdminCategoryResponse, error)
	AdminUpdate(ctx context.Context, p *authz.Principal, id uuid.UUID, req *dto.AdminUpdateCategoryRequest) (*dto.AdminCategoryResponse, error)
	AdminMerge(ctx context.Context, p *authz.Principal, sourceID uuid.UUID, req *dto.AdminMergeCategoryRequest) (*dto.AdminCategoryResponse, error)
}

type categoryService struct {
	categoryRepo  repository.CategoryRepo
	assetRepo     repository.AssetRepo
	platformOrgID uuid.UUID
}

func NewCategoryService(categoryRepo repository.CategoryRepo, assetRepo repository.AssetRepo, platformOrgID string) CategoryService {
	id, _ := uuid.Parse(platformOrgID)
	return &categoryService{
		categoryRepo:  categoryRepo,
		assetRepo:     assetRepo,
		platformOrgID: id,
	}
}

func (s *categoryService) isPlatform(p *authz.Principal) bool {
	return s.platformOrgID != uuid.Nil && p.OrgID == s.platformOrgID
}

// visible категория видна организации: общая L1 или своя
func visible(p *authz.Principal, c *models.Category) bool {
	return c.Level == 1 || (c.OrganizationID != nil && *c.OrganizationID == p.OrgID)
}

// editable админ организации правит свои категории, общие L1 — только организация платформы
func (s *categoryService) editable(p *authz.Principal, c *models.Category) bool {
	if c.Level == 1 {
		return s.isPlatform(p)
	}
	return c.OrganizationID != nil && *c.OrganizationID == p.OrgID
}

func (s *categoryService) orgForLevel(p *authz.Principal, level int16) (*uuid.UUID, error) {
	if level == 1 {
		if !s.isPlatform(p) {
			return nil, Forbidden("общие категории (уровень 1) правит только организация платформы")
		}
		return nil, nil
	}
	org := p.OrgID
	return &org, nil
}

func (s *categoryService) Search(ctx context.Context, p *authz.Principal, query string) ([]*models.Category, error) {
	return s.categoryRepo.Search(ctx, p.OrgID, query, p.AccountID, 20)
}

func (s *categoryService) Sync(ctx context.Context, p *authz.Principal, since *time.Time) (*dto.CategorySyncResponse, error) {
	categories, deletedUUIDs, err := s.categoryRepo.Sync(ctx, p.OrgID, p.AccountID, since)
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

func (s *categoryService) ResolveOrCreateCategory(ctx context.Context, p *authz.Principal, categoryUUID *uuid.UUID, categoryText *string) (*models.Category, error) {
	// 1. Если передан UUID категории
	if categoryUUID != nil && *categoryUUID != uuid.Nil {
		cat, err := s.categoryRepo.GetByUUID(ctx, *categoryUUID)
		if err == nil && cat != nil && visible(p, cat) {
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

	// Проверяем, существует ли уже категория с таким именем (L1, L2 организации или L3 пользователя)
	existing, err := s.categoryRepo.FindExactMatch(ctx, p.OrgID, normName, p.AccountID)
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
	org := p.OrgID
	newCat := &models.Category{
		UUID:              newID(),
		OrganizationID:    &org,
		Name:              cleanName,
		NormalizedName:    normName,
		Level:             3,
		Status:            models.CategoryStatusPending,
		CreatedBy:         &p.AccountID,
		CreatedByPersonID: &p.PersonID,
		UsageCount:        1,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.categoryRepo.Create(ctx, newCat); err != nil {
		// При возможном конфликте уникальности пробуем получить ещё раз
		retry, retryErr := s.categoryRepo.FindExactMatch(ctx, p.OrgID, normName, p.AccountID)
		if retryErr == nil && retry != nil {
			_ = s.categoryRepo.IncrementUsage(ctx, retry.UUID)
			return retry, nil
		}
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return newCat, nil
}

func (s *categoryService) ListPending(ctx context.Context, p *authz.Principal) ([]*models.Category, error) {
	return s.categoryRepo.ListPending(ctx, p.OrgID)
}

func (s *categoryService) Moderate(ctx context.Context, p *authz.Principal, id uuid.UUID, req *dto.ModerateCategoryRequest) (*models.Category, error) {
	cat, err := s.categoryRepo.GetByUUID(ctx, id)
	if err != nil || !s.editable(p, cat) {
		return nil, NotFound("category not found")
	}

	if cat.Level != 3 || cat.Status != models.CategoryStatusPending {
		return nil, BadRequest("only pending user categories (level 3) can be moderated")
	}

	switch req.Action {
	case "approve":
		cat.Status = models.CategoryStatusApproved
		level := int16(2) // По умолчанию переводим в Local (L2)
		if req.Level != nil && (*req.Level == 1 || *req.Level == 2) {
			level = *req.Level
		}
		org, err := s.orgForLevel(p, level)
		if err != nil {
			return nil, err
		}
		cat.Level = level
		cat.OrganizationID = org
		if req.AdminComment != nil {
			cat.AdminComment = *req.AdminComment
		}
		cat.UpdatedAt = time.Now()
		if err := s.categoryRepo.Update(ctx, cat); err != nil {
			return nil, err
		}

	case "merge":
		if req.MergedIntoUUID == nil || *req.MergedIntoUUID == "" {
			return nil, BadRequest("merged_into_uuid is required for merge action")
		}
		targetUUID, err := uuid.Parse(*req.MergedIntoUUID)
		if err != nil {
			return nil, BadRequest("invalid merged_into_uuid")
		}
		targetCat, err := s.categoryRepo.GetByUUID(ctx, targetUUID)
		if err != nil || targetCat.DeletedAt.Valid || !visible(p, targetCat) {
			return nil, BadRequest("target category not found or deleted")
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
		return nil, BadRequest(fmt.Sprintf("unknown moderation action: %s", req.Action))
	}

	return cat, nil
}

func adminResponse(c *models.Category, assets int64) *dto.AdminCategoryResponse {
	resp := &dto.AdminCategoryResponse{
		UUID:         c.UUID.String(),
		Name:         c.Name,
		Level:        c.Level,
		Status:       string(c.Status),
		UsageCount:   c.UsageCount,
		AdminComment: c.AdminComment,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.Format(time.RFC3339),
		AssetsCount:  assets,
	}
	if c.CreatedBy != nil {
		str := c.CreatedBy.String()
		resp.CreatedBy = &str
	}
	if c.CreatedByPersonID != nil {
		str := c.CreatedByPersonID.String()
		resp.CreatedByPersonID = &str
	}
	if c.MergedIntoUUID != nil {
		str := c.MergedIntoUUID.String()
		resp.MergedIntoUUID = &str
	}
	return resp
}

func (s *categoryService) ListAdmin(ctx context.Context, p *authz.Principal, level *int16, status *string, search string) ([]*dto.AdminCategoryResponse, error) {
	cats, err := s.categoryRepo.ListAdmin(ctx, p.OrgID, level, status, search)
	if err != nil {
		return nil, err
	}

	resps := make([]*dto.AdminCategoryResponse, 0, len(cats))
	for _, c := range cats {
		resps = append(resps, adminResponse(&c.Category, c.AssetsCount))
	}

	return resps, nil
}

func (s *categoryService) GetAffectedAssets(ctx context.Context, p *authz.Principal, id uuid.UUID) (*dto.CategoryAffectedResponse, error) {
	cat, err := s.categoryRepo.GetByUUID(ctx, id)
	if err != nil || !visible(p, cat) {
		return nil, NotFound("category not found")
	}

	count, err := s.assetRepo.CountAssetsByCategory(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	assets, err := s.assetRepo.GetAssetsByCategory(ctx, id, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sample assets: %w", err)
	}

	samples := make([]dto.SimpleAssetSample, 0, len(assets))
	for _, a := range assets {
		// пример — только активы своей организации
		if a.OrganizationID != p.OrgID {
			continue
		}
		samples = append(samples, dto.SimpleAssetSample{
			UUID:         a.UUID.String(),
			InventoryNum: a.InventoryNum,
			Name:         a.Name,
		})
	}

	return &dto.CategoryAffectedResponse{
		CategoryUUID: cat.UUID.String(),
		CategoryName: cat.Name,
		AssetsCount:  count,
		SampleAssets: samples,
	}, nil
}

func (s *categoryService) AdminCreate(ctx context.Context, p *authz.Principal, req *dto.AdminCreateCategoryRequest) (*dto.AdminCategoryResponse, error) {
	cleanName := strings.TrimSpace(req.Name)
	if cleanName == "" {
		return nil, BadRequest("название категории не может быть пустым")
	}
	if req.Level < 1 || req.Level > 2 {
		return nil, BadRequest("администратор может создавать категории только 1 (Global) или 2 (Local) уровня")
	}
	org, err := s.orgForLevel(p, req.Level)
	if err != nil {
		return nil, err
	}

	normName := strings.ToLower(cleanName)

	existing, err := s.categoryRepo.GetByNormalizedName(ctx, p.OrgID, normName, req.Level, nil)
	if err == nil && existing != nil && !existing.DeletedAt.Valid {
		return nil, Conflict(fmt.Sprintf("категория с названием «%s» уже существует на уровне %d", cleanName, req.Level))
	}

	cat := &models.Category{
		UUID:              newID(),
		OrganizationID:    org,
		Name:              cleanName,
		NormalizedName:    normName,
		Level:             req.Level,
		Status:            models.CategoryStatusApproved,
		CreatedBy:         &p.AccountID,
		CreatedByPersonID: &p.PersonID,
		UsageCount:        0,
		AdminComment:      req.AdminComment,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.categoryRepo.Create(ctx, cat); err != nil {
		return nil, fmt.Errorf("не удалось создать категорию: %w", err)
	}

	return adminResponse(cat, 0), nil
}

func (s *categoryService) AdminUpdate(ctx context.Context, p *authz.Principal, id uuid.UUID, req *dto.AdminUpdateCategoryRequest) (*dto.AdminCategoryResponse, error) {
	cat, err := s.categoryRepo.GetByUUID(ctx, id)
	if err != nil || cat.DeletedAt.Valid || !visible(p, cat) {
		return nil, NotFound("категория не найдена")
	}
	if !s.editable(p, cat) {
		return nil, Forbidden("общие категории (уровень 1) правит только организация платформы")
	}

	cleanName := strings.TrimSpace(req.Name)
	if cleanName == "" {
		return nil, BadRequest("название категории не может быть пустым")
	}
	normName := strings.ToLower(cleanName)

	if req.Level < 1 || req.Level > 3 {
		return nil, BadRequest("недопустимый уровень категории")
	}
	org, err := s.orgForLevel(p, req.Level)
	if err != nil {
		return nil, err
	}

	// Если изменилось имя или уровень, проверяем уникальность
	if normName != cat.NormalizedName || req.Level != cat.Level {
		existing, err := s.categoryRepo.GetByNormalizedName(ctx, p.OrgID, normName, req.Level, cat.CreatedBy)
		if err == nil && existing != nil && existing.UUID != cat.UUID && !existing.DeletedAt.Valid {
			return nil, Conflict(fmt.Sprintf("категория с названием «%s» уже существует на уровне %d", cleanName, req.Level))
		}
	}

	nameChanged := cleanName != cat.Name

	cat.Name = cleanName
	cat.NormalizedName = normName
	cat.Level = req.Level
	cat.OrganizationID = org
	if req.Status != nil && *req.Status != "" {
		cat.Status = models.CategoryStatus(*req.Status)
	}
	cat.AdminComment = req.AdminComment
	cat.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, cat); err != nil {
		return nil, fmt.Errorf("не удалось обновить категорию: %w", err)
	}

	// Если имя изменилось и запрошено обновление связанных полей
	if nameChanged && req.UpdateLinkedAssets {
		_ = s.assetRepo.UpdateCategoryName(ctx, cat.UUID, cleanName)
	}

	count, _ := s.assetRepo.CountAssetsByCategory(ctx, cat.UUID)
	return adminResponse(cat, count), nil
}

func (s *categoryService) AdminMerge(ctx context.Context, p *authz.Principal, sourceID uuid.UUID, req *dto.AdminMergeCategoryRequest) (*dto.AdminCategoryResponse, error) {
	targetUUID, err := uuid.Parse(req.TargetCategoryUUID)
	if err != nil {
		return nil, BadRequest("некорректный UUID целевой категории")
	}

	if sourceID == targetUUID {
		return nil, BadRequest("нельзя объединить категорию саму с собой")
	}

	source, err := s.categoryRepo.GetByUUID(ctx, sourceID)
	if err != nil || source.DeletedAt.Valid || !visible(p, source) {
		return nil, NotFound("исходная категория не найдена")
	}
	if !s.editable(p, source) {
		return nil, Forbidden("общие категории (уровень 1) правит только организация платформы")
	}

	target, err := s.categoryRepo.GetByUUID(ctx, targetUUID)
	if err != nil || target.DeletedAt.Valid || !visible(p, target) {
		return nil, NotFound("целевая категория не найдена")
	}

	if target.Status == models.CategoryStatusMerged {
		return nil, BadRequest("целевая категория уже объединена с другой категорией")
	}

	source.Status = models.CategoryStatusMerged
	source.MergedIntoUUID = &targetUUID
	if req.AdminComment != nil {
		source.AdminComment = *req.AdminComment
	}
	source.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(ctx, source); err != nil {
		return nil, fmt.Errorf("не удалось обновить статус категории: %w", err)
	}

	// Обновляем связанные поля в таблице assets, если запрошено
	if req.UpdateLinkedAssets {
		_ = s.assetRepo.ReassignCategory(ctx, source.UUID, target.UUID, target.Name)
	}

	count, _ := s.assetRepo.CountAssetsByCategory(ctx, source.UUID)
	return adminResponse(source, count), nil
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
