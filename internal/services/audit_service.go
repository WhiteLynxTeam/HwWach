package services

import (
	"HwWach/internal/authz"
	"HwWach/internal/models"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditEntry запись журнала go (3.4). Actor nil — действие системы
type AuditEntry struct {
	Org        uuid.UUID
	Actor      *authz.Principal
	Action     string
	ObjectType string
	ObjectID   *uuid.UUID
	Details    map[string]interface{}
}

type AuditService interface {
	Log(ctx context.Context, tx *gorm.DB, e AuditEntry) error
	List(ctx context.Context, org uuid.UUID, limit int, before *time.Time, objectID *uuid.UUID) ([]models.AuditLog, error)
}

type auditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) AuditService {
	return &auditService{db: db}
}

func (s *auditService) Log(ctx context.Context, tx *gorm.DB, e AuditEntry) error {
	if tx == nil {
		tx = s.db
	}
	id, err := uuid.NewV7()
	if err != nil {
		return err
	}
	row := models.AuditLog{
		ID:             id,
		OrganizationID: e.Org,
		ActorType:      "system",
		Action:         e.Action,
		ObjectID:       e.ObjectID,
		CreatedAt:      time.Now(),
	}
	if e.ObjectType != "" {
		t := e.ObjectType
		row.ObjectType = &t
	}
	if e.Actor != nil {
		row.ActorType = "user"
		row.AccountID = &e.Actor.AccountID
		row.ActorPersonID = &e.Actor.PersonID
		if e.Actor.IP != "" {
			ip := e.Actor.IP
			row.IP = &ip
		}
	}
	if len(e.Details) > 0 {
		b, err := json.Marshal(e.Details)
		if err != nil {
			return err
		}
		row.Details = models.JSONB(b)
	}
	return tx.WithContext(ctx).Create(&row).Error
}

func (s *auditService) List(ctx context.Context, org uuid.UUID, limit int, before *time.Time, objectID *uuid.UUID) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := s.db.WithContext(ctx).Where("organization_id = ?", org)
	if before != nil {
		q = q.Where("created_at < ?", *before)
	}
	if objectID != nil {
		q = q.Where("object_id = ?", *objectID)
	}
	var rows []models.AuditLog
	err := q.Order("created_at DESC").Limit(limit).Find(&rows).Error
	return rows, err
}
