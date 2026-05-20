package dto

import (
	"encoding/json"
	"time"
)

// CreateChangeRequestInput используется пользователем для создания заявки на изменение
type CreateChangeRequestInput struct {
	Type         string          `json:"type" binding:"required,oneof=update delete" example:"update"`
	ProposedData json.RawMessage `json:"proposed_data,omitempty" swaggertype:"object"`
	Reason       string          `json:"reason" binding:"required" example:"Неверно указан тип"`
}

// ApproveChangeRequestInput используется администратором для одобрения/отклонения
type ApproveChangeRequestInput struct {
	Status       string `json:"status" binding:"required,oneof=approved rejected" example:"approved"`
	AdminComment string `json:"admin_comment,omitempty" example:"Изменения приняты"`
}

// AssetChangeRequestResponse используется для отдачи данных наружу
type AssetChangeRequestResponse struct {
	UUID         string    `json:"uuid" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	AssetUUID    string    `json:"asset_uuid" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	UserUUID     string    `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type         string    `json:"type" example:"update"`
	Reason       string    `json:"reason" example:"Неверно указан тип"`
	AdminComment string    `json:"admin_comment,omitempty"`
	Status       string    `json:"status" example:"pending"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
