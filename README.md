# Hermit

> An intelligent, self-improving web crawler and archiver that powers a custom-trained chat AI.

---

## About The Project

Hermit is a powerful service designed to monitor, crawl, and archive websites. It systematically scrapes content from target sites, processes it into a structured format, and stores it for analysis and retrieval. The ultimate goal is to use this archived data to power a custom, fine-tuned large language model, enabling a sophisticated, context-aware chat and query interface for the scraped content.

This project is built with a focus on a robust, production-grade architecture from the ground up.

### Key Features

*   **Website Monitoring:** Add websites to a persistent monitoring list.
*   **Intelligent Crawling:** A `colly`-based crawler navigates and scrapes all pages of a target domain.
*   **Modern Data Pipeline:** Scraped content is stored in **MinIO**, with vector embeddings managed by **ChromaDB** for similarity search.
*   **Robust Backend:** A Go backend built with a clean, dependency-injected architecture using `uber-go/fx`.
*   **AI-Powered Chat:** RAG-based query system using Ollama for local LLM inference.

## Technology Stack

*   **Backend:** Go
*   **Framework:** Echo
*   **Database:** PostgreSQL (managed with `sqlx`)
*   **Object Storage:** MinIO
*   **Vector Store:** ChromaDB
*   **LLM & Embeddings:** Ollama (local inference)
*   **API Documentation:** Swagger (`echo-swagger`)
*   **Live Reloading:** Air
*   **Containerization:** Docker & Docker Compose

## Getting Started

This project is managed entirely through a comprehensive `Makefile`. Run `make help` to see all available commands.

### Prerequisites

*   Go (latest version)
*   Docker or Podman
*   `make`
*   Ollama (for local LLM inference)

### Installation & Running

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/harshpatel5940/hermit.git
    cd hermit
    ```

2.  **Set up environment variables:**
    Copy the `.env.example` file to `.env`. The default values are configured to work with the `docker-compose.yml` file and do not need to be changed for local development.

3.  **First-Time Setup:**
    This command starts all backend services (Postgres, MinIO, ChromaDB, Ollama) and runs the database migrations.
    ```sh
    make setup
    ```

4.  **Pull Ollama Models:**
    After starting the services, pull the required models for embeddings and chat:
    ```sh
    ./scripts/setup-ollama.sh
    ```
    This will download:
    - `mxbai-embed-large` (~500MB) - For text embeddings
    - `llama3.1` (~4.7GB) - For chat responses

5.  **Run the application in Development Mode:**
    This is the main command for development. It starts all services and runs the application with live-reloading. It will automatically rebuild the app and regenerate API documentation when you save a Go file.
    ```sh
    make dev
    ```

6.  **Access the API:**
    *   The API will be running at `http://localhost:8080`.
    *   API documentation (Swagger) is available at `http://localhost:8080/api/swagger/index.html`.

### API Endpoints

**Website Management:**
*   `POST /api/websites` - Add a new website to monitor
*   `GET /api/websites` - List all monitored websites
*   `GET /api/websites/{id}/status` - Get crawl status and statistics
*   `POST /api/websites/{id}/recrawl` - Manually trigger re-crawl

**Pages & Content:**
*   `GET /api/websites/{id}/pages` - List all crawled pages for a website

**AI Chat (RAG):**
*   `POST /api/websites/{id}/query` - Ask questions about website content

### Other Useful Commands

*   `make down`: Stop all running services.
*   `make logs`: Tail the logs from all running services.
*   `make test`: Run the test suite.
*   `make docs`: Manually regenerate the API documentation.
*   `make clean`: Clean up build artifacts and containers.

### Configuration

All settings can be configured via environment variables in `.env` file:

```env
# Server
PORT=8080

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/hermit?sslmode=disable

# MinIO Object Storage
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# ChromaDB Vector Store
CHROMA_DB_URL=http://localhost:8000

# Ollama LLM
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=mxbai-embed-large
OLLAMA_LLM_MODEL=llama3.1

# Crawler Settings
CRAWLER_MAX_DEPTH=10
CRAWLER_MAX_PAGES=1000
CRAWLER_DELAY_MS=500
CRAWLER_USER_AGENT=Hermit Crawler/1.0

# RAG Settings
RAG_TOP_K=5
RAG_CONTEXT_CHUNKS=3
```

See `.env.example` for all available options.
