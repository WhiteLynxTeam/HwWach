package services

import (
	"HwWach/internal/dto"
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetService interface {
	Create(ctx context.Context, asset *models.Asset) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Asset, error)
	IsInventoryNumUnique(ctx context.Context, inventoryNum string) (bool, error)
	GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error)
	ListPhotos(ctx context.Context, assetUUID uuid.UUID) ([]*models.Photo, error)
	ListRequests(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error)
	UpdatePending(ctx context.Context, userUUID uuid.UUID, assetUUID uuid.UUID, req *dto.UpdateAssetRequest) (*models.Asset, error)
	GetPaginated(ctx context.Context, userUUID *uuid.UUID, page, limit int) ([]*models.Asset, int64, error)
}

type assetService struct {
	assetRepo repository.AssetRepo
	photoRepo repository.PhotoRepo
}

func NewAssetService(assetRepo repository.AssetRepo, photoRepo repository.PhotoRepo) AssetService {
	return &assetService{
		assetRepo: assetRepo,
		photoRepo: photoRepo,
	}
}

func (s *assetService) Create(ctx context.Context, asset *models.Asset) error {
	return s.assetRepo.Create(ctx, asset)
}

func (s *assetService) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Asset, error) {
	return s.assetRepo.GetByUUID(ctx, uuid)
}

func (s *assetService) IsInventoryNumUnique(ctx context.Context, inventoryNum string) (bool, error) {
	_, err := s.assetRepo.GetByInventoryNum(ctx, inventoryNum)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (s *assetService) GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error) {
	return s.assetRepo.GetAllByUserUUID(ctx, userUUID)
}

func (s *assetService) ListPhotos(ctx context.Context, assetUUID uuid.UUID) ([]*models.Photo, error) {
	return s.assetRepo.ListPhotos(ctx, assetUUID)
}

func (s *assetService) ListRequests(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error) {
	return s.assetRepo.ListRequests(ctx, assetUUID)
}

func (s *assetService) UpdatePending(ctx context.Context, userUUID uuid.UUID, assetUUID uuid.UUID, req *dto.UpdateAssetRequest) (*models.Asset, error) {
	asset, err := s.assetRepo.GetByUUID(ctx, assetUUID)
	if err != nil {
		return nil, err // В реальном приложении лучше проверять на NotFound
	}

	if asset.UserUUID != userUUID {
		return nil, errors.New("you don't have permission to modify this asset")
	}

	if asset.ModerationStatus != models.ModerationPending {
		return nil, errors.New("asset already processed by admin, use change request instead")
	}

	if req.InventoryNum != nil {
		asset.InventoryNum = *req.InventoryNum
	}
	if req.Name != nil {
		asset.Name = *req.Name
	}
	if req.Category != nil {
		asset.Category = *req.Category
	}
	if req.Description != nil {
		asset.Description = *req.Description
	}
	if req.AssetStatus != nil {
		asset.AssetStatus = models.AssetStatus(*req.AssetStatus)
	}

	if err := s.assetRepo.Update(ctx, asset); err != nil {
		return nil, err
	}

	return asset, nil
}

func (s *assetService) GetPaginated(ctx context.Context, userUUID *uuid.UUID, page, limit int) ([]*models.Asset, int64, error) {
	return s.assetRepo.GetPaginated(ctx, userUUID, page, limit)
}
