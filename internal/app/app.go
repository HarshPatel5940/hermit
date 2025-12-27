package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"hermit/api/controllers"
	"hermit/api/middlewares"
	"hermit/api/routes"
	"hermit/internal/config"
	"hermit/internal/contentprocessor"
	"hermit/internal/crawler"
	"hermit/internal/database"
	"hermit/internal/jobs"
	"hermit/internal/llm"
	"hermit/internal/repositories"
	"hermit/internal/storage"
	"hermit/internal/vectorizer"

	"github.com/coder/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

type App struct {
	Echo   *echo.Echo
	DB     *sqlx.DB
	Logger *zap.Logger
}

func (a *App) WebsocketHandler(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()
	socket, err := websocket.Accept(w, r, nil)

	if err != nil {
		a.Logger.Error("could not open websocket", zap.Error(err))
		_, _ = w.Write([]byte("could not open websocket"))
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}

	defer socket.Close(websocket.StatusGoingAway, "server closing websocket")

	ctx := r.Context()
	socketCtx := socket.CloseRead(ctx)

	for {
		payload := fmt.Sprintf("server timestamp: %d", time.Now().UnixNano())
		err := socket.Write(socketCtx, websocket.MessageText, []byte(payload))
		if err != nil {
			break
		}
		time.Sleep(time.Second * 2)
	}
	return nil
}

func NewLogger() (*zap.Logger, error) {
	if os.Getenv("APP_ENV") == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func NewFxApp() *fx.App {
	return fx.New(
		fx.Provide(
			config.NewConfig,
			NewLogger,

			database.NewPostgresDB,
			database.NewGarageClient,
			database.NewChromaDBClient,

			storage.NewGarageStorage,

			repositories.NewWebsiteRepository,
			repositories.NewPageRepository,

			func(cfg *config.Config, logger *zap.Logger) *vectorizer.Embedder {
				return vectorizer.NewEmbedder(cfg.OllamaURL, cfg.OllamaModel, logger)
			},
			func(cfg *config.Config, logger *zap.Logger) (*vectorizer.ChromaRepository, error) {
				return vectorizer.NewChromaRepository(cfg.ChromaDBURL, logger)
			},
			vectorizer.NewService,

			func(cfg *config.Config, logger *zap.Logger) *llm.OllamaLLM {
				return llm.NewOllamaLLM(cfg.OllamaURL, cfg.OllamaLLMModel, logger)
			},
			func(vectorizerSvc *vectorizer.Service, ollamaLLM *llm.OllamaLLM, logger *zap.Logger, cfg *config.Config) *llm.RAGService {
				return llm.NewRAGService(vectorizerSvc, ollamaLLM, logger, cfg.RAGTopK, cfg.RAGContextChunks)
			},

			func(logger *zap.Logger) *contentprocessor.ContentProcessor {
				return contentprocessor.NewContentProcessor(logger)
			},
			func(cfg *config.Config, logger *zap.Logger) *contentprocessor.RobotsEnforcer {
				return contentprocessor.NewRobotsEnforcer(cfg.CrawlerUserAgent, logger)
			},

			crawler.NewCrawler,

			func(cfg *config.Config, logger *zap.Logger) (*jobs.Client, error) {
				return jobs.NewClient(cfg.RedisURL, logger)
			},

			controllers.NewWebsiteController,
			controllers.NewHealthController,
			func(logger *zap.Logger, cfg *config.Config) (*controllers.JobsController, error) {
				return controllers.NewJobsController(logger, cfg.RedisURL)
			},

			func() *echo.Echo {
				return echo.New()
			},

			func(e *echo.Echo, db *sqlx.DB, log *zap.Logger) *App {
				return &App{Echo: e, DB: db, Logger: log}
			},
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(middlewares.SetupMiddlewares),
		fx.Invoke(RegisterHooks),
		fx.Invoke(func(e *echo.Echo, app *App, wc *controllers.WebsiteController, hc *controllers.HealthController, jc *controllers.JobsController) {
			routes.SetupRoutes(e, app, wc, hc, jc)
		}),
		fx.Invoke(func(lc fx.Lifecycle, jobClient *jobs.Client) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					return jobClient.Close()
				},
			})
		}),
	)
}

func RegisterHooks(lc fx.Lifecycle, app *App, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				address := fmt.Sprintf(":%s", cfg.Port)
				if err := app.Echo.Start(address); err != nil && err != http.ErrServerClosed {
					app.Logger.Fatal("Error starting server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Echo.Shutdown(ctx)
		},
	})
}
