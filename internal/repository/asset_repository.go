package repository

import (
	"HwWach/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetRepo interface {
	Create(ctx context.Context, asset *models.Asset) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Asset, error)
	GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error)
	Update(ctx context.Context, asset *models.Asset) error
	Delete(ctx context.Context, uuid uuid.UUID) error
	UpdateStatus(ctx context.Context, uuid uuid.UUID, newStatus string) error
	ListPhotos(ctx context.Context, assetUUID uuid.UUID) ([]*models.Photo, error)
	ListRequests(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error)
}

type assetRepo struct {
	db *gorm.DB
}

func NewAssetRepo(db *gorm.DB) AssetRepo {
	return &assetRepo{db: db}
}

func (a assetRepo) Create(ctx context.Context, asset *models.Asset) error {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return err
	}
	asset.UUID = uuidV7
	return a.db.WithContext(ctx).Create(asset).Error
}

func (a assetRepo) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Asset, error) {
	//TODO implement me
	panic("implement me")
}

func (a assetRepo) GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error) {
	//TODO implement me
	panic("implement me")
}

func (a assetRepo) Update(ctx context.Context, asset *models.Asset) error {
	//TODO implement me
	panic("implement me")
}

func (a assetRepo) Delete(ctx context.Context, uuid uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (a assetRepo) UpdateStatus(ctx context.Context, uuid uuid.UUID, newStatus string) error {
	//TODO implement me
	panic("implement me")
}

func (a assetRepo) ListPhotos(ctx context.Context, assetUUID uuid.UUID) ([]*models.Photo, error) {
	//TODO implement me
	panic("implement me")
}

func (a assetRepo) ListRequests(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error) {
	//TODO implement me
	panic("implement me")
}
