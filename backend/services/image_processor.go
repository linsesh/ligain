package services

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"ligain/backend/models"

	"github.com/chai2010/webp"
	"golang.org/x/image/draw"
)

const (
	minDimension     = 100
	maxDimension     = 512
	webpQuality      = 85
	maxOutputSizeKB  = 500
)

// ImageProcessor defines the interface for avatar image processing
type ImageProcessor interface {
	// ProcessAvatar validates and transforms an avatar image
	// Input: raw image bytes (JPEG, PNG, or WebP)
	// Output: processed WebP image (max 512x512, square, stripped metadata)
	ProcessAvatar(imageData []byte) ([]byte, error)
}

// ImageProcessorImpl implements ImageProcessor
type ImageProcessorImpl struct{}

// NewImageProcessor creates a new ImageProcessor instance
func NewImageProcessor() ImageProcessor {
	return &ImageProcessorImpl{}
}

// ProcessAvatar processes an avatar image
func (p *ImageProcessorImpl) ProcessAvatar(imageData []byte) ([]byte, error) {
	// Validate input
	if len(imageData) == 0 {
		return nil, &models.ImageProcessingError{
			Code:   "INVALID_IMAGE",
			Reason: "image data is empty",
		}
	}

	// Decode image (try JPEG, PNG, then WebP)
	img, err := decodeImage(imageData)
	if err != nil {
		return nil, &models.ImageProcessingError{
			Code:   "INVALID_IMAGE",
			Reason: "unable to decode image: " + err.Error(),
		}
	}

	// Validate dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width < minDimension || height < minDimension {
		return nil, &models.ImageProcessingError{
			Code:   "IMAGE_TOO_SMALL",
			Reason: "image must be at least 100x100 pixels",
		}
	}

	// Resize if necessary (maintain aspect ratio, max 512 on longest side)
	img = resizeIfNeeded(img, maxDimension)

	// Center crop to square
	img = centerCropToSquare(img)

	// Encode as WebP
	output, err := encodeWebP(img, webpQuality)
	if err != nil {
		return nil, &models.ImageProcessingError{
			Code:   "INVALID_IMAGE",
			Reason: "failed to encode as WebP: " + err.Error(),
		}
	}

	// Check output size
	if len(output) > maxOutputSizeKB*1024 {
		return nil, &models.ImageProcessingError{
			Code:   "FILE_TOO_LARGE",
			Reason: "processed image exceeds 500KB limit",
		}
	}

	return output, nil
}

// decodeImage tries to decode image data as JPEG, PNG, or WebP
func decodeImage(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)

	// Try standard library decode which auto-detects format
	img, _, err := image.Decode(reader)
	if err == nil {
		return img, nil
	}

	// Try WebP explicitly
	reader.Reset(data)
	img, err = webp.Decode(reader)
	if err == nil {
		return img, nil
	}

	// Try JPEG explicitly
	reader.Reset(data)
	img, err = jpeg.Decode(reader)
	if err == nil {
		return img, nil
	}

	// Try PNG explicitly
	reader.Reset(data)
	img, err = png.Decode(reader)
	if err == nil {
		return img, nil
	}

	return nil, err
}

// resizeIfNeeded resizes the image if it exceeds maxSize on either dimension
func resizeIfNeeded(img image.Image, maxSize int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Check if resize is needed
	if width <= maxSize && height <= maxSize {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight int
	if width > height {
		newWidth = maxSize
		newHeight = height * maxSize / width
	} else {
		newHeight = maxSize
		newWidth = width * maxSize / height
	}

	// Create resized image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}

// centerCropToSquare crops the image to a square from the center
func centerCropToSquare(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Already square
	if width == height {
		return img
	}

	// Calculate crop dimensions
	var cropSize int
	var x, y int
	if width > height {
		cropSize = height
		x = (width - height) / 2
		y = 0
	} else {
		cropSize = width
		x = 0
		y = (height - width) / 2
	}

	// Create cropped image
	dst := image.NewRGBA(image.Rect(0, 0, cropSize, cropSize))
	draw.Copy(dst, image.Point{}, img, image.Rect(x, y, x+cropSize, y+cropSize), draw.Over, nil)

	return dst
}

// encodeWebP encodes an image as WebP with the specified quality
func encodeWebP(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	err := webp.Encode(&buf, img, &webp.Options{
		Lossless: false,
		Quality:  float32(quality),
	})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
