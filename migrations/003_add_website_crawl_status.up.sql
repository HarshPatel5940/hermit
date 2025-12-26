-- +goose Up
ALTER TABLE websites
ADD COLUMN crawl_status VARCHAR(50) DEFAULT 'idle',
ADD COLUMN crawl_started_at TIMESTAMPTZ,
ADD COLUMN crawl_completed_at TIMESTAMPTZ,
ADD COLUMN total_pages_crawled INTEGER DEFAULT 0,
ADD COLUMN total_pages_failed INTEGER DEFAULT 0,
ADD COLUMN last_error TEXT;

CREATE INDEX idx_websites_crawl_status ON websites(crawl_status);

-- +goose Down
DROP INDEX IF EXISTS idx_websites_crawl_status;

ALTER TABLE websites
DROP COLUMN IF EXISTS crawl_status,
DROP COLUMN IF EXISTS crawl_started_at,
DROP COLUMN IF EXISTS crawl_completed_at,
DROP COLUMN IF EXISTS total_pages_crawled,
DROP COLUMN IF EXISTS total_pages_failed,
DROP COLUMN IF EXISTS last_error;
