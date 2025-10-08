# Auto-detect container engine (podman or docker)
ifeq ($(shell command -v podman 2> /dev/null),)
  CONTAINER_CMD ?= docker
  # Check for docker compose v2 vs v1
  ifneq ($(shell docker compose version 2> /dev/null; echo $$?), 0)
    COMPOSE_CMD ?= docker-compose
  else
    COMPOSE_CMD ?= docker compose
  endif
else
  CONTAINER_CMD ?= podman
  COMPOSE_CMD ?= podman-compose
endif

# Default migration steps to 1 if not provided
N ?= 1

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "High-Level Targets:"
	@echo "  setup          - Run for the first time. Starts services and runs migrations."
	@echo "  dev            - Start all services and run the app in live-reload mode."
	@echo ""
	@echo "Service Management:"
	@echo "  up             - Start all services (Postgres, MinIO, etc.) in the background."
	@echo "  down           - Stop all services."
	@echo "  logs           - Tail the logs of all running services."
	@echo ""
	@echo "Application Lifecycle:"
	@echo "  build          - Build the Go application binary."
	@echo "  run            - Build and run the application."
	@echo "  watch          - Run the application in live-reload mode using Air."
	@echo "  clean          - Remove the binary and generated docs."
	@echo ""
	@echo "Database & Docs:"
	@echo "  migrate-up     - Apply all outstanding database migrations."
	@echo "  migrate-down   - Revert the last N migrations (default: 1). Usage: make migrate-down N=2"
	@echo "  docs           - Generate API documentation using swag."
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all Go tests."

# High-Level Targets
.PHONY: setup dev
setup: up migrate-up
	@echo "==> Setup complete. Run 'make dev' to start developing."
dev: up watch

# Service Management
.PHONY: up down logs
up:
	@echo "==> Starting services with $(COMPOSE_CMD)..."
	@$(COMPOSE_CMD) up -d
down:
	@echo "==> Stopping services with $(COMPOSE_CMD)..."
	@$(COMPOSE_CMD) down
logs:
	@$(COMPOSE_CMD) logs -f

# Application Lifecycle
.PHONY: build run watch clean
build: docs
	@echo "==> Building application binary..."
	@go build -o ./bin/hermit ./cmd/api
run: build
	@./bin/hermit
watch:
	@echo "==> Starting live-reload server with Air..."
	@air
clean:
	@echo "==> Cleaning up..."
	@rm -rf ./bin ./docs

# Database & Docs
.PHONY: migrate-up migrate-down docs
migrate-up:
	@echo "==> Applying database migrations..."
	@go run cmd/migrate/main.go up
migrate-down:
	@echo "==> Reverting last $(N) database migration(s)..."
	@go run cmd/migrate/main.go down $(N)
docs:
	@echo "==> Generating API documentation..."
	@swag init --generalInfo cmd/api/main.go

# Testing
.PHONY: test
test:
	@echo "==> Running tests..."
	@go test -v ./...
