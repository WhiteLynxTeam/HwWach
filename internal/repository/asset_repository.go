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
	var asset models.Asset
	if err := a.db.WithContext(ctx).First(&asset, "uuid = ?", uuid).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (a assetRepo) GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error) {
	var assets []*models.Asset
	if err := a.db.WithContext(ctx).Where("user_id = ?", userUUID).Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (a assetRepo) Update(ctx context.Context, asset *models.Asset) error {
	return a.db.WithContext(ctx).Save(asset).Error
}

func (a assetRepo) Delete(ctx context.Context, uuid uuid.UUID) error {
	return a.db.WithContext(ctx).Delete(&models.Asset{}, "uuid = ?", uuid).Error
}

func (a assetRepo) UpdateStatus(ctx context.Context, uuid uuid.UUID, newStatus string) error {
	return a.db.WithContext(ctx).Model(&models.Asset{}).
		Where("uuid = ?", uuid).
		Update("status", newStatus).Error
}

func (a assetRepo) ListPhotos(ctx context.Context, assetUUID uuid.UUID) ([]*models.Photo, error) {
	var photos []*models.Photo
	err := a.db.WithContext(ctx).
		Model(&models.Photo{}).
		Joins("join asset_photos on asset_photos.photo_uuid = photos.uuid").
		Where("asset_photos.asset_uuid = ?", assetUUID).
		Find(&photos).Error
	return photos, err
}

func (a assetRepo) ListRequests(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error) {
	var requests []*models.Request
	if err := a.db.WithContext(ctx).Where("asset_id = ?", assetUUID).Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}
