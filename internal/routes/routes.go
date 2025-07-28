package routes

import (
	"HwWach/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authHandler handlers.AuthHandler,
	userHandler handlers.UserHandler,
	deviceHandler handlers.DeviceHandler,
	photoHandler handlers.PhotoHandler,
	requestHandler handlers.RequestHandler,
	jwtMiddleware gin.HandlerFunc,
) {
	auth := r.Group("/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.POST("/register", authHandler.Register)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", authHandler.Refresh)
	}

	user := r.Group("/users", jwtMiddleware)
	{
		user.GET("", userHandler.GetProfile)
		user.PATCH("", userHandler.UpdateProfile)
		user.PATCH("/password", userHandler.ChangePassword)
		user.GET("/devices", userHandler.ListDevices)
	}

	dev := r.Group("/devices", jwtMiddleware)
	{
		dev.GET("", deviceHandler.ListUserDevices)
		dev.GET("/:id", deviceHandler.GetDevice)
		dev.GET("/:id/photos", deviceHandler.ListDevicePhotos)
	}

	photos := r.Group("/photos", jwtMiddleware)
	{
		photos.POST("/upload", photoHandler.UploadPhoto)
		photos.GET("/user", photoHandler.ListUserPhotos)
		photos.DELETE("/:id", photoHandler.DeletePhoto)
	}

	req := r.Group("/requests", jwtMiddleware)
	{
		req.POST("", requestHandler.CreateRequest)
		req.GET("/:id", requestHandler.GetRequest)
		req.DELETE("/:id", requestHandler.DeleteRequest)
	}
}
