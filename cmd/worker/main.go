package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"hermit/internal/config"
	"hermit/internal/contentprocessor"
	"hermit/internal/crawler"
	"hermit/internal/database"
	"hermit/internal/jobs"
	"hermit/internal/repositories"
	"hermit/internal/storage"
	"hermit/internal/vectorizer"

	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Hermit worker...")

	// Load configuration
	cfg := config.NewConfig()

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Garage storage
	garageClient, err := database.NewGarageClient(cfg)
	if err != nil {
		logger.Fatal("Failed to create Garage client", zap.Error(err))
	}
	garageStorage := storage.NewGarageStorage(garageClient, cfg, logger)

	// Initialize repositories
	websiteRepo := repositories.NewWebsiteRepository(db)
	pageRepo := repositories.NewPageRepository(db)

	// Initialize vectorizer components
	embedder := vectorizer.NewEmbedder(cfg.OllamaURL, cfg.OllamaModel, logger)
	chromaRepo, err := vectorizer.NewChromaRepository(cfg.ChromaDBURL, logger)
	if err != nil {
		logger.Fatal("Failed to create ChromaDB repository", zap.Error(err))
	}
	vectorizerSvc := vectorizer.NewService(embedder, chromaRepo, logger)

	// Initialize content processors
	contentProcessor := contentprocessor.NewContentProcessor(logger)
	robotsEnforcer := contentprocessor.NewRobotsEnforcer(cfg.CrawlerUserAgent, logger)

	// Initialize job client (for enqueueing sub-tasks)
	jobClient, err := jobs.NewClient(cfg.RedisURL, logger)
	if err != nil {
		logger.Fatal("Failed to create job client", zap.Error(err))
	}
	defer jobClient.Close()

	// Initialize crawler
	crawlerSvc := crawler.NewCrawler(
		logger,
		garageStorage,
		pageRepo,
		websiteRepo,
		vectorizerSvc,
		contentProcessor,
		robotsEnforcer,
		jobClient,
		cfg,
	)

	// Initialize job handlers
	handlers := jobs.NewHandlers(
		logger,
		crawlerSvc,
		vectorizerSvc,
		websiteRepo,
		pageRepo,
	)

	// Initialize job server
	serverCfg := jobs.ServerConfig{
		RedisURL:    cfg.RedisURL,
		Concurrency: 10, // TODO: make configurable
		Queues: map[string]int{
			"critical":    6,
			"crawl":       4,
			"vectorize":   3,
			"default":     2,
			"maintenance": 1,
		},
	}

	jobServer, err := jobs.NewServer(serverCfg, handlers, logger)
	if err != nil {
		logger.Fatal("Failed to create job server", zap.Error(err))
	}

	// Register handlers
	jobServer.RegisterHandlers()

	// Start job server in background
	if err := jobServer.Start(); err != nil {
		logger.Fatal("Failed to start job server", zap.Error(err))
	}

	logger.Info("Worker started successfully, processing jobs...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal, stopping worker...")

	// Graceful shutdown
	jobServer.Stop()

	logger.Info("Worker stopped successfully")
}

func initLogger() (*zap.Logger, error) {
	if os.Getenv("APP_ENV") == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
