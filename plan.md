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
    - [x] Implement robots.txt respect and parsing.
    - [x] Respect crawl-delay from robots.txt.
    - [x] URL normalization for duplicate detection.
    - [ ] Add user-agent rotation (future).

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

---

## Phase 2.6: Production Backend Features (Complete)

This phase focused on making the backend production-ready with essential services.

### Storage Migration
- [x] **MinIO → Garage Migration:**
    - [x] Replace MinIO with Garage (S3-compatible, AGPL-licensed).
    - [x] Update storage layer to use Garage client.
    - [x] Update docker-compose.yml with Garage service.
    - [x] Update configuration for Garage endpoints and credentials.

### Content Processing
- [x] **Content Quality Improvements:**
    - [x] Implement ContentProcessor using go-readability for main content extraction.
    - [x] Remove navigation, ads, and boilerplate from crawled content.
    - [x] Add content quality scoring.
    - [x] Add minimum content length and quality thresholds.
    - [x] Implement text cleaning and normalization.

### Robots.txt & URL Management
- [x] **RobotsEnforcer Service:**
    - [x] Parse and enforce robots.txt rules.
    - [x] Cache robots.txt with expiration.
    - [x] Respect crawl-delay directives.
    - [x] URL normalization (remove tracking params, fragments).
    - [x] Basic sitemap.xml parsing support.
    - [x] Integrate into crawler pipeline.

### Background Job Queue
- [x] **Asynq Job Queue Integration:**
    - [x] Add Redis to docker-compose.yml.
    - [x] Create job types (CrawlWebsite, VectorizePage, RecrawlWebsite, CleanupOldPages).
    - [x] Implement job client for enqueueing tasks.
    - [x] Implement job handlers for processing tasks.
    - [x] Create job server/worker for background processing.
    - [x] Update crawler to enqueue vectorization jobs.
    - [x] Update API controllers to enqueue crawl jobs instead of running directly.
    - [x] Create separate worker command (cmd/worker).
    - [x] Add retry logic and queue priorities.

### Health & Monitoring
- [x] **Health Check Endpoint:**
    - [x] Create HealthController with service checks.
    - [x] Add PostgreSQL health check.
    - [x] Add Garage (S3) health check.
    - [x] Add ChromaDB health check.
    - [x] Add Ollama health check.
    - [x] Return service status and latency metrics.
    - [x] Register GET /api/health endpoint.

### Infrastructure
- [x] **Docker Compose Updates:**
    - [x] Add Redis service for job queue.
    - [x] Replace MinIO with Garage.
    - [x] Configure service dependencies.
    - [x] Update environment variables.

- [x] **Database Constraints:**
    - [x] Add unique constraint on normalized URLs.
    - [x] Create migration for URL normalization column.
    - [x] Implement DB-level duplicate prevention.
    - [x] Update PageRepository to use normalized_url.

- [x] **Database Connection Pooling:**
    - [x] Configure SetMaxOpenConns for connection pool.
    - [x] Configure SetMaxIdleConns for idle connections.
    - [x] Configure SetConnMaxLifetime for connection recycling.
    - [x] Add configuration options to .env.example.
    - [x] Document optimal settings in config.

- [x] **Cleanup Job Implementation:**
    - [x] Implement framework for cleanup job logic.
    - [x] Query pages from database.
    - [x] Add support for storage and vector deletion.
    - [x] Add logging and error tracking.
    - [x] Framework ready for production use with safety checks.

---

## Phase 2.7: Production Hardening & Advanced Features (Future)

This phase covers authentication, security, monitoring, and scaling.

### Authentication & Authorization
- [ ] Implement JWT or API key authentication.
- [ ] Add user management (registration, login).
- [ ] Implement role-based access control (RBAC).
- [ ] Add per-user website limits and quotas.

### Rate Limiting & Security
- [ ] Add API rate limiting (per-user/per-IP).
- [ ] Implement abuse detection and prevention.
- [ ] Add input validation and sanitization.
- [ ] Implement CSRF protection.
- [ ] Add security headers middleware.

