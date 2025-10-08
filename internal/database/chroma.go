package database

import (
	"hermit/internal/config"
	"log"

	chromago "github.com/amikos-tech/chroma-go"
)

// NewChromaDBClient creates a new ChromaDB client.
// It defaults to connecting to http://localhost:8000.
func NewChromaDBClient(cfg *config.Config) (*chromago.Client, error) {
	client, err := chromago.NewClient()
	if err != nil {
		log.Fatalf("Failed to create ChromaDB client: %v", err)
		return nil, err
	}

	log.Println("ChromaDB client initialized.")
	return client, nil
}
