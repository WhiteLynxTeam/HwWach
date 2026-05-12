-- +goose Up
-- +goose StatementBegin

-- Переименование таблицы devices → assets
ALTER TABLE devices RENAME TO assets;

-- Переименование таблицы device_photos → asset_photos
ALTER TABLE device_photos RENAME TO asset_photos;

-- Переименование колонки device_id → asset_id в таблице requests
ALTER TABLE requests RENAME COLUMN device_id TO asset_id;

-- Переименование индексов
ALTER INDEX idx_devices_user_uuid RENAME TO idx_assets_user_uuid;

-- Переименование foreign key constraints
-- Сначала удаляем старые, потом создаём новые
ALTER TABLE asset_photos DROP CONSTRAINT fk_device_photos_device;
ALTER TABLE asset_photos DROP CONSTRAINT fk_device_photos_photo;

-- Переименование колонки
ALTER TABLE asset_photos RENAME COLUMN device_uuid TO asset_uuid;

ALTER TABLE asset_photos ADD CONSTRAINT fk_asset_photos_asset FOREIGN KEY (asset_uuid) REFERENCES assets(uuid) ON DELETE CASCADE;
ALTER TABLE asset_photos ADD CONSTRAINT fk_asset_photos_photo FOREIGN KEY (photo_uuid) REFERENCES photos(uuid) ON DELETE CASCADE;

ALTER TABLE requests DROP CONSTRAINT fk_devices_requests;
ALTER TABLE requests ADD CONSTRAINT fk_assets_requests FOREIGN KEY (asset_id) REFERENCES assets(uuid) ON UPDATE CASCADE ON DELETE SET NULL;

-- Обновление unique constraint для inventory_num
-- GORM создал индекс uni_devices_inventory_num, переименуем его через DROP/ADD
ALTER TABLE assets DROP CONSTRAINT IF EXISTS uni_devices_inventory_num;
ALTER TABLE assets ADD CONSTRAINT uni_assets_inventory_num UNIQUE (inventory_num);

-- Переименование первичных ключей (опционально, но для порядка)
ALTER INDEX devices_pkey RENAME TO assets_pkey;
ALTER INDEX device_photos_pkey RENAME TO asset_photos_pkey;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Откат переименования таблиц
ALTER TABLE assets RENAME TO devices;
ALTER TABLE asset_photos RENAME TO device_photos;

-- Откат переименования колонки
ALTER TABLE requests RENAME COLUMN asset_id TO device_id;

-- Откат переименования индексов
ALTER INDEX idx_assets_user_uuid RENAME TO idx_devices_user_uuid;

-- Откат foreign key constraints
ALTER TABLE device_photos DROP CONSTRAINT fk_asset_photos_asset;
ALTER TABLE device_photos DROP CONSTRAINT fk_asset_photos_photo;

-- Откат переименования колонки
ALTER TABLE device_photos RENAME COLUMN asset_uuid TO device_uuid;

ALTER TABLE device_photos ADD CONSTRAINT fk_device_photos_device FOREIGN KEY (device_uuid) REFERENCES devices(uuid) ON DELETE CASCADE;
ALTER TABLE device_photos ADD CONSTRAINT fk_device_photos_photo FOREIGN KEY (photo_uuid) REFERENCES photos(uuid) ON DELETE CASCADE;

ALTER TABLE requests DROP CONSTRAINT fk_assets_requests;
ALTER TABLE requests ADD CONSTRAINT fk_devices_requests FOREIGN KEY (device_id) REFERENCES devices(uuid) ON UPDATE CASCADE ON DELETE SET NULL;

-- Откат unique constraint
ALTER TABLE devices DROP CONSTRAINT IF EXISTS uni_assets_inventory_num;
ALTER TABLE devices ADD CONSTRAINT uni_devices_inventory_num UNIQUE (inventory_num);

-- Откат первичных ключей
ALTER INDEX assets_pkey RENAME TO devices_pkey;
ALTER INDEX asset_photos_pkey RENAME TO device_photos_pkey;

-- +goose StatementEnd