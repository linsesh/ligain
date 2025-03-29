.PHONY: test build clean format

# Default target
all: test build

# Run tests
test:
	go test ./backend/...

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