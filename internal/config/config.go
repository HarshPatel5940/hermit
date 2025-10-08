package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port            string
	DatabaseURL     string
	MinIOEndpoint   string
	MinIOAccessKey  string
	MinIOSecretKey  string
	MinIOBucketName string
	ChromaDBURL     string
}

// NewConfig creates a new Config struct
func NewConfig() *Config {
	if os.Getenv("APP_ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Println("No .env file found, using environment variables")
		}
	}

	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		MinIOEndpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey:  getEnv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey:  getEnv("MINIO_SECRET_KEY", ""),
		MinIOBucketName: getEnv("MINIO_BUCKET_NAME", "website-content"),
		ChromaDBURL:     getEnv("CHROMA_DB_URL", "http://localhost:8000"),
	}
}

// Simple helper function to read an environment variable or return a default value
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
