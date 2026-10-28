package handlers

import (
	"HwWach/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Внутренние запросы ts → go (5.1). Защищены ключом X-Internal-Key, в swagger не показываются
type internalHandler struct {
	roles services.RolesService
	orgs  services.OrgService
}

func NewInternalHandler(roles services.RolesService, orgs services.OrgService) InternalHandler {
	return &internalHandler{roles: roles, orgs: orgs}
}

// PersonAssetsCount за человеком числятся активы? Перед увольнением и снятием признака МОЛ
func (h *internalHandler) PersonAssetsCount(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	mol, user, err := h.roles.MolCount(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"mol_count": mol, "user_count": user})
}

// PersonDismissed увольнение: закрыть роли «работает с активом»
func (h *internalHandler) PersonDismissed(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	n, err := h.roles.PersonDismissed(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"closed": n})
}

// DeleteOrganization окончательное удаление организации: данные и фото. Повторный вызов безопасен
func (h *internalHandler) DeleteOrganization(c *gin.Context) {
	id, ok := parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.orgs.DeleteOrganizationData(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
