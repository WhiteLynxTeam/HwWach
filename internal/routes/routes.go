package routes

import (
	"HwWach/internal/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(
	r *gin.Engine,
	assetHandler handlers.AssetHandler,
	photoHandler handlers.PhotoHandler,
	requestHandler handlers.RequestHandler,
	changeRequestHandler handlers.AssetChangeRequestHandler,
	jwtMiddleware gin.HandlerFunc,
) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	assets := r.Group("/assets", jwtMiddleware)
	{
		assets.POST("", assetHandler.CreateAsset)
		assets.GET("", assetHandler.ListUserAssets)
		assets.GET("/paginated", assetHandler.ListUserAssetsPaginated)
		assets.GET("/check-inventory", assetHandler.CheckInventoryUnique)
		assets.GET("/:id", assetHandler.GetAsset)
		assets.PUT("/:id", assetHandler.UpdateAsset)
		assets.GET("/:id/photos", assetHandler.ListAssetPhotos)

		assets.POST("/:id/change-requests", changeRequestHandler.CreateRequest)
		assets.GET("/change-requests", changeRequestHandler.ListPending)
		assets.PATCH("/change-requests/:id", changeRequestHandler.ApproveRequest)
	}

	photos := r.Group("/photos", jwtMiddleware)
	{
		photos.POST("/upload-url", photoHandler.UploadPhoto)
		photos.POST("/complete-upload", photoHandler.ConfirmUpload)
		photos.GET("", photoHandler.ListUserPhotos)
		photos.DELETE("/:id", photoHandler.DeletePhoto)
	}

	req := r.Group("/requests", jwtMiddleware)
	{
		req.POST("", requestHandler.CreateRequest)
		req.GET("/:id", requestHandler.GetRequest)
		req.DELETE("/:id", requestHandler.DeleteRequest)
	}
}
