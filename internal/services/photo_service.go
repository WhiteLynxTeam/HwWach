package services

import (
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"HwWach/internal/storage"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PhotoService interface {
	CreatePendingPhoto(ctx context.Context, userUUID uuid.UUID, fileName string, fileSize int64, contentType string, clientID *uuid.UUID) (*models.Photo, error)
	CompletePhotoUpload(ctx context.Context, photoUUID uuid.UUID) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Photo, error)
	GetByClientID(ctx context.Context, clientID uuid.UUID) (*models.Photo, error)
	ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Photo, error)
	ListByDeviceUUID(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Photo, error)
	GetPresignedUploadURL(ctx context.Context, objectName, contentType string) (string, error)
	GetPublicURL(objectName string) string
}

type photoService struct {
	photoRepo  repository.PhotoRepo
	deviceRepo repository.DeviceRepo
	minioSVC   storage.Storage
}

func NewPhotoService(
	photoRepo repository.PhotoRepo,
	deviceRepo repository.DeviceRepo,
	minioSVC storage.Storage) PhotoService {
	return &photoService{
		photoRepo:  photoRepo,
		deviceRepo: deviceRepo,
		minioSVC:   minioSVC,
	}
}

func (s *photoService) CreatePendingPhoto(ctx context.Context, userUUID uuid.UUID, fileName string, fileSize int64, contentType string, clientID *uuid.UUID) (*models.Photo, error) {
	photoUUID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID v7: %w", err)
	}
	objectName := fmt.Sprintf("%s/%s_%s", userUUID.String(), photoUUID.String(), fileName)

	photo := &models.Photo{
		UUID:        photoUUID,
		UserUUID:    userUUID,
		URL:         objectName, // Сохраняем только относительный путь
		Status:      models.PhotoStatusPending,
		FileSize:    fileSize,
		FileName:    fileName,
		ContentType: contentType,
		ClientID:    clientID,
	}

	if err := s.photoRepo.Create(ctx, photo); err != nil {
		return nil, err
	}

	return photo, nil
}

func (s *photoService) CompletePhotoUpload(ctx context.Context, photoUUID uuid.UUID) error {
	return s.photoRepo.UpdateStatus(ctx, photoUUID, models.PhotoStatusCompleted)
}

func (s *photoService) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Photo, error) {
	return s.photoRepo.GetByUUID(ctx, uuid)
}

func (s *photoService) GetByClientID(ctx context.Context, clientID uuid.UUID) (*models.Photo, error) {
	return s.photoRepo.GetByClientID(ctx, clientID)
}

func (s *photoService) ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Photo, error) {
	return s.photoRepo.ListByUserUUID(ctx, userUUID)
}

func (s *photoService) ListByDeviceUUID(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Photo, error) {
	return s.photoRepo.ListByDeviceUUID(ctx, deviceUUID)
}

func (s *photoService) GetPresignedUploadURL(ctx context.Context, objectName, contentType string) (string, error) {
	// objectName теперь содержит чистый путь (user_id/photo_id_filename)
	return s.minioSVC.PresignedPutURL(ctx, objectName, contentType, 24*time.Hour)
}

func (s *photoService) GetPublicURL(objectName string) string {
	return s.minioSVC.GetPublicURL(objectName)
}
