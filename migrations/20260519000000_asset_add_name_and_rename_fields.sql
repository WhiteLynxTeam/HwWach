-- +goose Up
-- +goose StatementBegin

-- 1. ОЧИСТКА ТЕСТОВЫХ ДАННЫХ
-- Очищаем основные таблицы и связанные (если есть связи)
-- CASCADE удалит данные из связующих таблиц (например, asset_photos), если настроены Foreign Keys
TRUNCATE TABLE assets CASCADE;
TRUNCATE TABLE asset_change_requests CASCADE;

-- 2. Обновляем таблицу ASSETS
-- Теперь колонка name добавится в пустую таблицу, NOT NULL не вызовет конфликтов
ALTER TABLE assets ADD COLUMN name VARCHAR(255) NOT NULL;

-- Переименовываем колонки
ALTER TABLE assets RENAME COLUMN type TO category;
ALTER TABLE assets RENAME COLUMN specification TO description;
ALTER TABLE assets RENAME COLUMN status TO asset_status;

-- 3. Обновляем таблицу ASSET_CHANGE_REQUESTS
ALTER TABLE asset_change_requests RENAME COLUMN type TO request_type;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1. Откатываем таблицу ASSET_CHANGE_REQUESTS
ALTER TABLE asset_change_requests RENAME COLUMN request_type TO type;

-- 2. Откатываем таблицу ASSETS
ALTER TABLE assets RENAME COLUMN asset_status TO status;
ALTER TABLE assets RENAME COLUMN description TO specification;
ALTER TABLE assets RENAME COLUMN category TO type;
ALTER TABLE assets DROP COLUMN name;

-- +goose StatementEnd