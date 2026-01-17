package services

import (
	"context"
	"errors"
	"ligain/backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorageServiceForProfile implements StorageService for testing
type MockStorageServiceForProfile struct {
	UploadError      error
	SignedURLError   error
	DeleteError      error
	UploadedData     []byte
	LastObjectKey    string
	LastUserID       string
	GeneratedURL     string
	DeletedObjectKey string
}

func NewMockStorageServiceForProfile() *MockStorageServiceForProfile {
	return &MockStorageServiceForProfile{
		GeneratedURL: "https://storage.example.com/signed-url",
	}
}

func (m *MockStorageServiceForProfile) UploadAvatar(ctx context.Context, userID string, imageData []byte) (string, error) {
	if m.UploadError != nil {
		return "", m.UploadError
	}
	m.LastUserID = userID
	m.UploadedData = imageData
	m.LastObjectKey = "avatars/" + userID + "/test-uuid.webp"
	return m.LastObjectKey, nil
}

func (m *MockStorageServiceForProfile) GenerateSignedURL(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	if m.SignedURLError != nil {
		return "", m.SignedURLError
	}
	return m.GeneratedURL, nil
}

func (m *MockStorageServiceForProfile) DeleteAvatar(ctx context.Context, objectKey string) error {
	m.DeletedObjectKey = objectKey
	if m.DeleteError != nil {
		return m.DeleteError
	}
	return nil
}

// MockImageProcessorForProfile implements ImageProcessor for testing
type MockImageProcessorForProfile struct {
	ProcessError  error
	ProcessedData []byte
	LastInputData []byte
}

func NewMockImageProcessorForProfile() *MockImageProcessorForProfile {
	return &MockImageProcessorForProfile{
		ProcessedData: []byte("processed-webp-data"),
	}
}

func (m *MockImageProcessorForProfile) ProcessAvatar(imageData []byte) ([]byte, error) {
	m.LastInputData = imageData
	if m.ProcessError != nil {
		return nil, m.ProcessError
	}
	return m.ProcessedData, nil
}

// MockPlayerRepoForProfile implements repositories.PlayerRepository for testing
type MockPlayerRepoForProfile struct {
	players              map[string]*models.PlayerData
	updateAvatarErr      error
	clearAvatarErr       error
	updateSignedURLErr   error
	lastUpdatedObjectKey string
	lastUpdatedSignedURL string
}

func NewMockPlayerRepoForProfile() *MockPlayerRepoForProfile {
	return &MockPlayerRepoForProfile{
		players: make(map[string]*models.PlayerData),
	}
}

func (m *MockPlayerRepoForProfile) GetPlayer(playerId string) (models.Player, error) {
	if player, exists := m.players[playerId]; exists {
		return player, nil
	}
	return nil, nil
}

func (m *MockPlayerRepoForProfile) GetPlayers(gameId string) ([]models.Player, error) {
	return nil, nil
}

func (m *MockPlayerRepoForProfile) CreatePlayer(ctx context.Context, player *models.PlayerData) error {
	m.players[player.ID] = player
	return nil
}

func (m *MockPlayerRepoForProfile) GetPlayerByID(ctx context.Context, id string) (*models.PlayerData, error) {
	if player, exists := m.players[id]; exists {
		return player, nil
	}
	return nil, nil
}

func (m *MockPlayerRepoForProfile) GetPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error) {
	return nil, nil
}

func (m *MockPlayerRepoForProfile) GetPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error) {
	return nil, nil
}

func (m *MockPlayerRepoForProfile) GetPlayerByName(ctx context.Context, name string) (*models.PlayerData, error) {
	return nil, nil
}

func (m *MockPlayerRepoForProfile) UpdatePlayer(ctx context.Context, player *models.PlayerData) error {
	m.players[player.ID] = player
	return nil
}

func (m *MockPlayerRepoForProfile) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	return nil
}

func (m *MockPlayerRepoForProfile) GetAuthToken(ctx context.Context, token string) (*models.AuthToken, error) {
	return nil, nil
}

func (m *MockPlayerRepoForProfile) UpdateAuthToken(ctx context.Context, token *models.AuthToken) error {
	return nil
}

func (m *MockPlayerRepoForProfile) DeleteAuthToken(ctx context.Context, token string) error {
	return nil
}

func (m *MockPlayerRepoForProfile) DeleteExpiredTokens(ctx context.Context) error {
	return nil
}

func (m *MockPlayerRepoForProfile) DeletePlayer(ctx context.Context, playerID string) error {
	delete(m.players, playerID)
	return nil
}

func (m *MockPlayerRepoForProfile) UpdateAvatar(ctx context.Context, playerID string, objectKey string, signedURL string, expiresAt time.Time) error {
	if m.updateAvatarErr != nil {
		return m.updateAvatarErr
	}
	m.lastUpdatedObjectKey = objectKey
	m.lastUpdatedSignedURL = signedURL
	if player, exists := m.players[playerID]; exists {
		player.AvatarObjectKey = &objectKey
		player.AvatarSignedURL = &signedURL
		player.AvatarSignedURLExpiresAt = &expiresAt
	}
	return nil
}

