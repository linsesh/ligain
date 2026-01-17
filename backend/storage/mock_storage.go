package storage

import (
	"context"
	"fmt"
	"time"
)

// MockBlobStorage is an in-memory implementation of BlobStorage for testing.
type MockBlobStorage struct {
	objects map[string][]byte

	// Error injection for testing error paths
	UploadError    error
	SignedURLError error
	DeleteError    error
}

// NewMockBlobStorage creates a new MockBlobStorage instance.
func NewMockBlobStorage() *MockBlobStorage {
	return &MockBlobStorage{
		objects: make(map[string][]byte),
	}
}

// Upload stores data at the given path.
func (m *MockBlobStorage) Upload(ctx context.Context, objectPath string, data []byte, contentType string) error {
	if m.UploadError != nil {
		return m.UploadError
	}

	m.objects[objectPath] = data
	return nil
}

// GenerateSignedURL returns a predictable URL for testing.
func (m *MockBlobStorage) GenerateSignedURL(ctx context.Context, objectPath string, ttl time.Duration) (string, error) {
	if m.SignedURLError != nil {
		return "", m.SignedURLError
	}

	return fmt.Sprintf("https://mock-storage.example.com/%s?expires=%.0f", objectPath, ttl.Seconds()), nil
}

// Delete removes an object (idempotent).
func (m *MockBlobStorage) Delete(ctx context.Context, objectPath string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}

	delete(m.objects, objectPath)
	return nil
}

// GetObject retrieves stored data for testing verification.
func (m *MockBlobStorage) GetObject(path string) ([]byte, bool) {
	data, exists := m.objects[path]
	return data, exists
}

// ObjectCount returns the number of stored objects.
func (m *MockBlobStorage) ObjectCount() int {
	return len(m.objects)
}

// Clear removes all stored objects.
func (m *MockBlobStorage) Clear() {
	m.objects = make(map[string][]byte)
}
