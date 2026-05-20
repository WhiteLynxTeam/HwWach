package models

import (
	"gorm.io/gorm"
	"time"

	"github.com/google/uuid"
)

// AssetStatus статус актива
type AssetStatus string
type ModerationStatus string

const (
	AssetStatusActive         AssetStatus = "active"         // В эксплуатации
	AssetStatusInactive       AssetStatus = "inactive"       // Не используется
	AssetStatusMaintenance    AssetStatus = "maintenance"    // На обслуживании
	AssetStatusRepair         AssetStatus = "repair"         // В ремонте
	AssetStatusDecommissioned AssetStatus = "decommissioned" // Списан
	AssetStatusLost           AssetStatus = "lost"           // Утерян
)

const (
	ModerationPending  ModerationStatus = "pending"
	ModerationApproved ModerationStatus = "approved"
	ModerationRejected ModerationStatus = "rejected"
)

// Asset актив в системе
// @Description Актив представляет собой физический объект с инвентарным номером
type Asset struct {
	UUID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"uuid" swaggertype:"string" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`                         // UUID актива (v7)
	ClientID     *uuid.UUID `gorm:"type:uuid;uniqueIndex" json:"client_id,omitempty" swaggertype:"string" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`         // UUID сгенерированный клиентом (для оптимистичного UI)
	InventoryNum string     `gorm:"unique;not null" json:"inventory_num"`                                                                                         // Инвентарный номер
	Name         string     `gorm:"not null" json:"name"`                                                                                                         // Название актива
	Category     string     `gorm:"not null" json:"category"`                                                                                                     // Категория актива
	Description  string     `gorm:"type:text" json:"description"`                                                                                                 // Описание актива
	UserUUID     uuid.UUID  `gorm:"type:uuid;not null;index;column:user_id" json:"user_uuid" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"` // UUID владельца (из auth-сервиса)

	AssetStatus      AssetStatus      `gorm:"type:varchar(20);not null;default:'active'" json:"asset_status" example:"active"` // Статус актива
	ModerationStatus ModerationStatus `gorm:"type:varchar(20);default:'pending'" json:"moderation_status"`
	VerifiedAt       *time.Time       `json:"verified_at,omitempty"`                                  // Дата верификации админом
	AdminComment     string           `gorm:"type:text" json:"admin_comment,omitempty"`               // Комментарий администратора
	CreatedAt        time.Time        `json:"created_at"`                                             // Дата создания
	UpdatedAt        time.Time        `json:"updated_at"`                                             // Дата обновления
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"deleted_at,omitempty" swaggertype:"string"` // Дата удаления (soft delete)

	Photos   []Photo   `gorm:"many2many:asset_photos;joinForeignKey:AssetUUID;joinReferences:PhotoUUID" json:"photos,omitempty"`
	Requests []Request `gorm:"foreignKey:AssetUUID" json:"requests,omitempty"`
}
