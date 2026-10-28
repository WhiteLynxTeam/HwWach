package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// assetAccess актив организации вошедшего и его действующие бизнес-роли
type assetAccess struct {
	Asset *models.Asset
	Mol   *models.AssetRole
	Users []models.AssetRole
	// вошедший — МОЛ этого актива / работает с ним
	IsMol  bool
	IsUser bool
}

// loadAsset актив только своей организации. forUpdate — блокировка строки на время транзакции
func loadAsset(ctx context.Context, db *gorm.DB, p *authz.Principal, id uuid.UUID, forUpdate bool) (*assetAccess, error) {
	q := db.WithContext(ctx)
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var a models.Asset
	if err := q.Where("uuid = ? AND organization_id = ?", id, p.OrgID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("Актив не найден")
		}
		return nil, err
	}
	acc := &assetAccess{Asset: &a}
	var roles []models.AssetRole
	if err := db.WithContext(ctx).Where("asset_id = ? AND ended_at IS NULL", id).Order("started_at").Find(&roles).Error; err != nil {
		return nil, err
	}
	for i := range roles {
		r := roles[i]
		if r.Role == models.RoleMol {
			acc.Mol = &r
			if r.PersonID == p.PersonID {
				acc.IsMol = true
			}
		} else {
			acc.Users = append(acc.Users, r)
			if r.PersonID == p.PersonID {
				acc.IsUser = true
			}
		}
	}
	return acc, nil
}

// canView все активы — с assets.view_all; иначе свои: МОЛ, работает с ним или завёл карточку
func (a *assetAccess) canView(p *authz.Principal) bool {
	return p.Has(authz.PermAssetsViewAll) || a.IsMol || a.IsUser || a.Asset.CreatedByPersonID == p.PersonID
}

// canManageRoles админ (asset_roles.assign) или МОЛ этого актива (transfers.own)
func (a *assetAccess) canManageRoles(p *authz.Principal) bool {
	return p.Has(authz.PermAssetRolesAssign) || (a.IsMol && p.Has(authz.PermTransfersOwn))
}

// canChangeStatus админ / менеджер (assets.status) или МОЛ этого актива
func (a *assetAccess) canChangeStatus(p *authz.Principal) bool {
	return p.Has(authz.PermAssetsStatus) || a.IsMol
}

// visibleAssetsScope активы, которые видит вошедший (для списков)
func visibleAssetsScope(p *authz.Principal) func(*gorm.DB) *gorm.DB {
	return func(q *gorm.DB) *gorm.DB {
		q = q.Where("assets.organization_id = ?", p.OrgID)
		if p.Has(authz.PermAssetsViewAll) {
			return q
		}
		return q.Where(
			"assets.created_by_person_id = ? OR EXISTS (SELECT 1 FROM asset_roles r WHERE r.asset_id = assets.uuid AND r.ended_at IS NULL AND r.person_id = ?)",
			p.PersonID, p.PersonID,
		)
	}
}

func newID() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
