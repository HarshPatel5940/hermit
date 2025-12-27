package contentprocessor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
	"go.uber.org/zap"
)

// RobotsEnforcer handles robots.txt parsing and enforcement.
type RobotsEnforcer struct {
	logger      *zap.Logger
	cache       map[string]*robotsCacheEntry
	cacheMutex  sync.RWMutex
	userAgent   string
	httpTimeout time.Duration
}

// robotsCacheEntry represents a cached robots.txt entry.
type robotsCacheEntry struct {
	data      *robotstxt.RobotsData
	expiresAt time.Time
}

// NewRobotsEnforcer creates a new RobotsEnforcer.
func NewRobotsEnforcer(userAgent string, logger *zap.Logger) *RobotsEnforcer {
	return &RobotsEnforcer{
		logger:      logger,
		cache:       make(map[string]*robotsCacheEntry),
		userAgent:   userAgent,
		httpTimeout: 10 * time.Second,
	}
}

// CanFetch checks if the given URL can be crawled according to robots.txt.
func (r *RobotsEnforcer) CanFetch(ctx context.Context, pageURL string) (bool, error) {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return false, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Get robots.txt for this domain
	robotsData, err := r.getRobotsData(ctx, parsedURL)
	if err != nil {
		// If robots.txt cannot be fetched, allow crawling by default
		r.logger.Warn("Failed to fetch robots.txt, allowing by default",
			zap.String("url", pageURL),
			zap.Error(err),
		)
		return true, nil
	}

	// Check if URL is allowed
	group := robotsData.FindGroup(r.userAgent)
	if group == nil {
		// No specific rules for this user agent, allow
		return true, nil
	}

	allowed := group.Test(parsedURL.Path)

	r.logger.Debug("Robots.txt check",
		zap.String("url", pageURL),
		zap.String("userAgent", r.userAgent),
		zap.Bool("allowed", allowed),
	)

	return allowed, nil
}

// GetCrawlDelay returns the crawl delay specified in robots.txt.
func (r *RobotsEnforcer) GetCrawlDelay(ctx context.Context, pageURL string) (time.Duration, error) {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return 0, fmt.Errorf("failed to parse URL: %w", err)
	}

	robotsData, err := r.getRobotsData(ctx, parsedURL)
	if err != nil {
		// Return 0 if robots.txt cannot be fetched
		return 0, nil
	}

	group := robotsData.FindGroup(r.userAgent)
	if group == nil {
		return 0, nil
	}

	delay := group.CrawlDelay
	if delay > 0 {
		r.logger.Debug("Robots.txt crawl delay found",
			zap.String("domain", parsedURL.Host),
			zap.Duration("delay", delay),
		)
	}

	return delay, nil
}

// getRobotsData fetches and parses robots.txt for a domain.
func (r *RobotsEnforcer) getRobotsData(ctx context.Context, pageURL *url.URL) (*robotstxt.RobotsData, error) {
	domain := pageURL.Scheme + "://" + pageURL.Host

	// Check cache first
	r.cacheMutex.RLock()
	entry, exists := r.cache[domain]
	r.cacheMutex.RUnlock()

	if exists && time.Now().Before(entry.expiresAt) {
		r.logger.Debug("Using cached robots.txt",
			zap.String("domain", domain),
		)
		return entry.data, nil
	}

	// Fetch robots.txt
	robotsURL := domain + "/robots.txt"
	r.logger.Debug("Fetching robots.txt",
		zap.String("url", robotsURL),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", r.userAgent)

	client := &http.Client{
		Timeout: r.httpTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Limit redirects
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch robots.txt: %w", err)
	}
	defer resp.Body.Close()

	// Parse robots.txt
	var robotsData *robotstxt.RobotsData

	if resp.StatusCode == http.StatusOK {
		robotsData, err = robotstxt.FromResponse(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse robots.txt: %w", err)
		}

		r.logger.Info("Successfully fetched robots.txt",
			zap.String("domain", domain),
		)
	} else if resp.StatusCode == http.StatusNotFound {
		// No robots.txt means everything is allowed
		robotsData = &robotstxt.RobotsData{}
		r.logger.Debug("No robots.txt found, allowing all",
			zap.String("domain", domain),
		)
	} else {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Cache the result (cache for 24 hours)
	r.cacheMutex.Lock()
	r.cache[domain] = &robotsCacheEntry{
		data:      robotsData,
		expiresAt: time.Now().Add(24 * time.Hour),
	}
	r.cacheMutex.Unlock()

	return robotsData, nil
}

// ClearCache clears the robots.txt cache.
func (r *RobotsEnforcer) ClearCache() {
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	r.cache = make(map[string]*robotsCacheEntry)
	r.logger.Info("Robots.txt cache cleared")
}

// ClearDomainCache clears the cache for a specific domain.
func (r *RobotsEnforcer) ClearDomainCache(pageURL string) error {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	domain := parsedURL.Scheme + "://" + parsedURL.Host

	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	delete(r.cache, domain)
	r.logger.Debug("Cleared robots.txt cache for domain",
		zap.String("domain", domain),
	)

	return nil
}

// NormalizeURL normalizes a URL for duplicate detection.
func NormalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Convert scheme and host to lowercase
	parsedURL.Scheme = strings.ToLower(parsedURL.Scheme)
	parsedURL.Host = strings.ToLower(parsedURL.Host)

	// Remove fragment
	parsedURL.Fragment = ""

	// Remove common tracking parameters
	if parsedURL.RawQuery != "" {
		query := parsedURL.Query()
		trackingParams := []string{
			"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
			"fbclid", "gclid", "mc_cid", "mc_eid",
			"ref", "source", "campaign",
		}
		for _, param := range trackingParams {
			query.Del(param)
		}
		parsedURL.RawQuery = query.Encode()
	}

	// Remove trailing slash for consistency (except for root path)
	path := parsedURL.Path
	if path != "/" && strings.HasSuffix(path, "/") {
		parsedURL.Path = strings.TrimSuffix(path, "/")
	}

	// Ensure root path has slash
	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	}

	return parsedURL.String(), nil
}

// GetSitemapURLs extracts URLs from a sitemap.xml.
func (r *RobotsEnforcer) GetSitemapURLs(ctx context.Context, sitemapURL string) ([]string, error) {
	r.logger.Info("Fetching sitemap",
		zap.String("url", sitemapURL),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", sitemapURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", r.userAgent)

	client := &http.Client{
		Timeout: r.httpTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sitemap returned status %d", resp.StatusCode)
	}

	// Parse sitemap XML (basic implementation)
	// For production, use encoding/xml for proper parsing
	var urls []string

	// This is a simplified parser - in production use proper XML parsing
	buf := new(strings.Builder)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read sitemap: %w", err)
	}
	buf.Write(body)

	content := buf.String()

	// Extract URLs from <loc> tags
	for {
		start := strings.Index(content, "<loc>")
		if start == -1 {
			break
		}
		start += 5
		end := strings.Index(content[start:], "</loc>")
		if end == -1 {
			break
		}

		urlStr := strings.TrimSpace(content[start : start+end])
		if urlStr != "" {
			urls = append(urls, urlStr)
		}

		content = content[start+end+6:]
	}

	r.logger.Info("Parsed sitemap",
		zap.String("url", sitemapURL),
		zap.Int("urlCount", len(urls)),
	)

	return urls, nil
}
