package dto

// CreateAssetRequest запрос на создание asset
type CreateAssetRequest struct {
	ClientID      *string `json:"client_id,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"` // UUID сгенерированный клиентом
	InventoryNum  string  `json:"inventory_num" binding:"required" example:"ИНВ-001"`
	Type          string  `json:"type" binding:"required" example:"ноутбук"`
	Specification string  `json:"specification" example:"MacBook Pro 16"`
	Status        string  `json:"status" example:"active"`
}

// UpdateAssetRequest запрос на частичное обновление asset до проверки админом
type UpdateAssetRequest struct {
	InventoryNum  *string `json:"inventory_num,omitempty" example:"ИНВ-002"`
	Type          *string `json:"type,omitempty" example:"ноутбук"`
	Specification *string `json:"specification,omitempty" example:"MacBook Pro 16 2023"`
	Status        *string `json:"status,omitempty" example:"inactive"`
}

// AssetResponse ответ с информацией об asset (без вложений)
type AssetResponse struct {
	UUID          string  `json:"uuid" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	ClientID      *string `json:"client_id,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	InventoryNum  string  `json:"inventory_num"`
	Type          string  `json:"type"`
	Specification string  `json:"specification"`
	Status        string  `json:"status"`
	UserUUID      string  `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	VerifiedAt    *string `json:"verified_at,omitempty" example:"2026-05-11T12:00:00Z"`
	AdminComment  string  `json:"admin_comment,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// AssetListResponse ответ со списком assets
type AssetListResponse struct {
	UserUUID string          `json:"user_uuid"`
	Assets   []AssetResponse `json:"assets"`
}
