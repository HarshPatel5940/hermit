package vectorizer

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Service orchestrates the vectorization pipeline.
// It handles chunking text, generating embeddings, and storing them in ChromaDB.
type Service struct {
	embedder   *Embedder
	chromaRepo *ChromaRepository
	logger     *zap.Logger
}

// NewService creates a new vectorization service.
func NewService(
	embedder *Embedder,
	chromaRepo *ChromaRepository,
	logger *zap.Logger,
) *Service {
	return &Service{
		embedder:   embedder,
		chromaRepo: chromaRepo,
		logger:     logger,
	}
}

// ProcessPageContent processes page content through the full vectorization pipeline.
// It chunks the text, generates embeddings, and stores them in ChromaDB.
func (s *Service) ProcessPageContent(
	ctx context.Context,
	websiteID uint,
	pageID uint,
	pageURL string,
	content string,
) error {
	s.logger.Info("Starting vectorization process",
		zap.Uint("websiteID", websiteID),
		zap.Uint("pageID", pageID),
		zap.String("pageURL", pageURL),
		zap.Int("contentLength", len(content)),
	)

	// Step 1: Chunk the text
	chunks := ChunkText(content)
	if len(chunks) == 0 {
		s.logger.Warn("No chunks generated from content",
			zap.Uint("pageID", pageID),
		)
		return fmt.Errorf("no chunks generated from content")
	}

	s.logger.Info("Text chunked",
		zap.Int("numChunks", len(chunks)),
		zap.Uint("pageID", pageID),
	)

	// Step 2: Generate embeddings for all chunks
	embeddings, err := s.embedder.EmbedChunks(ctx, chunks)
	if err != nil {
		s.logger.Error("Failed to generate embeddings",
			zap.Uint("pageID", pageID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	s.logger.Info("Embeddings generated",
		zap.Int("numEmbeddings", len(embeddings)),
		zap.Uint("pageID", pageID),
	)

	// Step 3: Store chunks and embeddings in ChromaDB
	err = s.chromaRepo.StoreChunks(ctx, websiteID, pageID, pageURL, chunks, embeddings)
	if err != nil {
		s.logger.Error("Failed to store chunks in ChromaDB",
			zap.Uint("pageID", pageID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to store chunks: %w", err)
	}

	s.logger.Info("Vectorization completed successfully",
		zap.Uint("websiteID", websiteID),
		zap.Uint("pageID", pageID),
		zap.Int("totalChunks", len(chunks)),
	)

	return nil
}

// QuerySimilarContent performs semantic search to find similar content.
func (s *Service) QuerySimilarContent(
	ctx context.Context,
	websiteID uint,
	query string,
	topK int,
) ([]QueryResult, error) {
	s.logger.Info("Querying similar content",
		zap.Uint("websiteID", websiteID),
		zap.String("query", query),
		zap.Int("topK", topK),
	)

	// Generate embedding for the query
	queryEmbedding, err := s.embedder.EmbedText(ctx, query)
	if err != nil {
		s.logger.Error("Failed to embed query",
			zap.String("query", query),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Query ChromaDB for similar chunks
	results, err := s.chromaRepo.Query(ctx, websiteID, queryEmbedding, topK)
	if err != nil {
		s.logger.Error("Failed to query ChromaDB",
			zap.Uint("websiteID", websiteID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to query ChromaDB: %w", err)
	}

	s.logger.Info("Query completed",
		zap.Uint("websiteID", websiteID),
		zap.Int("resultsFound", len(results)),
	)

	return results, nil
}

// DeletePageVectors removes all vectors for a specific page.
func (s *Service) DeletePageVectors(ctx context.Context, websiteID uint, pageID uint) error {
	s.logger.Info("Deleting page vectors",
		zap.Uint("websiteID", websiteID),
		zap.Uint("pageID", pageID),
	)

	err := s.chromaRepo.DeletePageChunks(ctx, websiteID, pageID)
	if err != nil {
		s.logger.Error("Failed to delete page vectors",
			zap.Uint("pageID", pageID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Page vectors deleted successfully",
		zap.Uint("pageID", pageID),
	)

	return nil
}

// DeleteWebsiteVectors removes all vectors for a website.
func (s *Service) DeleteWebsiteVectors(ctx context.Context, websiteID uint) error {
	s.logger.Info("Deleting website vectors",
		zap.Uint("websiteID", websiteID),
	)

	err := s.chromaRepo.DeleteCollection(ctx, websiteID)
	if err != nil {
		s.logger.Error("Failed to delete website vectors",
			zap.Uint("websiteID", websiteID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Website vectors deleted successfully",
		zap.Uint("websiteID", websiteID),
	)

	return nil
}

// GetWebsiteVectorCount returns the number of vectors stored for a website.
func (s *Service) GetWebsiteVectorCount(ctx context.Context, websiteID uint) (int, error) {
	count, err := s.chromaRepo.GetCollectionCount(ctx, websiteID)
	if err != nil {
		return 0, err
	}

	return count, nil
}
