FROM --platform=linux/amd64 golang:1.24-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the backend source code
COPY backend/ ./backend/

# Build the application
RUN go build -o ligain ./backend

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./ligain"] 