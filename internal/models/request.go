package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RequestStatus статус заявки на ремонт (3.6)
type RequestStatus string

const (
	RequestNew      RequestStatus = "new"       // оформлена
	RequestInRepair RequestStatus = "in_repair" // по ней актив в ремонте
	RequestDone     RequestStatus = "done"
	RequestRejected RequestStatus = "rejected"
)

// Request заявка на ремонт
// @Description Заявку подаёт любой, кто работает с активом; видят МОЛ актива и админ
type Request struct {
	UUID                 uuid.UUID      `gorm:"type:uuid;primaryKey" json:"uuid" swaggertype:"string"`
	OrganizationID       uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id" swaggertype:"string"`
	AssetUUID            uuid.UUID      `gorm:"type:uuid;not null;index;column:asset_id" json:"asset_uuid" swaggertype:"string"`
	AccountUUID          uuid.UUID      `gorm:"type:uuid;not null;column:account_id" json:"account_id" swaggertype:"string"`
	PersonUUID           uuid.UUID      `gorm:"type:uuid;not null;column:person_id" json:"person_id" swaggertype:"string"`
	Message              string         `gorm:"type:text;not null" json:"message"`
	PhotoURL             *string        `json:"photo_url,omitempty"`
	Status               RequestStatus  `gorm:"type:varchar(20);not null;default:'new'" json:"status"`
	AdminComment         string         `gorm:"type:text" json:"admin_comment,omitempty"`
	ProcessedByAccountID *uuid.UUID     `gorm:"type:uuid" json:"processed_by_account_id,omitempty" swaggertype:"string"`
	ProcessedByPersonID  *uuid.UUID     `gorm:"type:uuid" json:"processed_by_person_id,omitempty" swaggertype:"string"`
	ProcessedAt          *time.Time     `json:"processed_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-" swaggertype:"string"`

	Asset *Asset `gorm:"foreignKey:AssetUUID" json:"asset,omitempty"`
}
