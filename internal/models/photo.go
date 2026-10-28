package models

import (
	"time"

	"github.com/google/uuid"
)

// PhotoStatus статус загрузки фотографии
type PhotoStatus string

const (
	PhotoStatusPending   PhotoStatus = "pending"   // Ожидает загрузки (выдан presigned URL)
	PhotoStatusCompleted PhotoStatus = "completed" // Загрузка завершена
)

// Photo фотография устройства
// @Description Фотография загруженная пользователем и привязанная к assets
type Photo struct {
	UUID           uuid.UUID   `gorm:"type:uuid;primaryKey" json:"uuid" swaggertype:"string"`
	OrganizationID uuid.UUID   `gorm:"type:uuid;not null;index" json:"organization_id" swaggertype:"string"`
	ClientID       *uuid.UUID  `gorm:"type:uuid;uniqueIndex" json:"client_id,omitempty" swaggertype:"string"` // UUID сгенерированный клиентом (для оптимистичного UI)
	AccountUUID    uuid.UUID   `gorm:"type:uuid;not null;index;column:account_id" json:"account_id" swaggertype:"string"`
	PersonUUID     uuid.UUID   `gorm:"type:uuid;not null;column:person_id" json:"person_id" swaggertype:"string"` // человек в момент съёмки
	URL            string      `gorm:"not null" json:"url"`                                                       // Путь к фото в хранилище
	Status         PhotoStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	FileSize       int64       `gorm:"not null" json:"file_size"`
	FileName       string      `gorm:"not null" json:"file_name"`
	ContentType    string      `gorm:"not null" json:"content_type"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	DeletedAt      *time.Time  `json:"deleted_at,omitempty"`

	Assets []Asset `gorm:"many2many:asset_photos;joinForeignKey:PhotoUUID;joinReferences:AssetUUID" json:"assets,omitempty"` // Активы на фото (many-to-many)
}
