package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"hermit/internal/config"
	"hermit/internal/repositories"
	"hermit/internal/storage"
	"hermit/internal/vectorizer"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

// Crawler manages the website crawling process.
type Crawler struct {
	logger        *zap.Logger
	storage       *storage.MinIOStorage
	pageRepo      *repositories.PageRepository
	websiteRepo   *repositories.WebsiteRepository
	vectorizerSvc *vectorizer.Service
	config        *config.Config
}

// NewCrawler creates a new Crawler service.
func NewCrawler(
	logger *zap.Logger,
	storage *storage.MinIOStorage,
	pageRepo *repositories.PageRepository,
	websiteRepo *repositories.WebsiteRepository,
	vectorizerSvc *vectorizer.Service,
	cfg *config.Config,
) *Crawler {
	return &Crawler{
		logger:        logger,
		storage:       storage,
		pageRepo:      pageRepo,
		websiteRepo:   websiteRepo,
		vectorizerSvc: vectorizerSvc,
		config:        cfg,
	}
}

// Crawl starts the crawling process for a given URL.
func (cr *Crawler) Crawl(websiteID uint, startURL string) {
	cr.logger.Info("Crawling started", zap.String("url", startURL), zap.Uint("websiteID", websiteID))

	// Ensure MinIO bucket exists
	ctx := context.Background()
	if err := cr.storage.EnsureBucket(ctx); err != nil {
		cr.logger.Error("Failed to ensure MinIO bucket", zap.Error(err))
		cr.websiteRepo.FailCrawl(ctx, websiteID, "Failed to ensure MinIO bucket: "+err.Error())
		return
	}

	// Mark crawl as started
	if err := cr.websiteRepo.StartCrawl(ctx, websiteID); err != nil {
		cr.logger.Error("Failed to update crawl status", zap.Error(err))
	}

	// Parse the starting URL to extract the domain
	parsedURL, err := url.Parse(startURL)
	if err != nil {
		cr.logger.Error("Failed to parse URL", zap.String("url", startURL), zap.Error(err))
		cr.websiteRepo.FailCrawl(ctx, websiteID, "Failed to parse URL: "+err.Error())
		return
	}

	// Create collector with allowed domain and configuration
	c := colly.NewCollector(
		colly.AllowedDomains(parsedURL.Host),
		colly.MaxDepth(cr.config.CrawlerMaxDepth),
		colly.UserAgent(cr.config.CrawlerUserAgent),
	)

	// Set up rate limiting with delay
	if cr.config.CrawlerDelayMS > 0 {
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Delay:       time.Duration(cr.config.CrawlerDelayMS) * time.Millisecond,
			RandomDelay: time.Duration(cr.config.CrawlerDelayMS/2) * time.Millisecond,
		})
	}

	// Track page count and stats
	pageCount := 0
	successCount := 0
	failureCount := 0
	maxPages := cr.config.CrawlerMaxPages

	// Extract text content from the body
	c.OnHTML("body", func(e *colly.HTMLElement) {
		pageURL := e.Request.URL.String()
		text := strings.TrimSpace(e.Text)

		cr.logger.Info("Extracted text",
			zap.String("url", pageURL),
			zap.Int("length", len(text)),
		)

		// Create or update page record
		page, err := cr.pageRepo.Upsert(ctx, websiteID, pageURL)
		if err != nil {
			cr.logger.Error("Failed to upsert page", zap.String("url", pageURL), zap.Error(err))
			failureCount++
			cr.websiteRepo.IncrementPageCount(ctx, websiteID, false)
			return
		}

		// Generate content hash
		contentHash := hashContent(text)

		// Save content to MinIO
		objectKey, err := cr.storage.SavePageContent(ctx, int(websiteID), pageURL, text)
		if err != nil {
			cr.logger.Error("Failed to save content to MinIO", zap.String("url", pageURL), zap.Error(err))
			// Update page with error status
			cr.pageRepo.UpdateError(ctx, page.ID, err.Error())
			failureCount++
			cr.websiteRepo.IncrementPageCount(ctx, websiteID, false)
			return
		}

		// Update page with success status
		err = cr.pageRepo.UpdateSuccess(ctx, page.ID, objectKey, contentHash)
		if err != nil {
			cr.logger.Error("Failed to update page status", zap.String("url", pageURL), zap.Error(err))
			failureCount++
			cr.websiteRepo.IncrementPageCount(ctx, websiteID, false)
			return
		}

		successCount++
		cr.websiteRepo.IncrementPageCount(ctx, websiteID, true)

		cr.logger.Info("Successfully saved page",
			zap.String("url", pageURL),
			zap.String("objectKey", objectKey),
		)

		// Step 4: Vectorize the content (async)
		go func() {
			err := cr.vectorizerSvc.ProcessPageContent(ctx, websiteID, page.ID, pageURL, text)
			if err != nil {
				cr.logger.Error("Failed to vectorize page content",
					zap.String("url", pageURL),
					zap.Uint("pageID", page.ID),
					zap.Error(err),
				)
				return
			}
			cr.logger.Info("Successfully vectorized page",
				zap.String("url", pageURL),
				zap.Uint("pageID", page.ID),
			)
		}()
	})

	// Find and visit all same-domain links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// Check if max pages limit reached
		if maxPages > 0 && pageCount >= maxPages {
			cr.logger.Info("Max pages limit reached, stopping crawler",
				zap.Int("maxPages", maxPages),
			)
			return
		}

		link := e.Attr("href")
		// Visit the link (colly handles same-domain filtering)
		e.Request.Visit(link)
	})

	c.OnRequest(func(r *colly.Request) {
		pageCount++
		cr.logger.Info("Visiting",
			zap.String("url", r.URL.String()),
			zap.Int("pageCount", pageCount),
			zap.Int("maxPages", maxPages),
		)
	})

	c.OnError(func(r *colly.Response, err error) {
		cr.logger.Error("Request failed",
			zap.String("url", r.Request.URL.String()),
			zap.Error(err),
		)
	})

	c.Visit(startURL)

	// Mark crawl as completed
	if err := cr.websiteRepo.CompleteCrawl(ctx, websiteID, successCount, failureCount); err != nil {
		cr.logger.Error("Failed to update crawl completion status", zap.Error(err))
	}

	cr.logger.Info("Crawling completed",
		zap.String("url", startURL),
		zap.Int("totalPages", pageCount),
		zap.Int("successCount", successCount),
		zap.Int("failureCount", failureCount),
	)
}

// hashContent creates a SHA256 hash of content.
func hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
