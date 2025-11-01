-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE posts ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE posts ADD COLUMN version INT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE posts DROP CONSTRAINT fk_user;
ALTER TABLE posts DROP COLUMN version;
-- +goose StatementEnd
