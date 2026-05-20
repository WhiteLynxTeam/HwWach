package models

import (
	"github.com/google/uuid"
	"time"
)

type RequestType string

const (
	RequestTypeUpdate RequestType = "update"
	RequestTypeDelete RequestType = "delete"
)

// AssetChangeRequest заявка на изменение актива
type AssetChangeRequest struct {
	UUID        uuid.UUID   `gorm:"type:uuid;primaryKey"`
	AssetUUID   uuid.UUID   `gorm:"type:uuid;index"`
	UserUUID    uuid.UUID   `gorm:"type:uuid"`
	RequestType RequestType `gorm:"column:request_type;type:varchar(20)"` // update или delete

	// Новые данные, которые пользователь хочет внести (в формате JSON)
	// Если Type == delete, это поле может быть пустым
	ProposedData []byte `gorm:"type:jsonb"`

	Reason       string           `gorm:"type:text"` // Комментарий пользователя "Зачем это нужно"
	AdminComment string           `gorm:"type:text"` // Ответ админа при отклонении/одобрении
	Status       ModerationStatus `gorm:"type:varchar(20);default:'pending'"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
