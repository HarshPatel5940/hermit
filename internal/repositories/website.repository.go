package repositories

import (
	"context"
	"database/sql"
	"hermit/internal/schema"
	"time"

	"github.com/jmoiron/sqlx"
)

// WebsiteRepository handles database operations for websites.
type WebsiteRepository struct {
	db *sqlx.DB
}

// NewWebsiteRepository creates a new WebsiteRepository.
func NewWebsiteRepository(db *sqlx.DB) *WebsiteRepository {
	return &WebsiteRepository{db: db}
}

// Create adds a new website to the database.
func (r *WebsiteRepository) Create(ctx context.Context, url string) (*schema.Website, error) {
	query := `
		INSERT INTO websites (url, is_monitored, crawl_status)
		VALUES ($1, $2, $3)
		RETURNING id, url, user_id, is_monitored, crawl_status, crawl_started_at, crawl_completed_at,
		          total_pages_crawled, total_pages_failed, last_error, created_at, updated_at
	`

	var website schema.Website
	err := r.db.QueryRowxContext(ctx, query, url, true, "idle").StructScan(&website)
	if err != nil {
		return nil, err
	}

	return &website, nil
}

// List retrieves all websites from the database.
func (r *WebsiteRepository) List(ctx context.Context) ([]schema.Website, error) {
	var websites []schema.Website
	query := `
		SELECT id, url, user_id, is_monitored, crawl_status, crawl_started_at, crawl_completed_at,
		       total_pages_crawled, total_pages_failed, last_error, created_at, updated_at
		FROM websites
	`

	err := r.db.SelectContext(ctx, &websites, query)
	if err != nil {
		return nil, err
	}

	return websites, nil
}

// GetByID retrieves a website by ID.
func (r *WebsiteRepository) GetByID(ctx context.Context, id uint) (*schema.Website, error) {
	var website schema.Website
	query := `
		SELECT id, url, user_id, is_monitored, crawl_status, crawl_started_at, crawl_completed_at,
		       total_pages_crawled, total_pages_failed, last_error, created_at, updated_at
		FROM websites
		WHERE id = $1
	`

	err := r.db.QueryRowxContext(ctx, query, id).StructScan(&website)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &website, nil
}

// Update updates a website in the database.
func (r *WebsiteRepository) Update(ctx context.Context, website *schema.Website) error {
	query := `
		UPDATE websites
		SET url = $1, user_id = $2, is_monitored = $3, crawl_status = $4,
		    crawl_started_at = $5, crawl_completed_at = $6,
		    total_pages_crawled = $7, total_pages_failed = $8,
		    last_error = $9, updated_at = NOW()
		WHERE id = $10
	`

	_, err := r.db.ExecContext(ctx, query,
		website.URL,
		website.UserID,
		website.IsMonitored,
		website.CrawlStatus,
		website.CrawlStartedAt,
		website.CrawlCompletedAt,
		website.TotalPagesCrawled,
		website.TotalPagesFailed,
		website.LastError,
		website.ID,
	)
	return err
}

// UpdateCrawlStatus updates the crawl status of a website.
func (r *WebsiteRepository) UpdateCrawlStatus(ctx context.Context, id uint, status string) error {
	query := `
		UPDATE websites
		SET crawl_status = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// StartCrawl marks a website as currently crawling.
func (r *WebsiteRepository) StartCrawl(ctx context.Context, id uint) error {
	query := `
		UPDATE websites
		SET crawl_status = 'crawling',
		    crawl_started_at = $1,
		    crawl_completed_at = NULL,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// CompleteCrawl marks a website crawl as completed with statistics.
func (r *WebsiteRepository) CompleteCrawl(ctx context.Context, id uint, totalPages, failedPages int) error {
	query := `
		UPDATE websites
		SET crawl_status = 'completed',
		    crawl_completed_at = $1,
		    total_pages_crawled = $2,
		    total_pages_failed = $3,
		    updated_at = NOW()
		WHERE id = $4
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), totalPages, failedPages, id)
	return err
}

// FailCrawl marks a website crawl as failed with error message.
func (r *WebsiteRepository) FailCrawl(ctx context.Context, id uint, errorMsg string) error {
	query := `
		UPDATE websites
		SET crawl_status = 'failed',
		    last_error = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, errorMsg, id)
	return err
}

// IncrementPageCount increments the total pages crawled counter.
func (r *WebsiteRepository) IncrementPageCount(ctx context.Context, id uint, success bool) error {
	var query string
	if success {
		query = `
			UPDATE websites
			SET total_pages_crawled = total_pages_crawled + 1,
			    updated_at = NOW()
			WHERE id = $1
		`
	} else {
		query = `
			UPDATE websites
			SET total_pages_failed = total_pages_failed + 1,
			    updated_at = NOW()
			WHERE id = $1
		`
	}

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
