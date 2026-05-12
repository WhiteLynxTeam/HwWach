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
	cfg       *config.Config
	db        *gorm.DB
	minioSvc  storage.Storage
	router    *gin.Engine

	assetH handlers.AssetHandler
	photoH handlers.PhotoHandler
	reqH   handlers.RequestHandler
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
		&models.Asset{},
		&models.Photo{},
		&models.Request{},
	); err != nil {
		return nil, err
	}

	minioSvc, err := storage.NewMinioStorage(
		cfg.MinioEndpoint,  // внутренний: minio:9000
		cfg.MinioPublicURL, // внешний: http://149.154.65.57:9000
		cfg.MinioAccessKey,
		cfg.MinioSecretKey,
		cfg.MinioUseSSL,
		cfg.MinioBucket,
	)
	if err != nil {
		return nil, err
	}

	assetRepo := repository.NewAssetRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	requestRepo := repository.NewRequestRepo(db)

	assetSvc := services.NewAssetService(assetRepo, photoRepo)
	photoSvc := services.NewPhotoService(photoRepo, assetRepo, minioSvc)
	reqSvc := services.NewRequestService(requestRepo, assetRepo, photoRepo)

	assetH := handlers.NewAssetHandler(assetSvc, photoSvc)
	photoH := handlers.NewPhotoHandler(photoSvc)
	reqH := handlers.NewRequestHandler(reqSvc)

	router := gin.Default()
	jwtMW := middleware.JWTMiddleware([]byte(cfg.JWTSecret))
	routes.SetupRoutes(router, assetH, photoH, reqH, jwtMW)

	return &App{
		cfg:      cfg,
		db:       db,
		minioSvc: minioSvc,
		router:   router,
		assetH:   assetH,
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
