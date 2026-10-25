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
	ListByAccount(ctx context.Context, org, accountID uuid.UUID) ([]*models.Photo, error)
	CountByOrg(ctx context.Context, org uuid.UUID) (int64, error)
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

func (p photoRepo) ListByAccount(ctx context.Context, org, accountID uuid.UUID) ([]*models.Photo, error) {
	var photos []*models.Photo
	if err := p.db.WithContext(ctx).Where("organization_id = ? AND account_id = ?", org, accountID).Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}

func (p photoRepo) CountByOrg(ctx context.Context, org uuid.UUID) (int64, error) {
	var n int64
	err := p.db.WithContext(ctx).Model(&models.Photo{}).Where("organization_id = ?", org).Count(&n).Error
	return n, err
}
