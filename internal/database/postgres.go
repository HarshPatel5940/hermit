package database

import (
	"hermit/internal/config"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewPostgresDB creates a new PostgreSQL database connection using sqlx.
func NewPostgresDB(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to postgres database: %v", err)
		return nil, err
	}

	log.Println("Successfully connected to PostgreSQL database.")
	return db, nil
}
