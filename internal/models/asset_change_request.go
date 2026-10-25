package models

import (
	"time"

	"github.com/google/uuid"
)

type RequestType string

const (
	RequestTypeUpdate RequestType = "update"
	RequestTypeDelete RequestType = "delete"
)

// AssetChangeRequest заявка на изменение актива
type AssetChangeRequest struct {
	UUID           uuid.UUID   `gorm:"type:uuid;primaryKey"`
	OrganizationID uuid.UUID   `gorm:"type:uuid;not null"`
	ClientID       *uuid.UUID  `gorm:"type:uuid;uniqueIndex" json:"client_id,omitempty"`
	AssetUUID      uuid.UUID   `gorm:"type:uuid;index"`
	AccountUUID    uuid.UUID   `gorm:"type:uuid;column:account_uuid"`
	PersonUUID     uuid.UUID   `gorm:"type:uuid;column:person_uuid"` // человек в момент заявки
	RequestType    RequestType `gorm:"column:request_type;type:varchar(20)"` // update или delete

	// Новые данные, которые пользователь хочет внести (в формате JSON)
	// Если Type == delete, это поле может быть пустым
	ProposedData []byte `gorm:"type:jsonb"`

	Reason               string           `gorm:"type:text"` // Комментарий пользователя "Зачем это нужно"
	AdminComment         string           `gorm:"type:text"` // Ответ админа при отклонении/одобрении
	Status               ModerationStatus `gorm:"type:varchar(20);default:'pending'"`
	ProcessedByAccountID *uuid.UUID       `gorm:"type:uuid"`
	ProcessedByPersonID  *uuid.UUID       `gorm:"type:uuid"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
