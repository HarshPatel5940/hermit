-- +goose Up
-- Add user_id column to websites table
ALTER TABLE websites ADD COLUMN user_id VARCHAR(26) REFERENCES users(id) ON DELETE CASCADE;

-- Create index on user_id for faster lookups
CREATE INDEX idx_websites_user_id ON websites(user_id);

-- +goose Down
-- Remove user_id column from websites table
DROP INDEX IF EXISTS idx_websites_user_id;
ALTER TABLE websites DROP COLUMN IF EXISTS user_id;
