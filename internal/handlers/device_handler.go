package handlers

import (
	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type deviceHandler struct {
	deviceSvc services.DeviceService
	photoSvc  services.PhotoService
}

func NewDeviceHandler(deviceSvc services.DeviceService, photoSvc services.PhotoService) DeviceHandler {
	return &deviceHandler{
		deviceSvc: deviceSvc,
		photoSvc:  photoSvc,
	}
}

// ListUserDevices godoc
// @Summary      Список устройств пользователя
// @Description  Получение всех устройств авторизованного пользователя
// @Tags         devices
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Device
// @Failure      401  {object}  map[string]string
// @Router       /devices [get]
// @Security     BearerAuth
func (d deviceHandler) ListUserDevices(c *gin.Context) {
	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	devices, err := d.deviceSvc.GetAllByUserUUID(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_uuid": userUUID.String(),
		"devices":   devices,
	})
}

// GetDevice godoc
// @Summary      Получить устройство по ID
// @Description  Получение информации об устройстве по идентификатору
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "ID устройства (uuid)"
// @Success      200  {object}  models.Device
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /devices/{id} [get]
// @Security     BearerAuth
func (d deviceHandler) GetDevice(c *gin.Context) {
	deviceUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return
	}

	device, err := d.deviceSvc.GetByUUID(c.Request.Context(), deviceUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

// ListDevicePhotos godoc
// @Summary      Список фотографий устройства
// @Description  Получение всех фотографий для указанного устройства
// @Tags         devices,photos
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "ID устройства (uuid)"
// @Success      200  {array}   models.Photo
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /devices/{id}/photos [get]
// @Security     BearerAuth
func (d deviceHandler) ListDevicePhotos(c *gin.Context) {
	deviceUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return
	}

	photos, err := d.deviceSvc.ListPhotos(c.Request.Context(), deviceUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Конвертируем модели в DTO с полными URL
	photoResponses := make([]dto.PhotoResponse, 0, len(photos))
	for _, photo := range photos {
		photoResp := dto.PhotoResponse{
			UUID:      photo.UUID.String(),
			URL:       d.photoSvc.GetPublicURL(photo.URL),
			CreatedAt: photo.CreatedAt.Format(time.RFC3339),
		}
		if photo.ClientID != nil {
			clientIDStr := photo.ClientID.String()
			photoResp.ClientID = &clientIDStr
		}
		photoResponses = append(photoResponses, photoResp)
	}

	c.JSON(http.StatusOK, gin.H{
		"device_uuid": deviceUUID.String(),
		"photos":      photoResponses,
	})
}
