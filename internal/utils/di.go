package utils

import (
	"HwWach/internal/handlers"
	"HwWach/internal/repository"
	"HwWach/internal/services"
	"HwWach/internal/storage"

	"gorm.io/gorm"
)

type DI struct {
	AuthHandler    handlers.AuthHandler
	UserHandler    handlers.UserHandler
	DeviceHandler  handlers.DeviceHandler
	PhotoHandler   handlers.PhotoHandler
	RequestHandler handlers.RequestHandler
}

func InitializeDI(
	db *gorm.DB,
	jwtSecret []byte,
	minioStorage storage.Storage,
) *DI {
	userRepo := repository.NewUserRepo(db)
	deviceRepo := repository.NewDeviceRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	requestRepo := repository.NewRequestRepo(db)

	authSvc := services.NewAuthService(userRepo, jwtSecret)
	userSvc := services.NewUserService(userRepo)
	deviceSvc := services.NewDeviceService(deviceRepo, photoRepo)
	photoSvc := services.NewPhotoService(photoRepo, deviceRepo, minioStorage)
	reqSvc := services.NewRequestService(requestRepo, deviceRepo, photoRepo)

	authH := handlers.NewAuthHandler(authSvc)
	userH := handlers.NewUserHandler(userSvc)
	deviceH := handlers.NewDeviceHandler(deviceSvc)
	photoH := handlers.NewPhotoHandler(photoSvc)
	reqH := handlers.NewRequestHandler(reqSvc)

	return &DI{
		AuthHandler:    authH,
		UserHandler:    userH,
		DeviceHandler:  deviceH,
		PhotoHandler:   photoH,
		RequestHandler: reqH,
	}
}
