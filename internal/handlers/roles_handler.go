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

type rolesHandler struct {
	roles services.RolesService
}

func NewRolesHandler(roles services.RolesService) RolesHandler {
	return &rolesHandler{roles: roles}
}

func bindAssign(c *gin.Context) (uuid.UUID, uuid.UUID, string, bool) {
	assetID, ok := parseUUIDParam(c, "id")
	if !ok {
		return uuid.Nil, uuid.Nil, "", false
	}
	var req dto.AssignPersonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return uuid.Nil, uuid.Nil, "", false
	}
	personID, err := uuid.Parse(req.PersonID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person_id"})
		return uuid.Nil, uuid.Nil, "", false
	}
	return assetID, personID, req.Comment, true
}

// AssignMol godoc
// @Summary      Передать актив МОЛ (сразу, «без подтверждения»)
// @Description  Админ — любому МОЛ; МОЛ актива — свой актив другому МОЛ. Получатель потом подтверждает или отказывается
// @Tags         asset-roles
// @Accept       json
// @Produce      json
// @Param        id       path  string                   true  "ID asset"
// @Param        request  body  dto.AssignPersonRequest  true  "Новый МОЛ"
// @Success      200  {object}  dto.AssetRoleResponse
// @Router       /assets/{id}/mol [post]
// @Security     BearerAuth
func (h *rolesHandler) AssignMol(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	assetID, personID, comment, ok := bindAssign(c)
	if !ok {
		return
	}
	r, err := h.roles.AssignMol(c.Request.Context(), p, assetID, personID, comment)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, roleToResponse(r))
}

// RemoveMol godoc
// @Summary      Снять МОЛ без передачи (только админ) — актив попадёт в «Без МОЛ»
// @Tags         asset-roles
// @Param        id  path  string  true  "ID asset"
// @Success      200  {object}  map[string]string
// @Router       /assets/{id}/mol [delete]
// @Security     BearerAuth
func (h *rolesHandler) RemoveMol(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	assetID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.roles.RemoveMol(c.Request.Context(), p, assetID, c.Query("comment")); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

// AddUser godoc
// @Summary      Назначить работающего с активом (админ или МОЛ актива)
// @Description  Человеку учётная запись не нужна
// @Tags         asset-roles
// @Accept       json
// @Produce      json
// @Param        id       path  string                   true  "ID asset"
// @Param        request  body  dto.AssignPersonRequest  true  "Человек"
// @Success      200  {object}  dto.AssetRoleResponse
// @Router       /assets/{id}/users [post]
// @Security     BearerAuth
func (h *rolesHandler) AddUser(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	assetID, personID, comment, ok := bindAssign(c)
	if !ok {
		return
	}
	r, err := h.roles.AddUser(c.Request.Context(), p, assetID, personID, comment)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, roleToResponse(r))
}

// RemoveUser godoc
// @Summary      Снять работающего с активом (админ или МОЛ актива)
// @Tags         asset-roles
// @Param        id         path  string  true  "ID asset"
// @Param        person_id  path  string  true  "ID человека"
// @Success      200  {object}  map[string]string
// @Router       /assets/{id}/users/{person_id} [delete]
// @Security     BearerAuth
func (h *rolesHandler) RemoveUser(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	assetID, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	personID, ok := parseUUIDParam(c, "person_id")
	if !ok {
		return
	}
	if err := h.roles.RemoveUser(c.Request.Context(), p, assetID, personID, c.Query("comment")); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "removed"})
}

func assignmentsToResponse(roles []models.AssetRole) []dto.AssignmentResponse {
	out := make([]dto.AssignmentResponse, 0, len(roles))
	for i := range roles {
		item := dto.AssignmentResponse{AssetRoleResponse: roleToResponse(&roles[i])}
		if roles[i].Asset != nil {
			a := assetToResponse(roles[i].Asset)
			item.Asset = &a
		}
		out = append(out, item)
	}
	return out
}

