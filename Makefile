.PHONY: test test-integration test-all build clean format mobile test-frontend docker-up docker-down docker-build db-start db-stop db-create db-drop db-init docker-push docker-build-push docker-deploy deploy deploy-prd build-prd push-prd build-push-prd setup-metrics help

# Default target
all: help

# Show help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development commands:"
	@echo "  make test              - Run tests"
	@echo "  make test-frontend     - Run frontend tests"
	@echo "  make test-integration  - Run integration tests"
	@echo "  make build             - Build the application"
	@echo "  make clean             - Clean build artifacts"
	@echo "  make format            - Format Go files"
	@echo "  make deps              - Install dependencies"
	@echo ""
	@echo "Docker commands (dev):"
	@echo "  make docker-build      - Build Docker image for dev"
	@echo "  make docker-push       - Push Docker image to dev GCR"
	@echo "  make docker-build-push - Build and push Docker image for dev"
	@echo "  make docker-deploy     - Deploy dev service"
	@echo "  make deploy            - Complete dev deployment workflow"
	@echo ""
	@echo "Production commands:"
	@echo "  make deploy-prd        - Complete production deployment workflow"
	@echo "  make build-prd         - Build Docker image for production"
	@echo "  make push-prd          - Push Docker image to production GCR"
	@echo "  make build-push-prd    - Build and push Docker image for production"
	@echo ""
	@echo "Database commands:"
	@echo "  make db-start          - Start PostgreSQL container"
	@echo "  make db-stop           - Stop PostgreSQL container"
	@echo "  make db-create         - Create database"
	@echo "  make db-drop           - Drop database"
	@echo "  make db-init           - Initialize database with schema and test data"
	@echo "  make migrate-up        - Run migrations"
	@echo "  make migrate-down      - Rollback migrations"
	@echo ""
	@echo "Monitoring commands:"
	@echo "  make setup-metrics     - Setup Cloud Monitoring metrics (auto-run on deploy)"
	@echo ""
	@echo "Environment variables:"
	@echo "  ENV=prd               - Use production environment (default: dev)"
	@echo ""
	@echo "Examples:"
	@echo "  make deploy-prd        - Deploy to production"
	@echo "  make ENV=prd deploy    - Alternative way to deploy to production"

# Run tests
test:
	go test ./... -short || true
	cd frontend/ligain && npm test || true

# Run frontend tests
test-frontend:
	cd frontend/ligain && npm test || true

# Run integration tests
test-integration:
	$(eval include .env)
	INTEGRATION_TESTS=true SPORTSMONK_API_TOKEN=${SPORTSMONK_API_TOKEN} go test ./... -v -run Integration

# Run all tests
test-all: test test-frontend test-integration

# Build the application
build:
	go build -o backend/bin/app ./backend/...

# Environment configuration
ENV ?= dev
PROJECT_ID := $(shell if [ "$(ENV)" = "prd" ]; then echo "prd-ligain"; else echo "woven-century-307314"; fi)
SERVICE_NAME := $(shell if [ "$(ENV)" = "prd" ]; then echo "server-prd"; else echo "server-dev"; fi)
GIT_SHA := $(shell git rev-parse --short HEAD)

# Build Docker image
docker-build:
	docker build -t gcr.io/$(PROJECT_ID)/$(SERVICE_NAME):$(GIT_SHA) -f backend/Dockerfile .

# Push Docker image to GCR
docker-push:
	docker push gcr.io/$(PROJECT_ID)/$(SERVICE_NAME):$(GIT_SHA)

# Build and push Docker image
docker-build-push:
	docker buildx build --platform linux/amd64 -t gcr.io/$(PROJECT_ID)/$(SERVICE_NAME):$(GIT_SHA) -f backend/Dockerfile . --push --provenance=false --sbom=false

# Deploy to GCP (requires image to be pushed first)
docker-deploy:
	cd infrastructure && pulumi stack select linsesh/$(ENV) && \
		pulumi config set ligain:image_tag $(GIT_SHA) && \
		pulumi up --yes
	@echo "Setting up Cloud Monitoring metrics..."
	@./scripts/setup-metrics.sh $(ENV)

# Complete deployment workflow
deploy: docker-build-push docker-deploy

# Production-specific commands
deploy-prd:
	$(MAKE) ENV=prd deploy

build-prd:
	$(MAKE) ENV=prd docker-build

push-prd:
	$(MAKE) ENV=prd docker-push

build-push-prd:
	$(MAKE) ENV=prd docker-build-push

# Clean build artifacts
clean:
	rm -rf backend/bin/
	go clean ./backend/...

# Install dependencies
deps:
	go mod tidy
	cd frontend/ligain && npm install

# Run with race detector
test-race:
	go test -race ./backend/...

# Format Go files
format:
	go fmt ./backend/...

# Run iOS app
mobile:
	cd frontend/ligain && npx expo run:ios --device 

# Docker commands
docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f 

# Database configuration
DB_USER := postgres
DB_PASSWORD := postgres
DB_NAME := ligain_test
DB_PORT := 5432
DB_HOST := localhost

# Docker configuration
POSTGRES_CONTAINER := ligain-postgres

# Start PostgreSQL in Docker
db-start:
	@echo "Starting PostgreSQL container..."
	@docker run --name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		-d postgres:16-alpine
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3

# Stop and remove PostgreSQL container
db-stop:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(POSTGRES_CONTAINER) || true
	@docker rm $(POSTGRES_CONTAINER) || true

# Create database
db-create:
	@echo "Creating database..."
	@PGPASSWORD=$(DB_PASSWORD) createdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) || true

# Drop database
db-drop:
	@echo "Dropping database..."
	@PGPASSWORD=$(DB_PASSWORD) dropdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) || true

# Install golang-migrate
install-migrate:
	@echo "Installing golang-migrate..."
	@if ! command -v migrate > /dev/null; then \
		go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
	fi

# Run migrations manually
migrate-up:
	@echo "Running migrations..."
	@PATH="$(shell go env GOPATH)/bin:$(PATH)" migrate -path backend/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

# Rollback migrations
migrate-down:
	@echo "Rolling back migrations..."
	@PATH="$(shell go env GOPATH)/bin:$(PATH)" migrate -path backend/migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

# Initialize database with schema and test data
db-init: install-migrate db-create
	@echo "Initializing database..."
	@go run scripts/init_db.go

# Setup Cloud Monitoring metrics
setup-metrics:
	@./scripts/setup-metrics.sh $(ENV) 