package web

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// SetupRoutes configures the routes for the web interface.
func SetupRoutes(e *echo.Echo) {
	// Use the embedded file system for static assets
	assetHandler := http.FileServer(http.FS(Files))
	e.GET("/assets/*", echo.WrapHandler(assetHandler))

	// Add the web routes
	e.GET("/web", echo.WrapHandler(templ.Handler(HelloForm())))
	e.POST("/hello", echo.WrapHandler(http.HandlerFunc(HelloWebHandler)))
}
