-- +goose Up
-- +goose StatementBegin

-- ========================================
-- Консолидированная инициализация схемы БД
-- Объединяет миграции:
--   20260511000000_rename_device_to_asset
--   20260515162608_add_asset_change_requests
--   20260518140500_add_missing_fields_to_assets_and_photos
--   20260519000000_asset_add_name_and_rename_fields
--   20260521000000_add_client_id_to_change_requests
-- ========================================

-- 1. Таблица assets (активы)
CREATE TABLE IF NOT EXISTS assets (
    uuid              UUID PRIMARY KEY,
    client_id         UUID,
    inventory_num     VARCHAR(255) NOT NULL,
    name              VARCHAR(255) NOT NULL,
    category          VARCHAR(255) NOT NULL,
    description       TEXT,
    user_id           UUID NOT NULL,
    asset_status      VARCHAR(20) NOT NULL DEFAULT 'active',
    moderation_status VARCHAR(20) DEFAULT 'pending',
    verified_at       TIMESTAMP WITH TIME ZONE,
    admin_comment     TEXT,
    created_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at        TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS uni_assets_inventory_num ON assets(inventory_num);
CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_client_id ON assets(client_id);
CREATE INDEX IF NOT EXISTS idx_assets_user_uuid ON assets(user_id);
CREATE INDEX IF NOT EXISTS idx_assets_deleted_at ON assets(deleted_at);

-- 2. Таблица photos (фотографии)
CREATE TABLE IF NOT EXISTS photos (
    uuid         UUID PRIMARY KEY,
    client_id    UUID,
    user_id      UUID NOT NULL,
    url          VARCHAR(255) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    file_size    BIGINT NOT NULL,
    file_name    VARCHAR(255) NOT NULL,
    content_type VARCHAR(255) NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at   TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_photos_client_id ON photos(client_id);
CREATE INDEX IF NOT EXISTS idx_photos_user_id ON photos(user_id);

-- 3. Таблица asset_photos (связь many-to-many: assets <-> photos)
CREATE TABLE IF NOT EXISTS asset_photos (
    asset_uuid UUID NOT NULL,
    photo_uuid UUID NOT NULL,
    PRIMARY KEY (asset_uuid, photo_uuid),
    CONSTRAINT fk_asset_photos_asset FOREIGN KEY (asset_uuid) REFERENCES assets(uuid) ON DELETE CASCADE,
    CONSTRAINT fk_asset_photos_photo FOREIGN KEY (photo_uuid) REFERENCES photos(uuid) ON DELETE CASCADE
);

-- 4. Таблица requests (запросы на обслуживание)
CREATE TABLE IF NOT EXISTS requests (
    uuid       UUID PRIMARY KEY,
    asset_id   UUID NOT NULL,
    user_id    UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message    TEXT NOT NULL,
    photo_url  VARCHAR(255),
    status     VARCHAR(20) NOT NULL DEFAULT 'Оформлена',
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT fk_assets_requests FOREIGN KEY (asset_id) REFERENCES assets(uuid) ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_requests_asset_id ON requests(asset_id);
CREATE INDEX IF NOT EXISTS idx_requests_user_id ON requests(user_id);

-- 5. Таблица asset_change_requests (заявки на изменение активов)
CREATE TABLE IF NOT EXISTS asset_change_requests (
    uuid          UUID PRIMARY KEY,
    client_id     UUID,
    asset_uuid    UUID NOT NULL,
    user_uuid     UUID NOT NULL,
    request_type  VARCHAR(20) NOT NULL,
    proposed_data JSONB,
    reason        TEXT NOT NULL,
    admin_comment TEXT,
    status        VARCHAR(20) DEFAULT 'pending',
    created_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_change_requests_client_id ON asset_change_requests(client_id);
CREATE INDEX IF NOT EXISTS idx_asset_change_requests_asset_uuid ON asset_change_requests(asset_uuid);
CREATE INDEX IF NOT EXISTS idx_asset_change_requests_status ON asset_change_requests(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS asset_change_requests;
DROP TABLE IF EXISTS requests;
DROP TABLE IF EXISTS asset_photos;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS assets;

-- +goose StatementEnd
