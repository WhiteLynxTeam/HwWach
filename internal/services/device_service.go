package services

import "HwWach/internal/repository"

type DeviceService interface{}

type deviceService struct {
	deviceRepo repository.DeviceRepo
	photoRepo  repository.PhotoRepo
}

func NewDeviceService(deviceRepo repository.DeviceRepo, photoRepo repository.PhotoRepo) DeviceService {
	return &deviceService{
		deviceRepo: deviceRepo,
		photoRepo:  photoRepo,
	}
}
