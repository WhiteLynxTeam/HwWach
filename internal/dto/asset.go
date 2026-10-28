package dto

// CreateAssetRequest запрос на создание asset
type CreateAssetRequest struct {
	ClientID       *string  `json:"client_id,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"` // UUID сгенерированный клиентом
	InventoryNum   *string  `json:"inventory_num,omitempty" example:"ИНВ-001"`
	Name           string   `json:"name" binding:"required" example:"Ноутбук служебный"`
	Category       string   `json:"category,omitempty" example:"ноутбук"`
	CategoryUUID   *string  `json:"category_uuid,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	CategoryText   *string  `json:"category_text,omitempty" example:"ноутбук"`
	Description    string   `json:"description" example:"MacBook Pro 16"`
	// in_use / in_stock / lost; по умолчанию in_use
	AccountingStatus string `json:"accounting_status,omitempty" example:"in_use"`
	// ok / needs_attention / faulty / unusable; по умолчанию ok
	Condition      string   `json:"condition,omitempty" example:"ok"`
	PhotoClientIDs []string `json:"photo_client_ids" example:"[\"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx\"]"`
	// МОЛ: не передан — создатель, если он МОЛ; "" — без МОЛ; uuid — этот человек (другого — только админ)
	MolPersonID *string `json:"mol_person_id,omitempty"`
}

// UpdateAssetRequest частичное обновление карточки (статусы — через /status и ремонт)
type UpdateAssetRequest struct {
	InventoryNum *string `json:"inventory_num,omitempty" example:"ИНВ-002"`
	Name         *string `json:"name,omitempty" example:"Ноутбук служебный"`
	Category     *string `json:"category,omitempty" example:"ноутбук"`
	CategoryUUID *string `json:"category_uuid,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	CategoryText *string `json:"category_text,omitempty" example:"ноутбук"`
	Description  *string `json:"description,omitempty" example:"MacBook Pro 16 2023"`
}

// AssetRoleResponse действующая бизнес-роль на активе
type AssetRoleResponse struct {
	ID        string  `json:"id"`
	PersonID  string  `json:"person_id"`
	Role      string  `json:"role"` // mol / user
	StartedAt string  `json:"started_at"`
	EndedAt   *string `json:"ended_at,omitempty"`
	EndReason *string `json:"end_reason,omitempty"`
	// pending — «без подтверждения», accepted, disputed — отказ (на разбор админу)
	AckStatus         string  `json:"ack_status"`
	AckReason         *string `json:"ack_reason,omitempty"`
	AckAt             *string `json:"ack_at,omitempty"`
	DisputeResolvedAt *string `json:"dispute_resolved_at,omitempty"`
	AssignedByPerson  *string `json:"assigned_by_person_id,omitempty"`
}

// RepairResponse период ремонта
type RepairResponse struct {
	ID               string  `json:"id"`
	Kind             string  `json:"kind"`
	StartedAt        string  `json:"started_at"`
	ExpectedReturnAt *string `json:"expected_return_at,omitempty"`
	ReturnedAt       *string `json:"returned_at,omitempty"`
	Overdue          bool    `json:"overdue"`
	StatusBefore     string  `json:"status_before"`
	ResultCondition  *string `json:"result_condition,omitempty"`
	WorkDone         *string `json:"work_done,omitempty"`
	Comment          *string `json:"comment,omitempty"`
	RequestID        *string `json:"request_id,omitempty"`
	CreatedByPerson  string  `json:"created_by_person_id"`
}

