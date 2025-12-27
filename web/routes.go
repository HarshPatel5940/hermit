package web

import (
	"net/http"

	"hermit/internal/auth"
	"hermit/internal/repositories"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// SetupRoutes configures the routes for the web interface.
func SetupRoutes(
	e *echo.Echo,
	authService *auth.Service,
	websiteRepo *repositories.WebsiteRepository,
	apiKeyRepo *repositories.APIKeyRepository,
	userRepo *repositories.UserRepository,
) {
	// Create handlers
	h := NewHandlers(authService, websiteRepo, apiKeyRepo, userRepo)

	// Use the embedded file system for static assets
	assetHandler := http.FileServer(http.FS(Files))
	e.GET("/assets/*", echo.WrapHandler(assetHandler))

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/login")
	})
	e.GET("/login", h.ShowLogin)
	e.POST("/login", h.HandleLogin)
	e.GET("/register", h.ShowRegister)
	e.POST("/register", h.HandleRegister)
	e.POST("/logout", h.HandleLogout)

	// Protected routes (require authentication)
	protected := e.Group("")
	protected.Use(h.AuthMiddleware)
	protected.GET("/chat", h.ShowChat)
	protected.GET("/websites", h.ShowWebsites)
	protected.GET("/api-keys", h.ShowAPIKeys)

	// Admin routes (require admin role)
	admin := e.Group("")
	admin.Use(h.AuthMiddleware)
	admin.Use(h.AdminMiddleware)
	admin.GET("/jobs", h.ShowJobs)

	// Legacy hello route
	e.GET("/web", echo.WrapHandler(templ.Handler(HelloForm())))
	e.POST("/hello", echo.WrapHandler(http.HandlerFunc(HelloWebHandler)))
}
