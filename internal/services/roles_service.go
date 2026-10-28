package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/models"
	"HwWach/internal/tsclient"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Бизнес-роли и передачи (3.2, 3.3). Всё применяется сразу, с пометкой «без подтверждения»;
// получатель потом подтверждает или отказывается с причиной — актив остаётся, где поставили
type RolesService interface {
	AssignMol(ctx context.Context, p *authz.Principal, assetID, personID uuid.UUID, comment string) (*models.AssetRole, error)
	RemoveMol(ctx context.Context, p *authz.Principal, assetID uuid.UUID, comment string) error
	AddUser(ctx context.Context, p *authz.Principal, assetID, personID uuid.UUID, comment string) (*models.AssetRole, error)
	RemoveUser(ctx context.Context, p *authz.Principal, assetID, personID uuid.UUID, comment string) error

	MyAssignments(ctx context.Context, p *authz.Principal) ([]models.AssetRole, error)
	Ack(ctx context.Context, p *authz.Principal, roleID uuid.UUID, accept bool, reason string) (*models.AssetRole, error)
	Disputes(ctx context.Context, p *authz.Principal) ([]models.AssetRole, error)
	Resolve(ctx context.Context, p *authz.Principal, roleID uuid.UUID, resolution string) (*models.AssetRole, error)
	Transfers(ctx context.Context, p *authz.Principal, assetID uuid.UUID) ([]models.AssetTransfer, []models.AssetRole, error)

	// для ts (5.1)
	MolCount(ctx context.Context, personID uuid.UUID) (int64, int64, error)
	PersonDismissed(ctx context.Context, personID uuid.UUID) (int, error)

	// первичное назначение при создании / одобрении карточки
	assignInTx(ctx context.Context, tx *gorm.DB, p *authz.Principal, asset *models.Asset, current *models.AssetRole, role models.RoleKind, personID uuid.UUID, reason, comment string) (*models.AssetRole, error)
	checkPerson(ctx context.Context, p *authz.Principal, personID uuid.UUID, needMol bool) error
}

type rolesService struct {
	db    *gorm.DB
	ts    tsclient.Client
	audit AuditService
}

func NewRolesService(db *gorm.DB, ts tsclient.Client, audit AuditService) RolesService {
	return &rolesService{db: db, ts: ts, audit: audit}
}

// checkPerson можно ли назначить человека (5.1): своя организация, не уволен, для МОЛ — признак МОЛ
func (s *rolesService) checkPerson(ctx context.Context, p *authz.Principal, personID uuid.UUID, needMol bool) error {
	person, err := s.ts.GetPerson(ctx, personID)
	if err != nil {
		if errors.Is(err, tsclient.ErrPersonNotFound) {
			return NotFound("Человек не найден")
		}
		return Unavailable("Не удалось проверить человека, повторите")
	}
	if person.OrganizationID != p.OrgID {
		return NotFound("Человек не найден")
	}
	if !person.IsActive {
		return BadRequest("Человек уволен")
	}
	if needMol && !person.IsMol {
		return BadRequest("Актив может числиться только за МОЛ — сначала админ отмечает человека как МОЛ")
	}
	return nil
}

