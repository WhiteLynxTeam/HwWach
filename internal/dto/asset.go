package dto

// CreateAssetRequest запрос на создание asset
type CreateAssetRequest struct {
	ClientID       *string  `json:"client_id,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"` // UUID сгенерированный клиентом
	InventoryNum   string   `json:"inventory_num" binding:"required" example:"ИНВ-001"`
	Name           string   `json:"name" binding:"required" example:"Ноутбук служебный"`
	Category       string   `json:"category" binding:"required" example:"ноутбук"`
	Description    string   `json:"description" example:"MacBook Pro 16"`
	AssetStatus    string   `json:"asset_status" example:"active"`
	PhotoClientIDs []string `json:"photo_client_ids" example:"[\"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx\"]"`
}

// UpdateAssetRequest запрос на частичное обновление asset до проверки админом
type UpdateAssetRequest struct {
	InventoryNum *string `json:"inventory_num,omitempty" example:"ИНВ-002"`
	Name         *string `json:"name,omitempty" example:"Ноутбук служебный"`
	Category     *string `json:"category,omitempty" example:"ноутбук"`
	Description  *string `json:"description,omitempty" example:"MacBook Pro 16 2023"`
	AssetStatus  *string `json:"asset_status,omitempty" example:"inactive"`
}

// AssetResponse ответ с информацией об asset (без вложений)
type AssetResponse struct {
	UUID         string  `json:"uuid" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	ClientID     *string `json:"client_id,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	InventoryNum string  `json:"inventory_num"`
	Name         string  `json:"name"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	AssetStatus  string  `json:"asset_status"`
	UserUUID     string  `json:"user_uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	VerifiedAt   *string `json:"verified_at,omitempty" example:"2026-05-11T12:00:00Z"`
	AdminComment string  `json:"admin_comment,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// AssetListResponse ответ со списком assets
type AssetListResponse struct {
	UserUUID string                    `json:"user_uuid"`
	Assets   []AssetWithPhotosResponse `json:"assets"`
}

// AssetWithPhotosResponse ответ с информацией об asset и его фотографиями
type AssetWithPhotosResponse struct {
	AssetResponse
	CreatedBy *string         `json:"created_by"`
	Photos    []PhotoResponse `json:"photos"`
}

// PaginatedAssetResponse ответ со списком assets с пагинацией
type PaginatedAssetResponse struct {
	Assets []AssetWithPhotosResponse `json:"assets"`
	Total  int64                     `json:"total"`
	Page   int                       `json:"page"`
	Limit  int                       `json:"limit"`
	Pages  int                       `json:"pages"`
}
