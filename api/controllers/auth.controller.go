package controllers

import (
	"net/http"

	"hermit/api/middlewares"
	"hermit/internal/auth"
	"hermit/internal/schema"

	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
)

// AuthController handles authentication endpoints
type AuthController struct {
	authService *auth.Service
}

// NewAuthController creates a new auth controller
func NewAuthController(authService *auth.Service) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (ctrl *AuthController) Register(c echo.Context) error {
	var req schema.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "email and password are required",
		})
	}

	// Register user
	user, err := ctrl.authService.Register(req.Email, req.Password)
	if err != nil {
		if err.Error() == "email already registered" {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": err.Error(),
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to register user",
		})
	}

	// Create default API key for the user
	_, plainKey, err := ctrl.authService.CreateAPIKey(
		user.ID,
		"Default API Key",
		[]string{},
		nil,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "user created but failed to generate API key",
		})
	}

	return c.JSON(http.StatusCreated, schema.LoginResponse{
		User:    user,
		APIKey:  plainKey,
		Message: "User registered successfully. Save your API key, it won't be shown again.",
	})
}

// Login handles user login
// POST /api/v1/auth/login
func (ctrl *AuthController) Login(c echo.Context) error {
	var req schema.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "email and password are required",
		})
	}

	// Login user
	user, err := ctrl.authService.Login(req.Email, req.Password)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "invalid credentials",
		})
	}

	// Create a new session API key
	_, plainKey, err := ctrl.authService.CreateAPIKey(
		user.ID,
		"Session Key",
		[]string{},
		nil,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "login successful but failed to generate API key",
		})
	}

	return c.JSON(http.StatusOK, schema.LoginResponse{
		User:    user,
		APIKey:  plainKey,
		Message: "Login successful. Save your API key, it won't be shown again.",
	})
}

// GetMe returns the authenticated user's information
// GET /api/v1/auth/me
func (ctrl *AuthController) GetMe(c echo.Context) error {
	user := middlewares.GetUser(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
	}

	return c.JSON(http.StatusOK, user.ToResponse())
}

// CreateAPIKey creates a new API key for the authenticated user
// POST /api/v1/auth/api-keys
func (ctrl *AuthController) CreateAPIKey(c echo.Context) error {
	userID, err := middlewares.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
	}

	var req schema.CreateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Validate request
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
	}

	// Create API key
	apiKey, plainKey, err := ctrl.authService.CreateAPIKey(
		userID,
		req.Name,
		req.Scopes,
		req.ExpiresAt,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to create API key",
		})
	}

	return c.JSON(http.StatusCreated, schema.CreateAPIKeyResponse{
		APIKey:   apiKey,
		PlainKey: plainKey,
		Message:  "API key created successfully. Save it securely, it won't be shown again.",
	})
}

// ListAPIKeys returns all API keys for the authenticated user
// GET /api/v1/auth/api-keys
func (ctrl *AuthController) ListAPIKeys(c echo.Context) error {
	userID, err := middlewares.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
	}

	apiKeys, err := ctrl.authService.GetUserAPIKeys(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve API keys",
		})
	}

	// Convert to response format
	var responses []*schema.APIKeyResponse
	for _, key := range apiKeys {
		responses = append(responses, key.ToResponse())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"api_keys": responses,
		"count":    len(responses),
	})
}

// GetAPIKey returns a specific API key by ID
// GET /api/v1/auth/api-keys/:id
func (ctrl *AuthController) GetAPIKey(c echo.Context) error {
	userID, err := middlewares.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
	}

	keyID, err := ulid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid API key ID",
		})
	}

	// Get all user's API keys and find the matching one
	apiKeys, err := ctrl.authService.GetUserAPIKeys(userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to retrieve API key",
		})
	}

	for _, key := range apiKeys {
		if key.ID == keyID {
			return c.JSON(http.StatusOK, key.ToResponse())
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{
		"error": "API key not found",
	})
}

// UpdateAPIKey updates an API key
// PUT /api/v1/auth/api-keys/:id
func (ctrl *AuthController) UpdateAPIKey(c echo.Context) error {
	userID, err := middlewares.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
	}

	keyID, err := ulid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid API key ID",
		})
	}

	var req schema.UpdateAPIKeyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// Update API key
	apiKey, err := ctrl.authService.UpdateAPIKey(
		keyID,
		userID,
		req.Name,
		req.Scopes,
		req.IsActive,
		req.ExpiresAt,
	)
	if err != nil {
		if err.Error() == "unauthorized" {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "you don't have permission to update this API key",
			})
		}
		if err.Error() == "API key not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "API key not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to update API key",
		})
	}

	return c.JSON(http.StatusOK, apiKey.ToResponse())
}

// RevokeAPIKey revokes (deletes) an API key
// DELETE /api/v1/auth/api-keys/:id
func (ctrl *AuthController) RevokeAPIKey(c echo.Context) error {
	userID, err := middlewares.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "authentication required",
		})
	}

	keyID, err := ulid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid API key ID",
		})
	}

	// Revoke API key
	err = ctrl.authService.RevokeAPIKey(keyID, userID)
	if err != nil {
		if err.Error() == "unauthorized" {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "you don't have permission to revoke this API key",
			})
		}
		if err.Error() == "API key not found" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "API key not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to revoke API key",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "API key revoked successfully",
	})
}
