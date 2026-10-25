package models

import (
	"time"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

// AccountingStatus учётный статус — где актив по учёту (3.6)
type AccountingStatus string

// Condition техническое состояние — что с активом физически (3.6)
type Condition string

type ModerationStatus string

const (
	AccountingInUse          AccountingStatus = "in_use"         // В эксплуатации
	AccountingInStock        AccountingStatus = "in_stock"       // На складе / не используется
	AccountingRepair         AccountingStatus = "repair"         // В ремонте
	AccountingLost           AccountingStatus = "lost"           // Утерян
	AccountingDecommissioned AccountingStatus = "decommissioned" // Списан
)

const (
	ConditionOK             Condition = "ok"              // Исправен
	ConditionNeedsAttention Condition = "needs_attention" // Требует внимания
	ConditionFaulty         Condition = "faulty"          // Неисправен
	ConditionUnusable       Condition = "unusable"        // Не подлежит ремонту
)

const (
	ModerationPending  ModerationStatus = "pending"
	ModerationApproved ModerationStatus = "approved"
	ModerationRejected ModerationStatus = "rejected"
)

func (s AccountingStatus) Valid() bool {
	switch s {
	case AccountingInUse, AccountingInStock, AccountingRepair, AccountingLost, AccountingDecommissioned:
		return true
	}
	return false
}

func (c Condition) Valid() bool {
	switch c {
	case ConditionOK, ConditionNeedsAttention, ConditionFaulty, ConditionUnusable:
		return true
	}
	return false
}

// Asset актив в системе
// @Description Актив — физический объект организации. Ответственность за него — в asset_roles
type Asset struct {
	UUID               uuid.UUID  `gorm:"type:uuid;primaryKey" json:"uuid" swaggertype:"string"`
	OrganizationID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id" swaggertype:"string"`
	ClientID           *uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"client_id,omitempty" swaggertype:"string"` // UUID сгенерированный клиентом (для оптимистичного UI)
	InventoryNum       *string    `json:"inventory_num,omitempty"`                                              // Инвентарный номер, уникален в организации
	Name               string     `gorm:"not null" json:"name"`
	Category           string     `gorm:"not null" json:"category"`
	CategoryUUID       *uuid.UUID `gorm:"type:uuid;index;column:category_uuid" json:"category_uuid,omitempty" swaggertype:"string"`
	CategoryObj        *Category  `gorm:"foreignKey:CategoryUUID" json:"category_obj,omitempty"`
	Description        string     `gorm:"type:text" json:"description"`
	CreatedByAccountID uuid.UUID  `gorm:"type:uuid;not null" json:"created_by_account_id" swaggertype:"string"` // кто завёл карточку
	CreatedByPersonID  uuid.UUID  `gorm:"type:uuid;not null" json:"created_by_person_id" swaggertype:"string"`

	AccountingStatus AccountingStatus `gorm:"type:varchar(20);not null;default:'in_use'" json:"accounting_status"`
	Condition        Condition        `gorm:"type:varchar(20);not null;default:'ok'" json:"condition"`
	ModerationStatus ModerationStatus `gorm:"type:varchar(20);default:'pending'" json:"moderation_status"`
	VerifiedAt       *time.Time       `json:"verified_at,omitempty"`
	AdminComment     string           `gorm:"type:text" json:"admin_comment,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty" swaggertype:"string"`

	Photos   []Photo   `gorm:"many2many:asset_photos;joinForeignKey:AssetUUID;joinReferences:PhotoUUID" json:"photos,omitempty"`
	Requests []Request `gorm:"foreignKey:AssetUUID" json:"requests,omitempty"`
}
