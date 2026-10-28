-- +goose Up
-- +goose StatementBegin

-- ========================================
-- Организации, бизнес-роли, передачи, состояние актива, журнал (org-roles-plan.md, раздел 3).
-- Рабочих данных ещё нет: прежние таблицы удаляются и создаются заново, без переноса.
-- Связей с базой ts нет: organization_id / person_id / account_id — обычные uuid (5.1).
-- ========================================

DROP TABLE IF EXISTS asset_change_requests CASCADE;
DROP TABLE IF EXISTS requests CASCADE;
DROP TABLE IF EXISTS asset_photos CASCADE;
DROP TABLE IF EXISTS photos CASCADE;
DROP TABLE IF EXISTS assets CASCADE;
DROP TABLE IF EXISTS categories CASCADE;

-- 1. Категории: L1 общие (organization_id NULL), L2 организации, L3 пользователя
CREATE TABLE categories (
    uuid                  UUID PRIMARY KEY,
    organization_id       UUID,
    name                  VARCHAR(255) NOT NULL,
    normalized_name       VARCHAR(255) NOT NULL,
    level                 SMALLINT NOT NULL CHECK (level IN (1, 2, 3)),
    status                VARCHAR(20) NOT NULL DEFAULT 'approved'
                          CHECK (status IN ('approved', 'pending', 'rejected', 'merged')),
    created_by            UUID,
    created_by_person_id  UUID,
    merged_into_uuid      UUID REFERENCES categories(uuid),
    usage_count           INTEGER DEFAULT 0,
    admin_comment         TEXT,
    created_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at            TIMESTAMP WITH TIME ZONE,
    CHECK ((level = 1) = (organization_id IS NULL))
);
CREATE INDEX idx_categories_org_level ON categories(organization_id, level);
CREATE INDEX idx_categories_status ON categories(status);
CREATE INDEX idx_categories_created_by ON categories(created_by);
CREATE INDEX idx_categories_deleted_at ON categories(deleted_at);
CREATE INDEX idx_categories_name_trgm ON categories USING gin (normalized_name gin_trgm_ops);
CREATE UNIQUE INDEX idx_categories_unique_global
    ON categories(normalized_name) WHERE level = 1 AND deleted_at IS NULL;
CREATE UNIQUE INDEX idx_categories_unique_local
    ON categories(organization_id, normalized_name) WHERE level = 2 AND deleted_at IS NULL;
CREATE UNIQUE INDEX idx_categories_unique_user
    ON categories(organization_id, created_by, normalized_name) WHERE level = 3 AND deleted_at IS NULL;

INSERT INTO categories (uuid, name, normalized_name, level, status, usage_count)
VALUES
    (gen_random_uuid(), 'Системный блок', 'системный блок', 1, 'approved', 0),
    (gen_random_uuid(), 'Монитор',        'монитор',        1, 'approved', 0),
    (gen_random_uuid(), 'Клавиатура',     'клавиатура',     1, 'approved', 0),
    (gen_random_uuid(), 'Мышь',           'мышь',           1, 'approved', 0),
    (gen_random_uuid(), 'Телефон',        'телефон',        1, 'approved', 0)
ON CONFLICT DO NOTHING;

