package database

import (
	"hermit/internal/config"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// NewGarageClient creates a new Garage S3 client.
// Garage is S3-compatible, so we use the same minio-go library.
func NewGarageClient(cfg *config.Config) (*minio.Client, error) {
	garageClient, err := minio.New(cfg.GarageEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.GarageAccessKey, cfg.GarageSecretKey, ""),
		Secure: false, // Set to true if using TLS
		Region: cfg.GarageRegion,
	})
	if err != nil {
		log.Fatalf("Failed to connect to Garage: %v", err)
		return nil, err
	}

	log.Println("Successfully connected to Garage S3 storage.")
	return garageClient, nil
}
