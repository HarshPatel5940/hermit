package routes

import (
	"hermit/api/controllers"

	"github.com/labstack/echo/v4"
)

// SetupWebsiteRoutes configures the routes for the website resource.
func SetupWebsiteRoutes(e *echo.Echo, wc *controllers.WebsiteController) {
	websites := e.Group("/websites")
	websites.POST("", wc.CreateWebsite)
	websites.GET("", wc.ListWebsites)
}
