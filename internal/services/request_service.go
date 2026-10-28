package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/models"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RequestFilter struct {
	Status  string
	AssetID *uuid.UUID
}

// RequestService заявки на ремонт (3.6): подаёт любой, кто работает с активом; видят МОЛ актива и админ
type RequestService interface {
	Create(ctx context.Context, p *authz.Principal, assetID uuid.UUID, message string, photoURL *string, markFaulty bool) (*models.Request, error)
	List(ctx context.Context, p *authz.Principal, f RequestFilter) ([]models.Request, error)
	Get(ctx context.Context, p *authz.Principal, id uuid.UUID) (*models.Request, error)
	SetStatus(ctx context.Context, p *authz.Principal, id uuid.UUID, status models.RequestStatus, comment string) (*models.Request, error)
	Delete(ctx context.Context, p *authz.Principal, id uuid.UUID) error
}

type requestService struct {
	db        *gorm.DB
	lifecycle LifecycleService
	audit     AuditService
}

func NewRequestService(db *gorm.DB, lifecycle LifecycleService, audit AuditService) RequestService {
	return &requestService{db: db, lifecycle: lifecycle, audit: audit}
}

func (s *requestService) Create(ctx context.Context, p *authz.Principal, assetID uuid.UUID, message string, photoURL *string, markFaulty bool) (*models.Request, error) {
	if !p.Has(authz.PermRequestsCreate) {
		return nil, errNoRights
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, BadRequest("Опишите неисправность")
	}
	var req models.Request
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		acc, err := loadAsset(ctx, tx, p, assetID, true)
		if err != nil {
			return err
		}
		if !acc.canView(p) {
			return errNoRights
		}
		if acc.Asset.AccountingStatus == models.AccountingDecommissioned {
			return BadRequest("Актив списан")
		}
		req = models.Request{
			UUID:           newID(),
			OrganizationID: p.OrgID,
			AssetUUID:      assetID,
			AccountUUID:    p.AccountID,
			PersonUUID:     p.PersonID,
			Message:        message,
			PhotoURL:       photoURL,
			Status:         models.RequestNew,
		}
		if err := tx.Create(&req).Error; err != nil {
			return err
		}
		// при подаче можно сразу отметить «неисправен»
		if markFaulty {
			if err := s.lifecycle.setCondition(ctx, tx, p, acc.Asset, models.ConditionFaulty, nil, message); err != nil {
				return err
			}
		}
		return s.audit.Log(ctx, tx, AuditEntry{Org: p.OrgID, Actor: p, Action: "request.create", ObjectType: "asset", ObjectID: &assetID,
			Details: map[string]interface{}{"request_id": req.UUID}})
	})
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// canSeeAll админ / менеджер видят все заявки организации; остальные — свои и по активам, где они МОЛ
func canSeeAllRequests(p *authz.Principal) bool {
	return p.Has(authz.PermAssetsManage) || p.Has(authz.PermAssetsStatus)
}

func (s *requestService) List(ctx context.Context, p *authz.Principal, f RequestFilter) ([]models.Request, error) {
	q := s.db.WithContext(ctx).Preload("Asset").Where("requests.organization_id = ?", p.OrgID)
	if !canSeeAllRequests(p) {
		q = q.Where("requests.person_id = ? OR EXISTS (SELECT 1 FROM asset_roles r WHERE r.asset_id = requests.asset_id AND r.ended_at IS NULL AND r.role = 'mol' AND r.person_id = ?)",
			p.PersonID, p.PersonID)
	}
	if f.Status != "" {
		q = q.Where("requests.status = ?", f.Status)
	}
	if f.AssetID != nil {
		q = q.Where("requests.asset_id = ?", *f.AssetID)
	}
	var list []models.Request
	err := q.Order("requests.created_at DESC").Limit(500).Find(&list).Error
	return list, err
}

func (s *requestService) load(ctx context.Context, tx *gorm.DB, p *authz.Principal, id uuid.UUID) (*models.Request, *assetAccess, error) {
	var r models.Request
	if err := tx.WithContext(ctx).Where("uuid = ? AND organization_id = ?", id, p.OrgID).First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, NotFound("Заявка не найдена")
		}
		return nil, nil, err
	}
	acc, err := loadAsset(ctx, tx, p, r.AssetUUID, false)
	if err != nil {
		return nil, nil, err
	}
	return &r, acc, nil
}

func (s *requestService) Get(ctx context.Context, p *authz.Principal, id uuid.UUID) (*models.Request, error) {
	r, acc, err := s.load(ctx, s.db, p, id)
	if err != nil {
		return nil, err
	}
	if !canSeeAllRequests(p) && !acc.IsMol && r.PersonUUID != p.PersonID {
		return nil, errNoRights
	}
	r.Asset = acc.Asset
	return r, nil
}

// SetStatus выполнена / отклонена — МОЛ актива или админ. «В ремонте» ставится через «В ремонт» с заявкой
func (s *requestService) SetStatus(ctx context.Context, p *authz.Principal, id uuid.UUID, status models.RequestStatus, comment string) (*models.Request, error) {
	if status != models.RequestDone && status != models.RequestRejected {
		return nil, BadRequest("Статус заявки: done или rejected")
	}
	var out *models.Request
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		r, acc, err := s.load(ctx, tx, p, id)
		if err != nil {
			return err
		}
		if !acc.canChangeStatus(p) {
			return errNoRights
		}
		if r.Status == models.RequestDone || r.Status == models.RequestRejected {
			return Conflict("Заявка уже закрыта")
		}
		now := time.Now()
		if err := tx.Model(r).Updates(map[string]interface{}{
			"status": status, "admin_comment": comment, "processed_at": now,
			"processed_by_account_id": p.AccountID, "processed_by_person_id": p.PersonID,
		}).Error; err != nil {
			return err
		}
		out = r
		return s.audit.Log(ctx, tx, AuditEntry{Org: p.OrgID, Actor: p, Action: "request." + string(status), ObjectType: "asset", ObjectID: &r.AssetUUID,
			Details: map[string]interface{}{"request_id": r.UUID, "comment": comment}})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, p, out.UUID)
}

// Delete свою новую заявку можно отозвать
func (s *requestService) Delete(ctx context.Context, p *authz.Principal, id uuid.UUID) error {
	r, _, err := s.load(ctx, s.db, p, id)
	if err != nil {
		return err
	}
	if r.PersonUUID != p.PersonID && !p.Has(authz.PermAssetsManage) {
		return errNoRights
	}
	if r.Status != models.RequestNew {
		return Conflict("Заявка уже в работе")
	}
	return s.db.WithContext(ctx).Delete(r).Error
}
