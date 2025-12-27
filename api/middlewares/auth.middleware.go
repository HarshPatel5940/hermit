package middlewares

import (
	"context"
	"net/http"
	"strings"

	"hermit/internal/auth"
	"hermit/internal/schema"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

// ContextKey type for context keys
type ContextKey string

const (
	// UserContextKey is the context key for the authenticated user
	UserContextKey ContextKey = "user"
	// APIKeyContextKey is the context key for the API key
	APIKeyContextKey ContextKey = "api_key"
)

// AuthMiddleware creates a middleware that validates API keys
func AuthMiddleware(authService *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get API key from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "missing authorization header",
				})
			}

			// Expected format: "Bearer hmt_xxxxx"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid authorization header format",
				})
			}

			apiKey := parts[1]

			// Validate API key
			user, key, err := authService.ValidateAPIKey(apiKey)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "invalid or expired API key",
				})
			}

			// Store user and API key in context
			ctx := context.WithValue(c.Request().Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, APIKeyContextKey, key)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// OptionalAuthMiddleware creates a middleware that validates API keys but doesn't require them
func OptionalAuthMiddleware(authService *auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get API key from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				// No auth header, continue without user context
				return next(c)
			}

			// Expected format: "Bearer hmt_xxxxx"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Invalid format, continue without user context
				return next(c)
			}

			apiKey := parts[1]

			// Validate API key
			user, key, err := authService.ValidateAPIKey(apiKey)
			if err != nil {
				// Invalid key, continue without user context
				return next(c)
			}

			// Store user and API key in context
			ctx := context.WithValue(c.Request().Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, APIKeyContextKey, key)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// RequireRole creates a middleware that checks if user has required role
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := GetUser(c)
			if user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			// Check if user has any of the required roles
			for _, role := range roles {
				if user.Role == role {
					return next(c)
				}
			}

			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "insufficient permissions",
			})
		}
	}
}

// RequireScope creates a middleware that checks if API key has required scope
func RequireScope(scope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := GetAPIKey(c)
			if apiKey == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			if !apiKey.HasScope(scope) {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "insufficient scope",
				})
			}

			return next(c)
		}
	}
}

// GetUser retrieves the authenticated user from context
func GetUser(c echo.Context) *schema.User {
	user := c.Request().Context().Value(UserContextKey)
	if user == nil {
		return nil
	}
	return user.(*schema.User)
}

// GetAPIKey retrieves the API key from context
func GetAPIKey(c echo.Context) *schema.APIKey {
	apiKey := c.Request().Context().Value(APIKeyContextKey)
	if apiKey == nil {
		return nil
	}
	return apiKey.(*schema.APIKey)
}

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c echo.Context) (ulid.ULID, error) {
	user := GetUser(c)
	if user == nil {
		return ulid.ULID{}, echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}
	return user.ID, nil
}

// IsAdmin checks if the authenticated user is an admin
func IsAdmin(c echo.Context) bool {
	user := GetUser(c)
	if user == nil {
		return false
	}
	return user.IsAdmin()
}
