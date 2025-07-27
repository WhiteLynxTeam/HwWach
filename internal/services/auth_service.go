package services

import "HwWach/internal/models"

type AuthService interface {
	Login(login, pass string) (string, error)
	Register(initialPwd, fio, phone string) (*models.User, error)
}
