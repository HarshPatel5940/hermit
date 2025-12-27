package schema

import (
	"time"

	"github.com/oklog/ulid/v2"
)

// User represents a user in the system
type User struct {
	ID           ulid.ULID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"` // Never send password hash to client
	Role         string    `db:"role" json:"role"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	WebsiteLimit int       `db:"website_limit" json:"website_limit"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// UserRole constants
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	User    *User  `json:"user"`
	APIKey  string `json:"api_key"`
	Message string `json:"message"`
}

// UpdateUserRequest represents the request to update user details
type UpdateUserRequest struct {
	Email        *string `json:"email,omitempty" validate:"omitempty,email"`
	Password     *string `json:"password,omitempty" validate:"omitempty,min=8"`
	Role         *string `json:"role,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	WebsiteLimit *int    `json:"website_limit,omitempty"`
}

// UserResponse represents user data returned to client (without sensitive fields)
type UserResponse struct {
	ID           ulid.ULID `json:"id"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	WebsiteLimit int       `json:"website_limit"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		Role:         u.Role,
		IsActive:     u.IsActive,
		WebsiteLimit: u.WebsiteLimit,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// CanCreateWebsite checks if user can create more websites
func (u *User) CanCreateWebsite(currentCount int) bool {
	return currentCount < u.WebsiteLimit
}
