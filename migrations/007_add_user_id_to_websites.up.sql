-- Add user_id column to websites table
ALTER TABLE websites ADD COLUMN user_id VARCHAR(26) REFERENCES users(id) ON DELETE CASCADE;

-- Create index on user_id for faster lookups
CREATE INDEX idx_websites_user_id ON websites(user_id);

-- Make user_id NOT NULL for new records (existing records can be migrated separately)
-- For now, allow NULL to support existing data
