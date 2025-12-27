package crawler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"hermit/internal/config"
	"hermit/internal/contentprocessor"
	"hermit/internal/repositories"
	"hermit/internal/storage"
	"hermit/internal/vectorizer"
	"net/url"
	"time"

	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

// Crawler manages the website crawling process.
type Crawler struct {
	logger           *zap.Logger
	storage          *storage.GarageStorage
	pageRepo         *repositories.PageRepository
	websiteRepo      *repositories.WebsiteRepository
	vectorizerSvc    *vectorizer.Service
	contentProcessor *contentprocessor.ContentProcessor
	robotsEnforcer   *contentprocessor.RobotsEnforcer
	jobClient        interface {
		EnqueueVectorizePage(ctx context.Context, websiteID, pageID uint, pageURL, content string) error
	}
	config *config.Config
}

// NewCrawler creates a new Crawler service.
func NewCrawler(
	logger *zap.Logger,
	storage *storage.GarageStorage,
	pageRepo *repositories.PageRepository,
	websiteRepo *repositories.WebsiteRepository,
	vectorizerSvc *vectorizer.Service,
	contentProcessor *contentprocessor.ContentProcessor,
	robotsEnforcer *contentprocessor.RobotsEnforcer,
	jobClient interface {
		EnqueueVectorizePage(ctx context.Context, websiteID, pageID uint, pageURL, content string) error
	},
	cfg *config.Config,
) *Crawler {
	return &Crawler{
		logger:           logger,
		storage:          storage,
		pageRepo:         pageRepo,
		websiteRepo:      websiteRepo,
		vectorizerSvc:    vectorizerSvc,
		contentProcessor: contentProcessor,
		robotsEnforcer:   robotsEnforcer,
		jobClient:        jobClient,
		config:           cfg,
	}
}

// Crawl starts the crawling process for a given URL.
func (cr *Crawler) Crawl(websiteID uint, startURL string) {
	cr.logger.Info("Crawling started", zap.String("url", startURL), zap.Uint("websiteID", websiteID))

	// Ensure Garage bucket exists
	ctx := context.Background()
	if err := cr.storage.EnsureBucket(ctx); err != nil {
		cr.logger.Error("Failed to ensure Garage bucket", zap.Error(err))
		cr.websiteRepo.FailCrawl(ctx, websiteID, "Failed to ensure Garage bucket: "+err.Error())
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
	visitedURLs := make(map[string]bool)

	// Extract and process HTML content
	c.OnHTML("html", func(e *colly.HTMLElement) {
		pageURL := e.Request.URL.String()
		htmlContent := e.Response.Body

		// Normalize URL to prevent duplicates
		normalizedURL, err := contentprocessor.NormalizeURL(pageURL)
		if err != nil {
			cr.logger.Error("Failed to normalize URL", zap.String("url", pageURL), zap.Error(err))
			failureCount++
			return
		}

		// Check if already visited (in-memory dedup)
		if visitedURLs[normalizedURL] {
			cr.logger.Debug("Skipping duplicate URL", zap.String("url", pageURL))
			return
		}
		visitedURLs[normalizedURL] = true

		cr.logger.Info("Processing page",
			zap.String("url", pageURL),
			zap.Int("htmlSize", len(htmlContent)),
		)

		// Extract main content using readability
		processed, err := cr.contentProcessor.ExtractMainContent(string(htmlContent), pageURL)
		if err != nil {
			cr.logger.Error("Failed to extract main content", zap.String("url", pageURL), zap.Error(err))
			failureCount++
			cr.websiteRepo.IncrementPageCount(ctx, websiteID, false)
			return
		}

		// Validate content quality
		if !cr.contentProcessor.IsContentValid(processed, cr.config.ContentMinLength, cr.config.ContentMinQuality) {
			cr.logger.Warn("Content quality too low, skipping",
				zap.String("url", pageURL),
				zap.Int("length", processed.Length),
				zap.Float64("quality", processed.Quality),
			)
			failureCount++
			cr.websiteRepo.IncrementPageCount(ctx, websiteID, false)
			return
		}

		// Clean text
		cleanedText := cr.contentProcessor.CleanText(processed.Content)

		cr.logger.Info("Extracted and cleaned content",
			zap.String("url", pageURL),
			zap.String("title", processed.Title),
			zap.Int("length", processed.Length),
			zap.Float64("quality", processed.Quality),
		)

		// Create or update page record
		page, err := cr.pageRepo.Upsert(ctx, websiteID, normalizedURL)
		if err != nil {
			cr.logger.Error("Failed to upsert page", zap.String("url", pageURL), zap.Error(err))
			failureCount++
			cr.websiteRepo.IncrementPageCount(ctx, websiteID, false)
			return
		}

		// Generate content hash
		contentHash := hashContent(cleanedText)

		// Save content to Garage
		objectKey, err := cr.storage.SavePageContent(ctx, int(websiteID), normalizedURL, cleanedText)
		if err != nil {
			cr.logger.Error("Failed to save content to Garage", zap.String("url", pageURL), zap.Error(err))
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

		// Vectorize the content via job queue or directly
		if cr.jobClient != nil {
			// Enqueue vectorization job
			err := cr.jobClient.EnqueueVectorizePage(ctx, websiteID, page.ID, normalizedURL, cleanedText)
			if err != nil {
				cr.logger.Error("Failed to enqueue vectorization job",
					zap.String("url", pageURL),
					zap.Uint("pageID", page.ID),
					zap.Error(err),
				)
			} else {
				cr.logger.Debug("Enqueued vectorization job",
					zap.String("url", pageURL),
					zap.Uint("pageID", page.ID),
				)
			}
		} else {
			// Fallback: vectorize directly (async)
			go func() {
				err := cr.vectorizerSvc.ProcessPageContent(ctx, websiteID, page.ID, normalizedURL, cleanedText)
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
		}
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
		absoluteURL := e.Request.AbsoluteURL(link)

		// Normalize URL before checking robots.txt
		normalizedURL, err := contentprocessor.NormalizeURL(absoluteURL)
		if err != nil {
			cr.logger.Debug("Failed to normalize link URL", zap.String("url", absoluteURL), zap.Error(err))
			return
		}

		// Check if already visited
		if visitedURLs[normalizedURL] {
			return
		}

		// Check robots.txt before visiting
		allowed, err := cr.robotsEnforcer.CanFetch(ctx, normalizedURL)
		if err != nil {
			cr.logger.Warn("Error checking robots.txt, skipping URL",
				zap.String("url", normalizedURL),
				zap.Error(err),
			)
			return
		}

		if !allowed {
			cr.logger.Debug("URL disallowed by robots.txt",
				zap.String("url", normalizedURL),
			)
			return
		}

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

		// Check crawl delay from robots.txt
		crawlDelay, err := cr.robotsEnforcer.GetCrawlDelay(ctx, r.URL.String())
		if err == nil && crawlDelay > 0 {
			// If robots.txt specifies a delay, respect it
			if crawlDelay > time.Duration(cr.config.CrawlerDelayMS)*time.Millisecond {
				cr.logger.Debug("Respecting robots.txt crawl delay",
					zap.String("url", r.URL.String()),
					zap.Duration("delay", crawlDelay),
				)
				time.Sleep(crawlDelay)
			}
		}
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
