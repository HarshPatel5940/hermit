# Hermit Project
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
	@echo "  build          - Build the server and worker binaries."
	@echo "  build-server   - Build only the server binary."
	@echo "  build-worker   - Build only the worker binary."
	@echo "  run            - Build and run the server."
	@echo "  run-worker     - Build and run the worker."
	@echo "  watch          - Run the server in live-reload mode using Air."
	@echo "  clean          - Remove binaries, generated docs, and frontend artifacts."
	@echo ""
	@echo "Database & Docs:"
	@echo "  migrate-up     - Apply all outstanding database migrations."
	@echo "  migrate-down   - Revert the last N migrations (default: 1). Usage: make migrate-down N=2"
	@echo "  docs           - Generate API documentation using swag."
	@echo ""
	@echo "Frontend:"
	@echo "  frontend       - Generate templ components and build Tailwind CSS."
	@echo "  templ-gen      - Generate templ components only."
	@echo "  tailwind-build - Build Tailwind CSS once."
	@echo "  tailwind-watch - Watch for CSS changes and rebuild automatically."
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all Go tests."


# High-Level Targets
.PHONY: setup dev
setup: install-tools up migrate-up
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
.PHONY: build build-server build-worker run run-worker watch clean
build: frontend build-server build-worker
build-server: frontend
	@echo "==> Building server binary..."
	@go build -o ./bin/hermit ./cmd/api
build-worker: frontend
	@echo "==> Building worker binary..."
	@go build -o ./bin/worker ./cmd/worker
run: build-server
	@./bin/hermit
run-worker: build-worker
	@./bin/worker
watch:
	@echo "==> Starting live-reload server with Air..."
	@air
clean:
	@echo "==> Cleaning up..."
	@rm -rf ./bin ./docs ./tailwindcss ./web/assets/css/output.css ./web/*_templ.go

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
	@swag init --generalInfo ./cmd/api/main.go --output ./docs || echo "Warning: swagger docs generation failed (non-critical)"

# Frontend
.PHONY: frontend templ-gen tailwind-install tailwind-build tailwind-watch
frontend: templ-gen tailwind-build
templ-gen:
	@echo "==> Generating templ components..."
	@templ generate ./web
tailwind-install:
	@if [ ! -f tailwindcss ]; then curl -sL https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-x64 -o tailwindcss; fi
	@chmod +x tailwindcss
tailwind-build: tailwind-install
	@echo "==> Building Tailwind CSS..."
	@./tailwindcss -i ./web/styles/input.css -o ./web/assets/css/output.css
tailwind-watch: tailwind-install
	@./tailwindcss -i ./web/styles/input.css -o ./web/assets/css/output.css --watch

# Tooling
.PHONY: install-tools
install-tools:
	@echo "==> Installing Go tools..."
	@go install github.com/a-h/templ/cmd/templ@latest
	@go install github.com/swaggo/swag/cmd/swag@latest

# Testing
.PHONY: test
test:
	@echo "==> Running tests..."
	@go test -v ./...
