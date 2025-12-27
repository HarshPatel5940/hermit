-- Add normalized_url column to pages table for better duplicate detection
ALTER TABLE pages ADD COLUMN IF NOT EXISTS normalized_url VARCHAR(2048);

-- Create index on normalized_url for faster lookups
CREATE INDEX IF NOT EXISTS idx_pages_normalized_url ON pages(normalized_url);

-- Add unique constraint on website_id + normalized_url combination
-- This prevents duplicate pages with same normalized URL
ALTER TABLE pages ADD CONSTRAINT unique_website_normalized_url
    UNIQUE (website_id, normalized_url);

-- Update existing rows to have normalized URLs (same as original URL for now)
-- In production, you might want to run a separate script to properly normalize existing URLs
UPDATE pages SET normalized_url = url WHERE normalized_url IS NULL;

-- Make normalized_url NOT NULL after populating existing rows
ALTER TABLE pages ALTER COLUMN normalized_url SET NOT NULL;
