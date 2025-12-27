-- +goose Up
CREATE TABLE IF NOT EXISTS pages (
    id SERIAL PRIMARY KEY,
    website_id INTEGER NOT NULL REFERENCES websites(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    minio_object_key TEXT,
    content_hash TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    error_message TEXT,
    crawled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(website_id, url)
);

CREATE INDEX idx_pages_website_id ON pages(website_id);
CREATE INDEX idx_pages_status ON pages(status);
CREATE INDEX idx_pages_url ON pages(url);

-- +goose Down
DROP TABLE IF EXISTS pages;
