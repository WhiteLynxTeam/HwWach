package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/models"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StatusInput смена учётного статуса и/или технического состояния (3.6)
type StatusInput struct {
	AccountingStatus *models.AccountingStatus
	Condition        *models.Condition
	Comment          string
}

type RepairStartInput struct {
	Kind             models.RepairKind
	ExpectedReturnAt *time.Time
	Comment          string
	RequestID        *uuid.UUID
}

type RepairFinishInput struct {
	ResultCondition models.Condition
	WorkDone        string
	StatusAfter     *models.AccountingStatus
	Comment         string
}

type AssetHistory struct {
	Status  []models.AssetStatusHistory `json:"status"`
	Repairs []models.AssetRepair        `json:"repairs"`
}

// LifecycleService состояние актива: учётный статус, техническое состояние, ремонт (3.6).
// Без подтверждений: всё сразу, с записью в историю
type LifecycleService interface {
	ChangeStatus(ctx context.Context, p *authz.Principal, assetID uuid.UUID, in StatusInput) error
	StartRepair(ctx context.Context, p *authz.Principal, assetID uuid.UUID, in RepairStartInput) (*models.AssetRepair, error)
	FinishRepair(ctx context.Context, p *authz.Principal, assetID uuid.UUID, in RepairFinishInput) (*models.AssetRepair, error)
	History(ctx context.Context, p *authz.Principal, assetID uuid.UUID) (*AssetHistory, error)
	setCondition(ctx context.Context, tx *gorm.DB, p *authz.Principal, a *models.Asset, to models.Condition, repairID *uuid.UUID, comment string) error
}

type lifecycleService struct {
	db    *gorm.DB
	roles *rolesService
	audit AuditService
}

func NewLifecycleService(db *gorm.DB, roles RolesService, audit AuditService) LifecycleService {
	return &lifecycleService{db: db, roles: roles.(*rolesService), audit: audit}
}

func (s *lifecycleService) history(tx *gorm.DB, p *authz.Principal, a *models.Asset, field string, from, to string, repairID *uuid.UUID, comment string) error {
	h := models.AssetStatusHistory{
		ID:             newID(),
		OrganizationID: a.OrganizationID,
		AssetID:        a.UUID,
		Field:          field,
		FromValue:      strPtr(from),
		ToValue:        to,
		RepairID:       repairID,
		Comment:        strPtr(comment),
		CreatedAt:      time.Now(),
	}
	if p != nil {
		h.AccountID = &p.AccountID
		h.PersonID = &p.PersonID
	}
	return tx.Create(&h).Error
}

func (s *lifecycleService) setAccounting(ctx context.Context, tx *gorm.DB, p *authz.Principal, a *models.Asset, to models.AccountingStatus, repairID *uuid.UUID, comment string) error {
	if a.AccountingStatus == to {
		return nil
	}
	from := a.AccountingStatus
	if err := tx.Model(&models.Asset{}).Where("uuid = ?", a.UUID).Update("accounting_status", to).Error; err != nil {
		return err
	}
	if err := s.history(tx, p, a, "accounting_status", string(from), string(to), repairID, comment); err != nil {
		return err
	}
	a.AccountingStatus = to
	return s.audit.Log(ctx, tx, AuditEntry{Org: a.OrganizationID, Actor: p, Action: "asset.status_change", ObjectType: "asset", ObjectID: &a.UUID,
		Details: map[string]interface{}{"field": "accounting_status", "from": from, "to": to, "comment": comment}})
}

func (s *lifecycleService) setCondition(ctx context.Context, tx *gorm.DB, p *authz.Principal, a *models.Asset, to models.Condition, repairID *uuid.UUID, comment string) error {
	if a.Condition == to {
		return nil
	}
	from := a.Condition
	if err := tx.Model(&models.Asset{}).Where("uuid = ?", a.UUID).Update("condition", to).Error; err != nil {
		return err
	}
	if err := s.history(tx, p, a, "condition", string(from), string(to), repairID, comment); err != nil {
		return err
	}
	a.Condition = to
	return s.audit.Log(ctx, tx, AuditEntry{Org: a.OrganizationID, Actor: p, Action: "asset.status_change", ObjectType: "asset", ObjectID: &a.UUID,
		Details: map[string]interface{}{"field": "condition", "from": from, "to": to, "comment": comment}})
}

