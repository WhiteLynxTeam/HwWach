package repository

import (
	"HwWach/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RequestRepo interface {
	Create(ctx context.Context, req *models.Request) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Request, error)
	ListByDeviceUUID(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Request, error)
	ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Request, error)
	Delete(ctx context.Context, uuid uuid.UUID) error
	UpdateStatus(ctx context.Context, uuid uuid.UUID, newStatus string) error
}

type requestRepo struct {
	db *gorm.DB
}

func NewRequestRepo(db *gorm.DB) RequestRepo {
	return &requestRepo{db: db}
}

func (r requestRepo) Create(ctx context.Context, req *models.Request) error {
	// Генерируем UUID v7 для запроса
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return err
	}
	req.UUID = uuidV7
	return r.db.WithContext(ctx).Create(req).Error
}

func (r requestRepo) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) ListByDeviceUUID(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) ListByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) Delete(ctx context.Context, uuid uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) UpdateStatus(ctx context.Context, uuid uuid.UUID, newStatus string) error {
	//TODO implement me
	panic("implement me")
}
