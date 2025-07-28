package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"HwWach/internal/config"
	"HwWach/internal/handlers"
	"HwWach/internal/middleware"
	"HwWach/internal/models"
	"HwWach/internal/repository"
	"HwWach/internal/routes"
	"HwWach/internal/services"
	"HwWach/internal/storage"
)

type App struct {
	cfg      *config.Config
	db       *gorm.DB
	minioSvc storage.Storage
	router   *gin.Engine

	authH   handlers.AuthHandler
	userH   handlers.UserHandler
	deviceH handlers.DeviceHandler
	photoH  handlers.PhotoHandler
	reqH    handlers.RequestHandler
}

func NewApp() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Device{},
		&models.Photo{},
		&models.Request{},
	); err != nil {
		return nil, err
	}

	minioCli, err := storage.NewMinioClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioUseSSL,
		cfg.MinioBucket,
	)
	if err != nil {
		return nil, err
	}
	minioSvc := storage.NewMinioStorage(minioCli, cfg.MinioBucket)

	userRepo := repository.NewUserRepo(db)
	deviceRepo := repository.NewDeviceRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	requestRepo := repository.NewRequestRepo(db)

	authSvc := services.NewAuthService(userRepo, []byte(cfg.JWTSecret))
	userSvc := services.NewUserService(userRepo)
	deviceSvc := services.NewDeviceService(deviceRepo, photoRepo)
	photoSvc := services.NewPhotoService(photoRepo, deviceRepo, minioSvc)
	reqSvc := services.NewRequestService(requestRepo, deviceRepo, photoRepo)

	authH := handlers.NewAuthHandler(authSvc)
	userH := handlers.NewUserHandler(userSvc)
	deviceH := handlers.NewDeviceHandler(deviceSvc)
	photoH := handlers.NewPhotoHandler(photoSvc)
	reqH := handlers.NewRequestHandler(reqSvc)

	router := gin.Default()
	jwtMW := middleware.JWTMiddleware([]byte(cfg.JWTSecret))
	routes.SetupRoutes(router, authH, userH, deviceH, photoH, reqH, jwtMW)

	return &App{
		cfg:      cfg,
		db:       db,
		minioSvc: minioSvc,
		router:   router,
		authH:    authH,
		userH:    userH,
		deviceH:  deviceH,
		photoH:   photoH,
		reqH:     reqH,
	}, nil
}

func (a *App) Run() {
	srv := &http.Server{
		Addr:    a.cfg.ServerAddress,
		Handler: a.router,
	}

	go func() {
		log.Printf("Server listening on %s\n", a.cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received; shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown error: %v", err)
	}

	log.Println("Server exited cleanly")
}
