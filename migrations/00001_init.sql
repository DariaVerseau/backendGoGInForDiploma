-- +goose Up
CREATE TABLE users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE images (
    id          UUID PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    style       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_images_user_id_created_at
    ON images(user_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_images_user_id_created_at;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS users;