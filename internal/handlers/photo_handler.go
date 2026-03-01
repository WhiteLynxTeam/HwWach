package handlers

import (
	"fmt"

	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/models"
	"HwWach/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type photoHandler struct {
	photoSvc services.PhotoService
}

func NewPhotoHandler(photoSvc services.PhotoService) PhotoHandler {
	return &photoHandler{
		photoSvc: photoSvc,
	}
}

// UploadPhoto godoc
// @Summary      Получить presigned URL для загрузки фотографии
// @Description  Возвращает временный URL для прямой загрузки файла в MinIO (минуя сервер). Макс. размер файла: 20 МБ
// @Tags         photos
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UploadPhotoRequest  true  "Имя файла, MIME тип и размер"
// @Success      200  {object}  dto.UploadPhotoResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /photos/upload-url [post]
// @Security     BearerAuth
func (p photoHandler) UploadPhoto(c *gin.Context) {
	var req dto.UploadPhotoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	// Проверка размера файла (макс. 20 МБ)
	if req.FileSize <= 0 || req.FileSize > dto.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("file size must be between 1 and %d bytes (%d MB)", dto.MaxFileSize, dto.MaxFileSize/1024/1024),
		})
		return
	}

	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	// Если передан client_id, проверяем, существует ли уже запись
	if req.ClientID != nil {
		existingPhoto, err := p.photoSvc.GetByClientID(c.Request.Context(), *req.ClientID)
		if err == nil {
			// Запись найдена
			if existingPhoto.UserUUID != userUUID {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied: photo belongs to another user"})
				return
			}

			if existingPhoto.Status == models.PhotoStatusCompleted {
				// Уже загружено — возвращаем информацию о фото
				c.JSON(http.StatusOK, dto.UploadPhotoResponse{
					UploadURL:   "",
					PhotoUUID:   existingPhoto.UUID.String(),
					ClientID:    req.ClientID.String(),
					Method:      "ALREADY_UPLOADED",
					ExpiresIn:   0,
					MaxFileSize: dto.MaxFileSize,
				})
				return
			}

			// Статус pending — возвращаем presigned URL для существующей записи
			uploadURL, err := p.photoSvc.GetPresignedUploadURL(c.Request.Context(), existingPhoto.URL, req.ContentType)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get upload URL: " + err.Error()})
				return
			}

			c.JSON(http.StatusOK, dto.UploadPhotoResponse{
				UploadURL:   uploadURL,
				PhotoUUID:   existingPhoto.UUID.String(),
				ClientID:    req.ClientID.String(),
				Method:      "PUT",
				ExpiresIn:   86400,
				MaxFileSize: dto.MaxFileSize,
			})
			return
		}
		// Ошибка (запись не найдена) — создаём новую
	}

	// Создаём запись в БД со статусом pending
	photo, err := p.photoSvc.CreatePendingPhoto(c.Request.Context(), userUUID, req.Filename, req.FileSize, req.ContentType, req.ClientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create photo record: " + err.Error()})
		return
	}

	// Получаем presigned URL для PUT запроса (загрузка)
	uploadURL, err := p.photoSvc.GetPresignedUploadURL(c.Request.Context(), photo.URL, req.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get upload URL: " + err.Error()})
		return
	}

	resp := dto.UploadPhotoResponse{
		UploadURL:   uploadURL,
		PhotoUUID:   photo.UUID.String(),
		Method:      "PUT",
		ExpiresIn:   86400, // 24 часа
		MaxFileSize: dto.MaxFileSize,
	}
	
	// Добавляем client_id, если он был передан
	if req.ClientID != nil {
		resp.ClientID = req.ClientID.String()
	}

	c.JSON(http.StatusOK, resp)
}

// ConfirmUpload godoc
// @Summary      Подтвердить загрузку фотографии
// @Description  Изменение статуса фотографии на completed после успешной загрузки в MinIO
// @Tags         photos
// @Accept       json
// @Produce      json
// @Param        request  body      dto.ConfirmUploadRequest  true  "UUID фотографии и опционально устройство"
// @Success      200  {object}  dto.PhotoResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /photos/complete-upload [post]
// @Security     BearerAuth
func (p photoHandler) ConfirmUpload(c *gin.Context) {
	var req dto.ConfirmUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	// Парсим UUID фотографии
	photoUUID, err := uuid.Parse(req.PhotoUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo_uuid format"})
		return
	}

	// Получаем фотографию из БД
	photo, err := p.photoSvc.GetByUUID(c.Request.Context(), photoUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "photo not found"})
		return
	}

	// Проверяем, что фото принадлежит пользователю
	if photo.UserUUID != userUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	// Проверяем, что фото ещё не загружено
	if photo.Status == models.PhotoStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "photo already uploaded"})
		return
	}

	// Меняем статус на completed
	if err := p.photoSvc.CompletePhotoUpload(c.Request.Context(), photoUUID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete upload: " + err.Error()})
		return
	}

	// Создаём ответ
	photoResp := dto.PhotoResponse{
		UUID:      photo.UUID.String(),
		UserUUID:  photo.UserUUID.String(),
		URL:       photo.URL,
		CreatedAt: photo.CreatedAt.Format(time.RFC3339),
	}
	
	// Добавляем client_id, если он был передан
	if photo.ClientID != nil {
		clientIDStr := photo.ClientID.String()
		photoResp.ClientID = &clientIDStr
	}

	c.JSON(http.StatusOK, photoResp)
}

// ListUserPhotos godoc
// @Summary      Список фотографий сделанных пользователем
// @Description  Получение всех фотографий авторизованного пользователя
// @Tags         photos
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.PhotoListResponse
// @Failure      401  {object}  map[string]string
// @Router       /photos [get]
// @Security     BearerAuth
func (p photoHandler) ListUserPhotos(c *gin.Context) {
	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	photos, err := p.photoSvc.ListByUserUUID(c.Request.Context(), userUUID)
	if err != nil {
		// Если таблица не существует — вернём пустой список с информацией
		c.JSON(http.StatusOK, gin.H{
			"user_uuid": userUUID.String(),
			"photos":    []string{},
			"note":      "Table not found or DB error: " + err.Error(),
		})
		return
	}

	// Конвертируем модели в DTO
	photoResponses := make([]dto.PhotoResponse, 0, len(photos))
	for _, photo := range photos {
		photoResp := dto.PhotoResponse{
			UUID:      photo.UUID.String(),
			UserUUID:  photo.UserUUID.String(),
			URL:       photo.URL,
			CreatedAt: photo.CreatedAt.Format(time.RFC3339),
		}
		if photo.ClientID != nil {
			clientIDStr := photo.ClientID.String()
			photoResp.ClientID = &clientIDStr
		}
		photoResponses = append(photoResponses, photoResp)
	}

	c.JSON(http.StatusOK, dto.PhotoListResponse{
		UserUUID: userUUID.String(),
		Photos:   photoResponses,
	})
}

// DeletePhoto godoc
// @Summary      Удалить фотографию
// @Description  Удаление фотографии по идентификатору
// @Tags         photos
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "ID фотографии (uuid)"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /photos/{id} [delete]
// @Security     BearerAuth
func (p photoHandler) DeletePhoto(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}
