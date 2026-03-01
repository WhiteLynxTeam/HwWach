package handlers

import "github.com/gin-gonic/gin"

type DeviceHandler interface {
	ListUserDevices(c *gin.Context)
	GetDevice(c *gin.Context)
	ListDevicePhotos(c *gin.Context)
}

type PhotoHandler interface {
	UploadPhoto(c *gin.Context)
	ConfirmUpload(c *gin.Context)
	ListUserPhotos(c *gin.Context)
	DeletePhoto(c *gin.Context)
}

type RequestHandler interface {
	CreateRequest(c *gin.Context)
	GetRequest(c *gin.Context)
	DeleteRequest(c *gin.Context)
}
