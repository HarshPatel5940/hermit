package llm

import (
	"context"
	"fmt"
	"hermit/internal/vectorizer"

	"go.uber.org/zap"
)

// RAGService orchestrates the Retrieval-Augmented Generation pipeline.
type RAGService struct {
	vectorizerSvc *vectorizer.Service
	llm           *OllamaLLM
	logger        *zap.Logger
	topK          int
	contextChunks int
}

// NewRAGService creates a new RAG service.
func NewRAGService(
	vectorizerSvc *vectorizer.Service,
	llm *OllamaLLM,
	logger *zap.Logger,
	topK int,
	contextChunks int,
) *RAGService {
	return &RAGService{
		vectorizerSvc: vectorizerSvc,
		llm:           llm,
		logger:        logger,
		topK:          topK,
		contextChunks: contextChunks,
	}
}

// QueryResponse represents the response from a RAG query.
type QueryResponse struct {
	Answer          string        `json:"answer"`
	Sources         []QuerySource `json:"sources"`
	RetrievedChunks int           `json:"retrieved_chunks"`
	Query           string        `json:"query"`
}

// QuerySource represents a source document used in the answer.
type QuerySource struct {
	PageURL    string  `json:"page_url"`
	ChunkText  string  `json:"chunk_text"`
	ChunkIndex int     `json:"chunk_index"`
	Similarity float32 `json:"similarity"`
	PageID     uint    `json:"page_id"`
}

// Query performs a RAG query against a website's content.
func (s *RAGService) Query(ctx context.Context, websiteID uint, query string) (*QueryResponse, error) {
	s.logger.Info("Processing RAG query",
		zap.Uint("websiteID", websiteID),
		zap.String("query", query),
	)

	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Step 1: Retrieve similar chunks from ChromaDB
	results, err := s.vectorizerSvc.QuerySimilarContent(ctx, websiteID, query, s.topK)
	if err != nil {
		s.logger.Error("Failed to retrieve similar content",
			zap.Uint("websiteID", websiteID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve content: %w", err)
	}

	if len(results) == 0 {
		s.logger.Warn("No similar content found",
			zap.Uint("websiteID", websiteID),
			zap.String("query", query),
		)
		return &QueryResponse{
			Answer:          "I couldn't find any relevant information to answer your question. The website might not have been crawled yet, or there's no content matching your query.",
			Sources:         []QuerySource{},
			RetrievedChunks: 0,
			Query:           query,
		}, nil
	}

	s.logger.Info("Retrieved similar chunks",
		zap.Int("count", len(results)),
	)

	// Step 2: Extract context chunks (limit to configured amount)
	contextLimit := s.contextChunks
	if contextLimit > len(results) {
		contextLimit = len(results)
	}

	contextChunks := make([]string, contextLimit)
	sources := make([]QuerySource, len(results))

	for i := 0; i < len(results); i++ {
		result := results[i]

		// Add to context if within limit
		if i < contextLimit {
			contextChunks[i] = result.Document
		}

		// Build source information
		source := QuerySource{
			ChunkText:  result.Document,
			Similarity: 1.0 - result.Distance, // Convert distance to similarity
		}

		// Extract metadata
		if result.Metadata != nil {
			if pageURL, ok := result.Metadata["page_url"].(string); ok {
				source.PageURL = pageURL
			}
			if chunkIndex, ok := result.Metadata["chunk_index"].(float64); ok {
				source.ChunkIndex = int(chunkIndex)
			}
			if pageID, ok := result.Metadata["page_id"].(float64); ok {
				source.PageID = uint(pageID)
			}
		}

		sources[i] = source
	}

	// Step 3: Generate answer using LLM with context
	s.logger.Info("Generating LLM response",
		zap.Int("contextChunks", len(contextChunks)),
	)

	answer, err := s.llm.GenerateWithContext(ctx, query, contextChunks)
	if err != nil {
		s.logger.Error("Failed to generate LLM response",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	s.logger.Info("RAG query completed successfully",
		zap.Uint("websiteID", websiteID),
		zap.Int("answerLength", len(answer)),
	)

	return &QueryResponse{
		Answer:          answer,
		Sources:         sources,
		RetrievedChunks: len(results),
		Query:           query,
	}, nil
}

// QueryWithCustomContext allows custom context to be provided.
func (s *RAGService) QueryWithCustomContext(ctx context.Context, query string, context []string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("query cannot be empty")
	}

	answer, err := s.llm.GenerateWithContext(ctx, query, context)
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	return answer, nil
}

// QueryStream performs a streaming RAG query against a website's content.
// The callback is called for each chunk of the LLM response.
func (s *RAGService) QueryStream(ctx context.Context, websiteID uint, query string, callback func(chunk string) error) (*QueryStreamMeta, error) {
	s.logger.Info("Processing streaming RAG query",
		zap.Uint("websiteID", websiteID),
		zap.String("query", query),
	)

	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// Step 1: Retrieve similar chunks from ChromaDB
	results, err := s.vectorizerSvc.QuerySimilarContent(ctx, websiteID, query, s.topK)
	if err != nil {
		s.logger.Error("Failed to retrieve similar content",
			zap.Uint("websiteID", websiteID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve content: %w", err)
	}

	if len(results) == 0 {
		s.logger.Warn("No similar content found",
			zap.Uint("websiteID", websiteID),
			zap.String("query", query),
		)
		// Send a complete message for no results
		err := callback("I couldn't find any relevant information to answer your question. The website might not have been crawled yet, or there's no content matching your query.")
		if err != nil {
			return nil, err
		}
		return &QueryStreamMeta{
			Sources:         []QuerySource{},
			RetrievedChunks: 0,
			Query:           query,
		}, nil
	}

	s.logger.Info("Retrieved similar chunks for streaming",
		zap.Int("count", len(results)),
	)

	// Step 2: Extract context chunks and build sources
	contextLimit := s.contextChunks
	if contextLimit > len(results) {
		contextLimit = len(results)
	}

	contextChunks := make([]string, contextLimit)
	sources := make([]QuerySource, len(results))

	for i := 0; i < len(results); i++ {
		result := results[i]

		if i < contextLimit {
			contextChunks[i] = result.Document
		}

		source := QuerySource{
			ChunkText:  result.Document,
			Similarity: 1.0 - result.Distance,
		}

		if result.Metadata != nil {
			if pageURL, ok := result.Metadata["page_url"].(string); ok {
				source.PageURL = pageURL
			}
			if chunkIndex, ok := result.Metadata["chunk_index"].(float64); ok {
				source.ChunkIndex = int(chunkIndex)
			}
			if pageID, ok := result.Metadata["page_id"].(float64); ok {
				source.PageID = uint(pageID)
			}
		}

		sources[i] = source
	}

	// Step 3: Generate streaming answer using LLM with context
	s.logger.Info("Generating streaming LLM response",
		zap.Int("contextChunks", len(contextChunks)),
	)

	err = s.llm.GenerateWithContextStream(ctx, query, contextChunks, callback)
	if err != nil {
		s.logger.Error("Failed to generate streaming LLM response",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to generate streaming answer: %w", err)
	}

	s.logger.Info("Streaming RAG query completed successfully",
		zap.Uint("websiteID", websiteID),
	)

	return &QueryStreamMeta{
		Sources:         sources,
		RetrievedChunks: len(results),
		Query:           query,
	}, nil
}

// QueryStreamMeta represents metadata from a streaming RAG query.
type QueryStreamMeta struct {
	Sources         []QuerySource `json:"sources"`
	RetrievedChunks int           `json:"retrieved_chunks"`
	Query           string        `json:"query"`
}
