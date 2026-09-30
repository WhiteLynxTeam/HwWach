package handlers

import (
	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/models"
	"HwWach/internal/services"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type categoryHandler struct {
	categorySvc services.CategoryService
}

func NewCategoryHandler(categorySvc services.CategoryService) CategoryHandler {
	return &categoryHandler{
		categorySvc: categorySvc,
	}
}

// Search godoc
// @Summary      Нечёткий поиск категорий
// @Description  Поиск категорий с автодополнением и нечётким сравнением (pg_trgm). Поиск выполняется начиная от 3 символов. При пустом запросе возвращаются популярные категории.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        q    query     string  false  "Поисковый запрос (минимум 3 символа)"
// @Success      200  {array}   dto.CategoryResponse
// @Failure      401  {object}  map[string]string
// @Router       /categories/search [get]
// @Security     BearerAuth
func (h *categoryHandler) Search(c *gin.Context) {
	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	query := strings.TrimSpace(c.Query("q"))
	runeCount := utf8.RuneCountInString(query)

	// Если введены 1 или 2 символа, поиск не запускаем, чтобы избежать нерелевантной выдачи
	if runeCount > 0 && runeCount < 3 {
		c.JSON(http.StatusOK, []dto.CategoryResponse{})
		return
	}

	categories, err := h.categorySvc.Search(c.Request.Context(), query, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search categories: " + err.Error()})
		return
	}

	responses := make([]dto.CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		responses = append(responses, catToResponse(cat))
	}

	c.JSON(http.StatusOK, responses)
}

// Sync godoc
// @Summary      Синхронизация категорий
// @Description  Инкрементальная или полная синхронизация категорий для локальной базы Room клиента.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        since  query     string  false  "Временная метка последней синхронизации (RFC3339)"
// @Success      200    {object}  dto.CategorySyncResponse
// @Failure      401    {object}  map[string]string
// @Router       /categories/sync [get]
// @Security     BearerAuth
func (h *categoryHandler) Sync(c *gin.Context) {
	userUUID, ok := middleware.RequireUserUUID(c)
	if !ok {
		return
	}

	var sinceTime *time.Time
	sinceStr := c.Query("since")
	if sinceStr != "" {
		parsed, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'since' timestamp format, expected RFC3339: " + err.Error()})
			return
		}
		sinceTime = &parsed
	}

	resp, err := h.categorySvc.Sync(c.Request.Context(), userUUID, sinceTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sync categories: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListPending godoc
// @Summary      Список пользовательских категорий на модерацию (Admin)
// @Description  Получение списка всех L3 категорий со статусом pending
// @Tags         categories
// @Accept       json
// @Produce      json
// @Success      200  {array}   dto.CategoryResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /categories/pending [get]
// @Security     BearerAuth
func (h *categoryHandler) ListPending(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	categories, err := h.categorySvc.ListPending(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list pending categories: " + err.Error()})
		return
	}

	responses := make([]dto.CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		responses = append(responses, catToResponse(cat))
	}

	c.JSON(http.StatusOK, responses)
}

// Moderate godoc
// @Summary      Модерация пользовательской категории (Admin)
// @Description  Одобрение (approve), объединение (merge) или отклонение (reject) пользовательской L3 категории
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id       path      string                      true  "UUID категории"
// @Param        request  body      dto.ModerateCategoryRequest  true  "Данные модерации"
// @Success      200      {object}  dto.CategoryResponse
// @Failure      400      {object}  map[string]string
// @Failure      401      {object}  map[string]string
// @Failure      403      {object}  map[string]string
// @Router       /categories/{id}/moderate [patch]
// @Security     BearerAuth
func (h *categoryHandler) Moderate(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	idStr := c.Param("id")
	categoryUUID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category uuid: " + err.Error()})
		return
	}

	var req dto.ModerateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	updated, err := h.categorySvc.Moderate(c.Request.Context(), categoryUUID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, catToResponse(updated))
}

func catToResponse(c *models.Category) dto.CategoryResponse {
	resp := dto.CategoryResponse{
		UUID:         c.UUID.String(),
		Name:         c.Name,
		Level:        c.Level,
		Status:       string(c.Status),
		UsageCount:   c.UsageCount,
		AdminComment: c.AdminComment,
		CreatedAt:    c.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    c.UpdatedAt.Format(time.RFC3339),
	}
	if c.CreatedBy != nil {
		createdByStr := c.CreatedBy.String()
		resp.CreatedBy = &createdByStr
	}
	return resp
}
