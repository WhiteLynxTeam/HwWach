package handlers

import (
	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/models"
	"HwWach/internal/services"
	"net/http"
	"strconv"
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
		InventoryNum: req.InventoryNum,
		Name:         req.Name,
		Category:     req.Category,
		Description:  req.Description,
		AssetStatus:  models.AssetStatus(req.AssetStatus),
		UserUUID:     userUUID,
	}

	if req.ClientID != nil {
		clientUUID, err := uuid.Parse(*req.ClientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id: " + err.Error()})
			return
		}
		asset.ClientID = &clientUUID
	}

	var photoClientUUIDs []uuid.UUID
	for _, idStr := range req.PhotoClientIDs {
		parsedUUID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid photo_client_id: " + err.Error()})
			return
		}
		photoClientUUIDs = append(photoClientUUIDs, parsedUUID)
	}

	if len(photoClientUUIDs) > 0 {
		photos, err := a.photoSvc.GetByClientIDs(c.Request.Context(), photoClientUUIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch photos: " + err.Error()})
			return
		}

		if len(photos) != len(photoClientUUIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "some photo_client_ids were not found"})
			return
		}

		var assetPhotos []models.Photo
		for _, photo := range photos {
			if photo.UserUUID != userUUID {
				c.JSON(http.StatusForbidden, gin.H{"error": "photo does not belong to the user"})
				return
			}
			assetPhotos = append(assetPhotos, *photo)
		}
		asset.Photos = assetPhotos
	}

	if req.AssetStatus == "" {
		asset.AssetStatus = models.AssetStatusActive
	}

	if err := a.assetSvc.Create(c.Request.Context(), asset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create asset: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, assetToResponse(asset))
}

// ListUserAssets godoc
// @Summary      Список assets пользователя
// @Description  Получение всех assets авторизованного пользователя с прикрепленными фотографиями
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

	// Конвертируем в DTO с фотографиями
	assetDTOs := make([]dto.AssetWithPhotosResponse, 0, len(assets))
	for _, asset := range assets {
		baseResp := assetToResponse(asset)

		photoResponses := make([]dto.PhotoResponse, 0, len(asset.Photos))
		for _, photo := range asset.Photos {
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

		assetDTOs = append(assetDTOs, dto.AssetWithPhotosResponse{
			AssetResponse: baseResp,
			Photos:        photoResponses,
		})
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

// UpdateAsset godoc
// @Summary      Редактировать asset (до проверки)
// @Description  Частичное обновление asset, если он еще не проверен администратором
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        id       path      string                  true  "ID asset (uuid)"
// @Param        request  body      dto.UpdateAssetRequest  true  "Данные для обновления"
// @Success      200      {object}  dto.AssetResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /assets/{id} [put]
// @Security     BearerAuth
func (a assetHandler) UpdateAsset(c *gin.Context) {
	assetUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid uuid format"})
		return
	}

	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	var req dto.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	updatedAsset, err := a.assetSvc.UpdatePending(c.Request.Context(), userUUID, assetUUID, &req)
	if err != nil {
		// Упрощенная обработка ошибок (в реальности можно проверять типы ошибок)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, assetToResponse(updatedAsset))
}

// CheckInventoryUnique godoc
// @Summary      Проверить уникальность инвентарного номера
// @Description  Проверка уникальности инвентарного номера по всей таблице assets
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        num      query     string  true  "Инвентарный номер"
// @Success      200      {object}  map[string]bool "Результат проверки, например {"unique": true}"
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /assets/check-inventory [get]
// @Security     BearerAuth
func (a assetHandler) CheckInventoryUnique(c *gin.Context) {
	inventoryNum := c.Query("num")
	if inventoryNum == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "num query parameter is required"})
		return
	}

	isUnique, err := a.assetSvc.IsInventoryNumUnique(c.Request.Context(), inventoryNum)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check uniqueness: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unique": isUnique})
}

// assetToResponse конвертирует модель Asset в DTO AssetResponse
func assetToResponse(asset *models.Asset) dto.AssetResponse {
	resp := dto.AssetResponse{
		UUID:         asset.UUID.String(),
		InventoryNum: asset.InventoryNum,
		Name:         asset.Name,
		Category:     asset.Category,
		Description:  asset.Description,
		AssetStatus:  string(asset.AssetStatus),
		UserUUID:     asset.UserUUID.String(),
		AdminComment: asset.AdminComment,
		CreatedAt:    asset.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    asset.UpdatedAt.Format(time.RFC3339),
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

// ListUserAssetsPaginated godoc
// @Summary      Список assets с пагинацией и фотографиями
// @Description  Получение списка assets с поддержкой пагинации и массивом всех прикрепленных фотографий к каждому asset. Админы видят все assets, обычные пользователи только свои.
// @Tags         assets
// @Accept       json
// @Produce      json
// @Param        page   query     int  false  "Номер страницы" default(1)
// @Param        limit  query     int  false  "Количество элементов на странице" default(10)
// @Success      200  {object}  dto.PaginatedAssetResponse
// @Failure      401  {object}  map[string]string
// @Router       /assets/paginated [get]
// @Security     BearerAuth
func (a assetHandler) ListUserAssetsPaginated(c *gin.Context) {
	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	isAdmin := middleware.IsAdmin(c)

	var filterUUID *uuid.UUID
	if !isAdmin {
		filterUUID = &userUUID
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	assets, total, err := a.assetSvc.GetPaginated(c.Request.Context(), filterUUID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	assetResponses := make([]dto.AssetWithPhotosResponse, 0, len(assets))
	for _, asset := range assets {
		// Конвертируем базовую инфо об asset
		baseResp := assetToResponse(asset)

		// Конвертируем фотографии с полными URL
		photoResponses := make([]dto.PhotoResponse, 0, len(asset.Photos))
		for _, photo := range asset.Photos {
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

		var createdBy *string
		if isAdmin {
			createdByUserStr := asset.UserUUID.String()
			createdBy = &createdByUserStr
		}

		assetResponses = append(assetResponses, dto.AssetWithPhotosResponse{
			AssetResponse: baseResp,
			CreatedBy:     createdBy,
			Photos:        photoResponses,
		})
	}

	pages := 0
	if limit > 0 {
		pages = int((total + int64(limit) - 1) / int64(limit))
	}

	c.JSON(http.StatusOK, dto.PaginatedAssetResponse{
		Assets: assetResponses,
		Total:  total,
		Page:   page,
		Limit:  limit,
		Pages:  pages,
	})
}
