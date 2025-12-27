package repositories

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"time"

	"hermit/internal/schema"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
)

// APIKeyRepository handles database operations for API keys
type APIKeyRepository struct {
	db *sqlx.DB
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *sqlx.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

// Create creates a new API key
func (r *APIKeyRepository) Create(ctx context.Context, apiKey *schema.APIKey) error {
	query := `
		INSERT INTO api_keys (id, user_id, key_hash, key_prefix, name, scopes, is_active, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	// Generate ULID
	entropy := ulid.Monotonic(rand.Reader, 0)
	apiKey.ID = ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		apiKey.ID.String(),
		apiKey.UserID.String(),
		apiKey.KeyHash,
		apiKey.KeyPrefix,
		apiKey.Name,
		apiKey.Scopes,
		apiKey.IsActive,
		apiKey.ExpiresAt,
		apiKey.CreatedAt,
		apiKey.UpdatedAt,
	).Scan(&apiKey.ID, &apiKey.CreatedAt, &apiKey.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetByID retrieves an API key by ID
func (r *APIKeyRepository) GetByID(ctx context.Context, id ulid.ULID) (*schema.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, key_prefix, name, scopes, is_active, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE id = $1
	`

	var apiKey schema.APIKey
	err := r.db.GetContext(ctx, &apiKey, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &apiKey, nil
}

// GetByKeyHash retrieves an API key by its hash
func (r *APIKeyRepository) GetByKeyHash(ctx context.Context, keyHash string) (*schema.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, key_prefix, name, scopes, is_active, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE key_hash = $1
	`

	var apiKey schema.APIKey
	err := r.db.GetContext(ctx, &apiKey, query, keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("API key not found")
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &apiKey, nil
}

// GetByUserID retrieves all API keys for a user
func (r *APIKeyRepository) GetByUserID(ctx context.Context, userID ulid.ULID) ([]*schema.APIKey, error) {
	query := `
		SELECT id, user_id, key_hash, key_prefix, name, scopes, is_active, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	var apiKeys []*schema.APIKey
	err := r.db.SelectContext(ctx, &apiKeys, query, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}

	return apiKeys, nil
}

// Update updates an API key
func (r *APIKeyRepository) Update(ctx context.Context, apiKey *schema.APIKey) error {
	query := `
		UPDATE api_keys
		SET name = $2, scopes = $3, is_active = $4, expires_at = $5, updated_at = $6
		WHERE id = $1
		RETURNING updated_at
	`

	apiKey.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		apiKey.ID.String(),
		apiKey.Name,
		apiKey.Scopes,
		apiKey.IsActive,
		apiKey.ExpiresAt,
		apiKey.UpdatedAt,
	).Scan(&apiKey.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("API key not found")
		}
		return fmt.Errorf("failed to update API key: %w", err)
	}

	return nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id ulid.ULID) error {
	query := `
		UPDATE api_keys
		SET last_used_at = $2
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id.String(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to update last used timestamp: %w", err)
	}

	return nil
}

// Delete deletes an API key by ID
func (r *APIKeyRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM api_keys WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// DeleteByUserID deletes all API keys for a user
func (r *APIKeyRepository) DeleteByUserID(ctx context.Context, userID ulid.ULID) error {
	query := `DELETE FROM api_keys WHERE user_id = $1`

	_, err := r.db.ExecContext(ctx, query, userID.String())
	if err != nil {
		return fmt.Errorf("failed to delete API keys: %w", err)
	}

	return nil
}

// List retrieves all API keys with pagination
func (r *APIKeyRepository) List(ctx context.Context, page, limit int) ([]*schema.APIKey, int, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM api_keys`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count API keys: %w", err)
	}

	// Get API keys
	query := `
		SELECT id, user_id, key_hash, key_prefix, name, scopes, is_active, last_used_at, expires_at, created_at, updated_at
		FROM api_keys
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var apiKeys []*schema.APIKey
	err = r.db.SelectContext(ctx, &apiKeys, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list API keys: %w", err)
	}

	return apiKeys, total, nil
}

// CleanupExpired deletes expired API keys
func (r *APIKeyRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM api_keys WHERE expires_at IS NOT NULL AND expires_at < $1`

	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired API keys: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