### Observability & Monitoring
- [ ] Add Prometheus metrics endpoint.
- [ ] Create Grafana dashboards.
- [ ] Implement distributed tracing (OpenTelemetry).
- [ ] Add structured logging with log levels.
- [ ] Implement alerting for service failures.
- [ ] Add APM (Application Performance Monitoring).

### Backup & Recovery (LAST PRIORITY)
- [ ] Implement PostgreSQL backup strategy.
- [ ] Implement Garage bucket backup/restore.
- [ ] Implement ChromaDB data export/import.
- [ ] Create disaster recovery procedures.
- [ ] Document backup schedules and retention.

### Scaling & Performance
- [ ] Load test the system (crawling, vectorization, RAG).
- [ ] Optimize database queries and add indexes.
- [ ] Implement caching layer (Redis).
- [ ] Add horizontal scaling for job workers.
- [ ] Implement multi-node crawling coordination.
- [ ] Optimize vector search performance.

### Advanced Job Queue Features (In Progress)
- [x] Implement streaming SSE for RAG query responses.
    - [x] Add streaming generation method to OllamaLLM.
    - [x] Add streaming query method to RAG service.
    - [x] Create POST /api/websites/{id}/query/stream endpoint with SSE.
    - [x] Send events: start, chunk, metadata, done, error.
- [x] Add job management API endpoints (list jobs, cancel job, job status).
    - [x] Create JobsController with asynq Inspector.
    - [x] Add GET /api/jobs/queues - list all queues with stats.
    - [x] Add GET /api/jobs/pending - list pending jobs.
    - [x] Add GET /api/jobs/active - list running jobs.
    - [x] Add GET /api/jobs/scheduled - list scheduled jobs.
    - [x] Add GET /api/jobs/retry - list jobs pending retry.
    - [x] Add GET /api/jobs/archived - list failed jobs.
    - [x] Add POST /api/jobs/{id}/cancel - cancel a job.
    - [x] Add POST /api/jobs/{id}/retry - retry a failed job.
    - [x] Add POST /api/jobs/queues/{queue}/pause - pause a queue.
    - [x] Add POST /api/jobs/queues/{queue}/resume - resume a queue.
- [x] Implement cleanup job logic (delete old pages from storage/vectors).
    - [x] Framework implemented with safety checks.
    - [x] Query pages from database.
    - [x] Support for storage/vector/both deletion modes.
    - [x] Ready for production use (requires additional safety measures).
- [ ] Add timeout and circuit breaker patterns for external services (skipped - larger effort).
- [x] Implement DB connection pooling configuration.
    - [x] Max open connections (default: 25).
    - [x] Max idle connections (default: 5).
    - [x] Connection max lifetime (default: 5 minutes).
- [ ] Add job queue monitoring dashboard.
- [ ] Implement job priority adjustment.

### Advanced Content Processing
- [ ] Add language detection.
- [ ] Implement duplicate detection across websites.
- [ ] Add OCR support for images/PDFs.
- [ ] Implement semantic deduplication.
- [ ] Add content categorization/tagging.

### API Enhancements
- [ ] Add pagination to all list endpoints.
- [ ] Implement filtering and sorting.
- [ ] Add GraphQL API option.
- [ ] Implement webhooks for crawl completion.

---

Essential fixes for production viability and legal compliance.

### Storage Migration

- [x] **MinIO → Garage Migration:**
    - [x] Replace MinIO with Garage in docker-compose
    - [x] Update storage client to use Garage (GarageStorage service)
    - [x] Update configuration to use Garage settings
    - [x] Update all references from MinIO to Garage
    - [ ] Add migration script for existing data
    - [x] Update documentation

### Critical Features

- [x] **HTML Content Cleaning:**
    - [x] Integrate readability library (codeberg.org/readeck/go-readability/v2)
    - [x] Extract main content, remove navigation/ads/footers
    - [x] Clean HTML tags and normalize text
    - [x] Add content quality scoring
    - [x] Add metadata extraction
    - [x] Add content validation

