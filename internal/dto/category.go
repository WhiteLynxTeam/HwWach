package dto

// CategoryResponse ответ с информацией о категории
type CategoryResponse struct {
	UUID         string  `json:"uuid" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	Name         string  `json:"name" example:"Монитор"`
	Level        int16   `json:"level" example:"1"` // 1: Global, 2: Local, 3: User
	Status       string  `json:"status" example:"approved"`
	UsageCount   int     `json:"usage_count" example:"5"`
	CreatedBy    *string `json:"created_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	AdminComment string  `json:"admin_comment,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// CategorySyncResponse ответ для инкрементальной синхронизации категорий
type CategorySyncResponse struct {
	Categories []CategoryResponse `json:"categories"`
	DeletedIDs []string           `json:"deleted_ids"`
	SyncedAt   string             `json:"synced_at"`
}

// CreateCategoryRequest запрос на создание категории
type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required" example:"Новая категория"`
}

// ModerateCategoryRequest запрос администратора на модерацию пользовательской категории L3
type ModerateCategoryRequest struct {
	Action         string  `json:"action" binding:"required" example:"approve"` // "approve", "merge", "reject"
	Level          *int16  `json:"level,omitempty" example:"2"`                // 1 или 2 (по умолчанию 2 для approve)
	MergedIntoUUID *string `json:"merged_into_uuid,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"`
	AdminComment   *string `json:"admin_comment,omitempty"`
}

// AdminCategoryResponse ответ с информацией о категории для админ-панели (включая количество связанных ассетов)
type AdminCategoryResponse struct {
	UUID           string  `json:"uuid"`
	Name           string  `json:"name"`
	Level          int16   `json:"level"`
	Status         string  `json:"status"`
	UsageCount     int     `json:"usage_count"`
	CreatedBy      *string `json:"created_by,omitempty"`
	MergedIntoUUID *string `json:"merged_into_uuid,omitempty"`
	AdminComment   string  `json:"admin_comment,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
	AssetsCount    int64   `json:"assets_count"`
}

// AdminCreateCategoryRequest запрос администратора на создание новой категории
type AdminCreateCategoryRequest struct {
	Name         string `json:"name" binding:"required"`
	Level        int16  `json:"level" binding:"required"` // 1 (Global) или 2 (Local)
	AdminComment string `json:"admin_comment,omitempty"`
}

// AdminUpdateCategoryRequest запрос администратора на редактирование категории
type AdminUpdateCategoryRequest struct {
	Name               string  `json:"name" binding:"required"`
	Level              int16   `json:"level" binding:"required"`
	Status             *string `json:"status,omitempty"`
	AdminComment       string  `json:"admin_comment,omitempty"`
	UpdateLinkedAssets bool    `json:"update_linked_assets"`
}

// AdminMergeCategoryRequest запрос администратора на объединение категорий
type AdminMergeCategoryRequest struct {
	TargetCategoryUUID string  `json:"target_category_uuid" binding:"required"`
	AdminComment       *string `json:"admin_comment,omitempty"`
	UpdateLinkedAssets bool    `json:"update_linked_assets"`
}

// CategoryAffectedResponse информация о связанных ассетах категории
type CategoryAffectedResponse struct {
	CategoryUUID string              `json:"category_uuid"`
	CategoryName string              `json:"category_name"`
	AssetsCount  int64               `json:"assets_count"`
	SampleAssets []SimpleAssetSample `json:"sample_assets"`
}

// SimpleAssetSample краткая информация об ассете
type SimpleAssetSample struct {
	UUID         string  `json:"uuid"`
	InventoryNum *string `json:"inventory_num,omitempty"`
	Name         string  `json:"name"`
}

