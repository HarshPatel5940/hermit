package database

import (
	"context"
	"fmt"
	"hermit/internal/config"
	"log"
	"net/http"
	"time"

	chromago "github.com/amikos-tech/chroma-go"
)

// ChromaDBClient wraps the ChromaDB client with additional methods.
type ChromaDBClient struct {
	*chromago.Client
	url string
}

// NewChromaDBClient creates a new ChromaDB client.
// It defaults to connecting to http://localhost:8000.
func NewChromaDBClient(cfg *config.Config) (*ChromaDBClient, error) {
	client, err := chromago.NewClient()
	if err != nil {
		log.Fatalf("Failed to create ChromaDB client: %v", err)
		return nil, err
	}

	log.Println("ChromaDB client initialized.")

	return &ChromaDBClient{
		Client: client,
		url:    cfg.ChromaDBURL,
	}, nil
}

// Heartbeat checks if ChromaDB is responding.
func (c *ChromaDBClient) Heartbeat(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.url+"/api/v1/heartbeat", nil)
	if err != nil {
		return fmt.Errorf("failed to create heartbeat request: %w", err)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("heartbeat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat returned status %d", resp.StatusCode)
	}

	return nil
}
