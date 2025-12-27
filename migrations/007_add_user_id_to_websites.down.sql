-- Remove user_id column from websites table
DROP INDEX IF EXISTS idx_websites_user_id;
ALTER TABLE websites DROP COLUMN IF EXISTS user_id;
