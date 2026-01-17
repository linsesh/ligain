package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockBlobStorage_Upload_StoresData(t *testing.T) {
	mock := NewMockBlobStorage()
	ctx := context.Background()

	err := mock.Upload(ctx, "test/path.webp", []byte("image data"), "image/webp")

	assert.NoError(t, err)
	data, exists := mock.GetObject("test/path.webp")
	assert.True(t, exists)
	assert.Equal(t, []byte("image data"), data)
}

func TestMockBlobStorage_Upload_ErrorInjection(t *testing.T) {
	mock := NewMockBlobStorage()
	mock.UploadError = errors.New("upload failed")
	ctx := context.Background()

	err := mock.Upload(ctx, "test/path.webp", []byte("data"), "image/webp")

	assert.Error(t, err)
	assert.Equal(t, "upload failed", err.Error())
}

func TestMockBlobStorage_GenerateSignedURL_ReturnsPredictableURL(t *testing.T) {
	mock := NewMockBlobStorage()
	ctx := context.Background()

	url, err := mock.GenerateSignedURL(ctx, "avatars/user123/abc.webp", 3600)

	assert.NoError(t, err)
	assert.Contains(t, url, "avatars/user123/abc.webp")
	assert.Contains(t, url, "mock-storage.example.com")
}

func TestMockBlobStorage_GenerateSignedURL_ErrorInjection(t *testing.T) {
	mock := NewMockBlobStorage()
	mock.SignedURLError = errors.New("signing failed")
	ctx := context.Background()

	url, err := mock.GenerateSignedURL(ctx, "test/path.webp", 3600)

	assert.Error(t, err)
	assert.Empty(t, url)
}

func TestMockBlobStorage_Delete_RemovesObject(t *testing.T) {
	mock := NewMockBlobStorage()
	ctx := context.Background()

	// Upload first
	_ = mock.Upload(ctx, "test/path.webp", []byte("data"), "image/webp")
	assert.Equal(t, 1, mock.ObjectCount())

	// Delete
	err := mock.Delete(ctx, "test/path.webp")

	assert.NoError(t, err)
	_, exists := mock.GetObject("test/path.webp")
	assert.False(t, exists)
	assert.Equal(t, 0, mock.ObjectCount())
}

func TestMockBlobStorage_Delete_Idempotent(t *testing.T) {
	mock := NewMockBlobStorage()
	ctx := context.Background()

	// Delete non-existent object - should not error
	err := mock.Delete(ctx, "nonexistent/path.webp")

	assert.NoError(t, err)
}

func TestMockBlobStorage_Delete_ErrorInjection(t *testing.T) {
	mock := NewMockBlobStorage()
	mock.DeleteError = errors.New("delete failed")
	ctx := context.Background()

	err := mock.Delete(ctx, "test/path.webp")

	assert.Error(t, err)
	assert.Equal(t, "delete failed", err.Error())
}

func TestMockBlobStorage_Clear_RemovesAllObjects(t *testing.T) {
	mock := NewMockBlobStorage()
	ctx := context.Background()

	_ = mock.Upload(ctx, "path1.webp", []byte("data1"), "image/webp")
	_ = mock.Upload(ctx, "path2.webp", []byte("data2"), "image/webp")
	assert.Equal(t, 2, mock.ObjectCount())

	mock.Clear()

	assert.Equal(t, 0, mock.ObjectCount())
}
