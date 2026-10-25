package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"HwWach/internal/storage"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PhotoService interface {
	CreatePendingPhoto(ctx context.Context, p *authz.Principal, fileName string, fileSize int64, contentType string, clientID *uuid.UUID) (*models.Photo, error)
	CompletePhotoUpload(ctx context.Context, p *authz.Principal, photoUUID uuid.UUID) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Photo, error)
	GetByClientID(ctx context.Context, clientID uuid.UUID) (*models.Photo, error)
	ListMine(ctx context.Context, p *authz.Principal) ([]*models.Photo, error)
	GetPresignedUploadURL(ctx context.Context, objectName, contentType string) (string, error)
	GetPresignedDownloadURL(ctx context.Context, objectName string) (string, error)
	GetPublicURL(objectName string) string
}

type photoService struct {
	photoRepo repository.PhotoRepo
	minioSVC  storage.Storage
	orgs      OrgService
}

func NewPhotoService(photoRepo repository.PhotoRepo, minioSVC storage.Storage, orgs OrgService) PhotoService {
	return &photoService{
		photoRepo: photoRepo,
		minioSVC:  minioSVC,
		orgs:      orgs,
	}
}

func (s *photoService) CreatePendingPhoto(ctx context.Context, p *authz.Principal, fileName string, fileSize int64, contentType string, clientID *uuid.UUID) (*models.Photo, error) {
	photoUUID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID v7: %w", err)
	}
	// фото организации — в её папке: при окончательном удалении организации их проще найти
	objectName := fmt.Sprintf("%s/%s/%s_%s", p.OrgID.String(), p.AccountID.String(), photoUUID.String(), fileName)

	photo := &models.Photo{
		UUID:           photoUUID,
		OrganizationID: p.OrgID,
		AccountUUID:    p.AccountID,
		PersonUUID:     p.PersonID,
		URL:            objectName, // Сохраняем только относительный путь
		Status:         models.PhotoStatusPending,
		FileSize:       fileSize,
		FileName:       fileName,
		ContentType:    contentType,
		ClientID:       clientID,
	}

	if err := s.photoRepo.Create(ctx, photo); err != nil {
		return nil, err
	}

	return photo, nil
}

func (s *photoService) CompletePhotoUpload(ctx context.Context, p *authz.Principal, photoUUID uuid.UUID) error {
	if err := s.photoRepo.UpdateStatus(ctx, photoUUID, models.PhotoStatusCompleted); err != nil {
		return err
	}
	// первое фото — организация «наполнена» (8.6)
	s.orgs.ReportActivity(p.OrgID, true)
	return nil
}

func (s *photoService) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Photo, error) {
	return s.photoRepo.GetByUUID(ctx, uuid)
}

func (s *photoService) GetByClientID(ctx context.Context, clientID uuid.UUID) (*models.Photo, error) {
	return s.photoRepo.GetByClientID(ctx, clientID)
}

func (s *photoService) ListMine(ctx context.Context, p *authz.Principal) ([]*models.Photo, error) {
	return s.photoRepo.ListByAccount(ctx, p.OrgID, p.AccountID)
}

func (s *photoService) GetPresignedUploadURL(ctx context.Context, objectName, contentType string) (string, error) {
	return s.minioSVC.PresignedPutURL(ctx, objectName, contentType, 24*time.Hour)
}

func (s *photoService) GetPresignedDownloadURL(ctx context.Context, objectName string) (string, error) {
	return s.minioSVC.PresignedGetURL(ctx, objectName, 3*time.Hour)
}

func (s *photoService) GetPublicURL(objectName string) string {
	return s.minioSVC.GetPublicURL(objectName)
}
