package repository

import (
	"HwWach/internal/models"
	"context"
	"gorm.io/gorm"
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

func NewPhotoRepo(db *gorm.DB) PhotoRepo {
	return &photoRepo{
		baseRepo: baseRepo{db: db},
	}
}

func (p photoRepo) Create(ctx context.Context, photo *models.Photo) error {
	//TODO implement me
	panic("implement me")
}

func (p photoRepo) GetByID(ctx context.Context, id uint) (*models.Photo, error) {
	//TODO implement me
	panic("implement me")
}

func (p photoRepo) ListByUser(ctx context.Context, userID uint) ([]*models.Photo, error) {
	//TODO implement me
	panic("implement me")
}

func (p photoRepo) ListByDevice(ctx context.Context, deviceID uint) ([]*models.Photo, error) {
	//TODO implement me
	panic("implement me")
}

func (p photoRepo) AttachToDevice(ctx context.Context, photoID, deviceID uint) error {
	//TODO implement me
	panic("implement me")
}

func (p photoRepo) Detach(ctx context.Context, photoID uint) error {
	//TODO implement me
	panic("implement me")
}

func (p photoRepo) Delete(ctx context.Context, id uint) error {
	//TODO implement me
	panic("implement me")
}
