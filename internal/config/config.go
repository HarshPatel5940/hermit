package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port             string
	DatabaseURL      string
	GarageEndpoint   string
	GarageRegion     string
	GarageAccessKey  string
	GarageSecretKey  string
	GarageBucketName string
	ChromaDBURL      string
	OllamaURL        string
	OllamaModel      string
	OllamaLLMModel   string
	// Redis settings
	RedisURL      string
	RedisPassword string
	RedisDB       int
	// Crawler settings
	CrawlerMaxDepth      int
	CrawlerMaxPages      int
	CrawlerDelayMS       int
	CrawlerRespectRobots bool
	CrawlerUserAgent     string
	// RAG settings
	RAGTopK          int
	RAGContextChunks int
	// Content processing
	ContentMinLength  int
	ContentMinQuality float64
	// HTTP timeouts
	HTTPTimeout     int
	CrawlerTimeout  int
	OllamaTimeout   int
	ChromaDBTimeout int
	// Database connection pool
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime int // in minutes
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
		Port:             getEnv("PORT", "8080"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		GarageEndpoint:   getEnv("GARAGE_ENDPOINT", "localhost:3902"),
		GarageRegion:     getEnv("GARAGE_REGION", "garage"),
		GarageAccessKey:  getEnv("GARAGE_ACCESS_KEY", ""),
		GarageSecretKey:  getEnv("GARAGE_SECRET_KEY", ""),
		GarageBucketName: getEnv("GARAGE_BUCKET_NAME", "website-content"),
		ChromaDBURL:      getEnv("CHROMA_DB_URL", "http://localhost:8000"),
		OllamaURL:        getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:      getEnv("OLLAMA_MODEL", "mxbai-embed-large"),
		OllamaLLMModel:   getEnv("OLLAMA_LLM_MODEL", "llama3.1"),
		// Redis settings
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),
		// Crawler settings
		CrawlerMaxDepth:      getEnvInt("CRAWLER_MAX_DEPTH", 10),
		CrawlerMaxPages:      getEnvInt("CRAWLER_MAX_PAGES", 1000),
		CrawlerDelayMS:       getEnvInt("CRAWLER_DELAY_MS", 500),
		CrawlerRespectRobots: getEnvBool("CRAWLER_RESPECT_ROBOTS_TXT", true),
		CrawlerUserAgent:     getEnv("CRAWLER_USER_AGENT", "Hermit Crawler/1.0"),
		// RAG settings
		RAGTopK:          getEnvInt("RAG_TOP_K", 5),
		RAGContextChunks: getEnvInt("RAG_CONTEXT_CHUNKS", 3),
		// Content processing
		ContentMinLength:  getEnvInt("CONTENT_MIN_LENGTH", 100),
		ContentMinQuality: getEnvFloat("CONTENT_MIN_QUALITY", 0.3),
		// HTTP timeouts
		HTTPTimeout:     getEnvInt("HTTP_TIMEOUT", 30),
		CrawlerTimeout:  getEnvInt("CRAWLER_TIMEOUT", 60),
		OllamaTimeout:   getEnvInt("OLLAMA_TIMEOUT", 120),
		ChromaDBTimeout: getEnvInt("CHROMADB_TIMEOUT", 30),
		// Database connection pool
		DBMaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 5), // 5 minutes default
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

// getEnvFloat reads an environment variable as a float64 or returns a default value
func getEnvFloat(key string, defaultValue float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}
