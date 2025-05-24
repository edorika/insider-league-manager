# Simple Makefile for a Go project

# Build the application
all: build test

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application locally
run:
	@go run cmd/api/main.go

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Integration Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# =============================================================================
# DOCKER COMMANDS
# =============================================================================

# Helper function for docker compose commands
define docker_compose
	@if docker compose $(1) 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose $(1); \
	fi
endef

# Helper function for database-only docker compose commands
define docker_compose_db
	@if docker compose -f docker-compose.db.yml $(1) 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose -f docker-compose.db.yml $(1); \
	fi
endef

# Start database only
db-up:
	@echo "Starting database..."
	$(call docker_compose_db,up --build -d)
	@echo "Database started at localhost:5432"

# Stop database only
db-down:
	@echo "Stopping database..."
	$(call docker_compose_db,down)

# Start complete application (database + API)
docker-up:
	@echo "Starting complete application with Docker..."
	$(call docker_compose,up --build -d)
	@echo "Application started successfully!"
	@echo "API Server: http://localhost:8080"
	@echo "Database: localhost:5432"

# Stop services (keep containers)
docker-stop:
	@echo "Stopping services..."
	$(call docker_compose,stop)

# Stop and remove containers, networks, volumes
docker-down:
	@echo "Stopping and removing containers..."
	$(call docker_compose,down)

# Restart all services
docker-restart: docker-stop docker-up

# Remove everything including volumes (clean slate)
docker-clean:
	@echo "Removing everything (containers, networks, volumes)..."
	$(call docker_compose,down -v)
	@docker system prune -f

# Build and restart only the API service (for development)
docker-rebuild:
	@echo "Rebuilding API service..."
	$(call docker_compose,up --build --force-recreate api -d)

# View logs from all services
docker-logs:
	$(call docker_compose,logs -f)

# View logs from API service only
docker-logs-api:
	$(call docker_compose,logs -f api)

# View logs from database service only
docker-logs-db:
	$(call docker_compose,logs -f psql_bp)

# Show status of all services
docker-status:
	@echo "Service Status:"
	$(call docker_compose,ps)

# =============================================================================

.PHONY: all build run test clean watch itest \
         db-up db-down \
         docker-up docker-stop docker-down docker-restart docker-clean docker-rebuild \
         docker-logs docker-logs-api docker-logs-db docker-status
