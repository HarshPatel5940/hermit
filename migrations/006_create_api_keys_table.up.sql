-- Create api_keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id VARCHAR(26) PRIMARY KEY,
    user_id VARCHAR(26) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    name VARCHAR(255) NOT NULL,
    scopes TEXT[] DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on user_id for faster lookups
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

-- Create index on key_hash for authentication lookups
CREATE INDEX idx_api_keys_key_hash ON api_keys(key_hash);

-- Create index on key_prefix for partial key display
CREATE INDEX idx_api_keys_key_prefix ON api_keys(key_prefix);

-- Create index on is_active for filtering active keys
CREATE INDEX idx_api_keys_is_active ON api_keys(is_active);