- [x] **Health Check System:**
    - [x] Add GET /health endpoint for API
    - [x] Check PostgreSQL connection
    - [x] Check Garage/S3 connection
    - [x] Check ChromaDB connection
    - [x] Check Ollama connection
    - [x] Return detailed status with service breakdown

- [x] **Robots.txt Enforcement:**
    - [x] Implement robots.txt parser (using temoto/robotstxt)
    - [x] Check robots.txt before crawling
    - [x] Respect crawl-delay directive
    - [x] Add robots.txt caching (24hr TTL)
    - [x] Cache management (clear, per-domain clear)

- [x] **Duplicate URL Prevention:**
    - [x] Add URL normalization (remove query params, fragments, tracking)
    - [x] Normalize scheme/host to lowercase
    - [x] Remove trailing slashes
    - [x] Remove common tracking parameters (utm_*, fbclid, etc)
    - [ ] Check if URL already exists before crawling (needs crawler integration)
    - [ ] Add unique constraint on normalized URLs (needs migration)
    - [ ] Skip already-crawled URLs in same session (needs crawler integration)

### High Priority Features

- [ ] **Streaming RAG Responses:** (Next Priority)
    - [ ] Implement Server-Sent Events (SSE) for query endpoint
    - [ ] Stream LLM tokens as they're generated
    - [ ] Add /query/stream endpoint variant
    - [ ] Update response format for streaming

- [ ] **HTTP Timeouts & Retry Logic:**
    - [ ] Add configurable timeouts for crawler
    - [ ] Add timeout for Ollama requests
    - [ ] Add timeout for ChromaDB requests
    - [ ] Implement exponential backoff retry
    - [ ] Add circuit breaker pattern for external services

- [ ] **Connection Pooling:**
    - [ ] Configure PostgreSQL connection pool (max/min connections)
    - [ ] Add connection lifetime limits
    - [ ] Monitor connection usage

- [x] **Sitemap Support:**
    - [x] Add sitemap.xml parser (basic implementation)
    - [x] GetSitemapURLs method in RobotsEnforcer
    - [ ] Prioritize sitemap URLs over crawling (needs crawler integration)
    - [ ] Support sitemap index files
    - [ ] Respect lastmod dates

- [ ] **Background Job Queue:** (Dependencies added)
    - [x] Add Redis to docker-compose
    - [x] Add Redis configuration
    - [x] Install asynq library
    - [ ] Integrate asynq job queue
    - [ ] Move crawler to background jobs
    - [ ] Move vectorization to background jobs
    - [ ] Add job status tracking
    - [ ] Add pause/resume/cancel functionality

- [ ] **Content Change Detection:**
    - [ ] Check content hash before re-crawling
    - [ ] Store last-modified headers
    - [ ] Skip unchanged pages
    - [ ] Update only modified content

---

## Phase 2.7: Production Readiness (Nice to Have)

Additional features for robust production deployment.

### Security & Authentication

- [ ] **API Authentication:**
    - [ ] Add JWT authentication middleware
    - [ ] Implement user registration/login
    - [ ] Add API key support
    - [ ] Role-based access control (admin/user)

- [ ] **Rate Limiting:**
    - [ ] Add rate limiting middleware (per IP)
    - [ ] Add rate limiting per user/API key
    - [ ] Configurable limits via env
    - [ ] Return proper 429 responses

- [ ] **Input Validation:**
    - [ ] Add request validation middleware
    - [ ] Validate URL formats
    - [ ] Sanitize user inputs
    - [ ] Add request size limits

### Monitoring & Observability

- [ ] **Metrics & Monitoring:**
    - [ ] Add Prometheus metrics endpoint
    - [ ] Track crawler metrics (pages/sec, errors)
    - [ ] Track RAG query latency
    - [ ] Track LLM token usage
    - [ ] Add Grafana dashboard configs

- [ ] **Logging & Tracing:**
    - [ ] Add request ID tracking
    - [ ] Add distributed tracing (OpenTelemetry)
    - [ ] Improve structured logging
    - [ ] Add log levels per component
    - [ ] Add audit logging

