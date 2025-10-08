-- +goose Up
CREATE TABLE IF NOT EXISTS websites (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL UNIQUE,
    is_monitored BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS websites;
