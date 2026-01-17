package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorageService implements services.StorageService for testing
type MockStorageService struct {
	UploadError      error
	SignedURLError   error
	DeleteError      error
	UploadedData     []byte
	LastObjectKey    string
	LastUserID       string
	GeneratedURL     string
	GeneratedExpiry  time.Time
}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		GeneratedURL:    "https://storage.example.com/signed-url",
		GeneratedExpiry: time.Now().Add(7 * 24 * time.Hour),
	}
}

func (m *MockStorageService) UploadAvatar(ctx context.Context, userID string, imageData []byte) (string, error) {
	if m.UploadError != nil {
		return "", m.UploadError
	}
	m.LastUserID = userID
	m.UploadedData = imageData
	m.LastObjectKey = "avatars/" + userID + "/test-uuid.webp"
	return m.LastObjectKey, nil
}

func (m *MockStorageService) GenerateSignedURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	if m.SignedURLError != nil {
		return "", m.SignedURLError
	}
	return m.GeneratedURL, nil
}

func (m *MockStorageService) DeleteAvatar(ctx context.Context, objectKey string) error {
	if m.DeleteError != nil {
		return m.DeleteError
	}
	return nil
}

// MockImageProcessor implements services.ImageProcessor for testing
type MockImageProcessor struct {
	ProcessError   error
	ProcessedData  []byte
	LastInputData  []byte
}

func NewMockImageProcessor() *MockImageProcessor {
	return &MockImageProcessor{
		// WebP magic bytes: RIFF....WEBP
		ProcessedData: []byte("RIFF\x00\x00\x00\x00WEBP"),
	}
}

func (m *MockImageProcessor) ProcessAvatar(imageData []byte) ([]byte, error) {
	m.LastInputData = imageData
	if m.ProcessError != nil {
		return nil, m.ProcessError
	}
	return m.ProcessedData, nil
}

// MockPlayerRepositoryForProfile is a mock player repository for profile tests
type MockPlayerRepositoryForProfile struct {
	players         map[string]*models.PlayerData
	updateAvatarErr error
	clearAvatarErr  error
}

func NewMockPlayerRepositoryForProfile() *MockPlayerRepositoryForProfile {
	return &MockPlayerRepositoryForProfile{
		players: make(map[string]*models.PlayerData),
	}
}

func (m *MockPlayerRepositoryForProfile) GetPlayer(playerId string) (models.Player, error) {
	if player, exists := m.players[playerId]; exists {
		return player, nil
	}
	return nil, nil
}

func (m *MockPlayerRepositoryForProfile) GetPlayers(gameId string) ([]models.Player, error) {
	return nil, nil
}

func (m *MockPlayerRepositoryForProfile) CreatePlayer(ctx context.Context, player *models.PlayerData) error {
	m.players[player.ID] = player
	return nil
}

func (m *MockPlayerRepositoryForProfile) GetPlayerByID(ctx context.Context, id string) (*models.PlayerData, error) {
	if player, exists := m.players[id]; exists {
		return player, nil
	}
	return nil, nil
}

func (m *MockPlayerRepositoryForProfile) GetPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error) {
	return nil, nil
}

func (m *MockPlayerRepositoryForProfile) GetPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error) {
	return nil, nil
}

func (m *MockPlayerRepositoryForProfile) GetPlayerByName(ctx context.Context, name string) (*models.PlayerData, error) {
	return nil, nil
}

func (m *MockPlayerRepositoryForProfile) UpdatePlayer(ctx context.Context, player *models.PlayerData) error {
	m.players[player.ID] = player
	return nil
}

func (m *MockPlayerRepositoryForProfile) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	return nil
}

func (m *MockPlayerRepositoryForProfile) GetAuthToken(ctx context.Context, token string) (*models.AuthToken, error) {
	return nil, nil
}

func (m *MockPlayerRepositoryForProfile) UpdateAuthToken(ctx context.Context, token *models.AuthToken) error {
	return nil
}

func (m *MockPlayerRepositoryForProfile) DeleteAuthToken(ctx context.Context, token string) error {
	return nil
}

func (m *MockPlayerRepositoryForProfile) DeleteExpiredTokens(ctx context.Context) error {
	return nil
}

func (m *MockPlayerRepositoryForProfile) DeletePlayer(ctx context.Context, playerID string) error {
	delete(m.players, playerID)
	return nil
}

func (m *MockPlayerRepositoryForProfile) UpdateAvatar(ctx context.Context, playerID string, objectKey string, signedURL string, expiresAt time.Time) error {
	if m.updateAvatarErr != nil {
		return m.updateAvatarErr
	}
	if player, exists := m.players[playerID]; exists {
		player.AvatarObjectKey = &objectKey
		player.AvatarSignedURL = &signedURL
		player.AvatarSignedURLExpiresAt = &expiresAt
	}
	return nil
}