- [ ] **Alerting:**
    - [ ] Set up alerting rules
    - [ ] Alert on high error rates
    - [ ] Alert on service downtime
    - [ ] Alert on disk space issues

### Reliability & Backup

- [ ] **Backup Strategy:**
    - [ ] Automated PostgreSQL backups
    - [ ] Garage/S3 bucket backups
    - [ ] ChromaDB collection backups
    - [ ] Backup restore procedures
    - [ ] Disaster recovery plan

- [ ] **Error Handling:**
    - [ ] Graceful degradation if Ollama down
    - [ ] Fallback if ChromaDB unavailable
    - [ ] Circuit breaker for all external services
    - [ ] Better error messages to users
    - [ ] Error tracking (Sentry integration)

### Performance & Scaling

- [ ] **Caching Layer:**
    - [ ] Add Redis for caching
    - [ ] Cache RAG query results
    - [ ] Cache embeddings for common queries
    - [ ] Cache crawled content metadata

- [ ] **Database Optimization:**
    - [ ] Add missing indexes
    - [ ] Optimize slow queries
    - [ ] Add database query logging
    - [ ] Consider read replicas

- [ ] **Parallel Crawling:**
    - [ ] Support multiple concurrent site crawls
    - [ ] Worker pool for crawling
    - [ ] Priority queue for URLs
    - [ ] Domain-specific rate limiting

### Content Quality

- [ ] **Advanced Content Processing:**
    - [ ] Language detection
    - [ ] Content quality scoring
    - [ ] Duplicate content detection across pages
    - [ ] Extract metadata (title, description, keywords)
    - [ ] Image/PDF text extraction (OCR)

- [ ] **Conversation Features:**
    - [ ] Multi-turn conversation support
    - [ ] Conversation history storage
    - [ ] Context window management
    - [ ] Follow-up question handling

### API Improvements

- [ ] **API Versioning:**
    - [ ] Change routes to /api/v1/
    - [ ] Support multiple API versions
    - [ ] Deprecation headers

- [ ] **Enhanced Endpoints:**
    - [ ] Pagination on /pages endpoint
    - [ ] Filtering by status, date
    - [ ] Sorting options
    - [ ] Bulk operations
    - [ ] Search across content

- [ ] **Webhooks:**
    - [ ] Webhook registration endpoint
    - [ ] Notify on crawl completion
    - [ ] Notify on errors
    - [ ] Webhook retry logic

### DevOps & CI/CD

- [ ] **Automation:**
    - [ ] GitHub Actions CI/CD pipeline
    - [ ] Automated testing on PR
    - [ ] Automated Docker builds
    - [ ] Automated deployments

- [ ] **Testing:**
    - [ ] Unit tests for all services
    - [ ] Integration tests
    - [ ] E2E tests
    - [ ] Load testing
    - [ ] Performance benchmarks

---

## Phase 3: AI-Powered Chat Interface

Frontend development after Phase 2.6/2.7 completion.

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

**Current API Endpoints:** 7
- POST /api/websites
- GET /api/websites
- GET /api/websites/{id}/status
- POST /api/websites/{id}/recrawl
- GET /api/websites/{id}/pages
- POST /api/websites/{id}/query
- GET /api/swagger/*

**Phase 2.6 Progress: 7/12 Critical Features Complete**

Completed:
- ✅ Garage migration (licensing) - Storage layer migrated
- ✅ HTML cleaning (content quality) - ContentProcessor service
- ✅ Health checks (reliability) - /health endpoint
- ✅ Robots.txt (compliance) - RobotsEnforcer service
- ✅ URL normalization (duplicate prevention foundation)
- ✅ Sitemap parser (basic implementation)
- ✅ Infrastructure (Redis, dependencies, config)

Next Priority:
- ⏳ Integrate content processor into crawler
- ⏳ Integrate robots.txt enforcer into crawler
- ⏳ Streaming RAG responses (SSE)
- ⏳ Background job queue (asynq integration)
- ⏳ HTTP timeouts & retry logic
