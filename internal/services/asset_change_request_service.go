package services

import (
	"HwWach/internal/dto"
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type AssetChangeRequestService interface {
	CreateRequest(ctx context.Context, userUUID, assetUUID uuid.UUID, req dto.CreateChangeRequestInput) (*models.AssetChangeRequest, error)
	ApproveRequest(ctx context.Context, adminUUID, reqUUID uuid.UUID, req dto.ApproveChangeRequestInput) error
	ListPending(ctx context.Context) ([]*models.AssetChangeRequest, error)
}

type assetChangeRequestService struct {
	repo      repository.AssetChangeRequestRepo
	assetRepo repository.AssetRepo
}

func NewAssetChangeRequestService(repo repository.AssetChangeRequestRepo, assetRepo repository.AssetRepo) AssetChangeRequestService {
	return &assetChangeRequestService{
		repo:      repo,
		assetRepo: assetRepo,
	}
}

func (s *assetChangeRequestService) CreateRequest(ctx context.Context, userUUID, assetUUID uuid.UUID, req dto.CreateChangeRequestInput) (*models.AssetChangeRequest, error) {
	// 1. Проверяем, существует ли актив
	asset, err := s.assetRepo.GetByUUID(ctx, assetUUID)
	if err != nil {
		return nil, errors.New("asset not found")
	}

	// 2. Проверяем права пользователя
	if asset.UserUUID != userUUID {
		return nil, errors.New("you don't have permission to modify this asset")
	}

	// 3. Проверяем наличие уже существующей активной заявки
	existingReq, err := s.repo.GetPendingByAssetID(ctx, assetUUID)
	if err != nil {
		return nil, err
	}
	if existingReq != nil {
		return nil, errors.New("asset already has a pending change request")
	}

	var clientUUID *uuid.UUID
	if req.ClientID != nil {
		parsed, err := uuid.Parse(*req.ClientID)
		if err != nil {
			return nil, errors.New("invalid client_id")
		}
		clientUUID = &parsed
	}

	newReq := &models.AssetChangeRequest{
		ClientID:     clientUUID,
		AssetUUID:    assetUUID,
		UserUUID:     userUUID,
		RequestType:  models.RequestType(req.Type),
		ProposedData: []byte(req.ProposedData),
		Reason:       req.Reason,
		Status:       models.ModerationPending,
	}

	if err := s.repo.Create(ctx, newReq); err != nil {
		return nil, err
	}

	return newReq, nil
}

func (s *assetChangeRequestService) ApproveRequest(ctx context.Context, adminUUID, reqUUID uuid.UUID, req dto.ApproveChangeRequestInput) error {
	changeReq, err := s.repo.GetByUUID(ctx, reqUUID)
	if err != nil {
		return errors.New("request not found")
	}

	if changeReq.Status != models.ModerationPending {
		return errors.New("request is not in pending status")
	}

	changeReq.Status = models.ModerationStatus(req.Status)
	changeReq.AdminComment = req.AdminComment

	if changeReq.Status == models.ModerationApproved {
		// Применяем изменения к активу
		asset, err := s.assetRepo.GetByUUID(ctx, changeReq.AssetUUID)
		if err != nil {
			return errors.New("asset not found")
		}

		if changeReq.RequestType == models.RequestTypeUpdate {
			// Обновляем актив полями из заявки
			var updateData dto.UpdateAssetRequest
			if len(changeReq.ProposedData) > 0 {
				if err := json.Unmarshal(changeReq.ProposedData, &updateData); err != nil {
					return errors.New("failed to parse proposed data: " + err.Error())
				}

				if updateData.Name != nil {
					asset.Name = *updateData.Name
				}
				if updateData.Category != nil {
					asset.Category = *updateData.Category
				}
				if updateData.Description != nil {
					asset.Description = *updateData.Description
				}
				if updateData.AssetStatus != nil {
					asset.AssetStatus = models.AssetStatus(*updateData.AssetStatus)
				}
			}

			if err := s.assetRepo.Update(ctx, asset); err != nil {
				return errors.New("failed to update asset")
			}
		} else if changeReq.RequestType == models.RequestTypeDelete {
			if err := s.assetRepo.Delete(ctx, changeReq.AssetUUID); err != nil {
				return errors.New("failed to delete asset")
			}
		}
	}

	return s.repo.UpdateStatus(ctx, changeReq)
}

func (s *assetChangeRequestService) ListPending(ctx context.Context) ([]*models.AssetChangeRequest, error) {
	return s.repo.ListPending(ctx)
}
