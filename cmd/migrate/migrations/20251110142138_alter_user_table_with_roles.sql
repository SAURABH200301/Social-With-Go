-- +goose Up
-- +goose StatementBegin

ALTER TABLE IF EXISTS users
ADD COLUMN role_id BIGSERIAL REFERENCES roles(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin


ALTER TABLE IF EXISTS users
DROP COLUMN role_id;

-- +goose StatementEnd
