package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"hermit/internal/repositories"
	"hermit/internal/schema"

	"github.com/oklog/ulid/v2"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication operations
type Service struct {
	userRepo   *repositories.UserRepository
	apiKeyRepo *repositories.APIKeyRepository
}

// NewService creates a new auth service
func NewService(userRepo *repositories.UserRepository, apiKeyRepo *repositories.APIKeyRepository) *Service {
	return &Service{
		userRepo:   userRepo,
		apiKeyRepo: apiKeyRepo,
	}
}

// Register creates a new user account
func (s *Service) Register(email, password string) (*schema.User, error) {
	// Check if email already exists
	exists, err := s.userRepo.EmailExists(context.TODO(), email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &schema.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         schema.RoleUser,
		IsActive:     true,
		WebsiteLimit: 10,
	}

	err = s.userRepo.Create(context.TODO(), user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns the user object
func (s *Service) Login(email, password string) (*schema.User, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(context.TODO(), email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("account is inactive")
	}

	// Verify password
	if !s.VerifyPassword(password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

// CreateAPIKey generates a new API key for a user
func (s *Service) CreateAPIKey(userID ulid.ULID, name string, scopes []string, expiresAt *time.Time) (*schema.APIKey, string, error) {
	// Generate random API key
	plainKey, err := s.GenerateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash the API key
	keyHash := s.HashAPIKey(plainKey)

	// Get key prefix (first 8 characters)
	keyPrefix := plainKey[:8]

	// Create API key record
	apiKey := &schema.APIKey{
		UserID:    userID,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		Name:      name,
		Scopes:    scopes,
		IsActive:  true,
		ExpiresAt: expiresAt,
	}

	err = s.apiKeyRepo.Create(context.TODO(), apiKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create API key: %w", err)
	}

	return apiKey, plainKey, nil
}

// ValidateAPIKey validates an API key and returns the associated user
func (s *Service) ValidateAPIKey(plainKey string) (*schema.User, *schema.APIKey, error) {
	// Hash the provided key
	keyHash := s.HashAPIKey(plainKey)

	// Get API key from database
	apiKey, err := s.apiKeyRepo.GetByKeyHash(context.TODO(), keyHash)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid API key")
	}

	// Check if key is valid (active and not expired)
	if !apiKey.IsValid() {
		return nil, nil, fmt.Errorf("API key is invalid or expired")
	}

	// Get associated user
	user, err := s.userRepo.GetByID(context.TODO(), apiKey.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("user not found")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, nil, fmt.Errorf("user account is inactive")
	}

	// Update last used timestamp (async, don't block)
	go s.apiKeyRepo.UpdateLastUsed(context.TODO(), apiKey.ID)

	return user, apiKey, nil
}

// GetUserAPIKeys retrieves all API keys for a user
func (s *Service) GetUserAPIKeys(userID ulid.ULID) ([]*schema.APIKey, error) {
	return s.apiKeyRepo.GetByUserID(context.TODO(), userID)
}

// RevokeAPIKey revokes (deletes) an API key
func (s *Service) RevokeAPIKey(keyID, userID ulid.ULID) error {
	// Get the API key to verify ownership
	apiKey, err := s.apiKeyRepo.GetByID(context.TODO(), keyID)
	if err != nil {
		return fmt.Errorf("API key not found")
	}

	// Verify the key belongs to the user
	if apiKey.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	// Delete the key
	return s.apiKeyRepo.Delete(context.TODO(), keyID)
}

// UpdateAPIKey updates an API key
func (s *Service) UpdateAPIKey(keyID, userID ulid.ULID, name *string, scopes []string, isActive *bool, expiresAt *time.Time) (*schema.APIKey, error) {
	// Get the API key to verify ownership
	apiKey, err := s.apiKeyRepo.GetByID(context.TODO(), keyID)
	if err != nil {
		return nil, fmt.Errorf("API key not found")
	}

	// Verify the key belongs to the user
	if apiKey.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Update fields
	if name != nil {
		apiKey.Name = *name
	}
	if scopes != nil {
		apiKey.Scopes = scopes
	}
	if isActive != nil {
		apiKey.IsActive = *isActive
	}
	if expiresAt != nil {
		apiKey.ExpiresAt = expiresAt
	}

	// Save changes
	err = s.apiKeyRepo.Update(context.TODO(), apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to update API key: %w", err)
	}

	return apiKey, nil
}

// HashPassword hashes a password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against a hash
func (s *Service) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateAPIKey generates a random API key
func (s *Service) GenerateAPIKey() (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode to base64 and add prefix
	key := base64.URLEncoding.EncodeToString(b)
	// Remove padding
	key = strings.TrimRight(key, "=")

	return fmt.Sprintf("hmt_%s", key), nil
}

// HashAPIKey hashes an API key using SHA256
func (s *Service) HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// CleanupExpiredAPIKeys removes expired API keys
func (s *Service) CleanupExpiredAPIKeys() (int64, error) {
	return s.apiKeyRepo.CleanupExpired(context.TODO())
}
