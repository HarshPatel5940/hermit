package crawler

import (
	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
)

// Crawler manages the website crawling process.
type Crawler struct {
	logger *zap.Logger
}

// NewCrawler creates a new Crawler service.
func NewCrawler(logger *zap.Logger) *Crawler {
	return &Crawler{logger: logger}
}

// Crawl starts the crawling process for a given URL.
func (cr *Crawler) Crawl(url string) {
	cr.logger.Info("Crawling started", zap.String("url", url))

	c := colly.NewCollector(
		// We'll only visit domains of the starting URL
		colly.AllowedDomains(url),
	)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		e.Request.Visit(e.Attr("href"))
	})

	c.OnRequest(func(r *colly.Request) {
		cr.logger.Info("Visiting", zap.String("url", r.URL.String()))
	})

	c.Visit(url)
}