// assignInTx закрыть прежнюю роль (если есть), записать передачу и открыть новую роль
func (s *rolesService) assignInTx(ctx context.Context, tx *gorm.DB, p *authz.Principal, asset *models.Asset, current *models.AssetRole, role models.RoleKind, personID uuid.UUID, reason, comment string) (*models.AssetRole, error) {
	now := time.Now()
	transfer := models.AssetTransfer{
		ID:                   newID(),
		OrganizationID:       asset.OrganizationID,
		AssetID:              asset.UUID,
		Role:                 role,
		ToPersonID:           &personID,
		InitiatedByAccountID: &p.AccountID,
		InitiatedByPersonID:  &p.PersonID,
		Reason:               strPtr(reason),
		Comment:              strPtr(comment),
		CreatedAt:            now,
	}
	if current != nil {
		transfer.FromPersonID = &current.PersonID
		endReason := "transfer"
		if err := tx.Model(&models.AssetRole{}).Where("id = ? AND ended_at IS NULL", current.ID).
			Updates(map[string]interface{}{"ended_at": now, "end_reason": endReason}).Error; err != nil {
			return nil, err
		}
	}
	if err := tx.Create(&transfer).Error; err != nil {
		return nil, err
	}
	r := models.AssetRole{
		ID:                  newID(),
		OrganizationID:      asset.OrganizationID,
		AssetID:             asset.UUID,
		PersonID:            personID,
		Role:                role,
		StartedAt:           now,
		AssignedByAccountID: &p.AccountID,
		AssignedByPersonID:  &p.PersonID,
		TransferID:          &transfer.ID,
		AckStatus:           models.AckPending,
	}
	// себе — подтверждать нечего (в журнале self_assigned)
	if personID == p.PersonID {
		r.AckStatus = models.AckAccepted
		r.AckAt = &now
	}
	if err := tx.Create(&r).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, Conflict("Роль уже назначена — обновите карточку")
		}
		return nil, err
	}
	details := map[string]interface{}{"role": role, "to_person_id": personID, "self_assigned": personID == p.PersonID}
	if current != nil {
		details["from_person_id"] = current.PersonID
	}
	if reason != "" {
		details["reason"] = reason
	}
	if err := s.audit.Log(ctx, tx, AuditEntry{Org: asset.OrganizationID, Actor: p, Action: "transfer.apply", ObjectType: "asset", ObjectID: &asset.UUID, Details: details}); err != nil {
		return nil, err
	}
	return &r, nil
}

// endRoleInTx снять роль без нового получателя
func (s *rolesService) endRoleInTx(ctx context.Context, tx *gorm.DB, p *authz.Principal, asset *models.Asset, r *models.AssetRole, reason, comment string) error {
	now := time.Now()
	if err := tx.Model(&models.AssetRole{}).Where("id = ? AND ended_at IS NULL", r.ID).
		Updates(map[string]interface{}{"ended_at": now, "end_reason": reason}).Error; err != nil {
		return err
	}
	t := models.AssetTransfer{
		ID:             newID(),
		OrganizationID: asset.OrganizationID,
		AssetID:        asset.UUID,
		Role:           r.Role,
		FromPersonID:   &r.PersonID,
		Reason:         strPtr(reason),
		Comment:        strPtr(comment),
		CreatedAt:      now,
	}
	if p != nil {
		t.InitiatedByAccountID = &p.AccountID
		t.InitiatedByPersonID = &p.PersonID
	}
	if err := tx.Create(&t).Error; err != nil {
		return err
	}
	return s.audit.Log(ctx, tx, AuditEntry{Org: asset.OrganizationID, Actor: p, Action: "asset_role.end", ObjectType: "asset", ObjectID: &asset.UUID,
		Details: map[string]interface{}{"role": r.Role, "person_id": r.PersonID, "reason": reason}})
}

// Списанный — передачи запрещены; утерянный — только админ (3.6)
func checkTransferable(p *authz.Principal, a *models.Asset) error {
	switch a.AccountingStatus {
	case models.AccountingDecommissioned:
		return BadRequest("Актив списан — передачи и назначения запрещены")
	case models.AccountingLost:
		if !p.Has(authz.PermAssetRolesAssign) {
			return BadRequest("Актив утерян — передать его может только администратор")
		}
	}
	return nil
}

func (s *rolesService) AssignMol(ctx context.Context, p *authz.Principal, assetID, personID uuid.UUID, comment string) (*models.AssetRole, error) {
	if err := s.checkPerson(ctx, p, personID, true); err != nil {
		return nil, err
	}
	var result *models.AssetRole
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, assetID, true)
		if err != nil {
			return err
		}
		// админ — кому угодно; МОЛ — свой актив другому МОЛ
		if !acc.canManageRoles(p) {
			return errNoRights
		}
		if err := checkTransferable(p, acc.Asset); err != nil {
			return err
		}
		if acc.Mol != nil && acc.Mol.PersonID == personID {
			return Conflict("Актив уже числится за этим человеком")
		}
		result, err = s.assignInTx(ctx, tx, p, acc.Asset, acc.Mol, models.RoleMol, personID, "", comment)
		return err
	})
	return result, err
}

