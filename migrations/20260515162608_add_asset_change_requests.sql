-- +goose Up
-- +goose StatementBegin

-- Изменение таблицы assets
ALTER TABLE assets ADD COLUMN IF NOT EXISTS moderation_status VARCHAR(20) DEFAULT 'pending';

-- DeletedAt от gorm обычно имеет тип timestamp with time zone (и индекс)
ALTER TABLE assets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;
CREATE INDEX IF NOT EXISTS idx_assets_deleted_at ON assets(deleted_at);

-- Создание таблицы asset_change_requests
CREATE TABLE IF NOT EXISTS asset_change_requests (
    uuid UUID PRIMARY KEY,
    asset_uuid UUID NOT NULL REFERENCES assets(uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    proposed_data JSONB,
    reason TEXT NOT NULL,
    admin_comment TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_asset_change_requests_asset_uuid ON asset_change_requests(asset_uuid);
CREATE INDEX IF NOT EXISTS idx_asset_change_requests_status ON asset_change_requests(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS asset_change_requests;
DROP INDEX IF EXISTS idx_assets_deleted_at;
ALTER TABLE assets DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE assets DROP COLUMN IF EXISTS moderation_status;
-- +goose StatementEnd
