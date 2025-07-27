package repository

import (
	"HwWach/internal/models"
	"context"
)

type DeviceRepo interface {
	Create(ctx context.Context, dev *models.Device) error
	GetByID(ctx context.Context, id uint) (*models.Device, error)
	GetAllByUser(ctx context.Context, userID uint) ([]*models.Device, error)
	Update(ctx context.Context, dev *models.Device) error
	Delete(ctx context.Context, id uint) error
	UpdateStatus(ctx context.Context, id uint, newStatus string) error
	AttachPhoto(ctx context.Context, deviceID, photoID uint) error
	ListPhotos(ctx context.Context, deviceID uint) ([]*models.Photo, error)
	ListRequests(ctx context.Context, deviceID uint) ([]*models.Request, error)
}
