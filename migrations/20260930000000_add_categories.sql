-- +goose Up
-- +goose StatementBegin

-- 1. Включаем расширение pg_trgm для нечёткого поиска
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 2. Таблица categories
CREATE TABLE IF NOT EXISTS categories (
    uuid             UUID PRIMARY KEY,
    name             VARCHAR(255) NOT NULL,
    normalized_name  VARCHAR(255) NOT NULL,
    level            SMALLINT NOT NULL CHECK (level IN (1, 2, 3)),
    status           VARCHAR(20) NOT NULL DEFAULT 'approved'
                     CHECK (status IN ('approved', 'pending', 'rejected', 'merged')),
    created_by       UUID,
    merged_into_uuid UUID REFERENCES categories(uuid),
    usage_count      INTEGER DEFAULT 0,
    admin_comment    TEXT,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at       TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_categories_level ON categories(level);
CREATE INDEX IF NOT EXISTS idx_categories_status ON categories(status);
CREATE INDEX IF NOT EXISTS idx_categories_created_by ON categories(created_by);
CREATE INDEX IF NOT EXISTS idx_categories_deleted_at ON categories(deleted_at);

-- GIN-индекс для нечёткого поиска по триграммам
CREATE INDEX IF NOT EXISTS idx_categories_name_trgm ON categories USING gin (normalized_name gin_trgm_ops);

-- Уникальность имени по уровням
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_unique_global
    ON categories(normalized_name) WHERE level = 1 AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_unique_local
    ON categories(normalized_name) WHERE level = 2 AND deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_unique_user
    ON categories(normalized_name, created_by) WHERE level = 3 AND deleted_at IS NULL;

-- 3. Начальные данные (seed)
INSERT INTO categories (uuid, name, normalized_name, level, status, usage_count)
VALUES
    (gen_random_uuid(), 'Системный блок', 'системный блок', 1, 'approved', 0),
    (gen_random_uuid(), 'Монитор',        'монитор',        1, 'approved', 0),
    (gen_random_uuid(), 'Клавиатура',      'клавиатура',      1, 'approved', 0),
    (gen_random_uuid(), 'Мышь',            'мышь',            1, 'approved', 0),
    (gen_random_uuid(), 'Телефон',        'телефон',        1, 'approved', 0)
ON CONFLICT DO NOTHING;

-- 4. Добавление внешнего ключа в assets
ALTER TABLE assets ADD COLUMN IF NOT EXISTS category_uuid UUID REFERENCES categories(uuid);
CREATE INDEX IF NOT EXISTS idx_assets_category_uuid ON assets(category_uuid);

-- Связываем существующие ассеты (если есть) по нормализованному названию
UPDATE assets a
SET category_uuid = c.uuid
FROM categories c
WHERE c.normalized_name = LOWER(TRIM(a.category))
  AND a.category_uuid IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE assets DROP COLUMN IF EXISTS category_uuid;
DROP TABLE IF EXISTS categories;

-- +goose StatementEnd
