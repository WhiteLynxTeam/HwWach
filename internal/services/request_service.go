package services

import (
	"HwWach/internal/repository"
)

type RequestService interface{}

type requestService struct {
	requestRepo repository.RequestRepo
	deviceRepo  repository.DeviceRepo
	photoRepo   repository.PhotoRepo
}

func NewRequestService(
	requestRepo repository.RequestRepo,
	deviceRepo repository.DeviceRepo,
	photoRepo repository.PhotoRepo) RequestService {
	return &requestService{
		requestRepo: requestRepo,
		deviceRepo:  deviceRepo,
		photoRepo:   photoRepo,
	}
}
