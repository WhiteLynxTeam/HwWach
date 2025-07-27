package handlers

import "github.com/gin-gonic/gin"

type AuthHandler interface {
	Login(c *gin.Context)
	Register(c *gin.Context)
	Logout(c *gin.Context)
	Refresh(c *gin.Context)
}

type UserHandler interface {
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	ChangePassword(c *gin.Context)
	ListDevices(c *gin.Context)
}

type DeviceHandler interface {
	ListUserDevices(c *gin.Context)
	GetDevice(c *gin.Context)
	ListDevicePhotos(c *gin.Context)
}

type PhotoHandler interface {
	UploadPhoto(c *gin.Context)
	ListUserPhotos(c *gin.Context)
	DeletePhoto(c *gin.Context)
}

type RequestHandler interface {
	CreateRequest(c *gin.Context)
	GetRequest(c *gin.Context)
	DeleteRequest(c *gin.Context)
}
