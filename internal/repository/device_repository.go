package repository

import (
	"HwWach/internal/models"
	"context"
	"gorm.io/gorm"
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

func NewDeviceRepo(db *gorm.DB) DeviceRepo {
	return &deviceRepo{
		baseRepo: baseRepo{db: db},
	}
}

func (d deviceRepo) Create(ctx context.Context, dev *models.Device) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) GetByID(ctx context.Context, id uint) (*models.Device, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) GetAllByUser(ctx context.Context, userID uint) ([]*models.Device, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) Update(ctx context.Context, dev *models.Device) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) Delete(ctx context.Context, id uint) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) UpdateStatus(ctx context.Context, id uint, newStatus string) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) AttachPhoto(ctx context.Context, deviceID, photoID uint) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) ListPhotos(ctx context.Context, deviceID uint) ([]*models.Photo, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) ListRequests(ctx context.Context, deviceID uint) ([]*models.Request, error) {
	//TODO implement me
	panic("implement me")
}
