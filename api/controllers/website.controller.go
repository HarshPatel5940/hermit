package controllers

import (
	"hermit/internal/crawler"
	"hermit/internal/llm"
	"hermit/internal/repositories"
	_ "hermit/internal/schema" // Used by swaggo
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// WebsiteController handles API requests for websites.
type WebsiteController struct {
	websiteRepo *repositories.WebsiteRepository
	pageRepo    *repositories.PageRepository
	crawler     *crawler.Crawler
	ragService  *llm.RAGService
}

// NewWebsiteController creates a new WebsiteController.
func NewWebsiteController(
	websiteRepo *repositories.WebsiteRepository,
	pageRepo *repositories.PageRepository,
	crawler *crawler.Crawler,
	ragService *llm.RAGService,
) *WebsiteController {
	return &WebsiteController{
		websiteRepo: websiteRepo,
		pageRepo:    pageRepo,
		crawler:     crawler,
		ragService:  ragService,
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

	// Start crawling in the background
	go wc.crawler.Crawl(uint(website.ID), website.URL)

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

	// Start crawling in the background
	go wc.crawler.Crawl(uint(websiteID), website.URL)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Re-crawl started",
		"status":  "crawling",
	})
}
