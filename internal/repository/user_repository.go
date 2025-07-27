package repository

import (
	"HwWach/internal/models"
	"context"

	"gorm.io/gorm"
)

type UserRepo interface {
	Create(ctx context.Context, u *models.User) error
	FindByLogin(ctx context.Context, login string) (*models.User, error)
	Update(ctx context.Context, u *models.User) error
	Delete(ctx context.Context, id uint) error
}

type userRepo struct {
	db *gorm.DB
}
