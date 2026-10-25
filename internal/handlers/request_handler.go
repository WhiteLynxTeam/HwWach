package handlers

import (
	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/models"
	"HwWach/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type requestHandler struct {
	reqSvc services.RequestService
}

func NewRequestHandler(reqSvc services.RequestService) RequestHandler {
	return &requestHandler{
		reqSvc: reqSvc,
	}
}

// CreateRequest godoc
// @Summary      Заявка на ремонт
// @Description  Подаёт любой, кто работает с активом; можно сразу отметить «неисправен»
// @Tags         requests
// @Accept       json
// @Produce      json
// @Param        request  body      dto.CreateRepairRequest  true  "Данные заявки"
// @Success      201  {object}  models.Request
// @Router       /requests [post]
// @Security     BearerAuth
func (r requestHandler) CreateRequest(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	var req dto.CreateRepairRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	assetID, err := uuid.Parse(req.AssetUUID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_uuid"})
		return
	}
	out, err := r.reqSvc.Create(c.Request.Context(), p, assetID, req.Message, req.PhotoURL, req.MarkFaulty)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// ListRequests godoc
// @Summary      Заявки на ремонт
// @Description  Админ / менеджер — все заявки организации; остальные — свои и по активам, где они МОЛ
// @Tags         requests
// @Produce      json
// @Param        status    query  string  false  "new / in_repair / done / rejected"
// @Param        asset_id  query  string  false  "Актив"
// @Success      200  {array}  models.Request
// @Router       /requests [get]
// @Security     BearerAuth
func (r requestHandler) ListRequests(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	f := services.RequestFilter{Status: c.Query("status")}
	if s := c.Query("asset_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			f.AssetID = &id
		}
	}
	list, err := r.reqSvc.List(c.Request.Context(), p, f)
	if err != nil {
		writeErr(c, err)
		return
	}
	if list == nil {
		list = []models.Request{}
	}
	c.JSON(http.StatusOK, list)
}

// GetRequest godoc
// @Summary      Заявка на ремонт по ID
// @Tags         requests
// @Produce      json
// @Param        id  path      string  true  "ID запроса (uuid)"
// @Success      200  {object}  models.Request
// @Router       /requests/{id} [get]
// @Security     BearerAuth
func (r requestHandler) GetRequest(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	out, err := r.reqSvc.Get(c.Request.Context(), p, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// UpdateRequestStatus godoc
// @Summary      Закрыть заявку: выполнена / отклонена (МОЛ актива или админ)
// @Description  «В ремонте» ставится действием «В ремонт» с этой заявкой
// @Tags         requests
// @Accept       json
// @Produce      json
// @Param        id       path  string                          true  "ID запроса"
// @Param        request  body  dto.UpdateRepairRequestStatus  true  "Статус"
// @Success      200  {object}  models.Request
// @Router       /requests/{id} [patch]
// @Security     BearerAuth
func (r requestHandler) UpdateRequestStatus(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateRepairRequestStatus
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := r.reqSvc.SetStatus(c.Request.Context(), p, id, models.RequestStatus(req.Status), req.Comment)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// DeleteRequest godoc
// @Summary      Отозвать свою новую заявку
// @Tags         requests
// @Param        id  path      string  true  "ID запроса (uuid)"
// @Success      200  {object}  map[string]string
// @Router       /requests/{id} [delete]
// @Security     BearerAuth
func (r requestHandler) DeleteRequest(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := r.reqSvc.Delete(c.Request.Context(), p, id); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
