package services

import (
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"context"

	"github.com/google/uuid"
)

type RequestService interface {
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Request, error)
	ListByAssetUUID(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error)
	ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Request, error)
}

type requestService struct {
	requestRepo repository.RequestRepo
	assetRepo  repository.AssetRepo
	photoRepo   repository.PhotoRepo
}

func NewRequestService(
	requestRepo repository.RequestRepo,
	assetRepo repository.AssetRepo,
	photoRepo repository.PhotoRepo) RequestService {
	return &requestService{
		requestRepo: requestRepo,
		assetRepo:  assetRepo,
		photoRepo:   photoRepo,
	}
}

func (s *requestService) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Request, error) {
	return s.requestRepo.GetByUUID(ctx, uuid)
}

func (s *requestService) ListByAssetUUID(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error) {
	return s.requestRepo.ListByAssetUUID(ctx, assetUUID)
}

func (s *requestService) ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Request, error) {
	return s.requestRepo.ListByUserUUID(ctx, userUUID)
}
