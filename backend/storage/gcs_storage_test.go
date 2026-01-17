package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for GCS - run with:
// INTEGRATION_TESTS=true GCS_TEST_BUCKET=ligain-avatars-test go test ./backend/storage/... -v -run Integration

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TESTS=true to run.")
	}
}

func getTestBucket(t *testing.T) string {
	bucket := os.Getenv("GCS_TEST_BUCKET")
	if bucket == "" {
		t.Fatal("GCS_TEST_BUCKET environment variable is required for integration tests")
	}
	return bucket
}

func TestGCS_Upload_Integration(t *testing.T) {
	skipIfNotIntegration(t)
	bucket := getTestBucket(t)

	ctx := context.Background()
	gcs, err := NewGCSBlobStorage(ctx, bucket)
	require.NoError(t, err)
	defer gcs.Close()

	objectPath := "test/" + uuid.New().String() + ".webp"

	// Cleanup after test
	t.Cleanup(func() {
		_ = gcs.Delete(ctx, objectPath)
	})

	// Upload test data
	testData := []byte("test image data for integration test")
	err = gcs.Upload(ctx, objectPath, testData, "image/webp")
	assert.NoError(t, err)

	// Verify object exists
	exists, err := gcs.Exists(ctx, objectPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Verify data can be downloaded
	downloaded, err := gcs.Download(ctx, objectPath)
	assert.NoError(t, err)
	assert.Equal(t, testData, downloaded)
}

func TestGCS_Delete_Integration(t *testing.T) {
	skipIfNotIntegration(t)
	bucket := getTestBucket(t)

	ctx := context.Background()
	gcs, err := NewGCSBlobStorage(ctx, bucket)
	require.NoError(t, err)
	defer gcs.Close()

	objectPath := "test/" + uuid.New().String() + ".webp"

	// Upload test data
	err = gcs.Upload(ctx, objectPath, []byte("data to delete"), "image/webp")
	require.NoError(t, err)

	// Delete
	err = gcs.Delete(ctx, objectPath)
	assert.NoError(t, err)

	// Verify object no longer exists
	exists, err := gcs.Exists(ctx, objectPath)
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestGCS_Delete_Idempotent_Integration(t *testing.T) {
	skipIfNotIntegration(t)
	bucket := getTestBucket(t)

	ctx := context.Background()
	gcs, err := NewGCSBlobStorage(ctx, bucket)
	require.NoError(t, err)
	defer gcs.Close()

	// Delete non-existent object - should not error
	objectPath := "test/nonexistent-" + uuid.New().String() + ".webp"
	err = gcs.Delete(ctx, objectPath)
	assert.NoError(t, err)
}

func TestGCS_GenerateSignedURL_Integration(t *testing.T) {
	skipIfNotIntegration(t)
	bucket := getTestBucket(t)

	ctx := context.Background()
	gcs, err := NewGCSBlobStorage(ctx, bucket)
	require.NoError(t, err)
	defer gcs.Close()

	objectPath := "test/" + uuid.New().String() + ".webp"

	// Cleanup after test
	t.Cleanup(func() {
		_ = gcs.Delete(ctx, objectPath)
	})

	// Upload test data
	err = gcs.Upload(ctx, objectPath, []byte("signed url test data"), "image/webp")
	require.NoError(t, err)

	// Generate signed URL
	signedURL, err := gcs.GenerateSignedURL(ctx, objectPath, time.Hour)
	assert.NoError(t, err)
	assert.NotEmpty(t, signedURL)
	assert.Contains(t, signedURL, objectPath)
	assert.Contains(t, signedURL, "storage.googleapis.com")
}
