package repository

import (
	"HwWach/internal/models"
	"context"
)

type PhotoRepo interface {
	Create(ctx context.Context, photo *models.Photo) error
	GetByID(ctx context.Context, id uint) (*models.Photo, error)
	ListByUser(ctx context.Context, userID uint) ([]*models.Photo, error)
	ListByDevice(ctx context.Context, deviceID uint) ([]*models.Photo, error)
	AttachToDevice(ctx context.Context, photoID, deviceID uint) error
	Detach(ctx context.Context, photoID uint) error
	Delete(ctx context.Context, id uint) error
}
