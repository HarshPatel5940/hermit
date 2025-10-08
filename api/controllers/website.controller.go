package controllers

import (
	"hermit/internal/crawler"
	"hermit/internal/repositories"
	_ "hermit/internal/schema" // Used by swaggo
	"net/http"

	"github.com/labstack/echo/v4"
)

// WebsiteController handles API requests for websites.
type WebsiteController struct {
	websiteRepo *repositories.WebsiteRepository
	crawler     *crawler.Crawler
}

// NewWebsiteController creates a new WebsiteController.
func NewWebsiteController(websiteRepo *repositories.WebsiteRepository, crawler *crawler.Crawler) *WebsiteController {
	return &WebsiteController{websiteRepo: websiteRepo, crawler: crawler}
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
	go wc.crawler.Crawl(website.URL)

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
