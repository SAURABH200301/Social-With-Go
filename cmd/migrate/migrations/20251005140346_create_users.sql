-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "citext";

SELECT 'up SQL query';
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(50) UNIQUE NOT NULL,
        email CITEXT UNIQUE NOT NULL,
        password bytea NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
    DROP TABLE IF EXISTS users;
-- +goose StatementEnd
