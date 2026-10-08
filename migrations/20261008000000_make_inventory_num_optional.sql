-- +goose Up
-- +goose StatementBegin
ALTER TABLE assets ALTER COLUMN inventory_num DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE assets ALTER COLUMN inventory_num SET NOT NULL;
-- +goose StatementEnd
