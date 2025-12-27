# Hermit

> An intelligent, self-improving web crawler and archiver that powers a custom-trained chat AI.

---

## About The Project

Hermit is a powerful service designed to monitor, crawl, and archive websites. It systematically scrapes content from target sites, processes it into a structured format, and stores it for analysis and retrieval. The ultimate goal is to use this archived data to power a custom, fine-tuned large language model, enabling a sophisticated, context-aware chat and query interface for the scraped content.

This project is built with a focus on a robust, production-grade architecture from the ground up.

### Key Features

*   **Modern Web UI:** Clean, dark-themed interface with session-based authentication
*   **AI-Powered Chat:** RAG-based query system with SSE streaming responses using Ollama
*   **Website Monitoring:** Add and manage websites with real-time crawl status tracking
*   **Intelligent Crawling:** `colly`-based crawler with robots.txt respect and content quality filtering
*   **API Key Management:** Secure API key creation, scoping, and revocation
*   **Job Monitoring:** Admin dashboard for background job queue visibility
*   **Modern Data Pipeline:** Garage (S3-compatible) storage with ChromaDB vector embeddings
*   **Robust Backend:** Go backend with clean architecture, DI using `uber-go/fx`
*   **Background Job Queue:** Asynchronous processing using **asynq** and **Redis**
*   **Content Processing:** Intelligent extraction using readability algorithms

## Technology Stack

### Backend
*   **Language:** Go 1.21+
*   **Framework:** Echo
*   **Database:** PostgreSQL (managed with `sqlx`)
*   **Object Storage:** Garage (S3-compatible)
*   **Vector Store:** ChromaDB
*   **Job Queue:** Redis + Asynq
*   **LLM & Embeddings:** Ollama (local inference)
*   **API Documentation:** Swagger (`echo-swagger`)
*   **Live Reloading:** Air
*   **Containerization:** Docker & Docker Compose

### Frontend
*   **Templates:** Templ (type-safe Go templates)
*   **Styling:** Tailwind CSS v4
*   **Interactivity:** HTMX + Alpine.js
*   **Auth:** Cookie-based sessions
*   **Streaming:** Server-Sent Events (SSE)

## Getting Started

This project is managed entirely through a comprehensive `Makefile`. Run `make help` to see all available commands.

### Prerequisites

*   Go 1.21+
*   Docker or Podman
*   `make`
*   Ollama (for local LLM inference)
*   Node.js (optional, for Tailwind CSS standalone binary download)

### Quick Start

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/yourusername/hermit.git
    cd hermit
    ```

2.  **Install Ollama models:**
    ```sh
    ollama pull mxbai-embed-large
    ollama pull llama3.1
    ```

3.  **First-time setup:**
    ```sh
    make setup
    ```
    This will:
    - Install required Go tools (templ, swag)
    - Start all services (PostgreSQL, Redis, Garage, ChromaDB, Ollama)
    - Run database migrations

4.  **Start development server:**
    ```sh
    make dev
    ```
    This starts the server with live-reload and all services.

5.  **Access the application:**
    - **Web UI:** http://localhost:8080
    - **API Docs:** http://localhost:8080/api/v1/swagger/index.html
    - **Health Check:** http://localhost:8080/api/v1/health

### Using the Web Interface

1.  **Register a new account:**
    - Navigate to http://localhost:8080/register
    - Create an account with email and password (min 8 characters)

2.  **Login:**
    - Navigate to http://localhost:8080/login
    - Enter your credentials
    - You'll be redirected to the chat interface

3.  **Add a website:**
    - Click "Websites" in the sidebar
    - Click "Add Website" button
    - Enter the URL, crawl depth, and max pages
    - The website will be queued for crawling

4.  **Chat with your data:**
    - Navigate to "Chat" in the sidebar
    - Ask questions about your indexed websites
    - Responses stream in real-time using SSE

5.  **Manage API keys:**
    - Navigate to "API Keys" in the sidebar
    - Create keys for programmatic access
    - Copy the key immediately (shown only once)
    - Use keys with `Authorization: Bearer hmt_xxxxx` header

6.  **Monitor jobs (admin only):**
    - Navigate to "Jobs" in the sidebar
    - View pending, processing, and completed jobs
    - Retry failed jobs or cancel pending ones

### Building for Production

1.  **Build all binaries:**
    ```sh
    make build
    ```

2.  **Run server:**
    ```sh
    ./bin/hermit
    ```

3.  **Run worker (separate terminal):**
    ```sh
    ./bin/worker
    ```

### Additional Commands

*   **Set up environment variables:**
    Copy the `.env.example` file to `.env`. The default values are configured to work with the `docker-compose.yml` file and do not need to be changed for local development.

*   **First-Time Setup (alternative to Quick Start):**
    This command starts all backend services (Postgres, Garage, ChromaDB, Redis, Ollama) and runs the database migrations.
    ```sh
    make setup
    ```
    
    **Note:** If you pull the latest changes, run migrations to update your database schema:
    ```sh
    make migrate-up
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

6.  **Run the Worker (Job Processor):**
    In a separate terminal, start the worker to process background jobs (crawling, vectorization):
    ```sh
    ./hermit-worker
    ```
    Or build and run:
    ```sh
    go run ./cmd/worker
    ```

7.  **Access the API:**
    *   The API will be running at `http://localhost:8080`.
    *   API documentation (Swagger) is available at `http://localhost:8080/api/swagger/index.html`.
    *   Health check: `http://localhost:8080/api/health`

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
*   `POST /api/websites/{id}/query/stream` - Ask questions with streaming SSE response

**Job Management:**
*   `GET /api/jobs/queues` - List all job queues with statistics
*   `GET /api/jobs/pending?queue=crawl&limit=50` - List pending jobs
*   `GET /api/jobs/active?queue=crawl` - List running jobs
*   `GET /api/jobs/scheduled?queue=crawl` - List scheduled jobs
*   `GET /api/jobs/retry?queue=crawl` - List jobs pending retry
*   `GET /api/jobs/archived?queue=crawl` - List failed jobs
*   `POST /api/jobs/{id}/cancel?queue=crawl` - Cancel a job
*   `POST /api/jobs/{id}/retry?queue=crawl` - Retry a failed job
*   `POST /api/jobs/queues/{queue}/pause` - Pause a queue
*   `POST /api/jobs/queues/{queue}/resume` - Resume a queue

**Health & Monitoring:**
*   `GET /api/health` - Check health of all services (Postgres, Garage, ChromaDB, Ollama)

### Other Useful Commands

*   `make down`: Stop all running services.
*   `make logs`: Tail the logs from all running services.
*   `make migrate-up`: Run database migrations.
*   `make migrate-down`: Rollback last database migration.
*   `make test`: Run the test suite.
*   `make docs`: Manually regenerate the API documentation.
*   `make clean`: Clean up build artifacts and containers.

### Database Migrations

The project uses database migrations to manage schema changes. Migrations are located in `migrations/` directory.

**Available migrations:**
- `001_create_websites_table` - Creates websites table
- `002_create_pages_table` - Creates pages table
- `003_add_website_crawl_status` - Adds crawl status tracking
- `004_add_normalized_url` - Adds normalized URL with unique constraint for duplicate prevention

**Commands:**
- `make migrate-up` - Apply all pending migrations
- `make migrate-down` - Rollback the last migration
- `make migrate-status` - Check migration status

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
