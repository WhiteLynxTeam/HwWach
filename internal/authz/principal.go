// Package authz права и данные вошедшего из access-токена ts (org-roles-plan.md, 2.4 и 5)
package authz

import "github.com/google/uuid"

// В коде проверяем право, а не роль
const (
	PermAssetsManage         = "assets.manage"
	PermAssetsViewAll        = "assets.view_all"
	PermAssetsPropose        = "assets.propose"
	PermAssetsStatus         = "assets.status"
	PermAssetsDecommission   = "assets.decommission"
	PermAssetRolesAssign     = "asset_roles.assign"
	PermTransfersOwn         = "transfers.own"
	PermCategoriesManage     = "categories.manage"
	PermChangeRequestsReview = "change_requests.review"
	PermPhotosReview         = "photos.review"
	PermRequestsCreate       = "requests.create"
	PermAuditView            = "audit.view"
)

// Principal вошедший. Человек и организация — в момент действия (из токена):
// после передачи админства старый токен подписывает действия прежним человеком (3.4)
type Principal struct {
	AccountID uuid.UUID
	PersonID  uuid.UUID
	OrgID     uuid.UUID
	Login     string
	Role      string
	Perms     map[string]bool
	IsMol     bool
	IP        string
}

func (p *Principal) Has(perm string) bool {
	return p != nil && p.Perms[perm]
}
