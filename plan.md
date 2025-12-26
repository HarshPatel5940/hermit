# Project Plan

This document outlines the development plan for the Hermit website monitoring and AI-powered chat application.

---

## Phase 1: Core API & Foundation (Complete)

- [x] Set up production-grade project structure with `fx` for dependency injection.
- [x] Configure connections to PostgreSQL, MinIO, and ChromaDB.
- [x] Implement the data access layer using `sqlx` for raw SQL control.
- [x] Create and apply database migrations using `goose`.
- [x] Build API endpoints (`POST`, `GET`) for managing monitored websites.
- [x] Integrate Swagger for automatic API documentation.
- [x] Configure a live-reloading development environment with `Air`.
- [x] Add a robust middleware stack (Logger, CORS, Secure, etc.).

---

## Phase 2: Crawler and Data Pipeline

This phase focuses on building the service that crawls websites and processes their content.

- [x] **Build the Crawler Service:**
    - [x] Add the `colly` library for web crawling.
    - [x] Create a `Crawler` service in `internal/crawler`.
    - [x] Implement logic to visit a URL, find same-domain links, and add them to a queue.
    - [x] Implement logic to extract text content from pages.
- [x] **Integrate Crawler with API:**
    - [x] Trigger the crawler asynchronously when a new website is added via the API.
- [x] **Build the Storage Pipeline:**
    - [x] Create a repository/service for interacting with MinIO.
    - [x] Save the extracted text content of each page to a MinIO bucket.
    - [x] Create pages table migration to track crawled pages.
    - [x] Create Page schema and PageRepository for database operations.
    - [x] Update Crawler to save content to MinIO and track in database.
    - [x] Add API endpoint to retrieve pages for a website.
- [x] **Build the Vectorization Pipeline:**
    - [x] Create a service to handle text embedding (using Ollama locally).
    - [x] Implement logic to split text into chunks with overlap.
    - [x] Use Ollama embedding model (mxbai-embed-large) to create vector embeddings for each chunk.
    - [x] Create a repository/service for interacting with ChromaDB.
    - [x] Save the vector embeddings and content metadata to ChromaDB.
    - [x] Integrate vectorization into crawler pipeline (async processing).
    - [x] Add configuration for Ollama URL and model selection.

---

## Phase 2.5: Backend Completion (Complete)

All critical and essential backend features have been implemented.

### Critical Features (Must Have)

- [x] **RAG Query Endpoint:**
    - [x] Create `POST /websites/{id}/query` endpoint for AI chat queries.
    - [x] Implement query embedding generation.
    - [x] Perform ChromaDB similarity search to retrieve relevant chunks.
    - [x] Integrate Ollama LLM service for text generation (e.g., llama3.1).
    - [x] Build RAG pipeline: query → embed → search → context → LLM → response.
    - [x] Add proper error handling and response formatting.

### Nice-to-Have Features (Highly Valuable)

- [x] **Crawl Management:**
    - [x] Add `GET /websites/{id}/status` endpoint to check crawl progress.
    - [x] Add `POST /websites/{id}/recrawl` endpoint to manually trigger re-crawling.
    - [x] Track crawl statistics (total pages, successful, failed, in-progress).
    - [x] Add crawl state management (idle, crawling, completed, failed).

- [x] **Crawler Enhancements:**
    - [x] Implement rate limiting to avoid overwhelming target websites.
    - [x] Add configurable delay between requests.
    - [x] Add user-agent configuration.
    - [x] Add max depth and max pages limits.
    - [ ] Add user-agent rotation.
    - [ ] Implement robots.txt respect.

- [x] **Configuration & Infrastructure:**
    - [x] Add all services to docker-compose.yml (including Ollama).
    - [x] Create .env.example with all configuration options.
    - [x] Make all settings configurable via environment variables.
    - [x] Add crawler configuration (max depth, max pages, delay, user agent).
    - [x] Add RAG configuration (top K, context chunks).

- [ ] **API Improvements (Future Enhancements):**
    - [ ] Add pagination to `GET /websites/{id}/pages` endpoint.
    - [ ] Add filtering by page status.
    - [ ] Improve error responses with proper status codes and messages.
    - [ ] Add request validation middleware.

- [ ] **Monitoring & Observability (Future Enhancements):**
    - [ ] Add metrics endpoint for crawl stats.
    - [ ] Add health check for Ollama and ChromaDB connections.
    - [ ] Improve structured logging for debugging.

---

## Phase 3: AI-Powered Chat Interface (Ready to Start)

Backend is complete. Ready to build the frontend chat interface.

- [ ] **Develop a Simple Frontend:**
    - [ ] Create a basic web page with a chat interface.
    - [ ] Use HTMX or simple JavaScript to call the Q&A API endpoint and display the results.
    - [ ] Add website management UI (list, add, status).
    - [ ] Display crawl progress and statistics.
    - [ ] Show source citations for AI answers.

---

## Summary of Completed Work

### Phase 1 ✅
- Production-grade Go backend with fx DI
- PostgreSQL + sqlx data layer
- MinIO and ChromaDB connections
- Swagger API documentation
- Docker Compose setup

### Phase 2 ✅
- Colly-based web crawler with rate limiting
- MinIO storage pipeline
- PostgreSQL page tracking
- Text chunking (800 chars, 100 overlap)
- Ollama embedding integration (mxbai-embed-large)
- ChromaDB vector storage

### Phase 2.5 ✅
- RAG query endpoint with Ollama LLM (llama3.1)
- Crawl status tracking and management
- Re-crawl trigger endpoint
- Full environment-based configuration
- Setup scripts for Ollama models
- Comprehensive documentation

**Total API Endpoints:** 7
- POST /api/websites
- GET /api/websites
- GET /api/websites/{id}/status
- POST /api/websites/{id}/recrawl
- GET /api/websites/{id}/pages
- POST /api/websites/{id}/query

**All services dockerized and configurable via .env**
