package routes

import (
	"HwWach/internal/handlers"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes(
	r *gin.Engine,
	deviceHandler handlers.DeviceHandler,
	photoHandler handlers.PhotoHandler,
	requestHandler handlers.RequestHandler,
	jwtMiddleware gin.HandlerFunc,
) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	dev := r.Group("/devices", jwtMiddleware)
	{
		dev.GET("", deviceHandler.ListUserDevices)
		dev.GET("/:id", deviceHandler.GetDevice)
		dev.GET("/:id/photos", deviceHandler.ListDevicePhotos)
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
