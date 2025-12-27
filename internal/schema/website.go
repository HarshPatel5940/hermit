package schema

import (
	"database/sql"
	"time"

	"github.com/oklog/ulid/v2"
)

// Website represents a website to be monitored in the database.
type Website struct {
	ID                uint           `db:"id"`
	URL               string         `db:"url"`
	UserID            *ulid.ULID     `db:"user_id"`
	IsMonitored       bool           `db:"is_monitored"`
	CrawlStatus       string         `db:"crawl_status"`
	CrawlStartedAt    sql.NullTime   `db:"crawl_started_at"`
	CrawlCompletedAt  sql.NullTime   `db:"crawl_completed_at"`
	TotalPagesCrawled int            `db:"total_pages_crawled"`
	TotalPagesFailed  int            `db:"total_pages_failed"`
	LastError         sql.NullString `db:"last_error"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
}
