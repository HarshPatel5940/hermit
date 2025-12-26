package routes

import (
	"net/http"

	"hermit/api/controllers"
	"hermit/web"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// AppForRoutes defines the interface required by the route setup functions.
type AppForRoutes interface {
	WebsocketHandler(c echo.Context) error
}

// SetupRoutes registers all the application routes.
func SetupRoutes(e *echo.Echo, app AppForRoutes, wc *controllers.WebsiteController) {
	// API Routes
	api := e.Group("/api")
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	api.GET("/swagger/*", echoSwagger.WrapHandler)

	websiteRoutes := api.Group("/websites")
	websiteRoutes.POST("", wc.CreateWebsite)
	websiteRoutes.GET("", wc.ListWebsites)
	websiteRoutes.GET("/:id/pages", wc.GetPages)
	websiteRoutes.POST("/:id/query", wc.QueryWebsite)
	websiteRoutes.GET("/:id/status", wc.GetWebsiteStatus)
	websiteRoutes.POST("/:id/recrawl", wc.RecrawlWebsite)

	// Web Routes
	e.Static("/assets", "web/assets")
	e.GET("/web", echo.WrapHandler(templ.Handler(web.HelloForm())))
	e.POST("/hello", echo.WrapHandler(http.HandlerFunc(web.HelloWebHandler)))

	// Websocket Route
	e.GET("/websocket", app.WebsocketHandler)

	// Root Route
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Hello World"})
	})
}
