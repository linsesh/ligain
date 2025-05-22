.PHONY: test test-integration test-all build clean format mobile test-frontend

# Default target
all: test test-frontend build

# Run tests
test:
	go test ./... -short

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