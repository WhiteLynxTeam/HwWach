package models

import (
	"time"

	"github.com/google/uuid"
)

// RoleKind бизнес-роль человека на активе (3.2)
type RoleKind string

const (
	RoleMol  RoleKind = "mol"  // числится за (материально ответственное лицо)
	RoleUser RoleKind = "user" // работает с активом
)

// AckStatus подтверждение назначения получателем: все назначения применяются сразу,
// получатель потом подтверждает или отказывается с причиной (0, 3.3)
type AckStatus string

const (
	AckPending  AckStatus = "pending" // «без подтверждения»
	AckAccepted AckStatus = "accepted"
	AckDisputed AckStatus = "disputed"
)

// AssetRole связь человека с активом. При передаче закрывается (ended_at), а не удаляется
type AssetRole struct {
	ID                         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID             uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	AssetID                    uuid.UUID  `gorm:"type:uuid;not null" json:"asset_id"`
	PersonID                   uuid.UUID  `gorm:"type:uuid;not null" json:"person_id"`
	Role                       RoleKind   `gorm:"type:varchar(10);not null" json:"role"`
	StartedAt                  time.Time  `json:"started_at"`
	EndedAt                    *time.Time `json:"ended_at,omitempty"`
	EndReason                  *string    `gorm:"type:varchar(30)" json:"end_reason,omitempty"`
	AssignedByAccountID        *uuid.UUID `gorm:"type:uuid" json:"assigned_by_account_id,omitempty"`
	AssignedByPersonID         *uuid.UUID `gorm:"type:uuid" json:"assigned_by_person_id,omitempty"`
	TransferID                 *uuid.UUID `gorm:"type:uuid" json:"transfer_id,omitempty"`
	AckStatus                  AckStatus  `gorm:"type:varchar(10);not null;default:'pending'" json:"ack_status"`
	AckReason                  *string    `gorm:"type:text" json:"ack_reason,omitempty"`
	AckAt                      *time.Time `json:"ack_at,omitempty"`
	DisputeResolvedAt          *time.Time `json:"dispute_resolved_at,omitempty"`
	DisputeResolvedByAccountID *uuid.UUID `gorm:"type:uuid" json:"dispute_resolved_by_account_id,omitempty"`
	DisputeResolvedByPersonID  *uuid.UUID `gorm:"type:uuid" json:"dispute_resolved_by_person_id,omitempty"`
	DisputeResolution          *string    `gorm:"type:text" json:"dispute_resolution,omitempty"`

	Asset *Asset `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
}

// AssetTransfer запись истории передачи. Применяется сразу, статусов ожидания нет (3.3)
type AssetTransfer struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID       uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	AssetID              uuid.UUID  `gorm:"type:uuid;not null" json:"asset_id"`
	Role                 RoleKind   `gorm:"type:varchar(10);not null" json:"role"`
	FromPersonID         *uuid.UUID `gorm:"type:uuid" json:"from_person_id,omitempty"`
	ToPersonID           *uuid.UUID `gorm:"type:uuid" json:"to_person_id,omitempty"` // NULL — снятие
	InitiatedByAccountID *uuid.UUID `gorm:"type:uuid" json:"initiated_by_account_id,omitempty"`
	InitiatedByPersonID  *uuid.UUID `gorm:"type:uuid" json:"initiated_by_person_id,omitempty"`
	Reason               *string    `gorm:"type:varchar(30)" json:"reason,omitempty"`
	Comment              *string    `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// RepairKind вид работ: плановое обслуживание — тоже «В ремонт» (3.6)
type RepairKind string

const (
	RepairKindRepair      RepairKind = "repair"
	RepairKindMaintenance RepairKind = "maintenance"
)

// AssetRepair период ремонта. Одна открытая запись на актив
type AssetRepair struct {
	ID                  uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID      uuid.UUID        `gorm:"type:uuid;not null" json:"organization_id"`
	AssetID             uuid.UUID        `gorm:"type:uuid;not null" json:"asset_id"`
	RequestID           *uuid.UUID       `gorm:"type:uuid" json:"request_id,omitempty"`
	Kind                RepairKind       `gorm:"type:varchar(20);not null" json:"kind"`
	StartedAt           time.Time        `json:"started_at"`
	ExpectedReturnAt    *time.Time       `json:"expected_return_at,omitempty"`
	ReturnedAt          *time.Time       `json:"returned_at,omitempty"`
	StatusBefore        AccountingStatus `gorm:"type:varchar(20);not null" json:"status_before"`
	ResultCondition     *Condition       `gorm:"type:varchar(20)" json:"result_condition,omitempty"`
	WorkDone            *string          `gorm:"type:text" json:"work_done,omitempty"`
	Comment             *string          `gorm:"type:text" json:"comment,omitempty"`
	CreatedByAccountID  uuid.UUID        `gorm:"type:uuid;not null" json:"created_by_account_id"`
	CreatedByPersonID   uuid.UUID        `gorm:"type:uuid;not null" json:"created_by_person_id"`
	ReturnedByAccountID *uuid.UUID       `gorm:"type:uuid" json:"returned_by_account_id,omitempty"`
	ReturnedByPersonID  *uuid.UUID       `gorm:"type:uuid" json:"returned_by_person_id,omitempty"`
}

// AssetStatusHistory изменение учётного статуса или состояния
type AssetStatusHistory struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID uuid.UUID  `gorm:"type:uuid;not null" json:"organization_id"`
	AssetID        uuid.UUID  `gorm:"type:uuid;not null" json:"asset_id"`
	Field          string     `gorm:"type:varchar(20);not null" json:"field"` // accounting_status / condition
	FromValue      *string    `gorm:"type:varchar(20)" json:"from_value,omitempty"`
	ToValue        string     `gorm:"type:varchar(20);not null" json:"to_value"`
	AccountID      *uuid.UUID `gorm:"type:uuid" json:"account_id,omitempty"`
	PersonID       *uuid.UUID `gorm:"type:uuid" json:"person_id,omitempty"`
	RepairID       *uuid.UUID `gorm:"type:uuid" json:"repair_id,omitempty"`
	Comment        *string    `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (AssetStatusHistory) TableName() string { return "asset_status_history" }

// AuditLog журнал go (3.4)
type AuditLog struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	OrganizationID uuid.UUID      `gorm:"type:uuid;not null" json:"organization_id"`
	ActorType      string         `gorm:"type:varchar(10);not null;default:'user'" json:"actor_type"`
	AccountID      *uuid.UUID     `gorm:"type:uuid" json:"account_id,omitempty"`
	ActorPersonID  *uuid.UUID     `gorm:"type:uuid" json:"actor_person_id,omitempty"`
	Action         string         `gorm:"type:varchar(50);not null" json:"action"`
	ObjectType     *string        `gorm:"type:varchar(30)" json:"object_type,omitempty"`
	ObjectID       *uuid.UUID     `gorm:"type:uuid" json:"object_id,omitempty"`
	Details        JSONB          `gorm:"type:jsonb" json:"details,omitempty" swaggertype:"object"`
	IP             *string        `gorm:"type:varchar(64)" json:"ip,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_log" }

// OrganizationState что go уже сообщил ts об организации (5.1)
type OrganizationState struct {
	OrganizationID   uuid.UUID  `gorm:"type:uuid;primaryKey"`
	FilledReportedAt *time.Time
	TouchReportedAt  *time.Time
}

func (OrganizationState) TableName() string { return "organization_state" }
