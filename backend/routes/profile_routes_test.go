package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/services"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockProfileService implements services.ProfileService for testing
type MockProfileService struct {
	uploadAvatarResult *services.AvatarResult
	uploadAvatarError  error
	deleteAvatarError  error
	getPlayerResult    *models.PlayerData
	getPlayerError     error

	// Track calls
	lastUploadPlayerID    string
	lastUploadOldKey      *string
	lastUploadImageData   []byte
	lastDeletePlayerID    string
	lastDeleteObjectKey   *string
	lastGetPlayerPlayerID string
}

func NewMockProfileService() *MockProfileService {
	return &MockProfileService{
		uploadAvatarResult: &services.AvatarResult{
			SignedURL: "https://storage.example.com/signed-url",
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		},
	}
}

func (m *MockProfileService) UploadAvatar(ctx context.Context, playerID string, oldObjectKey *string, imageData []byte) (*services.AvatarResult, error) {
	m.lastUploadPlayerID = playerID
	m.lastUploadOldKey = oldObjectKey
	m.lastUploadImageData = imageData
	if m.uploadAvatarError != nil {
		return nil, m.uploadAvatarError
	}
	return m.uploadAvatarResult, nil
}

func (m *MockProfileService) DeleteAvatar(ctx context.Context, playerID string, objectKey *string) error {
	m.lastDeletePlayerID = playerID
	m.lastDeleteObjectKey = objectKey
	return m.deleteAvatarError
}

func (m *MockProfileService) GetPlayerProfile(ctx context.Context, playerID string) (*models.PlayerData, error) {
	m.lastGetPlayerPlayerID = playerID
	if m.getPlayerError != nil {
		return nil, m.getPlayerError
	}
	return m.getPlayerResult, nil
}

// Helper to create a test JPEG image for upload tests
func createTestJPEGForUpload(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

// Helper to create a multipart form with an image file
func createMultipartForm(t *testing.T, fieldName string, fileName string, data []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	require.NoError(t, err)

	_, err = io.Copy(part, bytes.NewReader(data))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	return body, writer.FormDataContentType()
}

// Test GetPlayer endpoint
func TestGetPlayer_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	avatarURL := "https://storage.example.com/signed"
	profileService := NewMockProfileService()
	profileService.getPlayerResult = &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarSignedURL: &avatarURL,
	}
	authService := &MockAuthService{player: &models.PlayerData{ID: "player-1", Name: "Test Player"}}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.GET("/players/:id", middleware.PlayerAuth(authService), handler.GetPlayer)

	req, _ := http.NewRequest("GET", "/players/player-1", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	player := response["player"].(map[string]interface{})
	assert.Equal(t, "player-1", player["id"])
	assert.Equal(t, "Test Player", player["name"])
	assert.Equal(t, avatarURL, player["avatar_url"])
}

func TestGetPlayer_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	profileService.getPlayerError = &services.ProfileError{Code: "PLAYER_NOT_FOUND", Reason: "player not found"}
	authService := &MockAuthService{player: &models.PlayerData{ID: "player-1", Name: "Test Player"}}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.GET("/players/:id", middleware.PlayerAuth(authService), handler.GetPlayer)

	req, _ := http.NewRequest("GET", "/players/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetPlayer_NoAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	profileService.getPlayerResult = &models.PlayerData{
		ID:   "player-1",
		Name: "Test Player",
	}
	authService := &MockAuthService{player: &models.PlayerData{ID: "player-1", Name: "Test Player"}}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.GET("/players/:id", middleware.PlayerAuth(authService), handler.GetPlayer)

	req, _ := http.NewRequest("GET", "/players/player-1", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	player := response["player"].(map[string]interface{})
	_, hasAvatarURL := player["avatar_url"]
	assert.False(t, hasAvatarURL)
}

func TestGetPlayer_RefreshesExpiredURL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// The service handles URL refresh internally, we just verify we get the refreshed URL
	newURL := "https://storage.example.com/new-signed-url"
	profileService := NewMockProfileService()
	profileService.getPlayerResult = &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarSignedURL: &newURL,
	}
	authService := &MockAuthService{player: &models.PlayerData{ID: "player-1", Name: "Test Player"}}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.GET("/players/:id", middleware.PlayerAuth(authService), handler.GetPlayer)

	req, _ := http.NewRequest("GET", "/players/player-1", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	player := response["player"].(map[string]interface{})
	assert.Equal(t, "https://storage.example.com/new-signed-url", player["avatar_url"])
}

// Test UploadAvatar endpoint
func TestUploadAvatar_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.POST("/players/me/avatar", middleware.PlayerAuth(authService), handler.UploadAvatar)

	// Create test image
	imageData := createTestJPEGForUpload(200, 200)
	body, contentType := createMultipartForm(t, "avatar", "test.jpg", imageData)

	req, _ := http.NewRequest("POST", "/players/me/avatar", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "https://storage.example.com/signed-url", response["avatar_url"])

	// Verify service was called with correct params
	assert.Equal(t, "player-1", profileService.lastUploadPlayerID)
	assert.Nil(t, profileService.lastUploadOldKey)
}

func TestUploadAvatar_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.POST("/players/me/avatar", middleware.PlayerAuth(authService), handler.UploadAvatar)

	req, _ := http.NewRequest("POST", "/players/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INVALID_IMAGE", response["code"])
}

