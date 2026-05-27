-- +goose Up
-- +goose StatementBegin

-- Добавление client_id в таблицу asset_change_requests
ALTER TABLE asset_change_requests ADD COLUMN IF NOT EXISTS client_id UUID;
CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_change_requests_client_id ON asset_change_requests(client_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Откат изменений в таблице asset_change_requests
DROP INDEX IF EXISTS idx_asset_change_requests_client_id;
ALTER TABLE asset_change_requests DROP COLUMN IF EXISTS client_id;

-- +goose StatementEnd
