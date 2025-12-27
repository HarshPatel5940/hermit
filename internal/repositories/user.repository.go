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

// UserRepository handles database operations for users
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *schema.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, is_active, website_limit, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	// Generate ULID
	entropy := ulid.Monotonic(rand.Reader, 0)
	user.ID = ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if user.Role == "" {
		user.Role = schema.RoleUser
	}

	if user.WebsiteLimit == 0 {
		user.WebsiteLimit = 10
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID.String(),
		user.Email,
		user.PasswordHash,
		user.Role,
		user.IsActive,
		user.WebsiteLimit,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id ulid.ULID) (*schema.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_active, website_limit, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user schema.User
	err := r.db.GetContext(ctx, &user, query, id.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*schema.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_active, website_limit, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user schema.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *schema.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, role = $4, is_active = $5, website_limit = $6, updated_at = $7
		WHERE id = $1
		RETURNING updated_at
	`

	user.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID.String(),
		user.Email,
		user.PasswordHash,
		user.Role,
		user.IsActive,
		user.WebsiteLimit,
		user.UpdatedAt,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id ulid.ULID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves all users with pagination
func (r *UserRepository) List(ctx context.Context, page, limit int) ([]*schema.User, int, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.db.GetContext(ctx, &total, countQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	query := `
		SELECT id, email, password_hash, role, is_active, website_limit, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []*schema.User
	err = r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// GetWebsiteCount gets the count of websites for a user
func (r *UserRepository) GetWebsiteCount(ctx context.Context, userID ulid.ULID) (int, error) {
	query := `SELECT COUNT(*) FROM websites WHERE user_id = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID.String())
	if err != nil {
		return 0, fmt.Errorf("failed to count websites: %w", err)
	}

	return count, nil
}

// EmailExists checks if an email is already registered
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}
