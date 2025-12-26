package vectorizer

import (
	"context"
	"fmt"

	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/types"
	"go.uber.org/zap"
)

// ChromaRepository handles storing and querying vector embeddings in ChromaDB.
type ChromaRepository struct {
	client *chroma.Client
	logger *zap.Logger
}

// NewChromaRepository creates a new ChromaRepository.
func NewChromaRepository(chromaURL string, logger *zap.Logger) (*ChromaRepository, error) {
	client, err := chroma.NewClient(chroma.WithBasePath(chromaURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create ChromaDB client: %w", err)
	}

	logger.Info("Connected to ChromaDB", zap.String("url", chromaURL))

	return &ChromaRepository{
		client: client,
		logger: logger,
	}, nil
}

// getCollectionName generates a collection name for a website.
func (r *ChromaRepository) getCollectionName(websiteID uint) string {
	return fmt.Sprintf("website_%d", websiteID)
}

// EnsureCollection creates or retrieves a collection for a website.
func (r *ChromaRepository) EnsureCollection(ctx context.Context, websiteID uint) (*chroma.Collection, error) {
	collectionName := r.getCollectionName(websiteID)

	collection, err := r.client.GetCollection(ctx, collectionName, nil)
	if err != nil {
		// Collection doesn't exist, create it
		r.logger.Info("Creating new ChromaDB collection", zap.String("collection", collectionName))

		collection, err = r.client.CreateCollection(ctx, collectionName, map[string]interface{}{
			"hnsw:space": "cosine",
		}, true, nil, types.L2)
		if err != nil {
			return nil, fmt.Errorf("failed to create collection: %w", err)
		}

		r.logger.Info("Created ChromaDB collection", zap.String("collection", collectionName))
	}

	return collection, nil
}

// StoreChunks saves text chunks with their embeddings to ChromaDB.
func (r *ChromaRepository) StoreChunks(
	ctx context.Context,
	websiteID uint,
	pageID uint,
	pageURL string,
	chunks []string,
	embeddings [][]float32,
) error {
	if len(chunks) != len(embeddings) {
		return fmt.Errorf("chunks and embeddings length mismatch: %d vs %d", len(chunks), len(embeddings))
	}

	collection, err := r.EnsureCollection(ctx, websiteID)
	if err != nil {
		return err
	}

	// Prepare data for ChromaDB
	ids := make([]string, len(chunks))
	documents := make([]string, len(chunks))
	metadatas := make([]map[string]interface{}, len(chunks))
	embeddingTypes := make([]*types.Embedding, len(embeddings))

	for i, chunk := range chunks {
		// Generate unique ID for this chunk
		ids[i] = fmt.Sprintf("page_%d_chunk_%d", pageID, i)
		documents[i] = chunk

		// Convert float32 to float32[] for Embedding type
		embeddingFloat32 := make([]float32, len(embeddings[i]))
		for j, v := range embeddings[i] {
			embeddingFloat32[j] = v
		}
		embeddingTypes[i] = types.NewEmbeddingFromFloat32(embeddingFloat32)

		// Create metadata
		metadatas[i] = map[string]interface{}{
			"website_id":  websiteID,
			"page_id":     pageID,
			"page_url":    pageURL,
			"chunk_index": i,
			"chunk_size":  len(chunk),
		}
	}

	// Add documents to collection: Add(ctx, embeddings, metadatas, documents, ids)
	_, err = collection.Add(ctx, embeddingTypes, metadatas, documents, ids)
	if err != nil {
		return fmt.Errorf("failed to add documents to ChromaDB: %w", err)
	}

	r.logger.Info("Stored chunks in ChromaDB",
		zap.String("collection", r.getCollectionName(websiteID)),
		zap.Uint("websiteID", websiteID),
		zap.Uint("pageID", pageID),
		zap.Int("numChunks", len(chunks)),
	)

	return nil
}

// QueryResult represents a result from a similarity search.
type QueryResult struct {
	ID       string
	Document string
	Metadata map[string]interface{}
	Distance float32
}

// Query performs a similarity search using a query embedding.
func (r *ChromaRepository) Query(
	ctx context.Context,
	websiteID uint,
	queryEmbedding []float32,
	topK int,
) ([]QueryResult, error) {
	collection, err := r.client.GetCollection(ctx, r.getCollectionName(websiteID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Create Embedding type for query
	queryEmbeddingType := types.NewEmbeddingFromFloat32(queryEmbedding)

	// Query using QueryWithOptions for embedding-based search
	queryResults, err := collection.QueryWithOptions(
		ctx,
		types.WithQueryEmbedding(queryEmbeddingType),
		types.WithNResults(int32(topK)),
		types.WithInclude(types.IDocuments, types.IMetadatas, types.IDistances),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query ChromaDB: %w", err)
	}

	// Parse results
	var results []QueryResult

	if queryResults == nil || len(queryResults.Ids) == 0 {
		return results, nil
	}

	for i := 0; i < len(queryResults.Ids[0]); i++ {
		result := QueryResult{
			ID: queryResults.Ids[0][i],
		}

		if queryResults.Documents != nil && len(queryResults.Documents) > 0 && len(queryResults.Documents[0]) > i {
			result.Document = queryResults.Documents[0][i]
		}

		if queryResults.Metadatas != nil && len(queryResults.Metadatas) > 0 && len(queryResults.Metadatas[0]) > i {
			result.Metadata = queryResults.Metadatas[0][i]
		}

		if queryResults.Distances != nil && len(queryResults.Distances) > 0 && len(queryResults.Distances[0]) > i {
			result.Distance = float32(queryResults.Distances[0][i])
		}

		results = append(results, result)
	}

	r.logger.Info("Query completed",
		zap.String("collection", r.getCollectionName(websiteID)),
		zap.Int("resultsCount", len(results)),
	)

	return results, nil
}

// DeletePageChunks removes all chunks for a specific page.
func (r *ChromaRepository) DeletePageChunks(ctx context.Context, websiteID uint, pageID uint) error {
	collection, err := r.client.GetCollection(ctx, r.getCollectionName(websiteID), nil)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Query for all chunks belonging to this page
	where := map[string]interface{}{
		"page_id": pageID,
	}

	_, err = collection.Delete(ctx, nil, where, nil)
	if err != nil {
		return fmt.Errorf("failed to delete page chunks: %w", err)
	}

	r.logger.Info("Deleted page chunks",
		zap.String("collection", r.getCollectionName(websiteID)),
		zap.Uint("pageID", pageID),
	)

	return nil
}

// DeleteCollection removes an entire collection for a website.
func (r *ChromaRepository) DeleteCollection(ctx context.Context, websiteID uint) error {
	collectionName := r.getCollectionName(websiteID)

	_, err := r.client.DeleteCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	r.logger.Info("Deleted collection", zap.String("collection", collectionName))

	return nil
}

// GetCollectionCount returns the number of documents in a collection.
func (r *ChromaRepository) GetCollectionCount(ctx context.Context, websiteID uint) (int, error) {
	collection, err := r.client.GetCollection(ctx, r.getCollectionName(websiteID), nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get collection: %w", err)
	}

	count, err := collection.Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}

	return int(count), nil
}
