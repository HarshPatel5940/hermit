package controllers

import (
	"fmt"
	"hermit/internal/jobs"
	"hermit/internal/llm"
	"hermit/internal/repositories"
	_ "hermit/internal/schema" // Used by swaggo
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// WebsiteController handles API requests for websites.
type WebsiteController struct {
	websiteRepo *repositories.WebsiteRepository
	pageRepo    *repositories.PageRepository
	jobClient   *jobs.Client
	ragService  *llm.RAGService
	logger      *zap.Logger
}

// NewWebsiteController creates a new WebsiteController.
func NewWebsiteController(
	websiteRepo *repositories.WebsiteRepository,
	pageRepo *repositories.PageRepository,
	jobClient *jobs.Client,
	ragService *llm.RAGService,
	logger *zap.Logger,
) *WebsiteController {
	return &WebsiteController{
		websiteRepo: websiteRepo,
		pageRepo:    pageRepo,
		jobClient:   jobClient,
		ragService:  ragService,
		logger:      logger,
	}
}

// WebsiteCreateRequest defines the request body for creating a website.
type WebsiteCreateRequest struct {
	URL string `json:"url" example:"https://example.com"`
}

// CreateWebsite godoc
// @Summary      Create a new website
// @Description  Adds a new website to the monitoring list and starts the crawling process.
// @Tags         Websites
// @Accept       json
// @Produce      json
// @Param        website  body      WebsiteCreateRequest  true  "Website URL"
// @Success      201      {object}  schema.Website
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /websites [post]
func (wc *WebsiteController) CreateWebsite(c echo.Context) error {
	var req WebsiteCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	website, err := wc.websiteRepo.Create(c.Request().Context(), req.URL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create website"})
	}

	// Enqueue crawl job
	err = wc.jobClient.EnqueueCrawlWebsite(c.Request().Context(), uint(website.ID), website.URL)
	if err != nil {
		wc.logger.Error("Failed to enqueue crawl job", zap.Error(err))
		// Don't fail the request, website is created
	}

	return c.JSON(http.StatusCreated, website)
}

// ListWebsites godoc
// @Summary      List all websites
// @Description  Retrieves a list of all monitored websites.
// @Tags         Websites
// @Produce      json
// @Success      200  {array}   schema.Website
// @Failure      500  {object}  map[string]string
// @Router       /websites [get]
func (wc *WebsiteController) ListWebsites(c echo.Context) error {
	websites, err := wc.websiteRepo.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to list websites"})
	}

	return c.JSON(http.StatusOK, websites)
}

// GetPages godoc
// @Summary      Get pages for a website
// @Description  Retrieves all crawled pages for a specific website.
// @Tags         Websites
// @Produce      json
// @Param        id   path      int  true  "Website ID"
// @Success      200  {array}   schema.Page
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /websites/{id}/pages [get]
func (wc *WebsiteController) GetPages(c echo.Context) error {
	idParam := c.Param("id")
	websiteID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid website ID"})
	}

	pages, err := wc.pageRepo.GetByWebsiteID(c.Request().Context(), uint(websiteID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve pages"})
	}

	return c.JSON(http.StatusOK, pages)
}

// QueryRequest defines the request body for querying a website.
type QueryRequest struct {
	Query string `json:"query" example:"What is this website about?"`
}

// QueryWebsite godoc
// @Summary      Query website content using AI
// @Description  Performs a RAG-based query against the website's indexed content.
// @Tags         Websites
// @Accept       json
// @Produce      json
// @Param        id     path      int           true  "Website ID"
// @Param        query  body      QueryRequest  true  "Query"
// @Success      200    {object}  llm.QueryResponse
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /websites/{id}/query [post]
func (wc *WebsiteController) QueryWebsite(c echo.Context) error {
	idParam := c.Param("id")
	websiteID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid website ID"})
	}

	var req QueryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Query cannot be empty"})
	}

	response, err := wc.ragService.Query(c.Request().Context(), uint(websiteID), req.Query)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to process query"})
	}

	return c.JSON(http.StatusOK, response)
}

