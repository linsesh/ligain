package services

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"ligain/backend/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test JPEG image
func createTestJPEG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a solid color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

// Helper function to create a test PNG image
func createTestPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a solid color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// TestProcessAvatar_ValidJPEG tests that a valid JPEG returns processed WebP
func TestProcessAvatar_ValidJPEG(t *testing.T) {
	processor := NewImageProcessor()
	jpegData := createTestJPEG(200, 200)

	result, err := processor.ProcessAvatar(jpegData)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// WebP files start with "RIFF" magic bytes
	assert.True(t, len(result) >= 4)
	assert.Equal(t, "RIFF", string(result[:4]))
}

// TestProcessAvatar_ValidPNG tests that a valid PNG returns processed WebP
func TestProcessAvatar_ValidPNG(t *testing.T) {
	processor := NewImageProcessor()
	pngData := createTestPNG(200, 200)

	result, err := processor.ProcessAvatar(pngData)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// WebP files start with "RIFF" magic bytes
	assert.True(t, len(result) >= 4)
	assert.Equal(t, "RIFF", string(result[:4]))
}

// TestProcessAvatar_InvalidImage tests that invalid data returns INVALID_IMAGE error
func TestProcessAvatar_InvalidImage(t *testing.T) {
	processor := NewImageProcessor()
	invalidData := []byte("not an image")

	result, err := processor.ProcessAvatar(invalidData)

	assert.Nil(t, result)
	var imgErr *models.ImageProcessingError
	require.True(t, errors.As(err, &imgErr))
	assert.Equal(t, "INVALID_IMAGE", imgErr.Code)
}

// TestProcessAvatar_EmptyData tests that empty data returns INVALID_IMAGE error
func TestProcessAvatar_EmptyData(t *testing.T) {
	processor := NewImageProcessor()

	result, err := processor.ProcessAvatar([]byte{})

	assert.Nil(t, result)
	var imgErr *models.ImageProcessingError
	require.True(t, errors.As(err, &imgErr))
	assert.Equal(t, "INVALID_IMAGE", imgErr.Code)
}

// TestProcessAvatar_NilData tests that nil data returns INVALID_IMAGE error
func TestProcessAvatar_NilData(t *testing.T) {
	processor := NewImageProcessor()

	result, err := processor.ProcessAvatar(nil)

	assert.Nil(t, result)
	var imgErr *models.ImageProcessingError
	require.True(t, errors.As(err, &imgErr))
	assert.Equal(t, "INVALID_IMAGE", imgErr.Code)
}

// TestProcessAvatar_ImageTooSmall tests that images below 100x100 return IMAGE_TOO_SMALL error
func TestProcessAvatar_ImageTooSmall(t *testing.T) {
	processor := NewImageProcessor()
	smallJPEG := createTestJPEG(50, 50)

	result, err := processor.ProcessAvatar(smallJPEG)

	assert.Nil(t, result)
	var imgErr *models.ImageProcessingError
	require.True(t, errors.As(err, &imgErr))
	assert.Equal(t, "IMAGE_TOO_SMALL", imgErr.Code)
}

// TestProcessAvatar_ImageTooSmallWidth tests that images with width below 100 return IMAGE_TOO_SMALL
func TestProcessAvatar_ImageTooSmallWidth(t *testing.T) {
	processor := NewImageProcessor()
	smallJPEG := createTestJPEG(50, 200)

	result, err := processor.ProcessAvatar(smallJPEG)

	assert.Nil(t, result)
	var imgErr *models.ImageProcessingError
	require.True(t, errors.As(err, &imgErr))
	assert.Equal(t, "IMAGE_TOO_SMALL", imgErr.Code)
}

// TestProcessAvatar_ImageTooSmallHeight tests that images with height below 100 return IMAGE_TOO_SMALL
func TestProcessAvatar_ImageTooSmallHeight(t *testing.T) {
	processor := NewImageProcessor()
	smallJPEG := createTestJPEG(200, 50)

	result, err := processor.ProcessAvatar(smallJPEG)

	assert.Nil(t, result)
	var imgErr *models.ImageProcessingError
	require.True(t, errors.As(err, &imgErr))
	assert.Equal(t, "IMAGE_TOO_SMALL", imgErr.Code)
}

// TestProcessAvatar_ExactMinDimensions tests that exactly 100x100 images are accepted
func TestProcessAvatar_ExactMinDimensions(t *testing.T) {
	processor := NewImageProcessor()
	minJPEG := createTestJPEG(100, 100)

	result, err := processor.ProcessAvatar(minJPEG)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

// TestProcessAvatar_ResizesLargeImage tests that images larger than 512x512 are resized
func TestProcessAvatar_ResizesLargeImage(t *testing.T) {
	processor := NewImageProcessor()
	largeJPEG := createTestJPEG(1000, 800)

	result, err := processor.ProcessAvatar(largeJPEG)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// Result should be smaller than original (due to resize and WebP compression)
	assert.Less(t, len(result), len(largeJPEG))
}

// TestProcessAvatar_CropsToSquare tests that non-square images are center-cropped to square
func TestProcessAvatar_CropsToSquare(t *testing.T) {
	processor := NewImageProcessor()
	wideJPEG := createTestJPEG(400, 200)

	result, err := processor.ProcessAvatar(wideJPEG)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	// We can't easily verify the crop without decoding, but at least it shouldn't error
}

// TestProcessAvatar_TallImageCropsToSquare tests that tall images are center-cropped
func TestProcessAvatar_TallImageCropsToSquare(t *testing.T) {
	processor := NewImageProcessor()
	tallJPEG := createTestJPEG(200, 400)

	result, err := processor.ProcessAvatar(tallJPEG)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

// TestProcessAvatar_OutputSizeReasonable tests that output is not excessively large
func TestProcessAvatar_OutputSizeReasonable(t *testing.T) {
	processor := NewImageProcessor()
	// Create a larger image
	largeJPEG := createTestJPEG(512, 512)

	result, err := processor.ProcessAvatar(largeJPEG)

	assert.NoError(t, err)
	// Output should be under 500KB
	assert.LessOrEqual(t, len(result), 500*1024)
}
