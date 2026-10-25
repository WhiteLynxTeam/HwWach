package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
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
	"HwWach/internal/tsclient"
	"HwWach/migrations"
)

type App struct {
	cfg        *config.Config
	db         *gorm.DB
	storageSvc storage.Storage
	router     *gin.Engine
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

	ts := tsclient.New(strings.TrimRight(cfg.AuthInternalURL, "/"), cfg.InternalAPIKey)

	assetRepo := repository.NewAssetRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	changeReqRepo := repository.NewAssetChangeRequestRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)

	auditSvc := services.NewAuditService(db)
	orgSvc := services.NewOrgService(db, ts, storageSvc, auditSvc)
	rolesSvc := services.NewRolesService(db, ts, auditSvc)
	lifecycleSvc := services.NewLifecycleService(db, rolesSvc, auditSvc)
	categorySvc := services.NewCategoryService(categoryRepo, assetRepo, cfg.PlatformOrgID)
	assetSvc := services.NewAssetService(db, rolesSvc, orgSvc, auditSvc)
	photoSvc := services.NewPhotoService(photoRepo, storageSvc, orgSvc)
	reqSvc := services.NewRequestService(db, lifecycleSvc, auditSvc)
	changeReqSvc := services.NewAssetChangeRequestService(changeReqRepo, assetSvc)

	router := gin.Default()
	// IP клиента за прокси Render (8.7): без доверенных прокси все запросы «с одного адреса»
	if cfg.TrustedProxies != "" {
		if err := router.SetTrustedProxies(strings.Split(cfg.TrustedProxies, ",")); err != nil {
			return nil, err
		}
	} else {
		_ = router.SetTrustedProxies(nil)
	}
	if cfg.InternalAPIKey == "" {
		log.Println("WARNING: INTERNAL_API_KEY не задан — внутренние запросы ts ↔ go не работают")
	}
	routes.SetupRoutes(router, routes.Handlers{
		Asset:         handlers.NewAssetHandler(assetSvc, lifecycleSvc, photoSvc, categorySvc),
		Roles:         handlers.NewRolesHandler(rolesSvc),
		Photo:         handlers.NewPhotoHandler(photoSvc),
		Request:       handlers.NewRequestHandler(reqSvc),
		ChangeRequest: handlers.NewAssetChangeRequestHandler(changeReqSvc),
		Category:      handlers.NewCategoryHandler(categorySvc),
		Audit:         handlers.NewAuditHandler(auditSvc),
		Internal:      handlers.NewInternalHandler(rolesSvc, orgSvc),
	}, middleware.JWTMiddleware([]byte(cfg.JWTSecret)), middleware.InternalKeyMiddleware(cfg.InternalAPIKey))

	return &App{
		cfg:        cfg,
		db:         db,
		storageSvc: storageSvc,
		router:     router,
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
