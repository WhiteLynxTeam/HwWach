package utils

import (
	"HwWach/internal/handlers"
	"HwWach/internal/repository"
	"HwWach/internal/services"
	"HwWach/internal/storage"

	"gorm.io/gorm"
)

type DI struct {
	DeviceHandler  handlers.DeviceHandler
	PhotoHandler   handlers.PhotoHandler
	RequestHandler handlers.RequestHandler
}

func InitializeDI(
	db *gorm.DB,
	minioStorage storage.Storage,
) *DI {
	deviceRepo := repository.NewDeviceRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	requestRepo := repository.NewRequestRepo(db)

	deviceSvc := services.NewDeviceService(deviceRepo, photoRepo)
	photoSvc := services.NewPhotoService(photoRepo, deviceRepo, minioStorage)
	reqSvc := services.NewRequestService(requestRepo, deviceRepo, photoRepo)

	deviceH := handlers.NewDeviceHandler(deviceSvc)
	photoH := handlers.NewPhotoHandler(photoSvc)
	reqH := handlers.NewRequestHandler(reqSvc)

	return &DI{
		DeviceHandler:  deviceH,
		PhotoHandler:   photoH,
		RequestHandler: reqH,
	}
}
