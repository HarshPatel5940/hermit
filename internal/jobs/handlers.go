package jobs

import (
	"context"
	"fmt"

	"hermit/internal/crawler"
	"hermit/internal/repositories"
	"hermit/internal/vectorizer"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// Handlers contains all job handlers.
type Handlers struct {
	logger      *zap.Logger
	crawler     *crawler.Crawler
	vectorizer  *vectorizer.Service
	websiteRepo *repositories.WebsiteRepository
	pageRepo    *repositories.PageRepository
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(
	logger *zap.Logger,
	crawler *crawler.Crawler,
	vectorizer *vectorizer.Service,
	websiteRepo *repositories.WebsiteRepository,
	pageRepo *repositories.PageRepository,
) *Handlers {
	return &Handlers{
		logger:      logger,
		crawler:     crawler,
		vectorizer:  vectorizer,
		websiteRepo: websiteRepo,
		pageRepo:    pageRepo,
	}
}

// HandleCrawlWebsite handles the crawl website task.
func (h *Handlers) HandleCrawlWebsite(ctx context.Context, task *asynq.Task) error {
	payload, err := ParseCrawlWebsitePayload(task.Payload())
	if err != nil {
		h.logger.Error("Failed to parse crawl payload", zap.Error(err))
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	h.logger.Info("Starting crawl job",
		zap.Uint("websiteID", payload.WebsiteID),
		zap.String("startURL", payload.StartURL),
	)

	// Execute the crawl (this is synchronous and will block)
	h.crawler.Crawl(payload.WebsiteID, payload.StartURL)

	h.logger.Info("Crawl job completed",
		zap.Uint("websiteID", payload.WebsiteID),
		zap.String("startURL", payload.StartURL),
	)

	return nil
}

// HandleVectorizePage handles the vectorize page task.
func (h *Handlers) HandleVectorizePage(ctx context.Context, task *asynq.Task) error {
	payload, err := ParseVectorizePagePayload(task.Payload())
	if err != nil {
		h.logger.Error("Failed to parse vectorize payload", zap.Error(err))
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	h.logger.Info("Starting vectorize job",
		zap.Uint("websiteID", payload.WebsiteID),
		zap.Uint("pageID", payload.PageID),
		zap.String("pageURL", payload.PageURL),
	)

	err = h.vectorizer.ProcessPageContent(
		ctx,
		payload.WebsiteID,
		payload.PageID,
		payload.PageURL,
		payload.Content,
	)
	if err != nil {
		h.logger.Error("Failed to vectorize page",
			zap.Uint("websiteID", payload.WebsiteID),
			zap.Uint("pageID", payload.PageID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to vectorize page: %w", err)
	}

	h.logger.Info("Vectorize job completed",
		zap.Uint("websiteID", payload.WebsiteID),
		zap.Uint("pageID", payload.PageID),
	)

	return nil
}

// HandleRecrawlWebsite handles the recrawl website task.
func (h *Handlers) HandleRecrawlWebsite(ctx context.Context, task *asynq.Task) error {
	payload, err := ParseRecrawlWebsitePayload(task.Payload())
	if err != nil {
		h.logger.Error("Failed to parse recrawl payload", zap.Error(err))
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	h.logger.Info("Starting recrawl job",
		zap.Uint("websiteID", payload.WebsiteID),
	)

	// Get website details
	website, err := h.websiteRepo.GetByID(ctx, payload.WebsiteID)
	if err != nil {
		h.logger.Error("Failed to get website",
			zap.Uint("websiteID", payload.WebsiteID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get website: %w", err)
	}

	// Execute the crawl
	h.crawler.Crawl(payload.WebsiteID, website.URL)

	h.logger.Info("Recrawl job completed",
		zap.Uint("websiteID", payload.WebsiteID),
	)

	return nil
}

// HandleCleanupOldPages handles the cleanup old pages task.
func (h *Handlers) HandleCleanupOldPages(ctx context.Context, task *asynq.Task) error {
	payload, err := ParseCleanupOldPagesPayload(task.Payload())
	if err != nil {
		h.logger.Error("Failed to parse cleanup payload", zap.Error(err))
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	h.logger.Info("Starting cleanup job",
		zap.Uint("websiteID", payload.WebsiteID),
		zap.Int("daysOld", payload.DaysOld),
		zap.String("deleteFrom", payload.DeleteFrom),
	)

	// Query pages older than X days
	// For now, implementing basic logic. Can be extended with:
	// - Date-based filtering from database
	// - Batch processing for large datasets
	// - Transaction support for atomicity

	deleteCount := 0
	errorCount := 0

	// Get all pages for the website
	var pages []struct {
		ID        uint
		ObjectKey string
	}

	query := `SELECT id, object_key FROM pages WHERE website_id = $1 AND object_key IS NOT NULL`
	err = h.pageRepo.DB().SelectContext(ctx, &pages, query, payload.WebsiteID)
	if err != nil {
		h.logger.Error("Failed to query pages for cleanup",
			zap.Uint("websiteID", payload.WebsiteID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to query pages: %w", err)
	}

	h.logger.Info("Found pages for potential cleanup",
		zap.Int("count", len(pages)),
		zap.Uint("websiteID", payload.WebsiteID),
	)

	// Note: Actual date-based filtering should be added when pages table has created_at/updated_at
	// For now, this provides the framework for cleanup operations

	for _, page := range pages {
		shouldDelete := false

		// Delete from storage if requested
		if payload.DeleteFrom == "storage" || payload.DeleteFrom == "both" {
			// Storage deletion would go here
			// Skipping actual deletion for safety - needs storage service integration
			h.logger.Debug("Would delete from storage",
				zap.Uint("pageID", page.ID),
				zap.String("objectKey", page.ObjectKey),
			)
			shouldDelete = true
		}

		// Delete from vector store if requested
		if payload.DeleteFrom == "vectors" || payload.DeleteFrom == "both" {
			// Vector deletion would go here
			// Needs ChromaDB integration to delete by page_id metadata
			h.logger.Debug("Would delete vectors",
				zap.Uint("pageID", page.ID),
			)
			shouldDelete = true
		}

		if shouldDelete {
			deleteCount++
		}
	}

	h.logger.Info("Cleanup job completed",
		zap.Uint("websiteID", payload.WebsiteID),
		zap.Int("pagesProcessed", len(pages)),
		zap.Int("markedForDeletion", deleteCount),
		zap.Int("errors", errorCount),
	)

	// Return info message for now since we're not doing actual deletion yet
	// This provides the framework - actual deletion should be carefully implemented with:
	// 1. Database transaction support
	// 2. Rollback on failure
	// 3. Audit logging
	// 4. Confirmation requirements for production use

	return nil
}
