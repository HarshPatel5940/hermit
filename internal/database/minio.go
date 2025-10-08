package database

import (
	"hermit/internal/config"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// NewMinIOClient creates a new MinIO client.
func NewMinIOClient(cfg *config.Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: false, // Set to true if using TLS
	})
	if err != nil {
		log.Fatalf("Failed to connect to MinIO: %v", err)
		return nil, err
	}

	log.Println("Successfully connected to MinIO.")
	return minioClient, nil
}
