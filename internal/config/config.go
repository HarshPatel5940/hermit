package config

import (
	"log"
	"os"
	"strconv"

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
	OllamaURL       string
	OllamaModel     string
	OllamaLLMModel  string
	// Crawler settings
	CrawlerMaxDepth      int
	CrawlerMaxPages      int
	CrawlerDelayMS       int
	CrawlerRespectRobots bool
	CrawlerUserAgent     string
	// RAG settings
	RAGTopK          int
	RAGContextChunks int
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
		OllamaURL:       getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:     getEnv("OLLAMA_MODEL", "mxbai-embed-large"),
		OllamaLLMModel:  getEnv("OLLAMA_LLM_MODEL", "llama3.1"),
		// Crawler settings
		CrawlerMaxDepth:      getEnvInt("CRAWLER_MAX_DEPTH", 10),
		CrawlerMaxPages:      getEnvInt("CRAWLER_MAX_PAGES", 1000),
		CrawlerDelayMS:       getEnvInt("CRAWLER_DELAY_MS", 500),
		CrawlerRespectRobots: getEnvBool("CRAWLER_RESPECT_ROBOTS_TXT", true),
		CrawlerUserAgent:     getEnv("CRAWLER_USER_AGENT", "Hermit Crawler/1.0"),
		// RAG settings
		RAGTopK:          getEnvInt("RAG_TOP_K", 5),
		RAGContextChunks: getEnvInt("RAG_CONTEXT_CHUNKS", 3),
	}
}

// Simple helper function to read an environment variable or return a default value
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt reads an environment variable as an integer or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool reads an environment variable as a boolean or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
