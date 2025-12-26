package vectorizer

import (
	"regexp"
	"strings"
)

const (
	// ChunkSize defines the maximum size of each chunk in characters
	ChunkSize = 800
	// OverlapSize defines the overlap between chunks to maintain context
	OverlapSize = 100
)

// ChunkText splits text into overlapping chunks for better context preservation.
// Returns a slice of text chunks.
func ChunkText(text string) []string {
	// Clean and normalize text
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return []string{}
	}

	// Split by sentence boundaries (., !, ?)
	sentenceRegex := regexp.MustCompile(`[.!?]+\s+`)
	sentences := sentenceRegex.Split(text, -1)

	var chunks []string
	var currentChunk strings.Builder
	currentLen := 0

	for _, sent := range sentences {
		sent = strings.TrimSpace(sent)
		if len(sent) == 0 {
			continue
		}

		sentLen := len(sent)

		// If adding this sentence exceeds chunk size, finalize current chunk
		if currentLen+sentLen > ChunkSize && currentLen > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))

			// Create overlap for context preservation
			chunkStr := currentChunk.String()
			overlapStart := len(chunkStr) - OverlapSize
			if overlapStart < 0 {
				overlapStart = 0
			}

			// Find last space in overlap region for clean break
			lastSpace := strings.LastIndex(chunkStr[overlapStart:], " ")
			if lastSpace > 0 {
				overlapStart += lastSpace + 1
			}

			// Reset with overlap
			currentChunk.Reset()
			if overlapStart < len(chunkStr) {
				currentChunk.WriteString(chunkStr[overlapStart:])
				currentLen = len(chunkStr) - overlapStart
			} else {
				currentLen = 0
			}
		}

		// Add sentence to current chunk
		if currentLen > 0 {
			currentChunk.WriteString(" ")
			currentLen++
		}
		currentChunk.WriteString(sent)
		currentLen += sentLen
	}

	// Add final chunk if any content remains
	if currentLen > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	// If no chunks were created (e.g., no sentence boundaries), split by character limit
	if len(chunks) == 0 && len(text) > 0 {
		for i := 0; i < len(text); i += ChunkSize - OverlapSize {
			end := i + ChunkSize
			if end > len(text) {
				end = len(text)
			}
			chunks = append(chunks, text[i:end])
		}
	}

	return chunks
}

// ChunkWithMetadata represents a text chunk with its metadata.
type ChunkWithMetadata struct {
	Text  string
	Index int
	Start int
	End   int
}

// ChunkTextWithMetadata splits text into chunks and returns metadata for each chunk.
func ChunkTextWithMetadata(text string) []ChunkWithMetadata {
	chunks := ChunkText(text)
	result := make([]ChunkWithMetadata, len(chunks))

	currentPos := 0
	for i, chunk := range chunks {
		result[i] = ChunkWithMetadata{
			Text:  chunk,
			Index: i,
			Start: currentPos,
			End:   currentPos + len(chunk),
		}
		currentPos += len(chunk)
	}

	return result
}
