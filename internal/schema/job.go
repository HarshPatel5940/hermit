package schema

import (
	"time"
)

// Job represents a background job in the queue
type Job struct {
	ID          string     `db:"id" json:"id"`
	Type        string     `db:"type" json:"type"`
	Status      string     `db:"status" json:"status"`
	Payload     string     `db:"payload" json:"payload"`
	Error       string     `db:"error" json:"error,omitempty"`
	Progress    int        `db:"progress" json:"progress"`
	StartedAt   *time.Time `db:"started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}
