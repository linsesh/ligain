.PHONY: test test-integration test-all build clean format mobile

# Default target
all: test build

# Run tests
test:
	go test ./... -short

# Run integration tests
test-integration:
	$(eval include .env)
	INTEGRATION_TESTS=true SPORTSMONK_API_TOKEN=${SPORTSMONK_API_TOKEN} go test ./... -v -run Integration

# Run all tests
test-all:
	INTEGRATION_TESTS=true SPORTSMONK_API_TOKEN=${SPORTSMONK_API_TOKEN} go test ./... -v

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

# Run with race detector
test-race:
	go test -race ./backend/...

# Format Go files
format:
	go fmt ./backend/...

# Run iOS app
mobile:
	cd frontend/ligain && npx expo run:ios --device 