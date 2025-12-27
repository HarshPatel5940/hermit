package schema

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// APIKey represents an API key for authentication
type APIKey struct {
	ID         ulid.ULID  `db:"id" json:"id"`
	UserID     ulid.ULID  `db:"user_id" json:"user_id"`
	KeyHash    string     `db:"key_hash" json:"-"` // Never send key hash to client
	KeyPrefix  string     `db:"key_prefix" json:"key_prefix"`
	Name       string     `db:"name" json:"name"`
	Scopes     []string   `db:"scopes" json:"scopes"`
	IsActive   bool       `db:"is_active" json:"is_active"`
	LastUsedAt *time.Time `db:"last_used_at" json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}

// CreateAPIKeyRequest represents the request to create a new API key
type CreateAPIKeyRequest struct {
	Name      string     `json:"name" validate:"required,min=3,max=255"`
	Scopes    []string   `json:"scopes,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// CreateAPIKeyResponse represents the response after creating an API key
type CreateAPIKeyResponse struct {
	APIKey   *APIKey `json:"api_key"`
	PlainKey string  `json:"plain_key"` // Only returned once during creation
	Message  string  `json:"message"`
}

// UpdateAPIKeyRequest represents the request to update an API key
type UpdateAPIKeyRequest struct {
	Name      *string    `json:"name,omitempty" validate:"omitempty,min=3,max=255"`
	Scopes    []string   `json:"scopes,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// APIKeyResponse represents API key data returned to client (without sensitive fields)
type APIKeyResponse struct {
	ID         ulid.ULID  `json:"id"`
	UserID     ulid.ULID  `json:"user_id"`
	KeyPrefix  string     `json:"key_prefix"`
	Name       string     `json:"name"`
	Scopes     []string   `json:"scopes"`
	IsActive   bool       `json:"is_active"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ToResponse converts APIKey to APIKeyResponse
func (k *APIKey) ToResponse() *APIKeyResponse {
	return &APIKeyResponse{
		ID:         k.ID,
		UserID:     k.UserID,
		KeyPrefix:  k.KeyPrefix,
		Name:       k.Name,
		Scopes:     k.Scopes,
		IsActive:   k.IsActive,
		LastUsedAt: k.LastUsedAt,
		ExpiresAt:  k.ExpiresAt,
		CreatedAt:  k.CreatedAt,
		UpdatedAt:  k.UpdatedAt,
	}
}

// IsExpired checks if the API key has expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsValid checks if the API key is active and not expired
func (k *APIKey) IsValid() bool {
	return k.IsActive && !k.IsExpired()
}

// HasScope checks if the API key has a specific scope
func (k *APIKey) HasScope(scope string) bool {
	// Empty scopes means full access
	if len(k.Scopes) == 0 {
		return true
	}

	for _, s := range k.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}
