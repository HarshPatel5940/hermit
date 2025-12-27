package jobs

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Server wraps asynq.Server for processing tasks.
type Server struct {
	server   *asynq.Server
	mux      *asynq.ServeMux
	logger   *zap.Logger
	handlers *Handlers
}

// ServerConfig holds configuration for the job server.
type ServerConfig struct {
	RedisURL    string
	Concurrency int
	Queues      map[string]int
}

// NewServer creates a new job server.
func NewServer(cfg ServerConfig, handlers *Handlers, logger *zap.Logger) (*Server, error) {
	opt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// Set default queues if not provided
	queues := cfg.Queues
	if queues == nil {
		queues = map[string]int{
			"critical":    6, // Highest priority
			"crawl":       4,
			"vectorize":   3,
			"default":     2,
			"maintenance": 1, // Lowest priority
		}
	}

	// Set default concurrency
	concurrency := cfg.Concurrency
	if concurrency == 0 {
		concurrency = 10
	}

	server := asynq.NewServer(
		opt,
		asynq.Config{
			Concurrency:  concurrency,
			Queues:       queues,
			Logger:       NewAsynqLogger(logger),
			ErrorHandler: &errorHandler{logger: logger},
			// Retry failed tasks
			RetryDelayFunc: asynq.DefaultRetryDelayFunc,
		},
	)

	mux := asynq.NewServeMux()

	logger.Info("Job server initialized",
		zap.Int("concurrency", concurrency),
		zap.Any("queues", queues),
	)

	return &Server{
		server:   server,
		mux:      mux,
		logger:   logger,
		handlers: handlers,
	}, nil
}

// RegisterHandlers registers all task handlers.
func (s *Server) RegisterHandlers() {
	s.mux.HandleFunc(TypeCrawlWebsite, s.handlers.HandleCrawlWebsite)
	s.mux.HandleFunc(TypeVectorizePage, s.handlers.HandleVectorizePage)
	s.mux.HandleFunc(TypeRecrawlWebsite, s.handlers.HandleRecrawlWebsite)
	s.mux.HandleFunc(TypeCleanupOldPages, s.handlers.HandleCleanupOldPages)

	s.logger.Info("Job handlers registered",
		zap.Strings("types", []string{
			TypeCrawlWebsite,
			TypeVectorizePage,
			TypeRecrawlWebsite,
			TypeCleanupOldPages,
		}),
	)
}

// Start starts the job server.
func (s *Server) Start() error {
	s.logger.Info("Starting job server...")

	if err := s.server.Start(s.mux); err != nil {
		return fmt.Errorf("failed to start job server: %w", err)
	}

	s.logger.Info("Job server started successfully")
	return nil
}

// Stop gracefully shuts down the job server.
func (s *Server) Stop() {
	s.logger.Info("Stopping job server...")
	s.server.Shutdown()
	s.logger.Info("Job server stopped")
}

// AsynqLogger adapts zap.Logger to asynq.Logger interface.
type AsynqLogger struct {
	logger *zap.Logger
}

// NewAsynqLogger creates a new AsynqLogger.
func NewAsynqLogger(logger *zap.Logger) *AsynqLogger {
	return &AsynqLogger{logger: logger}
}

// Debug logs a debug message.
func (l *AsynqLogger) Debug(args ...interface{}) {
	l.logger.Sugar().Debug(args...)
}

// Info logs an info message.
func (l *AsynqLogger) Info(args ...interface{}) {
	l.logger.Sugar().Info(args...)
}

// Warn logs a warning message.
func (l *AsynqLogger) Warn(args ...interface{}) {
	l.logger.Sugar().Warn(args...)
}

// Error logs an error message.
func (l *AsynqLogger) Error(args ...interface{}) {
	l.logger.Sugar().Error(args...)
}

// Fatal logs a fatal message.
func (l *AsynqLogger) Fatal(args ...interface{}) {
	l.logger.Sugar().Fatal(args...)
}

// errorHandler implements asynq.ErrorHandler interface.
type errorHandler struct {
	logger *zap.Logger
}

// HandleError handles task processing errors.
func (h *errorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	h.logger.Error("Task processing failed",
		zap.String("type", task.Type()),
		zap.Error(err),
	)
}
