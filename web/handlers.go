package web

import (
	"net/http"
	"time"

	"hermit/internal/auth"
	"hermit/internal/repositories"
	"hermit/internal/schema"

	"github.com/labstack/echo/v4"
)

const (
	sessionCookieName = "hermit_session"
	sessionMaxAge     = 7 * 24 * 60 * 60 // 7 days
)

// Handlers holds all dependencies for web handlers
type Handlers struct {
	authService *auth.Service
	websiteRepo *repositories.WebsiteRepository
	apiKeyRepo  *repositories.APIKeyRepository
	userRepo    *repositories.UserRepository
}

// NewHandlers creates a new web handlers instance
func NewHandlers(
	authService *auth.Service,
	websiteRepo *repositories.WebsiteRepository,
	apiKeyRepo *repositories.APIKeyRepository,
	userRepo *repositories.UserRepository,
) *Handlers {
	return &Handlers{
		authService: authService,
		websiteRepo: websiteRepo,
		apiKeyRepo:  apiKeyRepo,
		userRepo:    userRepo,
	}
}

// getUserFromSession extracts user from session cookie
func (h *Handlers) getUserFromSession(c echo.Context) (*schema.User, error) {
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil {
		return nil, err
	}

	// Validate API key from cookie
	user, _, err := h.authService.ValidateAPIKey(cookie.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// setSessionCookie sets a session cookie with the API key
func (h *Handlers) setSessionCookie(c echo.Context, apiKey string) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    apiKey,
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}

// clearSessionCookie removes the session cookie
func (h *Handlers) clearSessionCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	}
	c.SetCookie(cookie)
}

// ShowLogin displays the login page
func (h *Handlers) ShowLogin(c echo.Context) error {
	// Check if already logged in
	if _, err := h.getUserFromSession(c); err == nil {
		return c.Redirect(http.StatusFound, "/chat")
	}
	return Login().Render(c.Request().Context(), c.Response().Writer)
}

// ShowRegister displays the registration page
func (h *Handlers) ShowRegister(c echo.Context) error {
	// Check if already logged in
	if _, err := h.getUserFromSession(c); err == nil {
		return c.Redirect(http.StatusFound, "/chat")
	}
	return Register().Render(c.Request().Context(), c.Response().Writer)
}

// HandleLogin processes login form submission
func (h *Handlers) HandleLogin(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	if email == "" || password == "" {
		return c.HTML(http.StatusBadRequest, `<div class="bg-red-900/50 border border-red-800 rounded-lg p-4 text-red-200 text-sm">Email and password are required</div>`)
	}

	// Login user
	user, err := h.authService.Login(email, password)
	if err != nil {
		return c.HTML(http.StatusUnauthorized, `<div class="bg-red-900/50 border border-red-800 rounded-lg p-4 text-red-200 text-sm">Invalid email or password</div>`)
	}

	// Create session API key
	_, plainKey, err := h.authService.CreateAPIKey(
		user.ID,
		"Web Session - "+time.Now().Format("2006-01-02 15:04:05"),
		[]string{"*"},
		nil,
	)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, `<div class="bg-red-900/50 border border-red-800 rounded-lg p-4 text-red-200 text-sm">Login successful but failed to create session</div>`)
	}

	// Set session cookie
	h.setSessionCookie(c, plainKey)

	// Redirect to chat
	c.Response().Header().Set("HX-Redirect", "/chat")
	return c.NoContent(http.StatusOK)
}

