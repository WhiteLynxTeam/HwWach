package repository

import (
	"HwWach/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PhotoRepo interface {
	Create(ctx context.Context, photo *models.Photo) error
	UpdateStatus(ctx context.Context, uuid uuid.UUID, status models.PhotoStatus) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Photo, error)
	GetByClientID(ctx context.Context, clientID uuid.UUID) (*models.Photo, error)
	ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Photo, error)
	ListByDeviceUUID(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Photo, error)
	Detach(ctx context.Context, photoUUID uuid.UUID) error
	Delete(ctx context.Context, uuid uuid.UUID) error
}

type photoRepo struct {
	db *gorm.DB
}

func NewPhotoRepo(db *gorm.DB) PhotoRepo {
	return &photoRepo{db: db}
}

func (p photoRepo) Create(ctx context.Context, photo *models.Photo) error {
	return p.db.WithContext(ctx).Create(photo).Error
}

func (p photoRepo) UpdateStatus(ctx context.Context, uuid uuid.UUID, status models.PhotoStatus) error {
	return p.db.WithContext(ctx).Model(&models.Photo{}).
		Where("uuid = ?", uuid).
		Update("status", status).Error
}

func (p photoRepo) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Photo, error) {
	var photo models.Photo
	if err := p.db.WithContext(ctx).First(&photo, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	return &photo, nil
}

func (p photoRepo) GetByClientID(ctx context.Context, clientID uuid.UUID) (*models.Photo, error) {
	var photo models.Photo
	if err := p.db.WithContext(ctx).Where("client_id = ?", clientID).First(&photo).Error; err != nil {
		return nil, err
	}
	return &photo, nil
}

func (p photoRepo) ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Photo, error) {
	var photos []*models.Photo
	if err := p.db.WithContext(ctx).Where("user_id = ?", userUUID).Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}

func (p photoRepo) ListByDeviceUUID(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Photo, error) {
	//TODO implement me
	panic("implement me")
}

func (p photoRepo) Detach(ctx context.Context, photoUUID uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (p photoRepo) Delete(ctx context.Context, uuid uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}
