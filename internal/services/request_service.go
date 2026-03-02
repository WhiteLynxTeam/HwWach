package services

import (
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"context"

	"github.com/google/uuid"
)

type RequestService interface {
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Request, error)
	ListByDeviceUUID(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Request, error)
	ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Request, error)
}

type requestService struct {
	requestRepo repository.RequestRepo
	deviceRepo  repository.DeviceRepo
	photoRepo   repository.PhotoRepo
}

func NewRequestService(
	requestRepo repository.RequestRepo,
	deviceRepo repository.DeviceRepo,
	photoRepo repository.PhotoRepo) RequestService {
	return &requestService{
		requestRepo: requestRepo,
		deviceRepo:  deviceRepo,
		photoRepo:   photoRepo,
	}
}

func (s *requestService) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Request, error) {
	return s.requestRepo.GetByUUID(ctx, uuid)
}

func (s *requestService) ListByDeviceUUID(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Request, error) {
	return s.requestRepo.ListByDeviceUUID(ctx, deviceUUID)
}

func (s *requestService) ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Request, error) {
	return s.requestRepo.ListByUserUUID(ctx, userUUID)
}
