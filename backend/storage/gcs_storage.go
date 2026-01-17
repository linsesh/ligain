package storage

import (
	"context"
	"errors"
	"io"
	"time"

	"cloud.google.com/go/storage"
)

// GCSBlobStorage implements BlobStorage using Google Cloud Storage.
type GCSBlobStorage struct {
	bucketName string
	client     *storage.Client
}

// NewGCSBlobStorage creates a new GCSBlobStorage instance.
// Uses Application Default Credentials (compatible with Cloud Run).
func NewGCSBlobStorage(ctx context.Context, bucketName string) (*GCSBlobStorage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &GCSBlobStorage{
		bucketName: bucketName,
		client:     client,
	}, nil
}

// Close releases resources held by the GCS client.
func (g *GCSBlobStorage) Close() error {
	return g.client.Close()
}

// Upload stores data at the given path with content type.
func (g *GCSBlobStorage) Upload(ctx context.Context, objectPath string, data []byte, contentType string) error {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	writer.CacheControl = "public, max-age=31536000" // 1 year cache

	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return err
	}

	return writer.Close()
}

// GenerateSignedURL creates a time-limited URL for reading an object.
// Uses V4 signing.
func (g *GCSBlobStorage) GenerateSignedURL(ctx context.Context, objectPath string, ttl time.Duration) (string, error) {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(ttl),
	}

	return g.client.Bucket(g.bucketName).SignedURL(objectPath, opts)
}

// Delete removes an object (idempotent - no error if doesn't exist).
func (g *GCSBlobStorage) Delete(ctx context.Context, objectPath string) error {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	err := obj.Delete(ctx)
	if err != nil {
		// Make delete idempotent - no error if object doesn't exist
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return err
	}

	return nil
}

// Exists checks if an object exists (useful for testing).
func (g *GCSBlobStorage) Exists(ctx context.Context, objectPath string) (bool, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	_, err := obj.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Download retrieves object data (useful for testing).
func (g *GCSBlobStorage) Download(ctx context.Context, objectPath string) ([]byte, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(objectPath)

	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}
