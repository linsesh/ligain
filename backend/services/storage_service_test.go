package services

import (
	"context"
	"errors"
	"ligain/backend/models"
	"ligain/backend/storage"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestUploadAvatar_Success tests that a valid upload returns an object key with correct format
func TestUploadAvatar_Success(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	objectKey, err := service.UploadAvatar(ctx, "user123", []byte("image data"))

	assert.NoError(t, err)
	assert.NotEmpty(t, objectKey)
	// Object key should be in format: avatars/{userID}/{uuid}.webp
	assert.True(t, strings.HasPrefix(objectKey, "avatars/user123/"))
	assert.True(t, strings.HasSuffix(objectKey, ".webp"))
	// Verify data was stored
	data, exists := mockStorage.GetObject(objectKey)
	assert.True(t, exists)
	assert.Equal(t, []byte("image data"), data)
}

// TestUploadAvatar_EmptyUserID tests that empty userID returns StorageError
func TestUploadAvatar_EmptyUserID(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	objectKey, err := service.UploadAvatar(ctx, "", []byte("image data"))

	assert.Empty(t, objectKey)
	var storageErr *models.StorageError
	assert.True(t, errors.As(err, &storageErr))
	assert.Contains(t, storageErr.Reason, "userID")
}

// TestUploadAvatar_EmptyImageData tests that empty image data returns StorageError
func TestUploadAvatar_EmptyImageData(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	objectKey, err := service.UploadAvatar(ctx, "user123", []byte{})

	assert.Empty(t, objectKey)
	var storageErr *models.StorageError
	assert.True(t, errors.As(err, &storageErr))
	assert.Contains(t, storageErr.Reason, "image data")
}

// TestUploadAvatar_NilImageData tests that nil image data returns StorageError
func TestUploadAvatar_NilImageData(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	objectKey, err := service.UploadAvatar(ctx, "user123", nil)

	assert.Empty(t, objectKey)
	var storageErr *models.StorageError
	assert.True(t, errors.As(err, &storageErr))
}

// TestUploadAvatar_BlobStorageError tests that blob storage errors are propagated
func TestUploadAvatar_BlobStorageError(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	mockStorage.UploadError = errors.New("GCS connection failed")
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	objectKey, err := service.UploadAvatar(ctx, "user123", []byte("image data"))

	assert.Empty(t, objectKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GCS connection failed")
}

// TestGenerateSignedURL_Success tests that a valid object key returns a signed URL
func TestGenerateSignedURL_Success(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	url, err := service.GenerateSignedURL(ctx, "avatars/user123/abc.webp", time.Hour)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	assert.Contains(t, url, "avatars/user123/abc.webp")
}

// TestGenerateSignedURL_EmptyObjectKey tests that empty object key returns StorageError
func TestGenerateSignedURL_EmptyObjectKey(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	url, err := service.GenerateSignedURL(ctx, "", time.Hour)

	assert.Empty(t, url)
	var storageErr *models.StorageError
	assert.True(t, errors.As(err, &storageErr))
	assert.Contains(t, storageErr.Reason, "object key")
}

// TestGenerateSignedURL_DefaultTTL tests that zero TTL uses default TTL
func TestGenerateSignedURL_DefaultTTL(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	url, err := service.GenerateSignedURL(ctx, "avatars/user123/abc.webp", 0)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	// The mock storage includes TTL in URL, we can verify default was used
	// Default is 7 days = 604800 seconds
	assert.Contains(t, url, "604800")
}

// TestGenerateSignedURL_ClampsTTL tests that TTL > 7 days is clamped to 7 days
func TestGenerateSignedURL_ClampsTTL(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	// Request 30 days TTL
	url, err := service.GenerateSignedURL(ctx, "avatars/user123/abc.webp", 30*24*time.Hour)

	assert.NoError(t, err)
	assert.NotEmpty(t, url)
	// Should be clamped to 7 days = 604800 seconds
	assert.Contains(t, url, "604800")
}

// TestGenerateSignedURL_BlobStorageError tests that blob storage errors are propagated
func TestGenerateSignedURL_BlobStorageError(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	mockStorage.SignedURLError = errors.New("signing failed")
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	url, err := service.GenerateSignedURL(ctx, "avatars/user123/abc.webp", time.Hour)

	assert.Empty(t, url)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signing failed")
}

// TestDeleteAvatar_Success tests that delete calls blob storage
func TestDeleteAvatar_Success(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	// First upload an object
	_ = mockStorage.Upload(ctx, "avatars/user123/abc.webp", []byte("data"), "image/webp")
	assert.Equal(t, 1, mockStorage.ObjectCount())

	// Delete it
	err := service.DeleteAvatar(ctx, "avatars/user123/abc.webp")

	assert.NoError(t, err)
	assert.Equal(t, 0, mockStorage.ObjectCount())
}

// TestDeleteAvatar_EmptyObjectKey tests that empty object key returns StorageError
func TestDeleteAvatar_EmptyObjectKey(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	err := service.DeleteAvatar(ctx, "")

	var storageErr *models.StorageError
	assert.True(t, errors.As(err, &storageErr))
	assert.Contains(t, storageErr.Reason, "object key")
}

// TestDeleteAvatar_BlobStorageError tests that blob storage errors are propagated
func TestDeleteAvatar_BlobStorageError(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	mockStorage.DeleteError = errors.New("delete failed")
	service := NewStorageService(mockStorage)
	ctx := context.Background()

	err := service.DeleteAvatar(ctx, "avatars/user123/abc.webp")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

// TestNewStorageServiceWithTTL tests custom TTL configuration
func TestNewStorageServiceWithTTL(t *testing.T) {
	mockStorage := storage.NewMockBlobStorage()
	customTTL := 24 * time.Hour
	service := NewStorageServiceWithTTL(mockStorage, customTTL)
	ctx := context.Background()

	url, err := service.GenerateSignedURL(ctx, "avatars/user123/abc.webp", 0)

	assert.NoError(t, err)
	// With 0 passed, should use custom default TTL of 24 hours = 86400 seconds
	assert.Contains(t, url, "86400")
}