func (m *MockPlayerRepoForProfile) UpdateAvatarSignedURL(ctx context.Context, playerID string, signedURL string, expiresAt time.Time) error {
	if m.updateSignedURLErr != nil {
		return m.updateSignedURLErr
	}
	m.lastUpdatedSignedURL = signedURL
	if player, exists := m.players[playerID]; exists {
		player.AvatarSignedURL = &signedURL
		player.AvatarSignedURLExpiresAt = &expiresAt
	}
	return nil
}

func (m *MockPlayerRepoForProfile) ClearAvatar(ctx context.Context, playerID string) error {
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

// Tests for ProfileService

func TestProfileService_UploadAvatar_Success(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	playerRepo.players["player-1"] = &models.PlayerData{ID: "player-1", Name: "Test Player"}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	result, err := service.UploadAvatar(ctx, "player-1", nil, []byte("raw-image-data"))

	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify image was processed
	assert.Equal(t, []byte("raw-image-data"), imageProcessor.LastInputData)

	// Verify processed image was uploaded
	assert.Equal(t, imageProcessor.ProcessedData, storageService.UploadedData)
	assert.Equal(t, "player-1", storageService.LastUserID)

	// Verify signed URL was returned
	assert.Equal(t, "https://storage.example.com/signed-url", result.SignedURL)

	// Verify database was updated
	assert.Equal(t, storageService.LastObjectKey, playerRepo.lastUpdatedObjectKey)
	assert.Equal(t, storageService.GeneratedURL, playerRepo.lastUpdatedSignedURL)
}

func TestProfileService_UploadAvatar_ReplacesExisting(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	oldObjectKey := "avatars/player-1/old-uuid.webp"
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarObjectKey: &oldObjectKey,
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	result, err := service.UploadAvatar(ctx, "player-1", &oldObjectKey, []byte("new-image-data"))

	require.NoError(t, err)
	require.NotNil(t, result)

	// Wait for fire-and-forget goroutine to complete
	time.Sleep(10 * time.Millisecond)

	// Verify old avatar deletion was triggered (fire-and-forget)
	assert.Equal(t, oldObjectKey, storageService.DeletedObjectKey)

	// Verify new avatar was uploaded
	assert.Equal(t, "player-1", storageService.LastUserID)
}

func TestProfileService_UploadAvatar_InvalidImage(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	imageProcessor.ProcessError = &models.ImageProcessingError{Code: "INVALID_IMAGE", Reason: "cannot decode"}
	playerRepo := NewMockPlayerRepoForProfile()

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	result, err := service.UploadAvatar(ctx, "player-1", nil, []byte("invalid-data"))

	require.Error(t, err)
	assert.Nil(t, result)

	var imgErr *models.ImageProcessingError
	assert.True(t, errors.As(err, &imgErr))
	assert.Equal(t, "INVALID_IMAGE", imgErr.Code)
}

func TestProfileService_UploadAvatar_StorageError(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	storageService.UploadError = &models.StorageError{Reason: "GCS failure"}
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	result, err := service.UploadAvatar(ctx, "player-1", nil, []byte("image-data"))

	require.Error(t, err)
	assert.Nil(t, result)

	var storageErr *models.StorageError
	assert.True(t, errors.As(err, &storageErr))
}

func TestProfileService_UploadAvatar_SignedURLError(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	storageService.SignedURLError = errors.New("signed URL generation failed")
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	result, err := service.UploadAvatar(ctx, "player-1", nil, []byte("image-data"))

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestProfileService_UploadAvatar_DatabaseError(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()
	playerRepo.updateAvatarErr = errors.New("database error")

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	result, err := service.UploadAvatar(ctx, "player-1", nil, []byte("image-data"))

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestProfileService_DeleteAvatar_Success(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	objectKey := "avatars/player-1/test.webp"
	signedURL := "https://storage.example.com/signed"
	expiresAt := time.Now().Add(48 * time.Hour)
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:                       "player-1",
		Name:                     "Test Player",
		AvatarObjectKey:          &objectKey,
		AvatarSignedURL:          &signedURL,
		AvatarSignedURLExpiresAt: &expiresAt,
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	err := service.DeleteAvatar(ctx, "player-1", &objectKey)

	require.NoError(t, err)

	// Wait for fire-and-forget goroutine to complete
	time.Sleep(10 * time.Millisecond)

	// Verify storage delete was called (fire-and-forget)
	assert.Equal(t, objectKey, storageService.DeletedObjectKey)

	// Verify database was cleared
	assert.Nil(t, playerRepo.players["player-1"].AvatarObjectKey)
	assert.Nil(t, playerRepo.players["player-1"].AvatarSignedURL)
}

func TestProfileService_DeleteAvatar_NoExistingAvatar(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	playerRepo.players["player-1"] = &models.PlayerData{
		ID:   "player-1",
		Name: "Test Player",
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	err := service.DeleteAvatar(ctx, "player-1", nil)

	// Should succeed gracefully
	require.NoError(t, err)

	// Storage delete should not have been called
	assert.Empty(t, storageService.DeletedObjectKey)
}

func TestProfileService_DeleteAvatar_StorageErrorIsIgnored(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	storageService.DeleteError = errors.New("storage error")
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	objectKey := "avatars/player-1/test.webp"
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarObjectKey: &objectKey,
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	err := service.DeleteAvatar(ctx, "player-1", &objectKey)

	// Should still succeed - storage errors are logged but not returned
	require.NoError(t, err)

	// Database should still be cleared
	assert.Nil(t, playerRepo.players["player-1"].AvatarObjectKey)
}

func TestProfileService_DeleteAvatar_DatabaseError(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()
	playerRepo.clearAvatarErr = errors.New("database error")

	objectKey := "avatars/player-1/test.webp"
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:              "player-1",
		Name:            "Test Player",
		AvatarObjectKey: &objectKey,
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	err := service.DeleteAvatar(ctx, "player-1", &objectKey)

	require.Error(t, err)
}

func TestProfileService_GetPlayerProfile_Success(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	avatarKey := "avatars/player-1/test.webp"
	avatarURL := "https://storage.example.com/signed"
	expiresAt := time.Now().Add(48 * time.Hour) // Not expiring soon
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:                       "player-1",
		Name:                     "Test Player",
		AvatarObjectKey:          &avatarKey,
		AvatarSignedURL:          &avatarURL,
		AvatarSignedURLExpiresAt: &expiresAt,
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	player, err := service.GetPlayerProfile(ctx, "player-1")

	require.NoError(t, err)
	require.NotNil(t, player)
	assert.Equal(t, "player-1", player.ID)
	assert.Equal(t, "Test Player", player.Name)
	assert.Equal(t, avatarURL, *player.AvatarSignedURL)
}

func TestProfileService_GetPlayerProfile_NotFound(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	player, err := service.GetPlayerProfile(ctx, "nonexistent")

	require.Error(t, err)
	assert.Nil(t, player)

	var profileErr *ProfileError
	assert.True(t, errors.As(err, &profileErr))
	assert.Equal(t, "PLAYER_NOT_FOUND", profileErr.Code)
}

func TestProfileService_GetPlayerProfile_RefreshesExpiredURL(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	storageService.GeneratedURL = "https://storage.example.com/new-signed-url"
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

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

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	player, err := service.GetPlayerProfile(ctx, "player-1")

	require.NoError(t, err)
	require.NotNil(t, player)

	// Should have refreshed URL
	assert.Equal(t, "https://storage.example.com/new-signed-url", *player.AvatarSignedURL)

	// Wait for fire-and-forget DB update goroutine to complete
	time.Sleep(10 * time.Millisecond)

	// Should have updated database (fire-and-forget)
	assert.Equal(t, "https://storage.example.com/new-signed-url", playerRepo.lastUpdatedSignedURL)
}

func TestProfileService_GetPlayerProfile_RefreshesNilExpiration(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	storageService.GeneratedURL = "https://storage.example.com/new-signed-url"
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	avatarKey := "avatars/player-1/test.webp"
	oldURL := "https://storage.example.com/old-signed-url"
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:                       "player-1",
		Name:                     "Test Player",
		AvatarObjectKey:          &avatarKey,
		AvatarSignedURL:          &oldURL,
		AvatarSignedURLExpiresAt: nil, // No expiration set
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	player, err := service.GetPlayerProfile(ctx, "player-1")

	require.NoError(t, err)
	require.NotNil(t, player)

	// Should have refreshed URL
	assert.Equal(t, "https://storage.example.com/new-signed-url", *player.AvatarSignedURL)
}

func TestProfileService_GetPlayerProfile_URLRefreshErrorDoesNotFail(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	storageService.SignedURLError = errors.New("failed to generate URL")
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	avatarKey := "avatars/player-1/test.webp"
	oldURL := "https://storage.example.com/old-signed-url"
	expiresAt := time.Now().Add(12 * time.Hour) // Within refresh window
	playerRepo.players["player-1"] = &models.PlayerData{
		ID:                       "player-1",
		Name:                     "Test Player",
		AvatarObjectKey:          &avatarKey,
		AvatarSignedURL:          &oldURL,
		AvatarSignedURLExpiresAt: &expiresAt,
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	player, err := service.GetPlayerProfile(ctx, "player-1")

	// Should still succeed, just with old URL
	require.NoError(t, err)
	require.NotNil(t, player)
	assert.Equal(t, oldURL, *player.AvatarSignedURL)
}

func TestProfileService_GetPlayerProfile_NoAvatar(t *testing.T) {
	storageService := NewMockStorageServiceForProfile()
	imageProcessor := NewMockImageProcessorForProfile()
	playerRepo := NewMockPlayerRepoForProfile()

	playerRepo.players["player-1"] = &models.PlayerData{
		ID:   "player-1",
		Name: "Test Player",
	}

	service := NewProfileService(storageService, imageProcessor, playerRepo)

	ctx := context.Background()
	player, err := service.GetPlayerProfile(ctx, "player-1")

	require.NoError(t, err)
	require.NotNil(t, player)
	assert.Nil(t, player.AvatarSignedURL)
}
