package handlers

import (
	"HwWach/internal/services"
	"github.com/gin-gonic/gin"
)

type deviceHandler struct {
	deviceSvc services.DeviceService
}

func NewDeviceHandler(deviceSvc services.DeviceService) DeviceHandler {
	return &deviceHandler{
		deviceSvc: deviceSvc,
	}
}

func (d deviceHandler) ListUserDevices(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (d deviceHandler) GetDevice(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (d deviceHandler) ListDevicePhotos(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
