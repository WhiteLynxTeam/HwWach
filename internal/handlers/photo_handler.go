package handlers

import (
	"HwWach/internal/services"
	"github.com/gin-gonic/gin"
)

type photoHandler struct {
	photoSvc services.PhotoService
}

func NewPhotoHandler(photoSvc services.PhotoService) PhotoHandler {
	return &photoHandler{
		photoSvc: photoSvc,
	}
}

func (p photoHandler) UploadPhoto(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (p photoHandler) ListUserPhotos(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (p photoHandler) DeletePhoto(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
