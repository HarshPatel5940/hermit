package repositories

import (
	"context"
	"database/sql"
	"hermit/internal/schema"
	"time"

	"github.com/jmoiron/sqlx"
)

// PageRepository handles database operations for pages.
type PageRepository struct {
	db *sqlx.DB
}

// NewPageRepository creates a new PageRepository.
func NewPageRepository(db *sqlx.DB) *PageRepository {
	return &PageRepository{db: db}
}

// DB returns the underlying database connection.
func (r *PageRepository) DB() *sqlx.DB {
	return r.db
}

// Create adds a new page to the database.
func (r *PageRepository) Create(ctx context.Context, websiteID uint, url string) (*schema.Page, error) {
	query := `
		INSERT INTO pages (website_id, url, normalized_url, status)
		VALUES ($1, $2, $2, $3)
		RETURNING id, website_id, url, minio_object_key, content_hash, status, error_message, crawled_at, created_at, updated_at
	`

	var page schema.Page
	err := r.db.QueryRowxContext(ctx, query, websiteID, url, "pending").StructScan(&page)
	if err != nil {
		return nil, err
	}

	return &page, nil
}

// Upsert creates or updates a page record.
func (r *PageRepository) Upsert(ctx context.Context, websiteID uint, url string) (*schema.Page, error) {
	query := `
		INSERT INTO pages (website_id, url, normalized_url, status)
		VALUES ($1, $2, $2, $3)
		ON CONFLICT (website_id, normalized_url)
		DO UPDATE SET url = EXCLUDED.url, updated_at = NOW()
		RETURNING id, website_id, url, minio_object_key, content_hash, status, error_message, crawled_at, created_at, updated_at
	`

	var page schema.Page
	err := r.db.QueryRowxContext(ctx, query, websiteID, url, "pending").StructScan(&page)
	if err != nil {
		return nil, err
	}

	return &page, nil
}

// UpdateSuccess updates a page with successful crawl data.
func (r *PageRepository) UpdateSuccess(ctx context.Context, pageID uint, minioObjectKey, contentHash string) error {
	query := `
		UPDATE pages
		SET minio_object_key = $1,
		    content_hash = $2,
		    status = $3,
		    crawled_at = $4,
		    updated_at = NOW()
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query, minioObjectKey, contentHash, "success", time.Now(), pageID)
	return err
}

// UpdateError updates a page with error information.
func (r *PageRepository) UpdateError(ctx context.Context, pageID uint, errorMessage string) error {
	query := `
		UPDATE pages
		SET status = $1,
		    error_message = $2,
		    updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, "error", errorMessage, pageID)
	return err
}

// GetByWebsiteID retrieves all pages for a specific website.
func (r *PageRepository) GetByWebsiteID(ctx context.Context, websiteID uint) ([]schema.Page, error) {
	var pages []schema.Page
	query := `
		SELECT id, website_id, url, minio_object_key, content_hash, status, error_message, crawled_at, created_at, updated_at
		FROM pages
		WHERE website_id = $1
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &pages, query, websiteID)
	if err != nil {
		return nil, err
	}

	return pages, nil
}

// GetByURL retrieves a page by website ID and URL.
func (r *PageRepository) GetByURL(ctx context.Context, websiteID uint, url string) (*schema.Page, error) {
	var page schema.Page
	query := `
		SELECT id, website_id, url, minio_object_key, content_hash, status, error_message, crawled_at, created_at, updated_at
		FROM pages
		WHERE website_id = $1 AND url = $2
	`

	err := r.db.QueryRowxContext(ctx, query, websiteID, url).StructScan(&page)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &page, nil
}

// List retrieves all pages with optional filtering.
func (r *PageRepository) List(ctx context.Context) ([]schema.Page, error) {
	var pages []schema.Page
	query := `
		SELECT id, website_id, url, minio_object_key, content_hash, status, error_message, crawled_at, created_at, updated_at
		FROM pages
		ORDER BY created_at DESC
	`

	err := r.db.SelectContext(ctx, &pages, query)
	if err != nil {
		return nil, err
	}

	return pages, nil
}
