package services

import (
	"HwWach/internal/models"
	"HwWach/internal/storage"
	"HwWach/internal/tsclient"
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrgService связь с ts по организации (5.1): «наполнена», активность, окончательное удаление
type OrgService interface {
	// ReportActivity первый актив или фото — «наполнена», остальное — активность не чаще раза в сутки.
	// Работает в фоне: пользовательский запрос не ждёт ts
	ReportActivity(org uuid.UUID, filled bool)
	// DeleteOrganizationData удаляет данные организации и фото в хранилище. Повторный вызов безопасен
	DeleteOrganizationData(ctx context.Context, org uuid.UUID) error
}

type orgService struct {
	db      *gorm.DB
	ts      tsclient.Client
	storage storage.Storage
	audit   AuditService
}

func NewOrgService(db *gorm.DB, ts tsclient.Client, st storage.Storage, audit AuditService) OrgService {
	return &orgService{db: db, ts: ts, storage: st, audit: audit}
}

func (s *orgService) ReportActivity(org uuid.UUID, filled bool) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		var st models.OrganizationState
		if err := s.db.WithContext(ctx).Where("organization_id = ?", org).First(&st).Error; err != nil {
			st = models.OrganizationState{OrganizationID: org}
		}
		now := time.Now()
		if filled && st.FilledReportedAt == nil {
			if err := s.ts.ReportFilled(ctx, org); err != nil {
				log.Printf("report filled %s: %v", org, err)
				return
			}
			st.FilledReportedAt = &now
			st.TouchReportedAt = &now
		} else if st.TouchReportedAt == nil || now.Sub(*st.TouchReportedAt) > 24*time.Hour {
			if err := s.ts.Touch(ctx, org); err != nil {
				log.Printf("report touch %s: %v", org, err)
				return
			}
			st.TouchReportedAt = &now
		} else {
			return
		}
		s.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&st)
	}()
}

func (s *orgService) DeleteOrganizationData(ctx context.Context, org uuid.UUID) error {
	var urls []string
	if err := s.db.WithContext(ctx).Model(&models.Photo{}).Where("organization_id = ?", org).Pluck("url", &urls).Error; err != nil {
		return err
	}
	for _, u := range urls {
		if err := s.storage.Delete(ctx, u); err != nil {
			log.Printf("delete photo %s: %v", u, err)
		}
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, q := range []string{
			"DELETE FROM asset_status_history WHERE organization_id = ?",
			"DELETE FROM asset_repairs WHERE organization_id = ?",
			"DELETE FROM asset_roles WHERE organization_id = ?",
			"DELETE FROM asset_transfers WHERE organization_id = ?",
			"DELETE FROM asset_change_requests WHERE organization_id = ?",
			"DELETE FROM requests WHERE organization_id = ?",
			"DELETE FROM asset_photos WHERE asset_uuid IN (SELECT uuid FROM assets WHERE organization_id = ?)",
			"DELETE FROM photos WHERE organization_id = ?",
			"DELETE FROM assets WHERE organization_id = ?",
			"DELETE FROM categories WHERE organization_id = ?",
			"DELETE FROM audit_log WHERE organization_id = ?",
			"DELETE FROM organization_state WHERE organization_id = ?",
		} {
			if err := tx.Exec(q, org).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
