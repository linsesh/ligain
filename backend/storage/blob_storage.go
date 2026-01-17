package storage

import (
	"context"
	"time"
)

// BlobStorage is a provider-agnostic interface for blob storage operations.
// Implementations can include GCS, S3, Azure Blob, or in-memory for testing.
type BlobStorage interface {
	// Upload stores data at the given path with content type
	Upload(ctx context.Context, objectPath string, data []byte, contentType string) error

	// GenerateSignedURL creates a time-limited URL for reading an object
	GenerateSignedURL(ctx context.Context, objectPath string, ttl time.Duration) (string, error)

	// Delete removes an object (idempotent - no error if doesn't exist)
	Delete(ctx context.Context, objectPath string) error
}