func (s *rolesService) RemoveMol(ctx context.Context, p *authz.Principal, assetID uuid.UUID, comment string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, assetID, true)
		if err != nil {
			return err
		}
		// снять МОЛ, не передав никому, может только админ: актив попадёт в «Без МОЛ»
		if !p.Has(authz.PermAssetRolesAssign) {
			return errNoRights
		}
		if acc.Mol == nil {
			return Conflict("Актив ни за кем не числится")
		}
		return s.endRoleInTx(ctx, tx, p, acc.Asset, acc.Mol, "removed", comment)
	})
}

func (s *rolesService) AddUser(ctx context.Context, p *authz.Principal, assetID, personID uuid.UUID, comment string) (*models.AssetRole, error) {
	if err := s.checkPerson(ctx, p, personID, false); err != nil {
		return nil, err
	}
	var result *models.AssetRole
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, assetID, true)
		if err != nil {
			return err
		}
		if !acc.canManageRoles(p) {
			return errNoRights
		}
		if err := checkTransferable(p, acc.Asset); err != nil {
			return err
		}
		for _, u := range acc.Users {
			if u.PersonID == personID {
				return Conflict("Человек уже работает с этим активом")
			}
		}
		result, err = s.assignInTx(ctx, tx, p, acc.Asset, nil, models.RoleUser, personID, "", comment)
		return err
	})
	return result, err
}

func (s *rolesService) RemoveUser(ctx context.Context, p *authz.Principal, assetID, personID uuid.UUID, comment string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, assetID, true)
		if err != nil {
			return err
		}
		if !acc.canManageRoles(p) {
			return errNoRights
		}
		for i := range acc.Users {
			if acc.Users[i].PersonID == personID {
				return s.endRoleInTx(ctx, tx, p, acc.Asset, &acc.Users[i], "removed", comment)
			}
		}
		return NotFound("Человек не работает с этим активом")
	})
}

// MyAssignments «Мне назначено»: действующие роли вошедшего с карточками активов
func (s *rolesService) MyAssignments(ctx context.Context, p *authz.Principal) ([]models.AssetRole, error) {
	var roles []models.AssetRole
	err := s.db.WithContext(ctx).Preload("Asset").
		Where("organization_id = ? AND person_id = ? AND ended_at IS NULL", p.OrgID, p.PersonID).
		Order("started_at DESC").Find(&roles).Error
	return roles, err
}

// Ack получатель подтверждает или отказывается с причиной. Актив не двигается; отказ — админу «на разбор»
func (s *rolesService) Ack(ctx context.Context, p *authz.Principal, roleID uuid.UUID, accept bool, reason string) (*models.AssetRole, error) {
	var r models.AssetRole
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND organization_id = ? AND ended_at IS NULL", roleID, p.OrgID).First(&r).Error; err != nil {
			return NotFound("Назначение не найдено")
		}
		if r.PersonID != p.PersonID {
			return Forbidden("Подтвердить может только тот, кому назначено")
		}
		now := time.Now()
		upd := map[string]interface{}{"ack_at": now}
		action := "asset_role.ack"
		if accept {
			upd["ack_status"] = models.AckAccepted
			upd["ack_reason"] = nil
		} else {
			if reason == "" {
				return BadRequest("Укажите причину отказа")
			}
			upd["ack_status"] = models.AckDisputed
			upd["ack_reason"] = reason
			upd["dispute_resolved_at"] = nil
			action = "asset_role.dispute"
		}
		if err := tx.Model(&r).Updates(upd).Error; err != nil {
			return err
		}
		return s.audit.Log(ctx, tx, AuditEntry{Org: p.OrgID, Actor: p, Action: action, ObjectType: "asset", ObjectID: &r.AssetID,
			Details: map[string]interface{}{"role_id": r.ID, "role": r.Role, "reason": reason}})
	})
	if err != nil {
		return nil, err
	}
	s.db.WithContext(ctx).Preload("Asset").First(&r, "id = ?", roleID)
	return &r, nil
}

