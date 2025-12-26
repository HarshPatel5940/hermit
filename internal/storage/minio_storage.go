package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hermit/internal/config"
	"net/url"
	"path"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

// MinIOStorage handles storing crawled content in MinIO.
type MinIOStorage struct {
	client     *minio.Client
	bucketName string
	logger     *zap.Logger
}

// NewMinIOStorage creates a new MinIOStorage service.
func NewMinIOStorage(client *minio.Client, cfg *config.Config, logger *zap.Logger) *MinIOStorage {
	return &MinIOStorage{
		client:     client,
		bucketName: cfg.MinIOBucketName,
		logger:     logger,
	}
}

// EnsureBucket creates the bucket if it doesn't exist.
func (s *MinIOStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		s.logger.Info("Created MinIO bucket", zap.String("bucket", s.bucketName))
	}

	return nil
}

// SavePageContent saves the content of a crawled page to MinIO.
// Returns the object key where the content was stored.
func (s *MinIOStorage) SavePageContent(ctx context.Context, websiteID int, pageURL string, content string) (string, error) {
	// Generate a unique key for this page
	objectKey := s.generateObjectKey(websiteID, pageURL)

	// Convert content to bytes
	contentBytes := []byte(content)
	reader := bytes.NewReader(contentBytes)

	// Upload to MinIO
	_, err := s.client.PutObject(
		ctx,
		s.bucketName,
		objectKey,
		reader,
		int64(len(contentBytes)),
		minio.PutObjectOptions{
			ContentType: "text/plain",
			UserMetadata: map[string]string{
				"website-id": fmt.Sprintf("%d", websiteID),
				"page-url":   pageURL,
			},
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to upload content to MinIO: %w", err)
	}

	s.logger.Info("Saved page content to MinIO",
		zap.String("objectKey", objectKey),
		zap.String("url", pageURL),
		zap.Int("size", len(contentBytes)),
	)

	return objectKey, nil
}

// generateObjectKey creates a unique key for storing page content.
// Format: websites/<website_id>/<url_hash>.txt
func (s *MinIOStorage) generateObjectKey(websiteID int, pageURL string) string {
	// Parse URL to get a clean path
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		// Fallback to hash if URL parsing fails
		return fmt.Sprintf("websites/%d/%s.txt", websiteID, hashString(pageURL))
	}

	// Create a hash of the full URL for uniqueness
	urlHash := hashString(pageURL)

	// Use domain and path for organization
	domain := parsedURL.Host
	urlPath := parsedURL.Path
	if urlPath == "" || urlPath == "/" {
		urlPath = "index"
	} else {
		// Clean the path
		urlPath = path.Clean(urlPath)
		// Remove leading slash
		if len(urlPath) > 0 && urlPath[0] == '/' {
			urlPath = urlPath[1:]
		}
	}

	// Combine into object key
	return fmt.Sprintf("websites/%d/%s/%s_%s.txt", websiteID, domain, urlPath, urlHash[:8])
}

// hashString creates a SHA256 hash of a string.
func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// GetPageContent retrieves content from MinIO by object key.
func (s *MinIOStorage) GetPageContent(ctx context.Context, objectKey string) (string, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get object from MinIO: %w", err)
	}
	defer object.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(object)
	if err != nil {
		return "", fmt.Errorf("failed to read object content: %w", err)
	}

	return buf.String(), nil
}
