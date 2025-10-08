package repositories

import (
	"context"
	"hermit/internal/schema"

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
	query := `INSERT INTO websites (url, is_monitored) VALUES ($1, $2) RETURNING id, url, is_monitored, created_at, updated_at`

	var website schema.Website
	err := r.db.QueryRowxContext(ctx, query, url, true).StructScan(&website)
	if err != nil {
		return nil, err
	}

	return &website, nil
}

// List retrieves all websites from the database.
func (r *WebsiteRepository) List(ctx context.Context) ([]schema.Website, error) {
	var websites []schema.Website
	query := `SELECT id, url, is_monitored, created_at, updated_at FROM websites`

	err := r.db.SelectContext(ctx, &websites, query)
	if err != nil {
		return nil, err
	}

	return websites, nil
}
