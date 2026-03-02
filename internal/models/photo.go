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
// @Description Фотография загруженная пользователем и привязанная к устройству
type Photo struct {
	UUID      uuid.UUID   `gorm:"type:uuid;primaryKey" json:"uuid" swaggertype:"string" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"` // UUID фотографии (v7, генерируется сервером)
	ClientID  *uuid.UUID  `gorm:"type:uuid;uniqueIndex" json:"client_id,omitempty" swaggertype:"string" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`                          // UUID сгенерированный клиентом (для оптимистичного UI)
	UserUUID  uuid.UUID   `gorm:"type:uuid;not null;index;column:user_id" json:"user_uuid" swaggertype:"string" example:"550e8400-e29b-41d4-a716-446655440000"`                                           // UUID владельца (из auth-сервиса)
	URL       string      `gorm:"not null" json:"url"`                                                                                                                                                      // Ссылка на фото в MinIO
	Status    PhotoStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status" example:"pending"`                                                                                              // Статус загрузки
	FileSize  int64       `gorm:"not null" json:"file_size" example:"1048576"`                                                                                                                              // Размер файла в байтах
	FileName  string      `gorm:"not null" json:"file_name" example:"photo.jpg"`                                                                                                                            // Имя файла
	ContentType string    `gorm:"not null" json:"content_type" example:"image/jpeg"`                                                                                                                        // MIME тип файла
	CreatedAt time.Time   `json:"created_at"`                                                                                                                                                               // Дата создания
	UpdatedAt time.Time   `json:"updated_at"`                                                                                                                                                               // Дата обновления
	DeletedAt time.Time   `json:"deleted_at,omitempty"`                                                                                                                                                     // Дата удаления (soft delete)

	Devices []Device `gorm:"many2many:device_photos;" json:"devices,omitempty"` // Устройства на фото (many-to-many)
}
