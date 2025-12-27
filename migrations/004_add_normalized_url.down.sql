-- Remove unique constraint
ALTER TABLE pages DROP CONSTRAINT IF EXISTS unique_website_normalized_url;

-- Drop index
DROP INDEX IF EXISTS idx_pages_normalized_url;

-- Remove normalized_url column
ALTER TABLE pages DROP COLUMN IF EXISTS normalized_url;
