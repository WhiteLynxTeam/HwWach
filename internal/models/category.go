package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryStatus string

const (
	CategoryStatusApproved CategoryStatus = "approved"
	CategoryStatusPending  CategoryStatus = "pending"
	CategoryStatusRejected CategoryStatus = "rejected"
	CategoryStatusMerged   CategoryStatus = "merged"
)

// Category представляет категорию оборудования (3 уровня: L1 общие для всех организаций,
// L2 своей организации, L3 пользовательские — на модерацию админу организации)
type Category struct {
	UUID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"uuid" swaggertype:"string"`
	OrganizationID    *uuid.UUID     `gorm:"type:uuid" json:"organization_id,omitempty" swaggertype:"string"` // NULL — общая L1
	Name              string         `gorm:"type:varchar(255);not null" json:"name" example:"Монитор"`
	NormalizedName    string         `gorm:"type:varchar(255);not null" json:"normalized_name"`
	Level             int16          `gorm:"type:smallint;not null" json:"level" example:"1"` // 1: Global, 2: Local, 3: User
	Status            CategoryStatus `gorm:"type:varchar(20);not null;default:'approved'" json:"status" example:"approved"`
	CreatedBy         *uuid.UUID     `gorm:"type:uuid;column:created_by" json:"created_by,omitempty" swaggertype:"string"` // учётная запись
	CreatedByPersonID *uuid.UUID     `gorm:"type:uuid;column:created_by_person_id" json:"created_by_person_id,omitempty" swaggertype:"string"`
	MergedIntoUUID    *uuid.UUID     `gorm:"type:uuid;column:merged_into_uuid" json:"merged_into_uuid,omitempty" swaggertype:"string"`
	UsageCount        int            `gorm:"column:usage_count;default:0" json:"usage_count"`
	AdminComment      string         `gorm:"type:text;column:admin_comment" json:"admin_comment,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty" swaggertype:"string"`
}
