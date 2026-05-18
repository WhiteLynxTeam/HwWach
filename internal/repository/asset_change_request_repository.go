package repository

import (
	"HwWach/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetChangeRequestRepo interface {
	Create(ctx context.Context, req *models.AssetChangeRequest) error
	GetPendingByAssetID(ctx context.Context, assetUUID uuid.UUID) (*models.AssetChangeRequest, error)
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.AssetChangeRequest, error)
	UpdateStatus(ctx context.Context, req *models.AssetChangeRequest) error
	ListPending(ctx context.Context) ([]*models.AssetChangeRequest, error)
}

type assetChangeRequestRepo struct {
	db *gorm.DB
}

func NewAssetChangeRequestRepo(db *gorm.DB) AssetChangeRequestRepo {
	return &assetChangeRequestRepo{db: db}
}

func (r assetChangeRequestRepo) Create(ctx context.Context, req *models.AssetChangeRequest) error {
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return err
	}
	req.UUID = uuidV7
	return r.db.WithContext(ctx).Create(req).Error
}

func (r assetChangeRequestRepo) GetPendingByAssetID(ctx context.Context, assetUUID uuid.UUID) (*models.AssetChangeRequest, error) {
	var req models.AssetChangeRequest
	err := r.db.WithContext(ctx).
		Where("asset_uuid = ? AND status = ?", assetUUID, models.ModerationPending).
		First(&req).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found - nil means no pending requests
		}
		return nil, err
	}
	return &req, nil
}

func (r assetChangeRequestRepo) GetByUUID(ctx context.Context, reqUUID uuid.UUID) (*models.AssetChangeRequest, error) {
	var req models.AssetChangeRequest
	err := r.db.WithContext(ctx).First(&req, "uuid = ?", reqUUID).Error
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r assetChangeRequestRepo) UpdateStatus(ctx context.Context, req *models.AssetChangeRequest) error {
	return r.db.WithContext(ctx).Save(req).Error
}

func (r assetChangeRequestRepo) ListPending(ctx context.Context) ([]*models.AssetChangeRequest, error) {
	var requests []*models.AssetChangeRequest
	err := r.db.WithContext(ctx).
		Where("status = ?", models.ModerationPending).
		Find(&requests).Error
	return requests, err
}
