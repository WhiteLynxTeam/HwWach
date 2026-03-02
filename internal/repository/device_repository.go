package repository

import (
	"HwWach/internal/models"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceRepo interface {
	Create(ctx context.Context, dev *models.Device) error
	GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Device, error)
	GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Device, error)
	Update(ctx context.Context, dev *models.Device) error
	Delete(ctx context.Context, uuid uuid.UUID) error
	UpdateStatus(ctx context.Context, uuid uuid.UUID, newStatus string) error
	ListPhotos(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Photo, error)
	ListRequests(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Request, error)
}

type deviceRepo struct {
	db *gorm.DB
}

func NewDeviceRepo(db *gorm.DB) DeviceRepo {
	return &deviceRepo{db: db}
}

func (d deviceRepo) Create(ctx context.Context, dev *models.Device) error {
	// Генерируем UUID v7 для устройства
	uuidV7, err := uuid.NewV7()
	if err != nil {
		return err
	}
	dev.UUID = uuidV7
	return d.db.WithContext(ctx).Create(dev).Error
}

func (d deviceRepo) GetByUUID(ctx context.Context, uuid uuid.UUID) (*models.Device, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) GetAllByUserUUID(ctx context.Context, userUUID uuid.UUID) ([]*models.Device, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) Update(ctx context.Context, dev *models.Device) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) Delete(ctx context.Context, uuid uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) UpdateStatus(ctx context.Context, uuid uuid.UUID, newStatus string) error {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) ListPhotos(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Photo, error) {
	//TODO implement me
	panic("implement me")
}

func (d deviceRepo) ListRequests(ctx context.Context, deviceUUID uuid.UUID) ([]*models.Request, error) {
	//TODO implement me
	panic("implement me")
}
