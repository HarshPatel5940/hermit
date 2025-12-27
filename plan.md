# Hermit - Web Crawler & RAG System

## Needed Business Backend Features

None - all core business features implemented.

## Frontend Plan

None - all core frontend features implemented.

## Completed Backend Features

### Core Business (Priority 1)

- [x] **Authentication & Authorization**
    - API key based authentication with ULID identifiers
    - User registration and login with bcrypt password hashing
    - Role-based access control (admin/user roles)
    - Per-user website limits and quotas
    - API key management (create, list, revoke, update)
    - Scope-based permissions for API keys
    - Website ownership verification
    - Auth middleware for protected routes

- [x] **API Versioning**
    - Versioned API structure at /api/v1/
    - Auth routes: /api/v1/auth/*
    - Website routes: /api/v1/auth-protected
    - Job management: /api/v1/jobs/* (admin only)
    - Legacy /api/* routes for backward compatibility

- [x] **Web Crawler**
    - Colly-based crawler with rate limiting
    - Configurable max depth and page limits
    - Per-domain crawl delays
    - User-agent customization
    - URL normalization to prevent duplicates
    - In-memory duplicate prevention during sessions
    - Timeout and error handling

- [x] **Content Processing**
    - HTML content extraction with go-readability
    - Content cleaning and quality filtering
    - Minimum length and quality thresholds
    - Metadata extraction (title, author, description)
    - Main content isolation from boilerplate

- [x] **Robots.txt Compliance**
    - Robots.txt parser and enforcer
    - Crawl-delay enforcement per domain
    - Robots cache with 24h TTL
    - User-agent specific rules
    - Disallow pattern matching

- [x] **RAG (Retrieval Augmented Generation)**
    - Query endpoint for AI-powered Q&A
    - Server-sent events (SSE) streaming responses
    - Context retrieval from ChromaDB
    - Ollama LLM integration (llama3.1)
    - Source citation in responses
    - Configurable top-k retrieval and context chunks

- [x] **Vector Storage & Embeddings**
    - Text chunking (800 chars, 100 overlap)
    - Ollama embedding generation (mxbai-embed-large)
    - ChromaDB vector storage
    - Metadata storage with chunks
    - Website-scoped collections

- [x] **Background Job Queue**
    - Asynq integration with Redis
    - Job types: crawl, vectorize, recrawl, cleanup
    - Configurable retry policies
    - Job prioritization
    - Worker process for background execution
    - Job persistence and reliability

### Operational (Priority 2)

- [x] **Job Management API**
    - List queues with stats
    - View pending/active/scheduled/retry/archived jobs
    - Cancel individual jobs
    - Retry failed jobs
    - Pause/resume queues
    - Admin-only access control

- [x] **Storage Layer**
    - Garage S3-compatible object storage
    - Replaced MinIO for licensing compliance
    - Configurable bucket and credentials
    - Error handling and retries

- [x] **Database**
    - PostgreSQL with pgx driver
    - Connection pooling (configurable max open/idle connections)
    - sqlx for query mapping
    - Migrations for schema versioning
    - Repositories: Website, Page, User, APIKey

- [x] **Rate Limiting & Security**
    - Per-IP rate limiting (configurable requests/min)
    - In-memory limiter with auto cleanup
    - Security headers (XSS, nosniff, X-Frame, HSTS, CSP)
    - CORS configuration
    - 429 Too Many Requests responses

- [x] **Health & Monitoring**
    - Health check endpoint at /api/health and /api/v1/health
    - PostgreSQL connection check
    - Garage storage check
    - ChromaDB availability check
    - Ollama service check
    - JSON response with status details

- [x] **API Features**
    - Pagination on all list endpoints (page/limit params)
    - Status filtering on pages endpoint
    - Max 100 items per page limit
    - Standardized response formats
    - Swagger/OpenAPI documentation

- [x] **Configuration**
    - Environment-based config with godotenv
    - All timeouts configurable (HTTP, crawler, Ollama, ChromaDB)
    - Crawler settings (delay, depth, pages, user-agent)
    - Content quality thresholds
    - RAG parameters (top-k, context chunks)
    - Database pooling settings
    - Redis and job queue config

### Infrastructure (Priority 3)

- [x] **Dependency Injection**
    - Uber-go/fx for DI container
    - Lifecycle management
    - Graceful shutdown
    - Logger integration

- [x] **Docker Setup**
    - Docker Compose with all services
    - PostgreSQL container
    - Redis container
    - Garage (S3) container
    - ChromaDB container
    - Ollama container
    - Volume persistence
    - Network configuration

- [x] **Logging**
    - Structured logging with zap
    - Development and production modes
    - Request/response logging middleware
    - Error logging throughout services

- [x] **Project Structure**
    - Clean architecture (controllers, services, repositories)
    - Separate API server and worker binaries
    - Middleware organization
    - Route versioning
    - Schema/model definitions

### Frontend (Priority 1)

- [x] **Modern UI Design**
    - Dark theme with Tailwind CSS
    - Responsive layout with collapsible sidebar
    - Alpine.js for reactive components
    - HTMX for dynamic updates without heavy JavaScript
    - Clean, minimal interface

- [x] **Authentication Pages**
    - Login page with email/password
    - Registration page with validation
    - Error message display
    - Responsive forms
    - Gradient background design

- [x] **Chat Interface**
    - Real-time chat UI with message bubbles
    - SSE streaming support for AI responses
    - Auto-scrolling message container
    - User/assistant message differentiation
    - Loading states and error handling
    - Keyboard shortcuts (Enter to send, Shift+Enter for newline)

- [x] **Website Management**
    - Grid view of all websites
    - Status badges (Indexed, Processing, Failed, Pending)
    - Real-time progress indicators
    - Add website modal with form
    - Configurable crawl depth and max pages
    - Reindex and delete actions
    - Empty state with call-to-action
    - Error message display per website

- [x] **API Key Management**
    - List all API keys with metadata
    - Create new API key modal
    - Scope selection (read/write)
    - Key preview with masked characters
    - One-time key display after creation
    - Copy to clipboard functionality
    - Revoke and delete actions
    - Active/inactive status badges
    - Security warnings for key handling

- [x] **Job Monitoring**
    - Job queue statistics dashboard
    - Status filters (pending, processing, completed, failed)
    - Job type filters (crawl, process, embed, cleanup)
    - Real-time job cards with metadata
    - Progress bars for active jobs
    - Retry and cancel actions
    - Job payload and error display
    - Icon-based job type indicators
    - Empty state handling

- [x] **App Layout**
    - Persistent sidebar navigation
    - Collapsible sidebar with toggle
    - Active page highlighting
    - User menu with logout
    - Responsive header with page title
    - Icon-based navigation
    - Smooth transitions and animations

- [x] **Build System**
    - Templ component generation
    - Tailwind CSS compilation
    - Makefile targets for frontend
    - Clean task for generated files
    - Integrated build pipeline
    - Watch mode for development

## API Endpoints (v1)

### Web Routes
- GET / - Redirect to login
- GET /login - Login page
- GET /register - Registration page
- GET /chat - Chat interface (protected)
- GET /websites - Website management (protected)
- GET /api-keys - API key management (protected)
- GET /jobs - Job monitoring (protected, admin)

### Public API
- POST /api/v1/auth/register - User registration
- POST /api/v1/auth/login - User login
- GET /api/v1/health - Health check

### Authenticated (API Key Required)
- GET /api/v1/auth/me - Get current user
- POST /api/v1/auth/api-keys - Create API key
- GET /api/v1/auth/api-keys - List API keys
- GET /api/v1/auth/api-keys/:id - Get API key
- PUT /api/v1/auth/api-keys/:id - Update API key
- DELETE /api/v1/auth/api-keys/:id - Revoke API key
- POST /api/v1/websites - Create website
- GET /api/v1/websites - List websites (user's only)
- GET /api/v1/websites/:id/pages - Get pages with pagination
- POST /api/v1/websites/:id/query - RAG query
- POST /api/v1/websites/:id/query/stream - RAG query (SSE streaming)
- GET /api/v1/websites/:id/status - Get crawl status
- POST /api/v1/websites/:id/recrawl - Trigger re-crawl

### Admin Only
- GET /api/v1/jobs/queues - List job queues
- GET /api/v1/jobs/pending - List pending jobs
- GET /api/v1/jobs/active - List active jobs
- GET /api/v1/jobs/scheduled - List scheduled jobs
- GET /api/v1/jobs/retry - List retry jobs
- GET /api/v1/jobs/archived - List archived jobs
- POST /api/v1/jobs/:id/cancel - Cancel job
- POST /api/v1/jobs/:id/retry - Retry job
- POST /api/v1/jobs/queues/:queue/pause - Pause queue
- POST /api/v1/jobs/queues/:queue/resume - Resume queue

## Tech Stack

### Backend
- Go 1.21+
- Echo web framework
- Uber-go/fx for dependency injection
- PostgreSQL with pgx/sqlx
- Redis for job queue
- Asynq for background jobs

### Frontend
- Templ for type-safe HTML templates
- Tailwind CSS v4 for styling
- HTMX for dynamic interactions
- Alpine.js for reactive components
- Server-side rendering
- SSE for real-time streaming

### Storage & AI
- Garage (S3-compatible object storage)
- ChromaDB for vector storage
- Ollama for embeddings and LLM
- Models: mxbai-embed-large, llama3.1

### Libraries
- Colly for web crawling
- go-readability for content extraction
- robotstxt for robots.txt parsing
- bcrypt for password hashing
- ULID for unique identifiers
- Zap for structured logging

## Database Schema

### users
- id (VARCHAR 26, ULID primary key)
- email (VARCHAR 255, unique)
- password_hash (VARCHAR 255)
- role (VARCHAR 50, default: 'user')
- is_active (BOOLEAN, default: true)
- website_limit (INTEGER, default: 10)
- created_at, updated_at (TIMESTAMP)

### api_keys
- id (VARCHAR 26, ULID primary key)
- user_id (VARCHAR 26, foreign key to users)
- key_hash (VARCHAR 255, unique)
- key_prefix (VARCHAR 20)
- name (VARCHAR 255)
- scopes (TEXT[])
- is_active (BOOLEAN, default: true)
- last_used_at (TIMESTAMP, nullable)
- expires_at (TIMESTAMP, nullable)
- created_at, updated_at (TIMESTAMP)

### websites
- id (SERIAL primary key)
- url (TEXT, unique)
- user_id (VARCHAR 26, foreign key to users, nullable)
- is_monitored (BOOLEAN)
- crawl_status (VARCHAR 50)
- crawl_started_at, crawl_completed_at (TIMESTAMP, nullable)
- total_pages_crawled, total_pages_failed (INTEGER)
- last_error (TEXT, nullable)
- created_at, updated_at (TIMESTAMP)

### pages
- id (SERIAL primary key)
- website_id (INTEGER, foreign key to websites)
- url (TEXT)
- normalized_url (TEXT, unique)
- title, content (TEXT)
- status (VARCHAR 50)
- error_message (TEXT, nullable)
- created_at, updated_at (TIMESTAMP)

## Environment Variables

### Server
- PORT (default: 8080)
- APP_ENV (development/production)
- HTTP_TIMEOUT (default: 30s)

### Database
- DATABASE_URL (PostgreSQL connection string)
- DB_MAX_OPEN_CONNS (default: 25)
- DB_MAX_IDLE_CONNS (default: 5)
- DB_CONN_MAX_LIFETIME (default: 5m)

### Redis
- REDIS_URL (default: localhost:6379)
- REDIS_PASSWORD
- REDIS_DB (default: 0)

### Storage (Garage)
- GARAGE_ENDPOINT
- GARAGE_REGION
- GARAGE_ACCESS_KEY
- GARAGE_SECRET_KEY
- GARAGE_BUCKET_NAME

### AI Services
- OLLAMA_URL (default: http://localhost:11434)
- OLLAMA_MODEL (embeddings, default: mxbai-embed-large)
- OLLAMA_LLM_MODEL (default: llama3.1)
- OLLAMA_TIMEOUT (default: 120s)
- CHROMADB_URL (default: http://localhost:8000)
- CHROMADB_TIMEOUT (default: 30s)

### RAG Configuration
- RAG_TOP_K (default: 5)
- RAG_CONTEXT_CHUNKS (default: 3)

### Crawler
- CRAWLER_MAX_DEPTH (default: 3)
- CRAWLER_MAX_PAGES (default: 100)
- CRAWLER_DELAY_MS (default: 1000)
- CRAWLER_USER_AGENT (default: HermitBot/1.0)
- CRAWLER_RESPECT_ROBOTS_TXT (default: true)
- CRAWLER_TIMEOUT (default: 30s)

### Content Processing
- CONTENT_MIN_LENGTH (default: 100)
- CONTENT_MIN_QUALITY (default: 0.3)

### Rate Limiting
- RATE_LIMIT_ENABLED (default: true)
- RATE_LIMIT_REQUESTS_PER_MIN (default: 60)
- RATE_LIMIT_BURST (default: 10)

## Running the System

### Prerequisites
```bash
# Install Ollama models
ollama pull mxbai-embed-large
ollama pull llama3.1
```

### Development
```bash
# First time setup
make setup

# Start development with live-reload
make dev

# Or manually:
# Start all services
make up

# Run migrations
make migrate-up

# Run API server with live-reload
make watch

# Run worker (separate terminal)
make run-worker

# Watch Tailwind CSS changes (optional, separate terminal)
make tailwind-watch
```

### Production
```bash
# Build all binaries and frontend
make build

# Run server
./bin/hermit

# Run worker (separate terminal)
./bin/worker
```

### Frontend Development
```bash
# Generate templ components
make templ-gen

# Build Tailwind CSS
make tailwind-build

# Build frontend (templ + CSS)
make frontend

# Watch CSS changes
make tailwind-watch
```

## Future Enhancements (Optional)

- [ ] Circuit breaker patterns for external services
- [ ] Prometheus metrics and Grafana dashboards
- [ ] Automated backups for PostgreSQL and Garage
- [ ] Horizontal scaling with multiple workers
- [ ] Webhooks for crawl completion events
- [ ] GraphQL API option
- [ ] Advanced content processing (language detection, OCR)
- [ ] Load testing and performance optimization
- [ ] CI/CD pipeline with GitHub Actions
- [ ] Unit and integration tests
- [ ] E2E testing