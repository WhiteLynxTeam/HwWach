-- +goose Down
-- +goose StatementBegin

-- Откат переименования таблиц
ALTER TABLE assets RENAME TO devices;
ALTER TABLE asset_photos RENAME TO device_photos;

-- Откат переименования колонки
ALTER TABLE requests RENAME COLUMN asset_id TO device_id;

-- Откат переименования индексов
ALTER INDEX idx_assets_user_id RENAME TO idx_devices_user_id;

-- Откат foreign key constraints
ALTER TABLE device_photos DROP CONSTRAINT fk_asset_photos_asset;
ALTER TABLE device_photos DROP CONSTRAINT fk_asset_photos_photo;
ALTER TABLE device_photos ADD CONSTRAINT fk_device_photos_device FOREIGN KEY (device_uuid) REFERENCES devices(uuid) ON DELETE CASCADE;
ALTER TABLE device_photos ADD CONSTRAINT fk_device_photos_photo FOREIGN KEY (photo_uuid) REFERENCES photos(uuid) ON DELETE CASCADE;

ALTER TABLE requests DROP CONSTRAINT fk_requests_asset;
ALTER TABLE requests ADD CONSTRAINT fk_requests_device FOREIGN KEY (device_id) REFERENCES devices(uuid) ON UPDATE CASCADE ON DELETE SET NULL;

-- Откат unique constraint
ALTER TABLE devices DROP CONSTRAINT assets_inventory_num_key;
ALTER TABLE devices ADD CONSTRAINT devices_inventory_num_key UNIQUE (inventory_num);

-- Откат первичных ключей
ALTER INDEX assets_pkey RENAME TO devices_pkey;
ALTER INDEX asset_photos_pkey RENAME TO device_photos_pkey;

-- +goose StatementEnd
