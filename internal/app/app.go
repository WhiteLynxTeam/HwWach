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
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"HwWach/internal/config"
	"HwWach/internal/handlers"
	"HwWach/internal/middleware"
	"HwWach/internal/repository"
	"HwWach/internal/routes"
	"HwWach/internal/services"
	"HwWach/internal/storage"
	"HwWach/migrations"
)

type App struct {
	cfg      *config.Config
	db       *gorm.DB
	storageSvc storage.Storage
	router   *gin.Engine

	assetH     handlers.AssetHandler
	photoH     handlers.PhotoHandler
	reqH       handlers.RequestHandler
	changeReqH handlers.AssetChangeRequestHandler
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

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		return nil, err
	}

	// Выбор хранилища на основе STORAGE_TYPE
	var storageSvc storage.Storage
	switch cfg.StorageType {
	case "yandex":
		storageSvc, err = storage.NewYandexDiskStorage(
			cfg.YandexEndpoint,
			cfg.YandexDiskToken,
			cfg.YandexDiskBucket,
		)
	default: // "minio"
		storageSvc, err = storage.NewMinioStorage(
			cfg.MinioEndpoint,  // внутренний: minio:9000
			cfg.MinioPublicURL, // внешний: http://149.154.65.57:9000
			cfg.MinioAccessKey,
			cfg.MinioSecretKey,
			cfg.MinioUseSSL,
			cfg.MinioBucket,
		)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("Storage type: %s\n", cfg.StorageType)

	assetRepo := repository.NewAssetRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	requestRepo := repository.NewRequestRepo(db)
	changeReqRepo := repository.NewAssetChangeRequestRepo(db)

	assetSvc := services.NewAssetService(assetRepo, photoRepo)
	photoSvc := services.NewPhotoService(photoRepo, assetRepo, storageSvc)
	reqSvc := services.NewRequestService(requestRepo, assetRepo, photoRepo)
	changeReqSvc := services.NewAssetChangeRequestService(changeReqRepo, assetRepo)

	assetH := handlers.NewAssetHandler(assetSvc, photoSvc)
	photoH := handlers.NewPhotoHandler(photoSvc)
	reqH := handlers.NewRequestHandler(reqSvc)
	changeReqH := handlers.NewAssetChangeRequestHandler(changeReqSvc)

	router := gin.Default()
	jwtMW := middleware.JWTMiddleware([]byte(cfg.JWTSecret))
	routes.SetupRoutes(router, assetH, photoH, reqH, changeReqH, jwtMW)

	return &App{
		cfg:        cfg,
		db:         db,
		storageSvc: storageSvc,
		router:     router,
		assetH:     assetH,
		photoH:     photoH,
		reqH:       reqH,
		changeReqH: changeReqH,
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
