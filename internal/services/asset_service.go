package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/dto"
	"HwWach/internal/models"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AssetView актив с действующими ролями и открытым ремонтом
type AssetView struct {
	Asset      *models.Asset
	Mol        *models.AssetRole
	Users      []models.AssetRole
	OpenRepair *models.AssetRepair
}

type CreateAssetInput struct {
	Asset       *models.Asset
	PhotoClientIDs []uuid.UUID // client_id фото с телефона
	MolPersonID *uuid.UUID // nil — по умолчанию (создатель, если он МОЛ)
	NoMol       bool       // явно «без МОЛ»
}

type AssetFilter struct {
	AccountingStatus string
	Condition        string
	Moderation       string
	MolPersonID      *uuid.UUID
	NoMol            bool
	Mine             bool
	Overdue          bool
	Query            string
	IncludeInactive  bool // списанные и утерянные по умолчанию скрыты
}

type AssetService interface {
	Create(ctx context.Context, p *authz.Principal, in CreateAssetInput) (*AssetView, error)
	Get(ctx context.Context, p *authz.Principal, id uuid.UUID) (*AssetView, error)
	List(ctx context.Context, p *authz.Principal, f AssetFilter, page, limit int) ([]*AssetView, int64, error)
	Update(ctx context.Context, p *authz.Principal, id uuid.UUID, req *dto.UpdateAssetRequest) (*AssetView, error)
	Delete(ctx context.Context, p *authz.Principal, id uuid.UUID) error
	Moderate(ctx context.Context, p *authz.Principal, id uuid.UUID, approve bool, comment string, molPersonID *uuid.UUID) (*AssetView, error)
	IsInventoryNumUnique(ctx context.Context, org uuid.UUID, num string, exclude *uuid.UUID) (bool, error)
	ListPhotos(ctx context.Context, p *authz.Principal, id uuid.UUID) ([]*models.Photo, error)
}

type assetService struct {
	db    *gorm.DB
	roles *rolesService
	orgs  OrgService
	audit AuditService
}

func NewAssetService(db *gorm.DB, roles RolesService, orgs OrgService, audit AuditService) AssetService {
	return &assetService{db: db, roles: roles.(*rolesService), orgs: orgs, audit: audit}
}

