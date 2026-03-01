package models

import (
	"time"

	"github.com/google/uuid"
)

// Request запрос на обслуживание устройства
// @Description Запрос создаётся пользователем для обслуживания или ремонта устройства
type Request struct {
	UUID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"uuid" swaggertype:"string" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"` // UUID запроса (v7)
	DeviceUUID uuid.UUID `gorm:"type:uuid;not null;index;column:device_id" json:"device_uuid" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"` // UUID устройства
	UserUUID   uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_uuid" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"`     // UUID владельца (из auth-сервиса)
	CreatedAt  time.Time `json:"created_at"`                                                                                                                       // Дата создания
	Message    string    `gorm:"type:text;not null" json:"message"`                                                                                                // Текст запроса
	PhotoURL   *string   `json:"photo_url,omitempty"`                                                                                                              // Ссылка на фото (опционально)
	Status     string    `gorm:"type:varchar(20);not null;default:'Оформлена'" json:"status"`                                                                      // Статус запроса
	UpdatedAt  time.Time `json:"updated_at"`                                                                                                                       // Дата обновления
	DeletedAt  time.Time `json:"deleted_at,omitempty"`                                                                                                             // Дата удаления (soft delete)

	Device Device `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"device,omitempty"` // Устройство (FK)
}
