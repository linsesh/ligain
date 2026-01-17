package services

import (
	"context"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/storage"
	"time"

	"github.com/google/uuid"
)

const (
	// DefaultSignedURLTTL is the default TTL for signed URLs (7 days)
	DefaultSignedURLTTL = 7 * 24 * time.Hour
	// MaxSignedURLTTL is the maximum allowed TTL for signed URLs (7 days)
	MaxSignedURLTTL = 7 * 24 * time.Hour
	// AvatarContentType is the content type for avatar images
	AvatarContentType = "image/webp"
)

// StorageService defines the interface for avatar storage operations.
type StorageService interface {
	// UploadAvatar uploads avatar image data and returns the object key
	UploadAvatar(ctx context.Context, userID string, imageData []byte) (objectKey string, err error)

	// GenerateSignedURL creates a time-limited URL for reading an object
	GenerateSignedURL(ctx context.Context, objectKey string, ttl time.Duration) (signedURL string, err error)

	// DeleteAvatar removes an avatar object
	DeleteAvatar(ctx context.Context, objectKey string) error
}

// StorageServiceImpl implements StorageService using a BlobStorage backend.
type StorageServiceImpl struct {
	blobStorage  storage.BlobStorage
	signedURLTTL time.Duration
}

// NewStorageService creates a new StorageServiceImpl with default TTL.
func NewStorageService(blobStorage storage.BlobStorage) StorageService {
	return &StorageServiceImpl{
		blobStorage:  blobStorage,
		signedURLTTL: DefaultSignedURLTTL,
	}
}

// NewStorageServiceWithTTL creates a new StorageServiceImpl with a custom default TTL.
func NewStorageServiceWithTTL(blobStorage storage.BlobStorage, ttl time.Duration) StorageService {
	return &StorageServiceImpl{
		blobStorage:  blobStorage,
		signedURLTTL: ttl,
	}
}

// UploadAvatar uploads avatar image data and returns the object key.
// Object path format: avatars/{userID}/{uuid}.webp
func (s *StorageServiceImpl) UploadAvatar(ctx context.Context, userID string, imageData []byte) (string, error) {
	if userID == "" {
		return "", &models.StorageError{Reason: "userID cannot be empty"}
	}

	if len(imageData) == 0 {
		return "", &models.StorageError{Reason: "image data cannot be empty"}
	}

	// Generate unique object path
	objectKey := fmt.Sprintf("avatars/%s/%s.webp", userID, uuid.New().String())

	err := s.blobStorage.Upload(ctx, objectKey, imageData, AvatarContentType)
	if err != nil {
		return "", err
	}

	return objectKey, nil
}

// GenerateSignedURL creates a time-limited URL for reading an object.
// TTL is clamped to MaxSignedURLTTL (7 days). If TTL is 0, uses default TTL.
func (s *StorageServiceImpl) GenerateSignedURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	if objectKey == "" {
		return "", &models.StorageError{Reason: "object key cannot be empty"}
	}

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = s.signedURLTTL
	}

	// Clamp TTL to maximum
	if ttl > MaxSignedURLTTL {
		ttl = MaxSignedURLTTL
	}

	return s.blobStorage.GenerateSignedURL(ctx, objectKey, ttl)
}

// DeleteAvatar removes an avatar object.
func (s *StorageServiceImpl) DeleteAvatar(ctx context.Context, objectKey string) error {
	if objectKey == "" {
		return &models.StorageError{Reason: "object key cannot be empty"}
	}

	return s.blobStorage.Delete(ctx, objectKey)
}
