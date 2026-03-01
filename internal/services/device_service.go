package services

import (
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"context"

	"github.com/google/uuid"
)

type DeviceService interface {
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Device, error)
	GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Device, error)
	ListPhotos(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Photo, error)
	ListRequests(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Request, error)
}

type deviceService struct {
	deviceRepo repository.DeviceRepo
	photoRepo  repository.PhotoRepo
}

func NewDeviceService(deviceRepo repository.DeviceRepo, photoRepo repository.PhotoRepo) DeviceService {
	return &deviceService{
		deviceRepo: deviceRepo,
		photoRepo:  photoRepo,
	}
}

func (s *deviceService) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Device, error) {
	return s.deviceRepo.GetByUUID(ctx, uuid)
}

func (s *deviceService) GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Device, error) {
	return s.deviceRepo.GetAllByUserUUID(ctx, userUUID)
}

func (s *deviceService) ListPhotos(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Photo, error) {
	return s.deviceRepo.ListPhotos(ctx, deviceUUID)
}

func (s *deviceService) ListRequests(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Request, error) {
	return s.deviceRepo.ListRequests(ctx, deviceUUID)
}
