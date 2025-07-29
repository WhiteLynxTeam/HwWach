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

func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db: db}
}

func (u2 userRepo) Create(ctx context.Context, u *models.User) error {
	//TODO implement me
	panic("implement me")
}

func (u2 userRepo) FindByLogin(ctx context.Context, login string) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u2 userRepo) Update(ctx context.Context, u *models.User) error {
	//TODO implement me
	panic("implement me")
}

func (u2 userRepo) Delete(ctx context.Context, id uint) error {
	//TODO implement me
	panic("implement me")
}