// Create карточка актива. Кто может назначить МОЛ сразу (3.2): админ — любого МОЛ, МОЛ — только себя.
// Тот, кто сфотографировал и завёл карточку, МОЛ автоматически не становится
func (s *assetService) Create(ctx context.Context, p *authz.Principal, in CreateAssetInput) (*AssetView, error) {
	a := in.Asset
	a.UUID = newID()
	a.OrganizationID = p.OrgID
	a.CreatedByAccountID = p.AccountID
	a.CreatedByPersonID = p.PersonID
	if a.AccountingStatus == "" {
		a.AccountingStatus = models.AccountingInUse
	}
	if a.Condition == "" {
		a.Condition = models.ConditionOK
	}
	if a.AccountingStatus == models.AccountingRepair || a.AccountingStatus == models.AccountingDecommissioned {
		return nil, BadRequest("Новый актив не может быть в ремонте или списан")
	}
	if p.Has(authz.PermAssetsManage) {
		now := time.Now()
		a.ModerationStatus = models.ModerationApproved
		a.VerifiedAt = &now
	} else {
		a.ModerationStatus = models.ModerationPending
	}

	var molID *uuid.UUID
	switch {
	case in.NoMol:
	case in.MolPersonID != nil:
		if *in.MolPersonID != p.PersonID && !p.Has(authz.PermAssetRolesAssign) {
			return nil, Forbidden("Назначить МОЛ другого человека может только администратор")
		}
		molID = in.MolPersonID
	case p.IsMol:
		molID = &p.PersonID
	}
	if molID != nil {
		if err := s.roles.checkPerson(ctx, p, *molID, true); err != nil {
			return nil, err
		}
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(in.PhotoClientIDs) > 0 {
			var photos []models.Photo
			if err := tx.Where("client_id IN ? AND organization_id = ?", in.PhotoClientIDs, p.OrgID).Find(&photos).Error; err != nil {
				return err
			}
			if len(photos) != len(in.PhotoClientIDs) {
				return BadRequest("some photo_client_ids were not found")
			}
			for _, ph := range photos {
				if ph.AccountUUID != p.AccountID {
					return Forbidden("photo does not belong to the user")
				}
			}
			a.Photos = photos
		}
		if err := tx.Create(a).Error; err != nil {
			if isUniqueViolation(err) {
				return Conflict("Инвентарный номер уже используется в организации")
			}
			return err
		}
		if err := s.audit.Log(ctx, tx, AuditEntry{Org: p.OrgID, Actor: p, Action: "asset.create", ObjectType: "asset", ObjectID: &a.UUID,
			Details: map[string]interface{}{"name": a.Name, "inventory_num": a.InventoryNum}}); err != nil {
			return err
		}
		if molID != nil {
			if _, err := s.roles.assignInTx(ctx, tx, p, a, nil, models.RoleMol, *molID, "initial", ""); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// первый актив — организация «наполнена» (8.6)
	s.orgs.ReportActivity(p.OrgID, true)
	return s.Get(ctx, p, a.UUID)
}

func (s *assetService) Get(ctx context.Context, p *authz.Principal, id uuid.UUID) (*AssetView, error) {
	acc, err := loadAsset(ctx, s.db, p, id, false)
	if err != nil {
		return nil, err
	}
	if !acc.canView(p) {
		return nil, errNoRights
	}
	if err := s.db.WithContext(ctx).Model(acc.Asset).Association("Photos").Find(&acc.Asset.Photos); err != nil {
		return nil, err
	}
	view := &AssetView{Asset: acc.Asset, Mol: acc.Mol, Users: acc.Users}
	var repair models.AssetRepair
	if err := s.db.WithContext(ctx).Where("asset_id = ? AND returned_at IS NULL", id).First(&repair).Error; err == nil {
		view.OpenRepair = &repair
	}
	return view, nil
}

func (s *assetService) List(ctx context.Context, p *authz.Principal, f AssetFilter, page, limit int) ([]*AssetView, int64, error) {
	q := s.db.WithContext(ctx).Model(&models.Asset{}).Scopes(visibleAssetsScope(p))
	if f.AccountingStatus != "" {
		q = q.Where("assets.accounting_status = ?", f.AccountingStatus)
	} else if !f.IncludeInactive {
		q = q.Where("assets.accounting_status NOT IN ?", []string{string(models.AccountingDecommissioned), string(models.AccountingLost)})
	}
	if f.Condition != "" {
		q = q.Where("assets.condition = ?", f.Condition)
	}
	if f.Moderation != "" {
		q = q.Where("assets.moderation_status = ?", f.Moderation)
	}
	if f.MolPersonID != nil {
		q = q.Where("EXISTS (SELECT 1 FROM asset_roles r WHERE r.asset_id = assets.uuid AND r.ended_at IS NULL AND r.role = 'mol' AND r.person_id = ?)", *f.MolPersonID)
	}
	// «Без МОЛ» у админа: списанные и утерянные туда не попадают
	if f.NoMol {
		q = q.Where("NOT EXISTS (SELECT 1 FROM asset_roles r WHERE r.asset_id = assets.uuid AND r.ended_at IS NULL AND r.role = 'mol')").
			Where("assets.accounting_status NOT IN ?", []string{string(models.AccountingDecommissioned), string(models.AccountingLost)})
	}
	if f.Mine {
		q = q.Where("EXISTS (SELECT 1 FROM asset_roles r WHERE r.asset_id = assets.uuid AND r.ended_at IS NULL AND r.person_id = ?)", p.PersonID)
	}
	if f.Overdue {
		q = q.Where("EXISTS (SELECT 1 FROM asset_repairs rp WHERE rp.asset_id = assets.uuid AND rp.returned_at IS NULL AND rp.expected_return_at < now())")
	}
	if qs := strings.TrimSpace(f.Query); qs != "" {
		like := "%" + strings.ToLower(qs) + "%"
		q = q.Where("(lower(assets.name) LIKE ? OR lower(coalesce(assets.inventory_num, '')) LIKE ? OR lower(assets.category) LIKE ?)", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 1000
	}
	if page < 1 {
		page = 1
	}
	var assets []*models.Asset
	if err := q.Order("assets.created_at DESC").Limit(limit).Offset((page - 1) * limit).Preload("Photos").Find(&assets).Error; err != nil {
		return nil, 0, err
	}
	views, err := s.withRoles(ctx, assets)
	return views, total, err
}

// withRoles роли и открытые ремонты списка одним запросом на таблицу
func (s *assetService) withRoles(ctx context.Context, assets []*models.Asset) ([]*AssetView, error) {
	views := make([]*AssetView, 0, len(assets))
	if len(assets) == 0 {
		return views, nil
	}
	ids := make([]uuid.UUID, 0, len(assets))
	byID := map[uuid.UUID]*AssetView{}
	for _, a := range assets {
		v := &AssetView{Asset: a}
		views = append(views, v)
		byID[a.UUID] = v
		ids = append(ids, a.UUID)
	}
	var roles []models.AssetRole
	if err := s.db.WithContext(ctx).Where("asset_id IN ? AND ended_at IS NULL", ids).Order("started_at").Find(&roles).Error; err != nil {
		return nil, err
	}
	for i := range roles {
		r := roles[i]
		v := byID[r.AssetID]
		if r.Role == models.RoleMol {
			v.Mol = &r
		} else {
			v.Users = append(v.Users, r)
		}
	}
	var repairs []models.AssetRepair
	if err := s.db.WithContext(ctx).Where("asset_id IN ? AND returned_at IS NULL", ids).Find(&repairs).Error; err != nil {
		return nil, err
	}
	for i := range repairs {
		r := repairs[i]
		byID[r.AssetID].OpenRepair = &r
	}
	return views, nil
}

// Update карточку до одобрения правит тот, кто её завёл; после — admin / менеджер (assets.manage).
// Статусы здесь не меняются — для них /status и ремонт
func (s *assetService) Update(ctx context.Context, p *authz.Principal, id uuid.UUID, req *dto.UpdateAssetRequest) (*AssetView, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, id, true)
		if err != nil {
			return err
		}
		a := acc.Asset
		canManage := p.Has(authz.PermAssetsManage)
		if !canManage {
			if a.CreatedByPersonID != p.PersonID {
				return Forbidden("you don't have permission to modify this asset")
			}
			if a.ModerationStatus != models.ModerationPending {
				return BadRequest("asset already processed by admin, use change request instead")
			}
		}
		if req.InventoryNum != nil {
			trimmed := strings.TrimSpace(*req.InventoryNum)
			if trimmed != "" {
				a.InventoryNum = &trimmed
			} else {
				a.InventoryNum = nil
			}
		}
		if req.Name != nil {
			a.Name = *req.Name
		}
		if req.Category != nil {
			a.Category = *req.Category
		}
		if req.CategoryUUID != nil && *req.CategoryUUID != "" {
			if parsed, err := uuid.Parse(*req.CategoryUUID); err == nil {
				a.CategoryUUID = &parsed
			}
		}
		if req.Description != nil {
			a.Description = *req.Description
		}
		if err := tx.Omit("Photos", "Requests", "CategoryObj").Save(a).Error; err != nil {
			if isUniqueViolation(err) {
				return Conflict("Инвентарный номер уже используется в организации")
			}
			return err
		}
		return s.audit.Log(ctx, tx, AuditEntry{Org: p.OrgID, Actor: p, Action: "asset.update", ObjectType: "asset", ObjectID: &a.UUID})
	})
	if err != nil {
		return nil, err
	}
	s.orgs.ReportActivity(p.OrgID, false)
	return s.Get(ctx, p, id)
}

