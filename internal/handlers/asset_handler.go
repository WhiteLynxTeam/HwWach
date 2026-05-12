package handlers

import (
	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/models"
	"HwWach/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type assetHandler struct {
	assetSvc services.AssetService
	photoSvc services.PhotoService
}

func NewAssetHandler(assetSvc services.AssetService, photoSvc services.PhotoService) AssetHandler {
	return &assetHandler{
		assetSvc: assetSvc,
		photoSvc: photoSvc,
	}
}

// CreateAsset godoc
// @Summary      Создать asset
// @Description  Создание нового asset в системе
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateAssetRequest  true  "Данные asset"
// @Success      201      {object}  dto.AssetResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /assets [post]
// @Security     BearerAuth
func (a assetHandler) CreateAsset(c *gin.Context) {
	var req dto.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	asset := &models.Asset{
		InventoryNum:  req.InventoryNum,
		Type:          req.Type,
		Specification: req.Specification,
		Status:        models.AssetStatus(req.Status),
		UserUUID:      userUUID,
	}

	if req.ClientID != nil {
		clientUUID, err := uuid.Parse(*req.ClientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id: " + err.Error()})
			return
		}
		asset.ClientID = &clientUUID
	}

	if req.Status == "" {
		asset.Status = models.AssetStatusActive
	}

	if err := a.assetSvc.Create(c.Request.Context(), asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create asset: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, assetToResponse(asset))
}

// ListUserAssets godoc
// @Summary      Список assets пользователя
// @Description  Получение всех assets авторизованного пользователя (без вложений)
// @Tags         assets
// @Accept       json
// @Produce      json
// @Success      200  {object}  dto.AssetListResponse
// @Failure      401  {object}  map[string]string
// @Router       /assets [get]
// @Security     BearerAuth
func (a assetHandler) ListUserAssets(c *gin.Context) {
	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	assets, err := a.assetSvc.GetAllByUserUUID(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Конвертируем в DTO без вложений
	assetDTOs := make([]dto.AssetResponse, 0, len(assets))
	for _, asset := range assets {
		assetDTOs = append(assetDTOs, assetToResponse(asset))
	}

	c.JSON(http.StatusOK, dto.AssetListResponse{
		UserUUID: userUUID.String(),
		Assets:   assetDTOs,
	})
}

// GetAsset godoc
// @Summary      Получить asset по ID
// @Description  Получение информации об asset по идентификатору
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "ID asset (uuid)"
// @Success      200  {object}  models.Asset
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /assets/{id} [get]
// @Security     BearerAuth
func (a assetHandler) GetAsset(c *gin.Context) {
	assetUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return
	}

	asset, err := a.assetSvc.GetByUUID(c.Request.Context(), assetUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	c.JSON(http.StatusOK, asset)
}

// ListAssetPhotos godoc
// @Summary      Список фотографий asset
// @Description  Получение всех фотографий для указанного asset
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "ID asset (uuid)"
// @Success      200  {array}   models.Photo
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Router       /assets/{id}/photos [get]
// @Security     BearerAuth
func (a assetHandler) ListAssetPhotos(c *gin.Context) {
	assetUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return
	}

	photos, err := a.assetSvc.ListPhotos(c.Request.Context(), assetUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Конвертируем модели в DTO с полными URL
	photoResponses := make([]dto.PhotoResponse, 0, len(photos))
	for _, photo := range photos {
		photoResp := dto.PhotoResponse{
			UUID:      photo.UUID.String(),
			URL:       a.photoSvc.GetPublicURL(photo.URL),
			CreatedAt: photo.CreatedAt.Format(time.RFC3339),
		}
		if photo.ClientID != nil {
			clientIDStr := photo.ClientID.String()
			photoResp.ClientID = &clientIDStr
		}
		photoResponses = append(photoResponses, photoResp)
	}

	c.JSON(http.StatusOK, gin.H{
		"asset_uuid": assetUUID.String(),
		"photos":     photoResponses,
	})
}

// assetToResponse конвертирует модель Asset в DTO AssetResponse
func assetToResponse(asset *models.Asset) dto.AssetResponse {
	resp := dto.AssetResponse{
		UUID:          asset.UUID.String(),
		InventoryNum:  asset.InventoryNum,
		Type:          asset.Type,
		Specification: asset.Specification,
		Status:        string(asset.Status),
		UserUUID:      asset.UserUUID.String(),
		AdminComment:  asset.AdminComment,
		CreatedAt:     asset.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     asset.UpdatedAt.Format(time.RFC3339),
	}

	if asset.ClientID != nil {
		clientIDStr := asset.ClientID.String()
		resp.ClientID = &clientIDStr
	}

	if asset.VerifiedAt != nil {
		verifiedStr := asset.VerifiedAt.Format(time.RFC3339)
		resp.VerifiedAt = &verifiedStr
	}

	return resp
}
