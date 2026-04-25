-- +goose Up
ALTER TABLE users ADD COLUMN jwt_token TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN jwt_token;