// Disputes «На разбор»: отказы, которые админ ещё не разобрал
func (s *rolesService) Disputes(ctx context.Context, p *authz.Principal) ([]models.AssetRole, error) {
	if !p.Has(authz.PermAssetRolesAssign) {
		return nil, errNoRights
	}
	var roles []models.AssetRole
	err := s.db.WithContext(ctx).Preload("Asset").
		Where("organization_id = ? AND ack_status = ? AND dispute_resolved_at IS NULL", p.OrgID, models.AckDisputed).
		Order("ack_at DESC").Find(&roles).Error
	return roles, err
}

// Resolve админ разобрал отказ: оставить как есть (переназначение — обычной передачей)
func (s *rolesService) Resolve(ctx context.Context, p *authz.Principal, roleID uuid.UUID, resolution string) (*models.AssetRole, error) {
	if !p.Has(authz.PermAssetRolesAssign) {
		return nil, errNoRights
	}
	var r models.AssetRole
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND organization_id = ?", roleID, p.OrgID).First(&r).Error; err != nil {
			return NotFound("Назначение не найдено")
		}
		if r.AckStatus != models.AckDisputed || r.DisputeResolvedAt != nil {
			return Conflict("Отказа на разбор нет")
		}
		now := time.Now()
		if err := tx.Model(&r).Updates(map[string]interface{}{
			"dispute_resolved_at":            now,
			"dispute_resolved_by_account_id": p.AccountID,
			"dispute_resolved_by_person_id":  p.PersonID,
			"dispute_resolution":             strPtr(resolution),
		}).Error; err != nil {
			return err
		}
		return s.audit.Log(ctx, tx, AuditEntry{Org: p.OrgID, Actor: p, Action: "asset_role.resolve", ObjectType: "asset", ObjectID: &r.AssetID,
			Details: map[string]interface{}{"role_id": r.ID, "resolution": resolution}})
	})
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Transfers история передач и всех ролей актива (включая закрытые)
func (s *rolesService) Transfers(ctx context.Context, p *authz.Principal, assetID uuid.UUID) ([]models.AssetTransfer, []models.AssetRole, error) {
	acc, err := loadAsset(ctx, s.db, p, assetID, false)
	if err != nil {
		return nil, nil, err
	}
	if !acc.canView(p) {
		return nil, nil, errNoRights
	}
	var transfers []models.AssetTransfer
	if err := s.db.WithContext(ctx).Where("asset_id = ?", assetID).Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, nil, err
	}
	var roles []models.AssetRole
	if err := s.db.WithContext(ctx).Where("asset_id = ?", assetID).Order("started_at DESC").Find(&roles).Error; err != nil {
		return nil, nil, err
	}
	return transfers, roles, nil
}

// MolCount действующие роли человека. Списанные не в счёт — при списании роли закрываются (3.6)
func (s *rolesService) MolCount(ctx context.Context, personID uuid.UUID) (int64, int64, error) {
	var mol, user int64
	if err := s.db.WithContext(ctx).Model(&models.AssetRole{}).
		Joins("JOIN assets a ON a.uuid = asset_roles.asset_id AND a.deleted_at IS NULL").
		Where("asset_roles.person_id = ? AND asset_roles.ended_at IS NULL AND asset_roles.role = ?", personID, models.RoleMol).
		Count(&mol).Error; err != nil {
		return 0, 0, err
	}
	if err := s.db.WithContext(ctx).Model(&models.AssetRole{}).
		Where("person_id = ? AND ended_at IS NULL AND role = ?", personID, models.RoleUser).
		Count(&user).Error; err != nil {
		return 0, 0, err
	}
	return mol, user, nil
}

// PersonDismissed увольнение: роли «работает с активом» закрываются автоматически (2.2)
func (s *rolesService) PersonDismissed(ctx context.Context, personID uuid.UUID) (int, error) {
	var roles []models.AssetRole
	if err := s.db.WithContext(ctx).Where("person_id = ? AND ended_at IS NULL AND role = ?", personID, models.RoleUser).Find(&roles).Error; err != nil {
		return 0, err
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range roles {
			asset := &models.Asset{UUID: roles[i].AssetID, OrganizationID: roles[i].OrganizationID}
			if err := s.endRoleInTx(ctx, tx, nil, asset, &roles[i], "dismissed", ""); err != nil {
				return err
			}
		}
		return nil
	})
	return len(roles), err
}