-- 2. Активы. Ответственность — в asset_roles, здесь только кто завёл карточку
CREATE TABLE assets (
    uuid                   UUID PRIMARY KEY,
    organization_id        UUID NOT NULL,
    client_id              UUID,
    inventory_num          VARCHAR(255),
    name                   VARCHAR(255) NOT NULL,
    category               VARCHAR(255) NOT NULL,
    category_uuid          UUID REFERENCES categories(uuid),
    description            TEXT,
    created_by_account_id  UUID NOT NULL,
    created_by_person_id   UUID NOT NULL,
    accounting_status      VARCHAR(20) NOT NULL DEFAULT 'in_use'
                           CHECK (accounting_status IN ('in_use', 'in_stock', 'repair', 'lost', 'decommissioned')),
    condition              VARCHAR(20) NOT NULL DEFAULT 'ok'
                           CHECK (condition IN ('ok', 'needs_attention', 'faulty', 'unusable')),
    moderation_status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    verified_at            TIMESTAMP WITH TIME ZONE,
    admin_comment          TEXT,
    created_at             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at             TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX idx_assets_client_id ON assets(client_id);
-- инвентарный номер уникален в пределах организации
CREATE UNIQUE INDEX uni_assets_org_inventory_num
    ON assets(organization_id, inventory_num) WHERE inventory_num IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_assets_org ON assets(organization_id, deleted_at);
CREATE INDEX idx_assets_created_by_person ON assets(created_by_person_id);
CREATE INDEX idx_assets_category_uuid ON assets(category_uuid);
CREATE INDEX idx_assets_deleted_at ON assets(deleted_at);

-- 3. Фото
CREATE TABLE photos (
    uuid             UUID PRIMARY KEY,
    organization_id  UUID NOT NULL,
    client_id        UUID,
    account_id       UUID NOT NULL,
    person_id        UUID NOT NULL,
    url              VARCHAR(255) NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'pending',
    file_size        BIGINT NOT NULL,
    file_name        VARCHAR(255) NOT NULL,
    content_type     VARCHAR(255) NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at       TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX idx_photos_client_id ON photos(client_id);
CREATE INDEX idx_photos_org ON photos(organization_id);
CREATE INDEX idx_photos_account ON photos(account_id);

CREATE TABLE asset_photos (
    asset_uuid UUID NOT NULL REFERENCES assets(uuid) ON DELETE CASCADE,
    photo_uuid UUID NOT NULL REFERENCES photos(uuid) ON DELETE CASCADE,
    PRIMARY KEY (asset_uuid, photo_uuid)
);

-- 4. Заявки на ремонт
CREATE TABLE requests (
    uuid                     UUID PRIMARY KEY,
    organization_id          UUID NOT NULL,
    asset_id                 UUID NOT NULL REFERENCES assets(uuid) ON UPDATE CASCADE ON DELETE CASCADE,
    account_id               UUID NOT NULL,
    person_id                UUID NOT NULL,
    message                  TEXT NOT NULL,
    photo_url                VARCHAR(255),
    status                   VARCHAR(20) NOT NULL DEFAULT 'new'
                             CHECK (status IN ('new', 'in_repair', 'done', 'rejected')),
    admin_comment            TEXT,
    processed_by_account_id  UUID,
    processed_by_person_id   UUID,
    processed_at             TIMESTAMP WITH TIME ZONE,
    created_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at               TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_requests_org_status ON requests(organization_id, status);
CREATE INDEX idx_requests_asset_id ON requests(asset_id);
CREATE INDEX idx_requests_person ON requests(person_id);

-- 5. Заявки на изменение актива
CREATE TABLE asset_change_requests (
    uuid                     UUID PRIMARY KEY,
    organization_id          UUID NOT NULL,
    client_id                UUID,
    asset_uuid               UUID NOT NULL REFERENCES assets(uuid) ON DELETE CASCADE,
    account_uuid             UUID NOT NULL,
    person_uuid              UUID NOT NULL,
    request_type             VARCHAR(20) NOT NULL,
    proposed_data            JSONB,
    reason                   TEXT NOT NULL,
    admin_comment            TEXT,
    status                   VARCHAR(20) DEFAULT 'pending',
    processed_by_account_id  UUID,
    processed_by_person_id   UUID,
    created_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX idx_asset_change_requests_client_id ON asset_change_requests(client_id);
CREATE INDEX idx_asset_change_requests_org_status ON asset_change_requests(organization_id, status);
CREATE INDEX idx_asset_change_requests_asset_uuid ON asset_change_requests(asset_uuid);

-- 6. Передачи — запись истории: применяются сразу, без ожидания (3.3)
CREATE TABLE asset_transfers (
    id                       UUID PRIMARY KEY,
    organization_id          UUID NOT NULL,
    asset_id                 UUID NOT NULL REFERENCES assets(uuid) ON DELETE CASCADE,
    role                     VARCHAR(10) NOT NULL CHECK (role IN ('mol', 'user')),
    from_person_id           UUID,
    to_person_id             UUID,
    -- NULL — системное действие (увольнение, списание)
    initiated_by_account_id  UUID,
    initiated_by_person_id   UUID,
    reason                   VARCHAR(30),
    comment                  TEXT,
    created_at               TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_asset_transfers_asset ON asset_transfers(asset_id, created_at);

-- 7. Бизнес-роли: «числится за» (mol) и «работает с активом» (user). Закрываются, а не удаляются (3.2)
CREATE TABLE asset_roles (
    id                               UUID PRIMARY KEY,
    organization_id                  UUID NOT NULL,
    asset_id                         UUID NOT NULL REFERENCES assets(uuid) ON DELETE CASCADE,
    person_id                        UUID NOT NULL,
    role                             VARCHAR(10) NOT NULL CHECK (role IN ('mol', 'user')),
    started_at                       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ended_at                         TIMESTAMP WITH TIME ZONE,
    end_reason                       VARCHAR(30),
    assigned_by_account_id           UUID,
    assigned_by_person_id            UUID,
    transfer_id                      UUID REFERENCES asset_transfers(id) ON DELETE SET NULL,
    ack_status                       VARCHAR(10) NOT NULL DEFAULT 'pending'
                                     CHECK (ack_status IN ('pending', 'accepted', 'disputed')),
    ack_reason                       TEXT,
    ack_at                           TIMESTAMP WITH TIME ZONE,
    dispute_resolved_at              TIMESTAMP WITH TIME ZONE,
    dispute_resolved_by_account_id   UUID,
    dispute_resolved_by_person_id    UUID,
    dispute_resolution               TEXT
);
-- у актива не больше одного действующего МОЛ
CREATE UNIQUE INDEX uni_asset_roles_one_mol ON asset_roles(asset_id) WHERE ended_at IS NULL AND role = 'mol';
CREATE UNIQUE INDEX uni_asset_roles_user ON asset_roles(asset_id, person_id, role) WHERE ended_at IS NULL;
CREATE INDEX idx_asset_roles_person ON asset_roles(person_id, ended_at);
CREATE INDEX idx_asset_roles_asset ON asset_roles(asset_id, ended_at);
CREATE INDEX idx_asset_roles_disputes ON asset_roles(organization_id) WHERE ack_status = 'disputed' AND dispute_resolved_at IS NULL;

-- 8. Периоды ремонта (3.6)
CREATE TABLE asset_repairs (
    id                      UUID PRIMARY KEY,
    organization_id         UUID NOT NULL,
    asset_id                UUID NOT NULL REFERENCES assets(uuid) ON DELETE CASCADE,
    request_id              UUID REFERENCES requests(uuid) ON DELETE SET NULL,
    kind                    VARCHAR(20) NOT NULL CHECK (kind IN ('repair', 'maintenance')),
    started_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expected_return_at      TIMESTAMP WITH TIME ZONE,
    returned_at             TIMESTAMP WITH TIME ZONE,
    status_before           VARCHAR(20) NOT NULL,
    result_condition        VARCHAR(20) CHECK (result_condition IN ('ok', 'faulty', 'unusable')),
    work_done               TEXT,
    comment                 TEXT,
    created_by_account_id   UUID NOT NULL,
    created_by_person_id    UUID NOT NULL,
    returned_by_account_id  UUID,
    returned_by_person_id   UUID
);
CREATE UNIQUE INDEX uni_asset_repairs_open ON asset_repairs(asset_id) WHERE returned_at IS NULL;
CREATE INDEX idx_asset_repairs_org ON asset_repairs(organization_id, returned_at);

-- 9. История статусов
CREATE TABLE asset_status_history (
    id          UUID PRIMARY KEY,
    organization_id UUID NOT NULL,
    asset_id    UUID NOT NULL REFERENCES assets(uuid) ON DELETE CASCADE,
    field       VARCHAR(20) NOT NULL CHECK (field IN ('accounting_status', 'condition')),
    from_value  VARCHAR(20),
    to_value    VARCHAR(20) NOT NULL,
    account_id  UUID,
    person_id   UUID,
    repair_id   UUID REFERENCES asset_repairs(id) ON DELETE SET NULL,
    comment     TEXT,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_asset_status_history_asset ON asset_status_history(asset_id, created_at);

-- 10. Журнал go (3.4): человек — в момент действия, из токена
CREATE TABLE audit_log (
    id               UUID PRIMARY KEY,
    organization_id  UUID NOT NULL,
    actor_type       VARCHAR(10) NOT NULL DEFAULT 'user' CHECK (actor_type IN ('user', 'system')),
    account_id       UUID,
    actor_person_id  UUID,
    action           VARCHAR(50) NOT NULL,
    object_type      VARCHAR(30),
    object_id        UUID,
    details          JSONB,
    ip               VARCHAR(64),
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (actor_type = 'system' OR (account_id IS NOT NULL AND actor_person_id IS NOT NULL))
);
CREATE INDEX idx_audit_log_org_created ON audit_log(organization_id, created_at DESC);
CREATE INDEX idx_audit_log_object ON audit_log(object_id);

-- 11. Что go уже сообщил ts об организации (5.1): «наполнена» и последняя активность
CREATE TABLE organization_state (
    organization_id     UUID PRIMARY KEY,
    filled_reported_at  TIMESTAMP WITH TIME ZONE,
    touch_reported_at   TIMESTAMP WITH TIME ZONE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS organization_state;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS asset_status_history;
DROP TABLE IF EXISTS asset_repairs;
DROP TABLE IF EXISTS asset_roles;
DROP TABLE IF EXISTS asset_transfers;
DROP TABLE IF EXISTS asset_change_requests;
DROP TABLE IF EXISTS requests;
DROP TABLE IF EXISTS asset_photos;
DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS categories;

-- +goose StatementEnd