func (s *lifecycleService) ChangeStatus(ctx context.Context, p *authz.Principal, assetID uuid.UUID, in StatusInput) error {
	if in.AccountingStatus == nil && in.Condition == nil {
		return BadRequest("Укажите статус или состояние")
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, assetID, true)
		if err != nil {
			return err
		}
		a := acc.Asset

		if in.AccountingStatus != nil {
			to := *in.AccountingStatus
			if !to.Valid() {
				return BadRequest("Неизвестный учётный статус")
			}
			if to == models.AccountingRepair || a.AccountingStatus == models.AccountingRepair {
				return BadRequest("Ремонт — действиями «В ремонт» / «Из ремонта»")
			}
			// списать и вернуть из списанных — только админ
			if to == models.AccountingDecommissioned || a.AccountingStatus == models.AccountingDecommissioned {
				if !p.Has(authz.PermAssetsDecommission) {
					return Forbidden("Списать или вернуть из списанных может только администратор")
				}
			} else if !acc.canChangeStatus(p) {
				return errNoRights
			}
			if err := s.setAccounting(ctx, tx, p, a, to, nil, in.Comment); err != nil {
				return err
			}
			// Списание закрывает роли: история остаётся, увольнение актив больше не держит
			if to == models.AccountingDecommissioned {
				if acc.Mol != nil {
					if err := s.roles.endRoleInTx(ctx, tx, p, a, acc.Mol, "decommissioned", in.Comment); err != nil {
						return err
					}
				}
				for i := range acc.Users {
					if err := s.roles.endRoleInTx(ctx, tx, p, a, &acc.Users[i], "decommissioned", in.Comment); err != nil {
						return err
					}
				}
			}
		}

		if in.Condition != nil {
			to := *in.Condition
			if !to.Valid() {
				return BadRequest("Неизвестное техническое состояние")
			}
			// «требует внимания» и «неисправен» может отметить любой, кто работает с активом
			minor := to == models.ConditionNeedsAttention || to == models.ConditionFaulty
			if !acc.canChangeStatus(p) && !(minor && (acc.IsUser || a.CreatedByPersonID == p.PersonID)) {
				return errNoRights
			}
			if err := s.setCondition(ctx, tx, p, a, to, nil, in.Comment); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *lifecycleService) openRepair(tx *gorm.DB, assetID uuid.UUID) (*models.AssetRepair, error) {
	var r models.AssetRepair
	err := tx.Where("asset_id = ? AND returned_at IS NULL", assetID).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &r, err
}

// StartRepair «В ремонт»: админ или МОЛ актива. Отдали мастеру — это обычная передача МОЛ → МОЛ,
// мастер сам ставит «В ремонт»
func (s *lifecycleService) StartRepair(ctx context.Context, p *authz.Principal, assetID uuid.UUID, in RepairStartInput) (*models.AssetRepair, error) {
	if in.Kind == "" {
		in.Kind = models.RepairKindRepair
	}
	if in.Kind != models.RepairKindRepair && in.Kind != models.RepairKindMaintenance {
		return nil, BadRequest("Вид работ: repair или maintenance")
	}
	var repair models.AssetRepair
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, assetID, true)
		if err != nil {
			return err
		}
		if !acc.canChangeStatus(p) {
			return errNoRights
		}
		a := acc.Asset
		switch a.AccountingStatus {
		case models.AccountingRepair:
			return Conflict("Актив уже в ремонте")
		case models.AccountingDecommissioned, models.AccountingLost:
			return BadRequest("Списанный или утерянный актив в ремонт не отправить")
		}
		repair = models.AssetRepair{
			ID:                 newID(),
			OrganizationID:     a.OrganizationID,
			AssetID:            a.UUID,
			RequestID:          in.RequestID,
			Kind:               in.Kind,
			StartedAt:          time.Now(),
			ExpectedReturnAt:   in.ExpectedReturnAt,
			StatusBefore:       a.AccountingStatus,
			Comment:            strPtr(in.Comment),
			CreatedByAccountID: p.AccountID,
			CreatedByPersonID:  p.PersonID,
		}
		if in.RequestID != nil {
			res := tx.Model(&models.Request{}).Where("uuid = ? AND asset_id = ? AND status = ?", *in.RequestID, a.UUID, models.RequestNew).
				Update("status", models.RequestInRepair)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return BadRequest("Заявка не найдена или уже обработана")
			}
		}
		if err := tx.Create(&repair).Error; err != nil {
			if isUniqueViolation(err) {
				return Conflict("Актив уже в ремонте")
			}
			return err
		}
		if err := s.setAccounting(ctx, tx, p, a, models.AccountingRepair, &repair.ID, in.Comment); err != nil {
			return err
		}
		return s.audit.Log(ctx, tx, AuditEntry{Org: a.OrganizationID, Actor: p, Action: "asset.repair_start", ObjectType: "asset", ObjectID: &a.UUID,
			Details: map[string]interface{}{"repair_id": repair.ID, "kind": in.Kind, "expected_return_at": in.ExpectedReturnAt}})
	})
	if err != nil {
		return nil, err
	}
	return &repair, nil
}

