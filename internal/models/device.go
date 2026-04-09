package models

import (
	"time"

	"github.com/google/uuid"
)

// Device устройство в системе
// @Description Устройство представляет собой физический объект с инвентарным номером
type Device struct {
	UUID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"uuid" swaggertype:"string" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"` // UUID устройства (v7)
	InventoryNum  string    `gorm:"unique;not null" json:"inventory_num"`                                                                                           // Инвентарный номер
	Type          string    `gorm:"not null" json:"type"`                                                                                                           // Тип устройства
	Specification string    `gorm:"type:text" json:"specification"`                                                                                                 // Спецификация
	//Нужен ли нам статус для девайсов, устройств на сервере. Что он нам дает? Какие статусы мы здесь отображаем?
	//Сервер один, клиентов много. В каком случае нам надо менять статусы.
	//[red flag] - рассмотреть возможность для удаления поля
	Status        string    `gorm:"not null" json:"status"`                                                                                                         // Статус
	UserUUID      uuid.UUID `gorm:"type:uuid;not null;index;column:user_id" json:"user_uuid" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"`   // UUID владельца (из auth-сервиса)
	CreatedAt     time.Time `json:"created_at"`                                                                                                                     // Дата создания
	UpdatedAt     time.Time `json:"updated_at"`                                                                                                                     // Дата обновления
	DeletedAt     time.Time `json:"deleted_at,omitempty"`                                                                                                           // Дата удаления (soft delete)

	Photos   []Photo   `gorm:"many2many:device_photos;joinForeignKey:DeviceUUID;joinReferences:PhotoUUID" json:"photos,omitempty"`
	Requests []Request `gorm:"foreignKey:DeviceUUID" json:"requests,omitempty"`
}
