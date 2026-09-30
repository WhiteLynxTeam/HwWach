package utils

import (
	"HwWach/internal/handlers"
	"HwWach/internal/repository"
	"HwWach/internal/services"
	"HwWach/internal/storage"

	"gorm.io/gorm"
)

type DI struct {
	AssetHandler    handlers.AssetHandler
	PhotoHandler    handlers.PhotoHandler
	RequestHandler  handlers.RequestHandler
	CategoryHandler handlers.CategoryHandler
}

func InitializeDI(
	db *gorm.DB,
	minioStorage storage.Storage,
) *DI {
	assetRepo := repository.NewAssetRepo(db)
	photoRepo := repository.NewPhotoRepo(db)
	requestRepo := repository.NewRequestRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)

	categorySvc := services.NewCategoryService(categoryRepo, assetRepo)
	assetSvc := services.NewAssetService(assetRepo, photoRepo)
	photoSvc := services.NewPhotoService(photoRepo, assetRepo, minioStorage)
	reqSvc := services.NewRequestService(requestRepo, assetRepo, photoRepo)

	categoryH := handlers.NewCategoryHandler(categorySvc)
	assetH := handlers.NewAssetHandler(assetSvc, photoSvc, categorySvc)
	photoH := handlers.NewPhotoHandler(photoSvc)
	reqH := handlers.NewRequestHandler(reqSvc)

	return &DI{
		AssetHandler:    assetH,
		PhotoHandler:    photoH,
		RequestHandler:  reqH,
		CategoryHandler: categoryH,
	}
}
