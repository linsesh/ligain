.PHONY: test test-integration test-all build clean format mobile test-frontend docker-up docker-down docker-build db-start db-stop db-create db-drop db-init docker-push docker-build-push docker-deploy deploy

# Default target
all: test test-frontend build

# Run tests
test:
	go test ./... -short
	cd frontend/ligain && npm test

# Run frontend tests
test-frontend:
	cd frontend/ligain && npm test

# Run integration tests
test-integration:
	$(eval include .env)
	INTEGRATION_TESTS=true SPORTSMONK_API_TOKEN=${SPORTSMONK_API_TOKEN} go test ./... -v -run Integration

# Run all tests
test-all: test test-frontend test-integration

# Build the application
build:
	go build -o backend/bin/app ./backend/...

# Build Docker image
docker-build:
	docker build -t gcr.io/woven-century-307314/server-dev:latest -f backend/Dockerfile .

# Push Docker image to GCR
docker-push:
	docker push gcr.io/woven-century-307314/server-dev:latest

# Build and push Docker image
docker-build-push:
	docker buildx build --platform linux/amd64 -t gcr.io/woven-century-307314/server-dev:latest -f backend/Dockerfile . --push

# Deploy to GCP (requires image to be pushed first)
docker-deploy:
	cd infrastructure && pulumi up --yes

# Complete deployment workflow
deploy: docker-build-push docker-deploy

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