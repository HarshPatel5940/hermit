package web

import (
	"net/http"

	"hermit/internal/schema"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// SetupRoutes configures the routes for the web interface.
func SetupRoutes(e *echo.Echo) {
	// Use the embedded file system for static assets
	assetHandler := http.FileServer(http.FS(Files))
	e.GET("/assets/*", echo.WrapHandler(assetHandler))

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/login")
	})
	e.GET("/login", func(c echo.Context) error {
		return Login().Render(c.Request().Context(), c.Response().Writer)
	})
	e.GET("/register", func(c echo.Context) error {
		return Register().Render(c.Request().Context(), c.Response().Writer)
	})

	// Protected routes (TODO: Add auth middleware)
	e.GET("/chat", func(c echo.Context) error {
		return Chat().Render(c.Request().Context(), c.Response().Writer)
	})
	e.GET("/websites", func(c echo.Context) error {
		// TODO: Fetch actual websites from database
		websites := []schema.Website{}
		return Websites(websites).Render(c.Request().Context(), c.Response().Writer)
	})
	e.GET("/api-keys", func(c echo.Context) error {
		// TODO: Fetch actual API keys from database
		keys := []schema.APIKey{}
		return APIKeys(keys).Render(c.Request().Context(), c.Response().Writer)
	})
	e.GET("/jobs", func(c echo.Context) error {
		// TODO: Fetch actual jobs from database
		jobs := []schema.Job{}
		return Jobs(jobs).Render(c.Request().Context(), c.Response().Writer)
	})

	// Legacy hello route
	e.GET("/web", echo.WrapHandler(templ.Handler(HelloForm())))
	e.POST("/hello", echo.WrapHandler(http.HandlerFunc(HelloWebHandler)))
}
