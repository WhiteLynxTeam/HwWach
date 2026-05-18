package handlers

import (
	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/models"
	"HwWach/internal/services"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type assetChangeRequestHandler struct {
	svc services.AssetChangeRequestService
}

func NewAssetChangeRequestHandler(svc services.AssetChangeRequestService) AssetChangeRequestHandler {
	return &assetChangeRequestHandler{
		svc: svc,
	}
}

// CreateRequest godoc
// @Summary      Создать заявку на изменение
// @Description  Пользователь создает заявку на редактирование или удаление своего актива
// @Tags         asset-change-requests
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "ID актива (uuid)"
// @Param        request  body      dto.CreateChangeRequestInput  true  "Данные заявки"
// @Success      201      {object}  dto.AssetChangeRequestResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      404      {object}  map[string]string
// @Router       /assets/{id}/change-requests [post]
// @Security     BearerAuth
func (h *assetChangeRequestHandler) CreateRequest(c *gin.Context) {
	assetUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset uuid"})
		return
	}

	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	var input dto.CreateChangeRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := h.svc.CreateRequest(c.Request.Context(), userUUID, assetUUID, input)
	if err != nil {
		// В реальном приложении можно различать 404 и 400
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, changeRequestToResponse(req))
}

// ListPending godoc
// @Summary      Список всех заявок (Admin only)
// @Description  Admin only. Получить список всех заявок со статусом pending (для администратора)
// @Tags         asset-change-requests
// @Accept       json
// @Produce      json
// @Success      200      {array}   dto.AssetChangeRequestResponse
// @Failure      401      {object}  map[string]string
// @Router       /assets/change-requests [get]
// @Security     BearerAuth
func (h *assetChangeRequestHandler) ListPending(c *gin.Context) {
	requests, err := h.svc.ListPending(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var res []dto.AssetChangeRequestResponse
	for _, r := range requests {
		res = append(res, changeRequestToResponse(r))
	}

	if res == nil {
		res = make([]dto.AssetChangeRequestResponse, 0)
	}

	c.JSON(http.StatusOK, res)
}

// ApproveRequest godoc
// @Summary      Одобрить/отклонить заявку (Admin only)
// @Description  Admin only. Администратор меняет статус заявки и применяет её (если approved)
// @Tags         asset-change-requests
// @Accept       json
// @Produce      json
// @Param        id       path      string                        true  "ID заявки (uuid)"
// @Param        request  body      dto.ApproveChangeRequestInput true  "Новый статус и комментарий"
// @Success      200      {object}  map[string]string
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Router       /assets/change-requests/{id} [patch]
// @Security     BearerAuth
func (h *assetChangeRequestHandler) ApproveRequest(c *gin.Context) {
	reqUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request uuid"})
		return
	}

	// Для админов можно доставать ID админа
	adminUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	var input dto.ApproveChangeRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.ApproveRequest(c.Request.Context(), adminUUID, reqUUID, input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func changeRequestToResponse(r *models.AssetChangeRequest) dto.AssetChangeRequestResponse {
	var proposedData json.RawMessage
	if len(r.ProposedData) > 0 {
		proposedData = json.RawMessage(r.ProposedData)
	}

	return dto.AssetChangeRequestResponse{
		UUID:         r.UUID.String(),
		AssetUUID:    r.AssetUUID.String(),
		UserUUID:     r.UserUUID.String(),
		Type:         string(r.Type),
		ProposedData: proposedData,
		Reason:       r.Reason,
		AdminComment: r.AdminComment,
		Status:       string(r.Status),
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}
