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
	GetByInventoryNum(ctx context.Context, inventoryNum string) (*models.Asset, error)
	GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error)
	Update(ctx context.Context, asset *models.Asset) error
	Delete(ctx context.Context, uuid uuid.UUID) error
	UpdateStatus(ctx context.Context, uuid uuid.UUID, newStatus string) error
	ListPhotos(ctx context.Context, assetUUID uuid.UUID) ([]*models.Photo, error)
	ListRequests(ctx context.Context, assetUUID uuid.UUID) ([]*models.Request, error)
	GetPaginated(ctx context.Context, userUUID *uuid.UUID, page, limit int) ([]*models.Asset, int64, error)
	ReassignCategory(ctx context.Context, oldCategoryUUID, newCategoryUUID uuid.UUID, newCategoryName string) error
	CountAssetsByCategory(ctx context.Context, categoryUUID uuid.UUID) (int64, error)
	GetAssetsByCategory(ctx context.Context, categoryUUID uuid.UUID, limit int) ([]*models.Asset, error)
	UpdateCategoryName(ctx context.Context, categoryUUID uuid.UUID, newName string) error
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

func (a assetRepo) GetByInventoryNum(ctx context.Context, inventoryNum string) (*models.Asset, error) {
	var asset models.Asset
	if err := a.db.WithContext(ctx).First(&asset, "inventory_num = ?", inventoryNum).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}

func (a assetRepo) GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Asset, error) {
	var assets []*models.Asset
	if err := a.db.WithContext(ctx).Preload("Photos").Where("user_id = ?", userUUID).Find(&assets).Error; err != nil {
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

func (a assetRepo) GetPaginated(ctx context.Context, userUUID *uuid.UUID, page, limit int) ([]*models.Asset, int64, error) {
	var total int64
	var assets []*models.Asset

	query := a.db.WithContext(ctx).Model(&models.Asset{})
	if userUUID != nil {
		query = query.Where("user_id = ?", *userUUID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Limit(limit).Offset(offset).Preload("Photos").Find(&assets).Error; err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (a assetRepo) ReassignCategory(ctx context.Context, oldCategoryUUID, newCategoryUUID uuid.UUID, newCategoryName string) error {
	return a.db.WithContext(ctx).Model(&models.Asset{}).
		Where("category_uuid = ?", oldCategoryUUID).
		Updates(map[string]interface{}{
			"category_uuid": newCategoryUUID,
			"category":      newCategoryName,
		}).Error
}

func (a assetRepo) CountAssetsByCategory(ctx context.Context, categoryUUID uuid.UUID) (int64, error) {
	var count int64
	err := a.db.WithContext(ctx).Model(&models.Asset{}).Where("category_uuid = ?", categoryUUID).Count(&count).Error
	return count, err
}

func (a assetRepo) GetAssetsByCategory(ctx context.Context, categoryUUID uuid.UUID, limit int) ([]*models.Asset, error) {
	var assets []*models.Asset
	if limit <= 0 {
		limit = 10
	}
	err := a.db.WithContext(ctx).Model(&models.Asset{}).Where("category_uuid = ?", categoryUUID).Limit(limit).Find(&assets).Error
	return assets, err
}

func (a assetRepo) UpdateCategoryName(ctx context.Context, categoryUUID uuid.UUID, newName string) error {
	return a.db.WithContext(ctx).Model(&models.Asset{}).
		Where("category_uuid = ?", categoryUUID).
		Update("category", newName).Error
}

