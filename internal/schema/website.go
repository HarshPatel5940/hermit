package schema

import (
	"time"
)

// Website represents a website to be monitored in the database.
type Website struct {
	ID          uint      `db:"id"`
	URL         string    `db:"url"`
	IsMonitored bool      `db:"is_monitored"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
