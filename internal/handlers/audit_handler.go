package handlers

import (
	"HwWach/internal/authz"
	"HwWach/internal/middleware"
	"HwWach/internal/services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type auditHandler struct {
	audit services.AuditService
}

func NewAuditHandler(audit services.AuditService) AuditHandler {
	return &auditHandler{audit: audit}
}

// List godoc
// @Summary      Журнал go: активы, роли, передачи, статусы (audit.view)
// @Description  Журнал ts (люди, учётки, регистрации) — GET /audit в ts; экран журнала сливает оба по времени
// @Tags         audit
// @Produce      json
// @Param        limit      query  int     false  "До 500"
// @Param        before     query  string  false  "RFC3339: записи раньше"
// @Param        object_id  query  string  false  "Актив или другой объект"
// @Success      200  {array}  models.AuditLog
// @Router       /audit [get]
// @Security     BearerAuth
func (h *auditHandler) List(c *gin.Context) {
	p, ok := middleware.RequirePerm(c, authz.PermAuditView)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	var before *time.Time
	if s := c.Query("before"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			before = &t
		}
	}
	var objectID *uuid.UUID
	if s := c.Query("object_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			objectID = &id
		}
	}
	rows, err := h.audit.List(c.Request.Context(), p.OrgID, limit, before, objectID)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, rows)
}
