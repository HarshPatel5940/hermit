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
func SetupRoutes(e *echo.Echo, app AppForRoutes, wc *controllers.WebsiteController, hc *controllers.HealthController, jc *controllers.JobsController) {
	// API Routes
	api := e.Group("/api")
	api.GET("/health", hc.GetHealth)
	api.GET("/swagger/*", echoSwagger.WrapHandler)

	websiteRoutes := api.Group("/websites")
	websiteRoutes.POST("", wc.CreateWebsite)
	websiteRoutes.GET("", wc.ListWebsites)
	websiteRoutes.GET("/:id/pages", wc.GetPages)
	websiteRoutes.POST("/:id/query", wc.QueryWebsite)
	websiteRoutes.POST("/:id/query/stream", wc.QueryWebsiteStream)
	websiteRoutes.GET("/:id/status", wc.GetWebsiteStatus)
	websiteRoutes.POST("/:id/recrawl", wc.RecrawlWebsite)

	// Job Management Routes
	jobRoutes := api.Group("/jobs")
	jobRoutes.GET("/queues", jc.ListQueues)
	jobRoutes.GET("/pending", jc.ListPendingJobs)
	jobRoutes.GET("/active", jc.ListActiveJobs)
	jobRoutes.GET("/scheduled", jc.ListScheduledJobs)
	jobRoutes.GET("/retry", jc.ListRetryJobs)
	jobRoutes.GET("/archived", jc.ListArchivedJobs)
	jobRoutes.POST("/:id/cancel", jc.CancelJob)
	jobRoutes.POST("/:id/retry", jc.RetryJob)
	jobRoutes.POST("/queues/:queue/pause", jc.PauseQueue)
	jobRoutes.POST("/queues/:queue/resume", jc.ResumeQueue)

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