// MyAssignments godoc
// @Summary      «Мне назначено»: что числится за мной и с чем я работаю
// @Tags         asset-roles
// @Produce      json
// @Success      200  {array}  dto.AssignmentResponse
// @Router       /me/assignments [get]
// @Security     BearerAuth
func (h *rolesHandler) MyAssignments(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	roles, err := h.roles.MyAssignments(c.Request.Context(), p)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, assignmentsToResponse(roles))
}

// Ack godoc
// @Summary      Подтвердить или отказаться от назначения (с причиной)
// @Description  Актив остаётся, где поставили; отказ попадает админу «на разбор»
// @Tags         asset-roles
// @Accept       json
// @Produce      json
// @Param        id       path  string          true  "ID назначения (asset_roles)"
// @Param        request  body  dto.AckRequest  true  "accept / dispute"
// @Success      200  {object}  dto.AssignmentResponse
// @Router       /asset-roles/{id}/ack [post]
// @Security     BearerAuth
func (h *rolesHandler) Ack(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.AckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	r, err := h.roles.Ack(c.Request.Context(), p, id, req.Action == "accept", req.Reason)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, assignmentsToResponse([]models.AssetRole{*r})[0])
}

// Disputes godoc
// @Summary      «На разбор»: отказы от назначений (asset_roles.assign)
// @Tags         asset-roles
// @Produce      json
// @Success      200  {array}  dto.AssignmentResponse
// @Router       /asset-roles/disputes [get]
// @Security     BearerAuth
func (h *rolesHandler) Disputes(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	roles, err := h.roles.Disputes(c.Request.Context(), p)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, assignmentsToResponse(roles))
}

// Resolve godoc
// @Summary      Отказ разобран (оставить как есть; переназначить — обычной передачей)
// @Tags         asset-roles
// @Accept       json
// @Produce      json
// @Param        id       path  string              true  "ID назначения"
// @Param        request  body  dto.ResolveRequest  false "Решение"
// @Success      200  {object}  dto.AssetRoleResponse
// @Router       /asset-roles/{id}/resolve [post]
// @Security     BearerAuth
func (h *rolesHandler) Resolve(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req dto.ResolveRequest
	_ = c.ShouldBindJSON(&req)
	r, err := h.roles.Resolve(c.Request.Context(), p, id, req.Resolution)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, roleToResponse(r))
}

type transferResponse struct {
	ID           string  `json:"id"`
	Role         string  `json:"role"`
	FromPersonID *string `json:"from_person_id,omitempty"`
	ToPersonID   *string `json:"to_person_id,omitempty"`
	ByPersonID   *string `json:"initiated_by_person_id,omitempty"`
	Reason       *string `json:"reason,omitempty"`
	Comment      *string `json:"comment,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// Transfers godoc
// @Summary      История передач и всех ролей актива
// @Tags         asset-roles
// @Produce      json
// @Param        id  path  string  true  "ID asset"
// @Success      200  {object}  map[string]interface{}
// @Router       /assets/{id}/transfers [get]
// @Security     BearerAuth
func (h *rolesHandler) Transfers(c *gin.Context) {
	p, ok := middleware.CurrentPrincipal(c)
	if !ok {
		return
	}
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	transfers, roles, err := h.roles.Transfers(c.Request.Context(), p, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	tr := make([]transferResponse, 0, len(transfers))
	for _, t := range transfers {
		tr = append(tr, transferResponse{
			ID: t.ID.String(), Role: string(t.Role), FromPersonID: uuidStr(t.FromPersonID), ToPersonID: uuidStr(t.ToPersonID),
			ByPersonID: uuidStr(t.InitiatedByPersonID), Reason: t.Reason, Comment: t.Comment, CreatedAt: t.CreatedAt.Format(time.RFC3339),
		})
	}
	rr := make([]dto.AssetRoleResponse, 0, len(roles))
	for i := range roles {
		rr = append(rr, roleToResponse(&roles[i]))
	}
	c.JSON(http.StatusOK, gin.H{"transfers": tr, "roles": rr})
}
