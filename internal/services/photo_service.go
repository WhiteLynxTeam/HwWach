package services

import (
	"HwWach/internal/repository"
	"HwWach/internal/storage"
)

type PhotoService interface{}

type photoService struct {
	photoRepo  repository.PhotoRepo
	deviceRepo repository.DeviceRepo
	minioSVC   storage.Storage
}

func NewPhotoService(
	photoRepo repository.PhotoRepo,
	deviceRepo repository.DeviceRepo,
	minioSVC storage.Storage) RequestService {
	return &photoService{
		photoRepo:  photoRepo,
		deviceRepo: deviceRepo,
		minioSVC:   minioSVC,
	}
}