// HandleRegister processes registration form submission
func (h *Handlers) HandleRegister(c echo.Context) error {
	email := c.FormValue("email")
	password := c.FormValue("password")
	confirmPassword := c.FormValue("confirm_password")

	if email == "" || password == "" {
		return c.HTML(http.StatusBadRequest, `<div class="bg-red-900/50 border border-red-800 rounded-lg p-4 text-red-200 text-sm">Email and password are required</div>`)
	}

	if password != confirmPassword {
		return c.HTML(http.StatusBadRequest, `<div class="bg-red-900/50 border border-red-800 rounded-lg p-4 text-red-200 text-sm">Passwords do not match</div>`)
	}

	if len(password) < 8 {
		return c.HTML(http.StatusBadRequest, `<div class="bg-red-900/50 border border-red-800 rounded-lg p-4 text-red-200 text-sm">Password must be at least 8 characters</div>`)
	}

	// Register user
	user, err := h.authService.Register(email, password)
	if err != nil {
		return c.HTML(http.StatusBadRequest, `<div class="bg-red-900/50 border border-red-800 rounded-lg p-4 text-red-200 text-sm">Registration failed: `+err.Error()+`</div>`)
	}

	// Create session API key
	_, plainKey, err := h.authService.CreateAPIKey(
		user.ID,
		"Web Session - "+time.Now().Format("2006-01-02 15:04:05"),
		[]string{"*"},
		nil,
	)
	if err != nil {
		return c.HTML(http.StatusInternalServerError, `<div class="bg-red-900/50 border border-red-800 rounded-lg p-4 text-red-200 text-sm">Registration successful but failed to create session</div>`)
	}

	// Set session cookie
	h.setSessionCookie(c, plainKey)

	// Redirect to chat
	c.Response().Header().Set("HX-Redirect", "/chat")
	return c.NoContent(http.StatusOK)
}

// HandleLogout logs out the user
func (h *Handlers) HandleLogout(c echo.Context) error {
	h.clearSessionCookie(c)
	c.Response().Header().Set("HX-Redirect", "/login")
	return c.NoContent(http.StatusOK)
}

// ShowChat displays the chat interface
func (h *Handlers) ShowChat(c echo.Context) error {
	user, err := h.getUserFromSession(c)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	_ = user // Use user if needed
	return Chat().Render(c.Request().Context(), c.Response().Writer)
}

// ShowWebsites displays the website management page
func (h *Handlers) ShowWebsites(c echo.Context) error {
	user, err := h.getUserFromSession(c)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	// Fetch all websites (TODO: filter by user_id when implemented in repo)
	websites, err := h.websiteRepo.List(c.Request().Context())
	if err != nil {
		websites = []schema.Website{}
	}

	// Filter by user ID
	userWebsites := []schema.Website{}
	for _, w := range websites {
		if w.UserID != nil && *w.UserID == user.ID {
			userWebsites = append(userWebsites, w)
		}
	}

	return Websites(userWebsites).Render(c.Request().Context(), c.Response().Writer)
}

// ShowAPIKeys displays the API key management page
func (h *Handlers) ShowAPIKeys(c echo.Context) error {
	user, err := h.getUserFromSession(c)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	// Fetch user's API keys
	keysPtr, err := h.apiKeyRepo.GetByUserID(c.Request().Context(), user.ID)
	if err != nil {
		return APIKeys([]schema.APIKey{}).Render(c.Request().Context(), c.Response().Writer)
	}

	// Convert []*schema.APIKey to []schema.APIKey
	keys := make([]schema.APIKey, len(keysPtr))
	for i, k := range keysPtr {
		keys[i] = *k
	}

	return APIKeys(keys).Render(c.Request().Context(), c.Response().Writer)
}

// ShowJobs displays the job monitoring page (admin only)
func (h *Handlers) ShowJobs(c echo.Context) error {
	user, err := h.getUserFromSession(c)
	if err != nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	// Check if user is admin
	if user.Role != "admin" {
		return c.Redirect(http.StatusFound, "/chat")
	}

	// TODO: Fetch jobs from asynq
	jobs := []schema.Job{}

	return Jobs(jobs).Render(c.Request().Context(), c.Response().Writer)
}

// AuthMiddleware checks if user is authenticated via session cookie
func (h *Handlers) AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := h.getUserFromSession(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}
		return next(c)
	}
}

// AdminMiddleware checks if user is an admin
func (h *Handlers) AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, err := h.getUserFromSession(c)
		if err != nil {
			return c.Redirect(http.StatusFound, "/login")
		}
		if user.Role != "admin" {
			return c.String(http.StatusForbidden, "Admin access required")
		}
		return next(c)
	}
}