func TestUploadAvatar_InvalidImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	profileService.uploadAvatarError = &models.ImageProcessingError{Code: "INVALID_IMAGE", Reason: "cannot decode"}

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.POST("/players/me/avatar", middleware.PlayerAuth(authService), handler.UploadAvatar)

	// Send invalid data
	body, contentType := createMultipartForm(t, "avatar", "test.jpg", []byte("not an image"))

	req, _ := http.NewRequest("POST", "/players/me/avatar", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "INVALID_IMAGE", response["code"])
}

func TestUploadAvatar_ImageTooSmall(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	profileService.uploadAvatarError = &models.ImageProcessingError{Code: "IMAGE_TOO_SMALL", Reason: "below 100x100"}

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.POST("/players/me/avatar", middleware.PlayerAuth(authService), handler.UploadAvatar)

	// Create small test image
	imageData := createTestJPEGForUpload(50, 50)
	body, contentType := createMultipartForm(t, "avatar", "test.jpg", imageData)

	req, _ := http.NewRequest("POST", "/players/me/avatar", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "IMAGE_TOO_SMALL", response["code"])
}

func TestUploadAvatar_StorageError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	profileService.uploadAvatarError = errors.New("storage error")

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.POST("/players/me/avatar", middleware.PlayerAuth(authService), handler.UploadAvatar)

	imageData := createTestJPEGForUpload(200, 200)
	body, contentType := createMultipartForm(t, "avatar", "test.jpg", imageData)

	req, _ := http.NewRequest("POST", "/players/me/avatar", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "UPLOAD_FAILED", response["code"])
}

func TestUploadAvatar_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	authService := &MockAuthService{shouldFail: true}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.POST("/players/me/avatar", middleware.PlayerAuth(authService), handler.UploadAvatar)

	imageData := createTestJPEGForUpload(200, 200)
	body, contentType := createMultipartForm(t, "avatar", "test.jpg", imageData)

	req, _ := http.NewRequest("POST", "/players/me/avatar", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUploadAvatar_ReplacesExistingAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()

	// Create player with existing avatar
	oldAvatarKey := "avatars/player-1/old.webp"
	player := &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarObjectKey: &oldAvatarKey,
	}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.POST("/players/me/avatar", middleware.PlayerAuth(authService), handler.UploadAvatar)

	imageData := createTestJPEGForUpload(200, 200)
	body, contentType := createMultipartForm(t, "avatar", "test.jpg", imageData)

	req, _ := http.NewRequest("POST", "/players/me/avatar", body)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify old key was passed to service
	assert.Equal(t, &oldAvatarKey, profileService.lastUploadOldKey)
}

// Test DeleteAvatar endpoint
func TestDeleteAvatar_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()

	avatarKey := "avatars/player-1/test.webp"
	player := &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarObjectKey: &avatarKey,
	}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.DELETE("/players/me/avatar", middleware.PlayerAuth(authService), handler.DeleteAvatar)

	req, _ := http.NewRequest("DELETE", "/players/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify service was called with correct params
	assert.Equal(t, "player-1", profileService.lastDeletePlayerID)
	assert.Equal(t, &avatarKey, profileService.lastDeleteObjectKey)
}

func TestDeleteAvatar_NoAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()

	player := &models.PlayerData{
		ID:   "player-1",
		Name: "Test Player",
	}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.DELETE("/players/me/avatar", middleware.PlayerAuth(authService), handler.DeleteAvatar)

	req, _ := http.NewRequest("DELETE", "/players/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should succeed even if no avatar exists
	assert.Equal(t, http.StatusOK, w.Code)

	// Service should not have been called (handler returns early)
	assert.Empty(t, profileService.lastDeletePlayerID)
}

func TestDeleteAvatar_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	authService := &MockAuthService{shouldFail: true}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.DELETE("/players/me/avatar", middleware.PlayerAuth(authService), handler.DeleteAvatar)

	req, _ := http.NewRequest("DELETE", "/players/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Test UpdateDisplayName endpoint
func TestUpdateDisplayName_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	authService := &MockAuthService{
		player: &models.PlayerData{ID: "player-1", Name: "Updated Name"},
	}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.PUT("/players/me/display-name", middleware.PlayerAuth(authService), handler.UpdateDisplayName)

	requestBody := map[string]string{"displayName": "Updated Name"}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/players/me/display-name", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	playerResp := response["player"].(map[string]interface{})
	assert.Equal(t, "Updated Name", playerResp["name"])
}

func TestUpdateDisplayName_InvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.PUT("/players/me/display-name", middleware.PlayerAuth(authService), handler.UpdateDisplayName)

	req, _ := http.NewRequest("PUT", "/players/me/display-name", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Invalid request body", response["error"])
}

func TestUpdateDisplayName_AuthServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	authService := &MockAuthService{player: player, shouldFail: true}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.PUT("/players/me/display-name", middleware.PlayerAuth(authService), handler.UpdateDisplayName)

	requestBody := map[string]string{"displayName": "Invalid Name"}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/players/me/display-name", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	// Need to manually set context since the middleware validates but we want to test handler error
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("player", player)

	handler.UpdateDisplayName(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"].(string), "mock update display name failed")
}

func TestUpdateDisplayName_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	profileService := NewMockProfileService()
	authService := &MockAuthService{shouldFail: true}

	handler := NewProfileHandler(profileService, authService)
	router := gin.New()
	router.PUT("/players/me/display-name", middleware.PlayerAuth(authService), handler.UpdateDisplayName)

	requestBody := map[string]string{"displayName": "New Name"}
	jsonBody, _ := json.Marshal(requestBody)

	req, _ := http.NewRequest("PUT", "/players/me/display-name", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer bad-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
