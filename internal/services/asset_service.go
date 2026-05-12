package services

import (
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"context"

	"github.com/google/uuid"
)

type AssetService interface {
	Create(ctx context.Context, asset *models.Asset) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Asset, error)
	GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error)
	ListPhotos(ctx context.Context, assetUUID uuid.UUID) ([]*models.Photo, error)
	ListRequests(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error)
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

func (s *assetService) GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error) {
	return s.assetRepo.GetAllByUserUUID(ctx, userUUID)
}

func (s *assetService) ListPhotos(ctx context.Context, assetUUID uuid.UUID) ([]*models.Photo, error) {
	return s.assetRepo.ListPhotos(ctx, assetUUID)
}

func (s *assetService) ListRequests(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error) {
	return s.assetRepo.ListRequests(ctx, assetUUID)
}
