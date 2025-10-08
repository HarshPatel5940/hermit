package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"hermit/api/controllers"
	"hermit/api/middlewares"
	"hermit/api/routes"
	"hermit/internal/config"
	"hermit/internal/crawler"
	"hermit/internal/database"
	"hermit/internal/repositories"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/swaggo/echo-swagger"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// App holds all the dependencies for the application
type App struct {
	Echo *echo.Echo
	DB   *sqlx.DB
}

// NewLogger creates a new zap logger that is environment-aware.
func NewLogger() (*zap.Logger, error) {
	if os.Getenv("APP_ENV") == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// NewFxApp creates the Fx application with all the dependencies
func NewFxApp() *fx.App {
	return fx.New(
		fx.Provide(
			// Provides the application configuration
			config.NewConfig,

			// Provides the zap logger
			NewLogger,

			// Provides the database clients
			database.NewPostgresDB,
			database.NewMinIOClient,
			database.NewChromaDBClient,

			// Provides the services
			crawler.NewCrawler,

			// Provides the repositories
			repositories.NewWebsiteRepository,

			// Provides the controllers
			controllers.NewWebsiteController,

			// Provides the Echo web server instance
			func() *echo.Echo {
				return echo.New()
			},

			// Provides the main App struct
			func(e *echo.Echo, db *sqlx.DB) *App {
				return &App{Echo: e, DB: db}
			},
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(middlewares.SetupMiddlewares),
		fx.Invoke(RegisterHooks),
		fx.Invoke(SetupRoutes),
	)
}

// RegisterHooks registers the application lifecycle hooks with Fx
func RegisterHooks(lc fx.Lifecycle, app *App, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				address := fmt.Sprintf(":%s", cfg.Port)
				if err := app.Echo.Start(address); err != nil && err != http.ErrServerClosed {
					log.Fatalf("Error starting server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Echo.Shutdown(ctx)
		},
	})
}

// SetupRoutes registers all the application routes.
// @description  Returns the health status of the server.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func SetupRoutes(e *echo.Echo, wc *controllers.WebsiteController) {
	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})
	e.GET("/api/swagger/*", echoSwagger.WrapHandler)
	routes.SetupWebsiteRoutes(e, wc)
}