// FinishRepair «Из ремонта»: итог и статус после (по умолчанию прежний). Ставит текущий МОЛ или админ,
// кто бы ни ставил «В ремонт»
func (s *lifecycleService) FinishRepair(ctx context.Context, p *authz.Principal, assetID uuid.UUID, in RepairFinishInput) (*models.AssetRepair, error) {
	switch in.ResultCondition {
	case models.ConditionOK, models.ConditionFaulty, models.ConditionUnusable:
	default:
		return nil, BadRequest("Итог ремонта: ok, faulty или unusable")
	}
	var repair *models.AssetRepair
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, assetID, true)
		if err != nil {
			return err
		}
		if !acc.canChangeStatus(p) {
			return errNoRights
		}
		repair, err = s.openRepair(tx, assetID)
		if err != nil {
			return err
		}
		if repair == nil {
			return Conflict("Актив не в ремонте")
		}
		statusAfter := repair.StatusBefore
		if in.StatusAfter != nil {
			statusAfter = *in.StatusAfter
		}
		if !statusAfter.Valid() || statusAfter == models.AccountingRepair || statusAfter == models.AccountingDecommissioned {
			return BadRequest("Статус после ремонта: в эксплуатации, на складе или утерян; списать — отдельным действием")
		}
		now := time.Now()
		cond := in.ResultCondition
		if err := tx.Model(&models.AssetRepair{}).Where("id = ?", repair.ID).Updates(map[string]interface{}{
			"returned_at":            now,
			"result_condition":       cond,
			"work_done":              strPtr(in.WorkDone),
			"returned_by_account_id": p.AccountID,
			"returned_by_person_id":  p.PersonID,
		}).Error; err != nil {
			return err
		}
		repair.ReturnedAt = &now
		repair.ResultCondition = &cond
		repair.WorkDone = strPtr(in.WorkDone)
		if err := s.setCondition(ctx, tx, p, acc.Asset, cond, &repair.ID, in.Comment); err != nil {
			return err
		}
		if err := s.setAccounting(ctx, tx, p, acc.Asset, statusAfter, &repair.ID, in.Comment); err != nil {
			return err
		}
		if repair.RequestID != nil {
			if err := tx.Model(&models.Request{}).Where("uuid = ? AND status = ?", *repair.RequestID, models.RequestInRepair).
				Updates(map[string]interface{}{"status": models.RequestDone, "processed_at": now, "processed_by_account_id": p.AccountID, "processed_by_person_id": p.PersonID}).Error; err != nil {
				return err
			}
		}
		return s.audit.Log(ctx, tx, AuditEntry{Org: acc.Asset.OrganizationID, Actor: p, Action: "asset.repair_finish", ObjectType: "asset", ObjectID: &acc.Asset.UUID,
			Details: map[string]interface{}{"repair_id": repair.ID, "result": cond, "status_after": statusAfter}})
	})
	if err != nil {
		return nil, err
	}
	return repair, nil
}

func (s *lifecycleService) History(ctx context.Context, p *authz.Principal, assetID uuid.UUID) (*AssetHistory, error) {
	acc, err := loadAsset(ctx, s.db, p, assetID, false)
	if err != nil {
		return nil, err
	}
	if !acc.canView(p) {
		return nil, errNoRights
	}
	h := &AssetHistory{}
	if err := s.db.WithContext(ctx).Where("asset_id = ?", assetID).Order("created_at DESC").Find(&h.Status).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Where("asset_id = ?", assetID).Order("started_at DESC").Find(&h.Repairs).Error; err != nil {
		return nil, err
	}
	return h, nil
}
