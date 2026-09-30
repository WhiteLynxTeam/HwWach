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
