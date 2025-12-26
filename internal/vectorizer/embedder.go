package vectorizer

import (
	"context"
	"fmt"

	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

// Embedder handles generating embeddings using Ollama.
type Embedder struct {
	client *api.Client
	model  string
	logger *zap.Logger
}

// NewEmbedder creates a new Embedder service.
// model should be the Ollama model name (e.g., "mxbai-embed-large", "nomic-embed-text")
func NewEmbedder(ollamaURL string, model string, logger *zap.Logger) *Embedder {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		logger.Warn("Failed to create Ollama client from environment, using default", zap.Error(err))
		// Fallback to default URL
		client = &api.Client{}
	}

	return &Embedder{
		client: client,
		model:  model,
		logger: logger,
	}
}

// EmbedText generates an embedding for a single text string.
// Returns the embedding vector and any error.
func (e *Embedder) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("cannot embed empty text")
	}

	req := &api.EmbedRequest{
		Model: e.model,
		Input: text,
	}

	resp, err := e.client.Embed(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("embedding failed: %w", err)
	}

	if len(resp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned from Ollama")
	}

	// Convert []float64 to []float32 for ChromaDB compatibility
	embedding := make([]float32, len(resp.Embeddings[0]))
	for i, v := range resp.Embeddings[0] {
		embedding[i] = float32(v)
	}

	e.logger.Debug("Generated embedding",
		zap.String("model", e.model),
		zap.Int("dimensions", len(embedding)),
		zap.Int("textLength", len(text)),
	)

	return embedding, nil
}

// EmbedChunks generates embeddings for multiple text chunks.
// Returns a slice of embedding vectors and any error.
func (e *Embedder) EmbedChunks(ctx context.Context, chunks []string) ([][]float32, error) {
	if len(chunks) == 0 {
		return nil, fmt.Errorf("no chunks provided")
	}

	embeddings := make([][]float32, len(chunks))

	for i, chunk := range chunks {
		embedding, err := e.EmbedText(ctx, chunk)
		if err != nil {
			e.logger.Error("Failed to embed chunk",
				zap.Int("chunkIndex", i),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}
		embeddings[i] = embedding

		e.logger.Debug("Embedded chunk",
			zap.Int("chunkIndex", i),
			zap.Int("chunkSize", len(chunk)),
			zap.Int("embeddingDimensions", len(embedding)),
		)
	}

	e.logger.Info("Successfully embedded all chunks",
		zap.Int("totalChunks", len(chunks)),
		zap.Int("dimensions", len(embeddings[0])),
	)

	return embeddings, nil
}

// GetModelInfo retrieves information about the current embedding model.
func (e *Embedder) GetModelInfo(ctx context.Context) (*api.ShowResponse, error) {
	req := &api.ShowRequest{
		Model: e.model,
	}

	resp, err := e.client.Show(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get model info: %w", err)
	}

	return resp, nil
}
