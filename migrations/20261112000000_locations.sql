-- +goose Up
-- +goose StatementBegin

-- Справочник мест организации — дерево: населённый пункт → объект/подразделение → помещение.
-- id может прийти с телефона: место заводят без сети вместе с активом
CREATE TABLE IF NOT EXISTS locations (
    id                   UUID PRIMARY KEY,
    organization_id      UUID NOT NULL,
    parent_id            UUID REFERENCES locations(id),
    name                 VARCHAR(255) NOT NULL,
    normalized_name      VARCHAR(255) NOT NULL,
    kind                 VARCHAR(20) NOT NULL DEFAULT 'other'
                         CHECK (kind IN ('settlement', 'site', 'room', 'other')),
    -- «Каленковичи / Ферма 700 / Кабинет зав. фермы» — для поиска и показа одной строкой
    path                 TEXT NOT NULL,
    created_by_person_id UUID,
    created_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at           TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_locations_org ON locations(organization_id);
CREATE INDEX IF NOT EXISTS idx_locations_parent ON locations(parent_id);
-- одно имя на одном уровне одной ветки
CREATE UNIQUE INDEX IF NOT EXISTS idx_locations_unique_name
    ON locations(organization_id, COALESCE(parent_id, '00000000-0000-0000-0000-000000000000'::uuid), normalized_name)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_locations_path_trgm ON locations USING gin (lower(path) gin_trgm_ops);

-- unset — не указано (старые карточки), place — в месте из справочника,
-- none — места нет (почта, лицензия), portable — носимое, «у сотрудника» (смартфон, ЭЦП)
ALTER TABLE assets ADD COLUMN IF NOT EXISTS location_kind VARCHAR(20) NOT NULL DEFAULT 'unset'
    CHECK (location_kind IN ('unset', 'place', 'none', 'portable'));
ALTER TABLE assets ADD COLUMN IF NOT EXISTS location_id UUID REFERENCES locations(id);
CREATE INDEX IF NOT EXISTS idx_assets_location ON assets(location_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_assets_location;
ALTER TABLE assets DROP COLUMN IF EXISTS location_id;
ALTER TABLE assets DROP COLUMN IF EXISTS location_kind;
DROP TABLE IF EXISTS locations;
-- +goose StatementEnd
