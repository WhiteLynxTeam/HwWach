package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// YandexDiskStorage реализует интерфейс Storage для Яндекс.Диска
// через REST API (cloud-api.yandex.net)
type YandexDiskStorage struct {
	endpoint   string // https://cloud-api.yandex.net/v1
	token      string // OAuth-токен
	bucket     string // базовая папка на диске (аналог бакета)
	httpClient *http.Client
}

// yandexLinkResponse — ответ от API Яндекс.Диска на запрос upload/download ссылки
type yandexLinkResponse struct {
	Href      string `json:"href"`
	Method    string `json:"method"`
	Templated bool   `json:"templated"`
}

// yandexErrorResponse — ответ ошибки от API Яндекс.Диска
type yandexErrorResponse struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Error       string `json:"error"`
}

func NewYandexDiskStorage(endpoint, token, bucket string) (Storage, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("YANDEX_ENDPOINT is required")
	}
	if token == "" {
		return nil, fmt.Errorf("YANDEX_DISK_TOKEN is required")
	}
	if bucket == "" {
		bucket = "photos"
	}

	// Убираем завершающий слэш из endpoint
	endpoint = strings.TrimRight(endpoint, "/")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	s := &YandexDiskStorage{
		endpoint:   endpoint,
		token:      token,
		bucket:     bucket,
		httpClient: client,
	}

	// Проверяем/создаём базовую папку (аналог бакета)
	if err := s.ensureBucketFolder(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket folder: %w", err)
	}

	return s, nil
}

// ensureBucketFolder создаёт папку-бакет на Яндекс.Диске, если она не существует
func (s *YandexDiskStorage) ensureBucketFolder() error {
	folderPath := fmt.Sprintf("disk:/%s", s.bucket)

	reqURL := fmt.Sprintf("%s/disk/resources?path=%s",
		s.endpoint, url.QueryEscape(folderPath))

	req, err := http.NewRequest(http.MethodPut, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "OAuth "+s.token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	defer resp.Body.Close()

	// 201 — папка создана, 409 — уже существует. Оба случая нормальные
	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		return nil
	}

	var errResp yandexErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		return fmt.Errorf("yandex disk error (status %d): %s - %s",
			resp.StatusCode, errResp.Error, errResp.Description)
	}

	return fmt.Errorf("unexpected status code %d", resp.StatusCode)
}

// objectPath формирует полный путь к объекту на Яндекс.Диске
func (s *YandexDiskStorage) objectPath(objectName string) string {
	return fmt.Sprintf("disk:/%s/%s", s.bucket, objectName)
}

// PresignedGetURL получает временную ссылку на скачивание файла с Яндекс.Диска
func (s *YandexDiskStorage) PresignedGetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	path := s.objectPath(objectName)

	reqURL := fmt.Sprintf("%s/disk/resources/download?path=%s",
		s.endpoint, url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("Authorization", "OAuth "+s.token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get download link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp yandexErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return "", fmt.Errorf("yandex disk download error (status %d): %s - %s",
				resp.StatusCode, errResp.Error, errResp.Description)
		}
		return "", fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var linkResp yandexLinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&linkResp); err != nil {
		return "", fmt.Errorf("failed to decode download link response: %w", err)
	}

	return linkResp.Href, nil
}

// PresignedPutURL получает временную ссылку на загрузку файла на Яндекс.Диск.
// Перед получением ссылки создаёт необходимые промежуточные папки.
func (s *YandexDiskStorage) PresignedPutURL(ctx context.Context, objectName, contentType string, expiry time.Duration) (string, error) {
	path := s.objectPath(objectName)

	// Создаём промежуточные папки (например, user_uuid/)
	if err := s.ensureParentFolders(ctx, objectName); err != nil {
		return "", fmt.Errorf("failed to ensure parent folders: %w", err)
	}

	reqURL := fmt.Sprintf("%s/disk/resources/upload?path=%s&overwrite=true",
		s.endpoint, url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Authorization", "OAuth "+s.token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get upload link: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp yandexErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return "", fmt.Errorf("yandex disk upload error (status %d): %s - %s",
				resp.StatusCode, errResp.Error, errResp.Description)
		}
		return "", fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var linkResp yandexLinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&linkResp); err != nil {
		return "", fmt.Errorf("failed to decode upload link response: %w", err)
	}

	return linkResp.Href, nil
}

// Delete удаляет файл с Яндекс.Диска навсегда (permanently=true)
func (s *YandexDiskStorage) Delete(ctx context.Context, objectName string) error {
	path := s.objectPath(objectName)

	reqURL := fmt.Sprintf("%s/disk/resources?path=%s&permanently=true",
		s.endpoint, url.QueryEscape(path))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.Header.Set("Authorization", "OAuth "+s.token)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	// 204 — успешно удалено, 404 — файл не найден (считаем OK)
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	var errResp yandexErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		return fmt.Errorf("yandex disk delete error (status %d): %s - %s",
			resp.StatusCode, errResp.Error, errResp.Description)
	}

	return fmt.Errorf("unexpected status code %d", resp.StatusCode)
}

// GetPublicURL возвращает путь к объекту.
// Для Яндекс.Диска прямой публичный URL недоступен, поэтому
// возвращаем относительный путь (как и у MinIO при отсутствии externalBaseURL)
func (s *YandexDiskStorage) GetPublicURL(objectName string) string {
	return fmt.Sprintf("%s/%s", s.bucket, objectName)
}

// ensureParentFolders создаёт все промежуточные папки для файла.
// Например, для objectName="user-uuid/photo.jpg" создаст disk:/photos/user-uuid/
func (s *YandexDiskStorage) ensureParentFolders(ctx context.Context, objectName string) error {
	parts := strings.Split(objectName, "/")
	if len(parts) <= 1 {
		return nil // нет промежуточных папок
	}

	// Создаём все папки кроме последнего элемента (это сам файл)
	currentPath := fmt.Sprintf("disk:/%s", s.bucket)
	for _, part := range parts[:len(parts)-1] {
		currentPath = fmt.Sprintf("%s/%s", currentPath, part)

		reqURL := fmt.Sprintf("%s/disk/resources?path=%s",
			s.endpoint, url.QueryEscape(currentPath))

		req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create folder request: %w", err)
		}
		req.Header.Set("Authorization", "OAuth "+s.token)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to create folder %s: %w", currentPath, err)
		}
		resp.Body.Close()

		// 201 — создано, 409 — уже существует
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
			return fmt.Errorf("failed to create folder %s: status %d", currentPath, resp.StatusCode)
		}
	}

	return nil
}
