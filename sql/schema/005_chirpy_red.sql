-- +goose Up
ALTER TABLE users
ADD COLUMN chirpy_red BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE users
DROP COLUMN chirpy_red;
