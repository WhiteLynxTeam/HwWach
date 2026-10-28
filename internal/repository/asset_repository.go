package repository

import (
	"HwWach/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetRepo запросы по активам для категорий. Остальная работа с активами — в services
// (транзакции через несколько таблиц: роли, передачи, история)
type AssetRepo interface {
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