func (s *assetService) Delete(ctx context.Context, p *authz.Principal, id uuid.UUID) error {
	if !p.Has(authz.PermAssetsManage) {
		return errNoRights
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, id, true)
		if err != nil {
			return err
		}
		// роли закрываются, иначе удалённая карточка держала бы человека (увольнение, снятие МОЛ)
		if acc.Mol != nil {
			if err := s.roles.endRoleInTx(ctx, tx, p, acc.Asset, acc.Mol, "deleted", ""); err != nil {
				return err
			}
		}
		for i := range acc.Users {
			if err := s.roles.endRoleInTx(ctx, tx, p, acc.Asset, &acc.Users[i], "deleted", ""); err != nil {
				return err
			}
		}
		if err := tx.Delete(acc.Asset).Error; err != nil {
			return err
		}
		return s.audit.Log(ctx, tx, AuditEntry{Org: p.OrgID, Actor: p, Action: "asset.delete", ObjectType: "asset", ObjectID: &id,
			Details: map[string]interface{}{"name": acc.Asset.Name}})
	})
}

// Moderate одобрение карточки — с первичным назначением МОЛ. Без МОЛ можно, актив попадёт в «Без МОЛ»
func (s *assetService) Moderate(ctx context.Context, p *authz.Principal, id uuid.UUID, approve bool, comment string, molPersonID *uuid.UUID) (*AssetView, error) {
	if !p.Has(authz.PermAssetsManage) {
		return nil, errNoRights
	}
	if molPersonID != nil {
		if err := s.roles.checkPerson(ctx, p, *molPersonID, true); err != nil {
			return nil, err
		}
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, id, true)
		if err != nil {
			return err
		}
		a := acc.Asset
		if a.ModerationStatus != models.ModerationPending {
			return Conflict("Карточка уже проверена")
		}
		now := time.Now()
		status := models.ModerationRejected
		if approve {
			status = models.ModerationApproved
		}
		if err := tx.Model(&models.Asset{}).Where("uuid = ?", a.UUID).Updates(map[string]interface{}{
			"moderation_status": status, "verified_at": now, "admin_comment": comment,
		}).Error; err != nil {
			return err
		}
		if err := s.audit.Log(ctx, tx, AuditEntry{Org: p.OrgID, Actor: p, Action: "asset.moderate", ObjectType: "asset", ObjectID: &a.UUID,
			Details: map[string]interface{}{"status": status, "comment": comment}}); err != nil {
			return err
		}
		if approve && molPersonID != nil && (acc.Mol == nil || acc.Mol.PersonID != *molPersonID) {
			if _, err := s.roles.assignInTx(ctx, tx, p, a, acc.Mol, models.RoleMol, *molPersonID, "initial", comment); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, p, id)
}

func (s *assetService) IsInventoryNumUnique(ctx context.Context, org uuid.UUID, num string, exclude *uuid.UUID) (bool, error) {
	q := s.db.WithContext(ctx).Model(&models.Asset{}).Where("organization_id = ? AND inventory_num = ?", org, num)
	if exclude != nil {
		q = q.Where("uuid <> ?", *exclude)
	}
	var n int64
	err := q.Count(&n).Error
	return n == 0, err
}

func (s *assetService) ListPhotos(ctx context.Context, p *authz.Principal, id uuid.UUID) ([]*models.Photo, error) {
	acc, err := loadAsset(ctx, s.db, p, id, false)
	if err != nil {
		return nil, err
	}
	if !acc.canView(p) {
		return nil, errNoRights
	}
	var photos []*models.Photo
	err = s.db.WithContext(ctx).Model(&models.Photo{}).
		Joins("join asset_photos on asset_photos.photo_uuid = photos.uuid").
		Where("asset_photos.asset_uuid = ?", id).
		Find(&photos).Error
	return photos, err
}
