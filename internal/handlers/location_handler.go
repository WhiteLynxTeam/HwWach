package handlers

import (
	"HwWach/internal/dto"
	"HwWach/internal/middleware"
	"HwWach/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type locationHandler struct {
	svc services.LocationService
}

func NewLocationHandler(svc services.LocationService) LocationHandler {
	return &locationHandler{svc: svc}
}

// List godoc
// @Summary      Справочник мест организации
// @Description  Все места организации, отсортированы по пути. q — поиск по пути («ферма 700»)
// @Tags         locations
// @Produce      json
// @Param        q  query  string  false  "Поиск по пути"
// @Success      200  {array}  dto.LocationResponse
// @Router       /locations [get]
// @Security     BearerAuth
func (h *locationHandler) List(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	list, err := h.svc.List(c.Request.Context(), p, c.Query("q"))
	if err != nil {
		writeErr(c, err)
		return
	}
	out := make([]dto.LocationResponse, 0, len(list))
	for i := range list {
		out = append(out, locationToResponse(&list[i]))
	}
	c.JSON(http.StatusOK, out)
}

// Create godoc
// @Summary      Завести место
// @Description  Цепочка от корня: готовые узлы находятся по id или имени, недостающие заводятся. Возвращает последний
// @Tags         locations
// @Accept       json
// @Produce      json
// @Param        request  body  dto.CreateLocationRequest  true  "Цепочка"
// @Success      200  {object}  dto.LocationResponse
// @Router       /locations [post]
// @Security     BearerAuth
func (h *locationHandler) Create(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	var req dto.CreateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	chain, err := parseLocationChain(req.Chain)
	if err != nil {
		writeErr(c, err)
		return
	}
	loc, err := h.svc.Create(c.Request.Context(), p, chain)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, locationToResponse(loc))
}
