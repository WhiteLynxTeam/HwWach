package repository

import (
	"HwWach/internal/models"
	"context"
	"gorm.io/gorm"
)

type RequestRepo interface {
	Create(ctx context.Context, req *models.Request) error
	GetByID(ctx context.Context, id uint) (*models.Request, error)
	ListByDevice(ctx context.Context, deviceID uint) ([]*models.Request, error)
	ListByUser(ctx context.Context, userID uint) ([]*models.Request, error)
	Delete(ctx context.Context, id uint) error
	UpdateStatus(ctx context.Context, id uint, newStatus string) error
}

func NewRequestRepo(db *gorm.DB) RequestRepo {
	return &requestRepo{
		baseRepo: baseRepo{db: db},
	}
}

func (r requestRepo) Create(ctx context.Context, req *models.Request) error {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) GetByID(ctx context.Context, id uint) (*models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) ListByDevice(ctx context.Context, deviceID uint) ([]*models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) ListByUser(ctx context.Context, userID uint) ([]*models.Request, error) {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) Delete(ctx context.Context, id uint) error {
	//TODO implement me
	panic("implement me")
}

func (r requestRepo) UpdateStatus(ctx context.Context, id uint, newStatus string) error {
	//TODO implement me
	panic("implement me")
}
