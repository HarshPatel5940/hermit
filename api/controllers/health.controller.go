package controllers

import (
	"context"
	"net/http"
	"time"

	"hermit/internal/config"
	"hermit/internal/database"
	"hermit/internal/storage"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// HealthController handles health check endpoints.
type HealthController struct {
	logger   *zap.Logger
	db       *sqlx.DB
	storage  *storage.GarageStorage
	chromaDB *database.ChromaDBClient
	config   *config.Config
}

// NewHealthController creates a new HealthController.
func NewHealthController(
	logger *zap.Logger,
	db *sqlx.DB,
	storage *storage.GarageStorage,
	chromaDB *database.ChromaDBClient,
	cfg *config.Config,
) *HealthController {
	return &HealthController{
		logger:   logger,
		db:       db,
		storage:  storage,
		chromaDB: chromaDB,
		config:   cfg,
	}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp string                   `json:"timestamp"`
	Services  map[string]ServiceHealth `json:"services"`
}

// ServiceHealth represents the health of a service.
type ServiceHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
}

// GetHealth handles GET /health
// @Summary Health check
// @Description Check health of all services
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *HealthController) GetHealth(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  make(map[string]ServiceHealth),
	}

	// Check PostgreSQL
	postgresHealth := h.checkPostgres(ctx)
	response.Services["postgres"] = postgresHealth
	if postgresHealth.Status != "healthy" {
		response.Status = "unhealthy"
	}

	// Check Garage (S3)
	garageHealth := h.checkGarage(ctx)
	response.Services["garage"] = garageHealth
	if garageHealth.Status != "healthy" {
		response.Status = "degraded"
	}

	// Check ChromaDB
	chromaHealth := h.checkChromaDB(ctx)
	response.Services["chromadb"] = chromaHealth
	if chromaHealth.Status != "healthy" {
		response.Status = "degraded"
	}

	// Check Ollama
	ollamaHealth := h.checkOllama(ctx)
	response.Services["ollama"] = ollamaHealth
	if ollamaHealth.Status != "healthy" {
		response.Status = "degraded"
	}

	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, response)
}

// checkPostgres checks PostgreSQL connection.
func (h *HealthController) checkPostgres(ctx context.Context) ServiceHealth {
	start := time.Now()

	err := h.db.PingContext(ctx)
	latency := time.Since(start)

	if err != nil {
		h.logger.Error("PostgreSQL health check failed", zap.Error(err))
		return ServiceHealth{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return ServiceHealth{
		Status:  "healthy",
		Latency: latency.String(),
	}
}

// checkGarage checks Garage S3 connection.
func (h *HealthController) checkGarage(ctx context.Context) ServiceHealth {
	start := time.Now()

	err := h.storage.EnsureBucket(ctx)
	latency := time.Since(start)

	if err != nil {
		h.logger.Error("Garage health check failed", zap.Error(err))
		return ServiceHealth{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return ServiceHealth{
		Status:  "healthy",
		Latency: latency.String(),
	}
}

// checkChromaDB checks ChromaDB connection.
func (h *HealthController) checkChromaDB(ctx context.Context) ServiceHealth {
	start := time.Now()

	// Simple heartbeat check
	err := h.chromaDB.Heartbeat(ctx)
	latency := time.Since(start)

	if err != nil {
		h.logger.Error("ChromaDB health check failed", zap.Error(err))
		return ServiceHealth{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}

	return ServiceHealth{
		Status:  "healthy",
		Latency: latency.String(),
	}
}

// checkOllama checks Ollama service.
func (h *HealthController) checkOllama(ctx context.Context) ServiceHealth {
	start := time.Now()

	// Make a simple HTTP request to Ollama
	req, err := http.NewRequestWithContext(ctx, "GET", h.config.OllamaURL+"/api/tags", nil)
	if err != nil {
		return ServiceHealth{
			Status:  "unhealthy",
			Message: "Failed to create request: " + err.Error(),
		}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		h.logger.Error("Ollama health check failed", zap.Error(err))
		return ServiceHealth{
			Status:  "unhealthy",
			Message: err.Error(),
			Latency: latency.String(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ServiceHealth{
			Status:  "unhealthy",
			Message: "Unexpected status code",
			Latency: latency.String(),
		}
	}

	return ServiceHealth{
		Status:  "healthy",
		Latency: latency.String(),
	}
}
