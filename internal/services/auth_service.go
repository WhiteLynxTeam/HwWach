package services

import (
	"HwWach/internal/models"
	"HwWach/internal/repository"
)

type AuthService interface {
	Login(login, pass string) (string, error)
	Register(initialPwd, fio, phone string) (*models.User, error)
}

type authService struct {
	userRepo  repository.UserRepo
	jwtSecret []byte
}

func NewAuthService(userRepo repository.UserRepo, jwtSecret []byte) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
	}
}

func (a authService) Login(login, pass string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (a authService) Register(initialPwd, fio, phone string) (*models.User, error) {
	//TODO implement me
	panic("implement me")
}
