package schema

import (
	"database/sql"
	"time"
)

// Page represents a crawled page in the database.
type Page struct {
	ID             uint           `db:"id"`
	WebsiteID      uint           `db:"website_id"`
	URL            string         `db:"url"`
	MinioObjectKey sql.NullString `db:"minio_object_key"`
	ContentHash    sql.NullString `db:"content_hash"`
	Status         string         `db:"status"`
	ErrorMessage   sql.NullString `db:"error_message"`
	CrawledAt      sql.NullTime   `db:"crawled_at"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}
