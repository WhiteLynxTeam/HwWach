// Package tsclient внутренние запросы go → ts (5.1)
package tsclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ErrUnavailable ts не ответил — действие не применяем, просим повторить
var ErrUnavailable = errors.New("не удалось проверить человека, повторите")

// ErrPersonNotFound человек не найден в ts
var ErrPersonNotFound = errors.New("человек не найден")

type Person struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	IsMol          bool      `json:"is_mol"`
	IsActive       bool      `json:"is_active"`
	FullName       string    `json:"full_name"`
	HasAccount     bool      `json:"has_account"`
}

type Client interface {
	GetPerson(ctx context.Context, id uuid.UUID) (*Person, error)
	ReportFilled(ctx context.Context, orgID uuid.UUID) error
	Touch(ctx context.Context, orgID uuid.UUID) error
}

type httpClient struct {
	baseURL string
	key     string
	http    *http.Client
}

// New тайм-аут ~60 с: запрос может разбудить спящий ts на Render
func New(baseURL, key string) Client {
	return &httpClient{baseURL: baseURL, key: key, http: &http.Client{Timeout: 60 * time.Second}}
}

func (c *httpClient) do(ctx context.Context, method, path string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("X-Internal-Key", c.key)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

func (c *httpClient) GetPerson(ctx context.Context, id uuid.UUID) (*Person, error) {
	body, status, err := c.do(ctx, http.MethodGet, "/internal/persons/"+id.String())
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, ErrPersonNotFound
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("%w: ts status %d", ErrUnavailable, status)
	}
	var p Person
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	return &p, nil
}

func (c *httpClient) post(ctx context.Context, path string) error {
	_, status, err := c.do(ctx, http.MethodPost, path)
	if err != nil {
		return err
	}
	if status >= 300 {
		return fmt.Errorf("ts status %d", status)
	}
	return nil
}

func (c *httpClient) ReportFilled(ctx context.Context, orgID uuid.UUID) error {
	return c.post(ctx, "/internal/organizations/"+orgID.String()+"/filled")
}

func (c *httpClient) Touch(ctx context.Context, orgID uuid.UUID) error {
	return c.post(ctx, "/internal/organizations/"+orgID.String()+"/touch")
}
