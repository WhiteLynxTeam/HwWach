package repository

import (
	"HwWach/internal/models"
	"context"
)

type RequestRepo interface {
	Create(ctx context.Context, req *models.Request) error
	GetByID(ctx context.Context, id uint) (*models.Request, error)
	ListByDevice(ctx context.Context, deviceID uint) ([]*models.Request, error)
	ListByUser(ctx context.Context, userID uint) ([]*models.Request, error)
	Delete(ctx context.Context, id uint) error
	UpdateStatus(ctx context.Context, id uint, newStatus string) error
}
