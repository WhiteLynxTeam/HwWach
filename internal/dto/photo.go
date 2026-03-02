package dto

import "github.com/google/uuid"

const MaxFileSize = 20 * 1024 * 1024 // 20 МБ

// UploadPhotoRequest запрос на получение presigned URL
type UploadPhotoRequest struct {
	Filename    string     `json:"filename" binding:"required" example:"photo.jpg"`
	ContentType string     `json:"content_type" example:"image/jpeg"`
	FileSize    int64      `json:"file_size" binding:"required" example:"1048576"`
	ClientID    *uuid.UUID `json:"client_id,omitempty" example:"0194f7b0-1234-7xxx-xxxx-xxxxxxxxxxxx"` // UUID сгенерированный клиентом (для оптимистичного UI)
}

// UploadPhotoResponse ответ с presigned URL для загрузки
type UploadPhotoResponse struct {
	UploadURL   string `json:"upload_url"`
	PhotoUUID   string `json:"photo_uuid"`
	ClientID    string `json:"client_id,omitempty"` // UUID сгенерированный клиентом
	Method      string `json:"method"`
	ExpiresIn   int    `json:"expires_in"`
	MaxFileSize int64  `json:"max_file_size"` // Макс. размер файла в байтах
}

// ConfirmUploadRequest запрос на подтверждение загрузки
type ConfirmUploadRequest struct {
	PhotoUUID string `json:"photo_uuid" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	DeviceID  string `json:"device_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// PhotoResponse ответ с информацией о фотографии
type PhotoResponse struct {
	UUID      string  `json:"uuid"`
	UserUUID  string  `json:"user_uuid"`
	URL       string  `json:"url"`
	CreatedAt string  `json:"created_at"`
	ClientID  *string `json:"client_id,omitempty"` // UUID сгенерированный клиентом (если был передан)
}

// PhotoListResponse ответ со списком фотографий
type PhotoListResponse struct {
	UserUUID string          `json:"user_uuid"`
	Photos   []PhotoResponse `json:"photos"`
}