// AssetResponse ответ с информацией об asset (без вложений)
type AssetResponse struct {
	UUID              string              `json:"uuid" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	OrganizationID    string              `json:"organization_id"`
	ClientID          *string             `json:"client_id,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	InventoryNum      *string             `json:"inventory_num,omitempty"`
	Name              string              `json:"name"`
	Category          string              `json:"category"`
	CategoryUUID      *string             `json:"category_uuid,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	Description       string              `json:"description"`
	AccountingStatus  string              `json:"accounting_status" example:"in_use"`
	Condition         string              `json:"condition" example:"ok"`
	ModerationStatus  string              `json:"moderation_status" example:"approved"`
	CreatedByAccount  string              `json:"created_by_account_id"`
	CreatedByPerson   string              `json:"created_by_person_id"`
	Mol               *AssetRoleResponse  `json:"mol"`
	Users             []AssetRoleResponse `json:"users"`
	OpenRepair        *RepairResponse     `json:"open_repair,omitempty"`
	VerifiedAt        *string             `json:"verified_at,omitempty" example:"2026-05-11T12:00:00Z"`
	AdminComment      string              `json:"admin_comment,omitempty"`
	CreatedAt         string              `json:"created_at"`
	UpdatedAt         string              `json:"updated_at"`
}

// AssetListResponse ответ со списком assets
type AssetListResponse struct {
	PersonID string                    `json:"person_id"`
	Assets   []AssetWithPhotosResponse `json:"assets"`
}

// AssetWithPhotosResponse ответ с информацией об asset и его фотографиями
type AssetWithPhotosResponse struct {
	AssetResponse
	Photos []PhotoResponse `json:"photos"`
}

// PaginatedAssetResponse ответ со списком assets с пагинацией
type PaginatedAssetResponse struct {
	Assets []AssetWithPhotosResponse `json:"assets"`
	Total  int64                     `json:"total"`
	Page   int                       `json:"page"`
	Limit  int                       `json:"limit"`
	Pages  int                       `json:"pages"`
}

// ModerateAssetRequest одобрение карточки: сразу можно назначить МОЛ
type ModerateAssetRequest struct {
	Status       string  `json:"status" binding:"required,oneof=approved rejected" example:"approved"`
	AdminComment string  `json:"admin_comment,omitempty"`
	MolPersonID  *string `json:"mol_person_id,omitempty"`
}

// ChangeStatusRequest учётный статус и/или техническое состояние (3.6)
type ChangeStatusRequest struct {
	AccountingStatus *string `json:"accounting_status,omitempty" example:"in_stock"`
	Condition        *string `json:"condition,omitempty" example:"needs_attention"`
	Comment          string  `json:"comment,omitempty"`
}

// RepairStartRequest «В ремонт»
type RepairStartRequest struct {
	Kind             string  `json:"kind,omitempty" example:"repair"` // repair / maintenance
	ExpectedReturnAt *string `json:"expected_return_at,omitempty" example:"2026-11-20T00:00:00Z"`
	Comment          string  `json:"comment,omitempty"`
	RequestID        *string `json:"request_id,omitempty"`
}

// RepairFinishRequest «Из ремонта»
type RepairFinishRequest struct {
	ResultCondition string  `json:"result_condition" binding:"required,oneof=ok faulty unusable" example:"ok"`
	WorkDone        string  `json:"work_done,omitempty"`
	StatusAfter     *string `json:"status_after,omitempty" example:"in_use"`
	Comment         string  `json:"comment,omitempty"`
}

// AssignPersonRequest передать МОЛ / назначить работающего с активом
type AssignPersonRequest struct {
	PersonID string `json:"person_id" binding:"required"`
	Comment  string `json:"comment,omitempty"`
}

// AckRequest подтвердить или отказаться от назначения
type AckRequest struct {
	Action string `json:"action" binding:"required,oneof=accept dispute" example:"accept"`
	Reason string `json:"reason,omitempty"`
}

// ResolveRequest админ разобрал отказ
type ResolveRequest struct {
	Resolution string `json:"resolution,omitempty"`
}

// AssignmentResponse «Мне назначено»
type AssignmentResponse struct {
	AssetRoleResponse
	Asset *AssetResponse `json:"asset,omitempty"`
}

// CreateRepairRequest заявка на ремонт
type CreateRepairRequest struct {
	AssetUUID  string  `json:"asset_uuid" binding:"required"`
	Message    string  `json:"message" binding:"required"`
	PhotoURL   *string `json:"photo_url,omitempty"`
	MarkFaulty bool    `json:"mark_faulty,omitempty"`
}

// UpdateRepairRequestStatus закрыть заявку
type UpdateRepairRequestStatus struct {
	Status  string `json:"status" binding:"required,oneof=done rejected"`
	Comment string `json:"comment,omitempty"`
}
