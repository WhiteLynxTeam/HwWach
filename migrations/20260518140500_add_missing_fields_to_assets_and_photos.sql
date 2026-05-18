-- +goose Up
-- +goose StatementBegin

-- Добавление client_id в таблицу assets
ALTER TABLE assets ADD COLUMN IF NOT EXISTS client_id UUID;
CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_client_id ON assets(client_id);

-- Добавление verified_at и admin_comment в таблицу assets
ALTER TABLE assets ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE assets ADD COLUMN IF NOT EXISTS admin_comment TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Откат изменений в таблице assets
ALTER TABLE assets DROP COLUMN IF EXISTS admin_comment;
ALTER TABLE assets DROP COLUMN IF EXISTS verified_at;
DROP INDEX IF EXISTS idx_assets_client_id;
ALTER TABLE assets DROP COLUMN IF EXISTS client_id;

-- +goose StatementEnd
