package contentprocessor

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
	"go.uber.org/zap"
)

// ContentProcessor handles HTML content cleaning and text extraction.
type ContentProcessor struct {
	logger *zap.Logger
}

// NewContentProcessor creates a new ContentProcessor.
func NewContentProcessor(logger *zap.Logger) *ContentProcessor {
	return &ContentProcessor{
		logger: logger,
	}
}

// ProcessedContent represents the cleaned and processed content.
type ProcessedContent struct {
	Title       string
	Content     string
	Excerpt     string
	Byline      string
	Length      int
	Quality     float64
	IsReadable  bool
	CleanedHTML string
}

// ExtractMainContent extracts the main content from HTML, removing navigation, ads, etc.
func (p *ContentProcessor) ExtractMainContent(htmlContent string, pageURL string) (*ProcessedContent, error) {
	if htmlContent == "" {
		return nil, fmt.Errorf("HTML content is empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		p.logger.Warn("Failed to parse URL, using readability without URL context",
			zap.String("url", pageURL),
			zap.Error(err),
		)
		parsedURL = nil
	}

	// Create readability parser
	article, err := readability.FromReader(strings.NewReader(htmlContent), parsedURL)
	if err != nil {
		p.logger.Error("Readability parsing failed",
			zap.String("url", pageURL),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract text content using RenderText
	var textBuf bytes.Buffer
	err = article.RenderText(&textBuf)
	if err != nil {
		p.logger.Error("Failed to render text content",
			zap.String("url", pageURL),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to render text: %w", err)
	}
	textContent := textBuf.String()

	if textContent == "" {
		p.logger.Warn("No text content extracted, using fallback",
			zap.String("url", pageURL),
		)
		textContent = p.fallbackExtraction(htmlContent)
	}

	// Extract HTML content
	var htmlBuf bytes.Buffer
	err = article.RenderHTML(&htmlBuf)
	if err != nil {
		p.logger.Warn("Failed to render HTML content",
			zap.String("url", pageURL),
			zap.Error(err),
		)
	}

	// Calculate quality score (simple heuristic)
	length := len(textContent)
	quality := p.calculateQualityScore(textContent, length)

	processed := &ProcessedContent{
		Title:       article.Title(),
		Content:     textContent,
		Excerpt:     article.Excerpt(),
		Byline:      article.Byline(),
		Length:      length,
		Quality:     quality,
		IsReadable:  quality >= 0.3,
		CleanedHTML: htmlBuf.String(),
	}

	p.logger.Debug("Content processed",
		zap.String("url", pageURL),
		zap.String("title", processed.Title),
		zap.Int("length", processed.Length),
		zap.Float64("quality", processed.Quality),
		zap.Bool("readable", processed.IsReadable),
	)

	return processed, nil
}

// CleanText performs additional text cleaning and normalization.
func (p *ContentProcessor) CleanText(text string) string {
	// Remove excessive whitespace
	text = strings.TrimSpace(text)

	// Replace multiple spaces with single space
	text = strings.Join(strings.Fields(text), " ")

	// Remove common noise patterns
	text = p.removeNoisePatterns(text)

	return text
}

// calculateQualityScore calculates a simple quality score for the content.
func (p *ContentProcessor) calculateQualityScore(content string, length int) float64 {
	if length == 0 {
		return 0.0
	}

	score := 0.0

	// Length score (prefer 500-5000 chars)
	if length >= 500 && length <= 5000 {
		score += 0.4
	} else if length > 5000 {
		score += 0.3
	} else if length > 200 {
		score += 0.2
	}

	// Word count score
	words := strings.Fields(content)
	wordCount := len(words)
	if wordCount > 100 {
		score += 0.3
	} else if wordCount > 50 {
		score += 0.2
	}

	// Sentence structure score (rough estimate)
	sentences := strings.Count(content, ".") + strings.Count(content, "!") + strings.Count(content, "?")
	if sentences > 5 {
		score += 0.2
	}

	// Avoid score > 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// removeNoisePatterns removes common noise patterns from text.
func (p *ContentProcessor) removeNoisePatterns(text string) string {
	// Common noise patterns
	noisePatterns := []string{
		"Click here",
		"Read more",
		"Subscribe now",
		"Sign up",
		"Advertisement",
		"Cookie policy",
		"Privacy policy",
		"Terms of service",
	}

	cleaned := text
	for _, pattern := range noisePatterns {
		// Case-insensitive replacement
		cleaned = strings.ReplaceAll(cleaned, pattern, "")
		cleaned = strings.ReplaceAll(cleaned, strings.ToLower(pattern), "")
		cleaned = strings.ReplaceAll(cleaned, strings.ToUpper(pattern), "")
	}

	return cleaned
}

// fallbackExtraction provides a basic fallback if readability fails.
func (p *ContentProcessor) fallbackExtraction(htmlContent string) string {
	// Very basic HTML tag removal as fallback
	text := htmlContent

	// Remove script and style tags with content
	text = removeTagsWithContent(text, "script")
	text = removeTagsWithContent(text, "style")
	text = removeTagsWithContent(text, "noscript")

	// Remove HTML tags
	text = strings.ReplaceAll(text, "<", " <")
	text = strings.ReplaceAll(text, ">", "> ")

	// Simple tag removal (not perfect but works for fallback)
	var result strings.Builder
	inTag := false
	for _, char := range text {
		if char == '<' {
			inTag = true
			continue
		}
		if char == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(char)
		}
	}

	return strings.TrimSpace(result.String())
}

// removeTagsWithContent removes HTML tags and their content.
func removeTagsWithContent(html, tag string) string {
	openTag := "<" + tag
	closeTag := "</" + tag + ">"

	for {
		start := strings.Index(strings.ToLower(html), openTag)
		if start == -1 {
			break
		}

		end := strings.Index(strings.ToLower(html[start:]), closeTag)
		if end == -1 {
			break
		}

		end += start + len(closeTag)
		html = html[:start] + html[end:]
	}

	return html
}

// IsContentValid checks if the processed content meets minimum quality standards.
func (p *ContentProcessor) IsContentValid(content *ProcessedContent, minLength int, minQuality float64) bool {
	if content == nil {
		return false
	}

	if content.Length < minLength {
		p.logger.Debug("Content too short",
			zap.Int("length", content.Length),
			zap.Int("minLength", minLength),
		)
		return false
	}

	if content.Quality < minQuality {
		p.logger.Debug("Content quality too low",
			zap.Float64("quality", content.Quality),
			zap.Float64("minQuality", minQuality),
		)
		return false
	}

	return true
}

// ExtractMetadata extracts metadata from HTML content.
func (p *ContentProcessor) ExtractMetadata(htmlContent string) map[string]string {
	metadata := make(map[string]string)

	// Simple meta tag extraction (basic implementation)
	// In production, use a proper HTML parser like golang.org/x/net/html

	lines := strings.Split(htmlContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Extract title
		if strings.Contains(strings.ToLower(line), "<title>") {
			start := strings.Index(strings.ToLower(line), "<title>") + 7
			end := strings.Index(strings.ToLower(line), "</title>")
			if start > 0 && end > start {
				metadata["title"] = strings.TrimSpace(line[start:end])
			}
		}

		// Extract meta description
		if strings.Contains(strings.ToLower(line), `name="description"`) {
			if contentIdx := strings.Index(strings.ToLower(line), `content="`); contentIdx != -1 {
				start := contentIdx + 9
				if end := strings.Index(line[start:], `"`); end != -1 {
					metadata["description"] = strings.TrimSpace(line[start : start+end])
				}
			}
		}

		// Extract meta keywords
		if strings.Contains(strings.ToLower(line), `name="keywords"`) {
			if contentIdx := strings.Index(strings.ToLower(line), `content="`); contentIdx != -1 {
				start := contentIdx + 9
				if end := strings.Index(line[start:], `"`); end != -1 {
					metadata["keywords"] = strings.TrimSpace(line[start : start+end])
				}
			}
		}
	}

	return metadata
}
