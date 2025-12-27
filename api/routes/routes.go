package routes

import (
	"net/http"

	"hermit/api/controllers"
	"hermit/api/middlewares"
	"hermit/internal/auth"
	"hermit/web"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// AppForRoutes defines the interface required by the route setup functions.
type AppForRoutes interface {
	WebsocketHandler(c echo.Context) error
}

// SetupRoutes registers all the application routes with API versioning.
func SetupRoutes(
	e *echo.Echo,
	app AppForRoutes,
	wc *controllers.WebsiteController,
	hc *controllers.HealthController,
	jc *controllers.JobsController,
	ac *controllers.AuthController,
	authService *auth.Service,
) {
	// Root Route
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"message":      "Hermit API",
			"version":      "1.0.0",
			"api_versions": []string{"v1"},
		})
	})

	// API Routes (legacy, without versioning - for backward compatibility)
	api := e.Group("/api")
	api.GET("/health", hc.GetHealth)
	api.GET("/swagger/*", echoSwagger.WrapHandler)

	// API v1 Routes (versioned)
	v1 := e.Group("/api/v1")

	// Health & Swagger
	v1.GET("/health", hc.GetHealth)
	v1.GET("/swagger/*", echoSwagger.WrapHandler)

	// Auth Routes (public, no auth required)
	authRoutes := v1.Group("/auth")
	authRoutes.POST("/register", ac.Register)
	authRoutes.POST("/login", ac.Login)

	// Auth Routes (protected, auth required)
	authProtectedRoutes := v1.Group("/auth")
	authProtectedRoutes.Use(middlewares.AuthMiddleware(authService))
	authProtectedRoutes.GET("/me", ac.GetMe)
	authProtectedRoutes.POST("/api-keys", ac.CreateAPIKey)
	authProtectedRoutes.GET("/api-keys", ac.ListAPIKeys)
	authProtectedRoutes.GET("/api-keys/:id", ac.GetAPIKey)
	authProtectedRoutes.PUT("/api-keys/:id", ac.UpdateAPIKey)
	authProtectedRoutes.DELETE("/api-keys/:id", ac.RevokeAPIKey)

	// Website Routes (protected)
	websiteRoutes := v1.Group("/websites")
	websiteRoutes.Use(middlewares.AuthMiddleware(authService))
	websiteRoutes.POST("", wc.CreateWebsite)
	websiteRoutes.GET("", wc.ListWebsites)
	websiteRoutes.GET("/:id/pages", wc.GetPages)
	websiteRoutes.POST("/:id/query", wc.QueryWebsite)
	websiteRoutes.POST("/:id/query/stream", wc.QueryWebsiteStream)
	websiteRoutes.GET("/:id/status", wc.GetWebsiteStatus)
	websiteRoutes.POST("/:id/recrawl", wc.RecrawlWebsite)

	// Job Management Routes (protected, admin only)
	jobRoutes := v1.Group("/jobs")
	jobRoutes.Use(middlewares.AuthMiddleware(authService))
	jobRoutes.Use(middlewares.RequireRole("admin"))
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

	// Web Routes (public)
	e.Static("/assets", "web/assets")
	e.GET("/web", echo.WrapHandler(templ.Handler(web.HelloForm())))
	e.POST("/hello", echo.WrapHandler(http.HandlerFunc(web.HelloWebHandler)))

	// Websocket Route (public for now, can add auth later)
	e.GET("/websocket", app.WebsocketHandler)
}
