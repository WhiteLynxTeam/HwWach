-- +goose Up
-- +goose StatementBegin

-- Переименование таблицы devices → assets
ALTER TABLE devices RENAME TO assets;

-- Переименование таблицы device_photos → asset_photos
ALTER TABLE device_photos RENAME TO asset_photos;

-- Переименование колонки device_id → asset_id в таблице requests
ALTER TABLE requests RENAME COLUMN device_id TO asset_id;

-- Переименование индексов
ALTER INDEX idx_devices_user_id RENAME TO idx_assets_user_id;

-- Переименование foreign key constraints
-- Сначала удаляем старые, потом создаём новые
ALTER TABLE asset_photos DROP CONSTRAINT fk_device_photos_device;
ALTER TABLE asset_photos DROP CONSTRAINT fk_device_photos_photo;
ALTER TABLE asset_photos ADD CONSTRAINT fk_asset_photos_asset FOREIGN KEY (device_uuid) REFERENCES assets(uuid) ON DELETE CASCADE;
ALTER TABLE asset_photos ADD CONSTRAINT fk_asset_photos_photo FOREIGN KEY (photo_uuid) REFERENCES photos(uuid) ON DELETE CASCADE;

ALTER TABLE requests DROP CONSTRAINT fk_requests_device;
ALTER TABLE requests ADD CONSTRAINT fk_requests_asset FOREIGN KEY (asset_id) REFERENCES assets(uuid) ON UPDATE CASCADE ON DELETE SET NULL;

-- Обновление unique constraint для inventory_num
ALTER TABLE assets DROP CONSTRAINT devices_inventory_num_key;
ALTER TABLE assets ADD CONSTRAINT assets_inventory_num_key UNIQUE (inventory_num);

-- Переименование первичных ключей (опционально, но для порядка)
ALTER INDEX devices_pkey RENAME TO assets_pkey;
ALTER INDEX device_photos_pkey RENAME TO asset_photos_pkey;

-- +goose StatementEnd