func (m *MockPlayerRepositoryForProfile) UpdateAvatarSignedURL(ctx context.Context, playerID string, signedURL string, expiresAt time.Time) error {
	if player, exists := m.players[playerID]; exists {
		player.AvatarSignedURL = &signedURL
		player.AvatarSignedURLExpiresAt = &expiresAt
	}
	return nil
}

func (m *MockPlayerRepositoryForProfile) ClearAvatar(ctx context.Context, playerID string) error {
	if m.clearAvatarErr != nil {
		return m.clearAvatarErr
	}
	if player, exists := m.players[playerID]; exists {
		player.AvatarObjectKey = nil
		player.AvatarSignedURL = nil
		player.AvatarSignedURLExpiresAt = nil
	}
	return nil
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()
	authService := &MockAuthService{player: &models.PlayerData{ID: "player-1", Name: "Test Player"}}

	// Create player with avatar
	avatarKey := "avatars/player-1/test.webp"
	avatarURL := "https://storage.example.com/signed"
	expiresAt := time.Now().Add(48 * time.Hour)
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:                       "player-1",
		Name:                     "Test Player",
		AvatarObjectKey:          &avatarKey,
		AvatarSignedURL:          &avatarURL,
		AvatarSignedURLExpiresAt: &expiresAt,
	}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()
	authService := &MockAuthService{player: &models.PlayerData{ID: "player-1", Name: "Test Player"}}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()
	authService := &MockAuthService{player: &models.PlayerData{ID: "player-1", Name: "Test Player"}}

	// Create player without avatar
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:   "player-1",
		Name: "Test Player",
	}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	storageService.GeneratedURL = "https://storage.example.com/new-signed-url"
	imageProcessor := NewMockImageProcessor()
	authService := &MockAuthService{player: &models.PlayerData{ID: "player-1", Name: "Test Player"}}

	// Create player with expired URL (expires in 12 hours, should refresh since < 24h)
	avatarKey := "avatars/player-1/test.webp"
	oldURL := "https://storage.example.com/old-signed-url"
	expiresAt := time.Now().Add(12 * time.Hour) // Within 24h refresh window
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:                       "player-1",
		Name:                     "Test Player",
		AvatarObjectKey:          &avatarKey,
		AvatarSignedURL:          &oldURL,
		AvatarSignedURLExpiresAt: &expiresAt,
	}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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
}

func TestUploadAvatar_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()
	imageProcessor.ProcessError = &models.ImageProcessingError{Code: "INVALID_IMAGE", Reason: "cannot decode"}

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()
	imageProcessor.ProcessError = &models.ImageProcessingError{Code: "IMAGE_TOO_SMALL", Reason: "below 100x100"}

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	storageService.UploadError = &models.StorageError{Reason: "GCS failure"}
	imageProcessor := NewMockImageProcessor()

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()
	authService := &MockAuthService{shouldFail: true}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()

	// Create player with existing avatar
	oldAvatarKey := "avatars/player-1/old.webp"
	player := &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarObjectKey: &oldAvatarKey,
	}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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
	// The old avatar deletion happens (fire and forget), new avatar is set
}

// Test DeleteAvatar endpoint
func TestDeleteAvatar_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()

	avatarKey := "avatars/player-1/test.webp"
	player := &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarObjectKey: &avatarKey,
	}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
	router := gin.New()
	router.DELETE("/players/me/avatar", middleware.PlayerAuth(authService), handler.DeleteAvatar)

	req, _ := http.NewRequest("DELETE", "/players/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify avatar was cleared
	assert.Nil(t, playerRepo.players["player-1"].AvatarObjectKey)
}

func TestDeleteAvatar_NoAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()

	player := &models.PlayerData{
		ID:   "player-1",
		Name: "Test Player",
	}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
	router := gin.New()
	router.DELETE("/players/me/avatar", middleware.PlayerAuth(authService), handler.DeleteAvatar)

	req, _ := http.NewRequest("DELETE", "/players/me/avatar", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should succeed even if no avatar exists
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteAvatar_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()
	authService := &MockAuthService{shouldFail: true}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{
		player: &models.PlayerData{ID: "player-1", Name: "Updated Name"},
	}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()

	player := &models.PlayerData{ID: "player-1", Name: "Test Player"}
	playerRepo.players["player-1"] = player
	authService := &MockAuthService{player: player, shouldFail: true}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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

	playerRepo := NewMockPlayerRepositoryForProfile()
	storageService := NewMockStorageService()
	imageProcessor := NewMockImageProcessor()
	authService := &MockAuthService{shouldFail: true}

	handler := NewProfileHandler(storageService, imageProcessor, playerRepo, authService)
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
