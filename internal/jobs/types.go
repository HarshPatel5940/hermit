package jobs

import (
	"encoding/json"
	"fmt"
)

// Task types
const (
	TypeCrawlWebsite    = "crawl:website"
	TypeVectorizePage   = "vectorize:page"
	TypeRecrawlWebsite  = "recrawl:website"
	TypeCleanupOldPages = "cleanup:old_pages"
)

// CrawlWebsitePayload represents the payload for crawling a website.
type CrawlWebsitePayload struct {
	WebsiteID uint   `json:"website_id"`
	StartURL  string `json:"start_url"`
}

// NewCrawlWebsitePayload creates a new CrawlWebsitePayload.
func NewCrawlWebsitePayload(websiteID uint, startURL string) ([]byte, error) {
	payload := CrawlWebsitePayload{
		WebsiteID: websiteID,
		StartURL:  startURL,
	}
	return json.Marshal(payload)
}

// ParseCrawlWebsitePayload parses a CrawlWebsitePayload from bytes.
func ParseCrawlWebsitePayload(data []byte) (*CrawlWebsitePayload, error) {
	var payload CrawlWebsitePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal crawl payload: %w", err)
	}
	return &payload, nil
}

// VectorizePagePayload represents the payload for vectorizing a page.
type VectorizePagePayload struct {
	WebsiteID uint   `json:"website_id"`
	PageID    uint   `json:"page_id"`
	PageURL   string `json:"page_url"`
	Content   string `json:"content"`
}

// NewVectorizePagePayload creates a new VectorizePagePayload.
func NewVectorizePagePayload(websiteID, pageID uint, pageURL, content string) ([]byte, error) {
	payload := VectorizePagePayload{
		WebsiteID: websiteID,
		PageID:    pageID,
		PageURL:   pageURL,
		Content:   content,
	}
	return json.Marshal(payload)
}

// ParseVectorizePagePayload parses a VectorizePagePayload from bytes.
func ParseVectorizePagePayload(data []byte) (*VectorizePagePayload, error) {
	var payload VectorizePagePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vectorize payload: %w", err)
	}
	return &payload, nil
}

// RecrawlWebsitePayload represents the payload for recrawling a website.
type RecrawlWebsitePayload struct {
	WebsiteID uint `json:"website_id"`
}

// NewRecrawlWebsitePayload creates a new RecrawlWebsitePayload.
func NewRecrawlWebsitePayload(websiteID uint) ([]byte, error) {
	payload := RecrawlWebsitePayload{
		WebsiteID: websiteID,
	}
	return json.Marshal(payload)
}

// ParseRecrawlWebsitePayload parses a RecrawlWebsitePayload from bytes.
func ParseRecrawlWebsitePayload(data []byte) (*RecrawlWebsitePayload, error) {
	var payload RecrawlWebsitePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal recrawl payload: %w", err)
	}
	return &payload, nil
}

// CleanupOldPagesPayload represents the payload for cleaning up old pages.
type CleanupOldPagesPayload struct {
	WebsiteID  uint   `json:"website_id,omitempty"`
	DaysOld    int    `json:"days_old"`
	DeleteFrom string `json:"delete_from"` // "storage", "vectors", "both"
}

// NewCleanupOldPagesPayload creates a new CleanupOldPagesPayload.
func NewCleanupOldPagesPayload(websiteID uint, daysOld int, deleteFrom string) ([]byte, error) {
	payload := CleanupOldPagesPayload{
		WebsiteID:  websiteID,
		DaysOld:    daysOld,
		DeleteFrom: deleteFrom,
	}
	return json.Marshal(payload)
}

// ParseCleanupOldPagesPayload parses a CleanupOldPagesPayload from bytes.
func ParseCleanupOldPagesPayload(data []byte) (*CleanupOldPagesPayload, error) {
	var payload CleanupOldPagesPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cleanup payload: %w", err)
	}
	return &payload, nil
}