// QueryWebsiteStream godoc
// @Summary      Query website content (streaming)
// @Description  Ask questions about website content using AI with Server-Sent Events streaming
// @Tags         Websites
// @Accept       json
// @Produce      text/event-stream
// @Param        id     path      int                   true  "Website ID"
// @Param        query  body      QueryRequest   true  "Query"
// @Success      200    {string}  string                "SSE stream of answer chunks"
// @Failure      400    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Router       /websites/{id}/query/stream [post]
func (wc *WebsiteController) QueryWebsiteStream(c echo.Context) error {
	idParam := c.Param("id")
	websiteID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid website ID"})
	}

	var req QueryRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if req.Query == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Query cannot be empty"})
	}

	// Set headers for SSE
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	c.Response().WriteHeader(http.StatusOK)

	// Send initial event
	fmt.Fprintf(c.Response(), "event: start\ndata: {\"query\":\"%s\"}\n\n", req.Query)
	c.Response().Flush()

	// Stream the response
	meta, err := wc.ragService.QueryStream(c.Request().Context(), uint(websiteID), req.Query, func(chunk string) error {
		// Send each chunk as SSE
		fmt.Fprintf(c.Response(), "event: chunk\ndata: %s\n\n", chunk)
		c.Response().Flush()
		return nil
	})

	if err != nil {
		wc.logger.Error("Failed to process streaming query", zap.Error(err))
		fmt.Fprintf(c.Response(), "event: error\ndata: {\"error\":\"Failed to process query\"}\n\n")
		c.Response().Flush()
		return nil
	}

	// Send metadata with sources
	fmt.Fprintf(c.Response(), "event: metadata\ndata: {\"retrieved_chunks\":%d,\"sources_count\":%d}\n\n",
		meta.RetrievedChunks, len(meta.Sources))
	c.Response().Flush()

	// Send done event
	fmt.Fprintf(c.Response(), "event: done\ndata: {\"status\":\"complete\"}\n\n")
	c.Response().Flush()

	return nil
}

// GetWebsiteStatus godoc
// @Summary      Get website crawl status
// @Description  Retrieves the current crawl status and statistics for a website.
// @Tags         Websites
// @Produce      json
// @Param        id   path      int  true  "Website ID"
// @Success      200  {object}  schema.Website
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /websites/{id}/status [get]
func (wc *WebsiteController) GetWebsiteStatus(c echo.Context) error {
	idParam := c.Param("id")
	websiteID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid website ID"})
	}

	website, err := wc.websiteRepo.GetByID(c.Request().Context(), uint(websiteID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve website"})
	}

	if website == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Website not found"})
	}

	return c.JSON(http.StatusOK, website)
}

// RecrawlWebsite godoc
// @Summary      Trigger website re-crawl
// @Description  Manually triggers a re-crawl of a website.
// @Tags         Websites
// @Produce      json
// @Param        id   path      int  true  "Website ID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Router       /websites/{id}/recrawl [post]
func (wc *WebsiteController) RecrawlWebsite(c echo.Context) error {
	idParam := c.Param("id")
	websiteID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid website ID"})
	}

	website, err := wc.websiteRepo.GetByID(c.Request().Context(), uint(websiteID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve website"})
	}

	if website == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Website not found"})
	}

	// Check if already crawling
	if website.CrawlStatus == "crawling" {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Website is already being crawled"})
	}

	// Enqueue recrawl job
	err = wc.jobClient.EnqueueRecrawlWebsite(c.Request().Context(), uint(websiteID))
	if err != nil {
		wc.logger.Error("Failed to enqueue recrawl job", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to enqueue recrawl job"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Re-crawl job enqueued",
		"status":  "pending",
	})
}
