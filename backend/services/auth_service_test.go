package services

import (
	"context"
	"errors"
	"fmt"
	"ligain/backend/models"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// frozenTime is a fixed time for consistent testing
var frozenTime = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

type MockPlayerRepository struct {
	players map[string]*models.PlayerData
	tokens  map[string]*models.AuthToken
}

func NewMockPlayerRepository() *MockPlayerRepository {
	return &MockPlayerRepository{
		players: make(map[string]*models.PlayerData),
		tokens:  make(map[string]*models.AuthToken),
	}
}

func (m *MockPlayerRepository) GetPlayer(playerId string) (models.Player, error) {
	if player, exists := m.players[playerId]; exists {
		return player, nil
	}
	return &models.PlayerData{}, nil
}

func (m *MockPlayerRepository) GetPlayers(gameId string) ([]models.Player, error) {
	var players []models.Player
	for _, player := range m.players {
		players = append(players, player)
	}
	return players, nil
}

func (m *MockPlayerRepository) CreatePlayer(ctx context.Context, player *models.PlayerData) error {
	if player.ID == "" {
		player.ID = "mock_id_" + frozenTime.String()
	}
	m.players[player.ID] = player
	return nil
}

func (m *MockPlayerRepository) GetPlayerByID(ctx context.Context, id string) (*models.PlayerData, error) {
	if player, exists := m.players[id]; exists {
		return player, nil
	}
	return nil, nil
}

func (m *MockPlayerRepository) GetPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error) {
	for _, player := range m.players {
		if player.Email != nil && *player.Email == email {
			return player, nil
		}
	}
	return nil, nil
}

func (m *MockPlayerRepository) GetPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error) {
	for _, player := range m.players {
		if player.Provider != nil && *player.Provider == provider &&
			player.ProviderID != nil && *player.ProviderID == providerID {
			return player, nil
		}
	}
	return nil, nil
}

func (m *MockPlayerRepository) GetPlayerByName(ctx context.Context, name string) (*models.PlayerData, error) {
	for _, player := range m.players {
		if player.Name == name {
			return player, nil
		}
	}
	return nil, nil
}

func (m *MockPlayerRepository) UpdatePlayer(ctx context.Context, player *models.PlayerData) error {
	m.players[player.ID] = player
	return nil
}

func (m *MockPlayerRepository) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	m.tokens[token.Token] = token
	return nil
}

func (m *MockPlayerRepository) GetAuthToken(ctx context.Context, token string) (*models.AuthToken, error) {
	if authToken, exists := m.tokens[token]; exists {
		return authToken, nil
	}
	return nil, nil
}

func (m *MockPlayerRepository) DeleteAuthToken(ctx context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func (m *MockPlayerRepository) DeleteExpiredTokens(ctx context.Context) error {
	now := frozenTime
	for token, authToken := range m.tokens {
		if now.After(authToken.ExpiresAt) {
			delete(m.tokens, token)
		}
	}
	return nil
}

func (m *MockPlayerRepository) UpdateAuthToken(ctx context.Context, token *models.AuthToken) error {
	// In-memory implementation - update in a simple map
	// Since we're using a map with token as key, we need to delete and recreate
	delete(m.tokens, token.Token)
	m.tokens[token.Token] = token
	return nil
}

func TestAuthService_Authenticate_NewUser(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()
	req := &models.AuthRequest{
		Provider: "google",
		Token:    "mock_google_token_123",
		Email:    "test@google.com",
		Name:     "Test Display Name",
	}

	response, err := authService.Authenticate(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Player.Name != "Test Display Name" {
		t.Errorf("Expected player name 'Test Display Name', got %s", response.Player.Name)
	}

	if response.Token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify player was created in repository
	player, err := mockRepo.GetPlayerByEmail(ctx, "test@google.com")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if player == nil {
		t.Error("Expected player to be created in repository")
	}

	if *player.Provider != "google" {
		t.Errorf("Expected provider 'google', got %s", *player.Provider)
	}

	if *player.ProviderID != "google_user_123" {
		t.Errorf("Expected provider ID 'google_user_123', got %s", *player.ProviderID)
	}
}

func TestAuthService_Authenticate_ExistingUser(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create existing player
	existingPlayer := &models.PlayerData{
		ID:         "existing_id",
		Name:       "Old Name",
		Email:      stringPtr("test@google.com"),
		Provider:   stringPtr("google"),
		ProviderID: stringPtr("google_user_123"),
	}
	mockRepo.CreatePlayer(ctx, existingPlayer)

	req := &models.AuthRequest{
		Provider: "google",
		Token:    "mock_google_token_123",
		Email:    "test@google.com",
		Name:     "New Display Name",
	}

	response, err := authService.Authenticate(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response.Player.Name != "New Display Name" {
		t.Errorf("Expected updated name 'New Display Name', got %s", response.Player.Name)
	}

	// Verify player was updated in repository
	player, err := mockRepo.GetPlayerByID(ctx, "existing_id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if player.Name != "New Display Name" {
		t.Errorf("Expected updated name in repository 'New Display Name', got %s", player.Name)
	}
}

func TestAuthService_Authenticate_ExistingUser_NoDisplayName(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create existing player
	existingPlayer := &models.PlayerData{
		ID:         "existing_id",
		Name:       "Existing Name",
		Email:      stringPtr("test@google.com"),
		Provider:   stringPtr("google"),
		ProviderID: stringPtr("google_user_123"),
	}
	mockRepo.CreatePlayer(ctx, existingPlayer)

	// Authenticate without providing a display name (should work for existing users)
	req := &models.AuthRequest{
		Provider: "google",
		Token:    "mock_google_token_123",
		Email:    "test@google.com",
		Name:     "", // Empty display name
	}

	response, err := authService.Authenticate(ctx, req)
	if err != nil {
		t.Fatalf("Expected no error for existing user without display name, got %v", err)
	}

	// Should return the existing player with their original name
	if response.Player.Name != "Existing Name" {
		t.Errorf("Expected existing name 'Existing Name', got %s", response.Player.Name)
	}

	// Verify player was NOT updated in repository (name should remain the same)
	player, err := mockRepo.GetPlayerByID(ctx, "existing_id")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if player.Name != "Existing Name" {
		t.Errorf("Expected unchanged name in repository 'Existing Name', got %s", player.Name)
	}
}

func TestAuthService_Authenticate_UnsupportedProvider(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()
	req := &models.AuthRequest{
		Provider: "facebook", // Unsupported provider
		Token:    "mock_token",
		Email:    "test@example.com",
		Name:     "Test User",
	}

	_, err := authService.Authenticate(ctx, req)
	if err == nil {
		t.Error("Expected error for unsupported provider")
	}

	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Errorf("Expected error containing 'unsupported provider', got %s", err.Error())
	}
}

func TestAuthService_Authenticate_OAuthVerificationFailure(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifierWithFailure(), func() time.Time { return frozenTime })

	ctx := context.Background()
	req := &models.AuthRequest{
		Provider: "google",
		Token:    "invalid_token",
		Email:    "test@example.com",
		Name:     "Test User",
	}

	_, err := authService.Authenticate(ctx, req)
	if err == nil {
		t.Error("Expected error for OAuth verification failure")
	}

	if !strings.Contains(err.Error(), "failed to verify token") {
		t.Errorf("Expected OAuth verification error, got %v", err)
	}
}

func TestAuthService_ValidateToken_ValidToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create a test player
	player := &models.PlayerData{
		ID:   "test_player_id",
		Name: "Test Player",
	}
	mockRepo.CreatePlayer(ctx, player)

	// Create a valid auth token
	authToken := &models.AuthToken{
		PlayerID:  player.ID,
		Token:     "valid_token",
		ExpiresAt: frozenTime.Add(1 * time.Hour),
	}
	mockRepo.CreateAuthToken(ctx, authToken)

	// Test valid token
	validatedPlayer, err := authService.ValidateToken(ctx, "valid_token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if validatedPlayer.ID != player.ID {
		t.Errorf("Expected player ID %s, got %s", player.ID, validatedPlayer.ID)
	}
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Test invalid token
	_, err := authService.ValidateToken(ctx, "invalid_token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}

	if !strings.Contains(err.Error(), "invalid token") {
		t.Errorf("Expected error containing 'invalid token', got %s", err.Error())
	}
}

func TestAuthService_ValidateToken_ExpiredToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create a test player
	player := &models.PlayerData{
		ID:   "test_player_id",
		Name: "Test Player",
	}
	mockRepo.CreatePlayer(ctx, player)

	// Create an expired auth token (expired relative to frozenTime)
	authToken := &models.AuthToken{
		PlayerID:  player.ID,
		Token:     "expired_token",
		ExpiresAt: frozenTime.Add(-25 * time.Hour), // Expired (more than 24 hours)
	}
	mockRepo.CreateAuthToken(ctx, authToken)

	_, err := authService.ValidateToken(ctx, "expired_token")
	if err == nil {
		t.Error("Expected error for expired token")
		return
	}

	if !strings.Contains(err.Error(), "token expired") {
		t.Errorf("Expected error containing 'token expired', got %s", err.Error())
	}
}

func TestAuthService_Logout(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create a test auth token
	authToken := &models.AuthToken{
		PlayerID:  "test_player_id",
		Token:     "test_token",
		ExpiresAt: frozenTime.Add(1 * time.Hour),
	}
	mockRepo.CreateAuthToken(ctx, authToken)

	// Test logout
	err := authService.Logout(ctx, "test_token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify token was deleted
	deletedToken, err := mockRepo.GetAuthToken(ctx, "test_token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if deletedToken != nil {
		t.Error("Expected token to be deleted")
	}
}

func TestAuthService_GenerateAuthToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()
	playerID := "test_player_id"

	token, err := authService.generateAuthToken(ctx, playerID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify token was created in repository
	authToken, err := mockRepo.GetAuthToken(ctx, token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if authToken == nil {
		t.Error("Expected auth token to be created in repository")
	}

	if authToken.PlayerID != playerID {
		t.Errorf("Expected player ID %s, got %s", playerID, authToken.PlayerID)
	}

	// Verify token expires in 24 hours
	expectedExpiry := frozenTime.Add(24 * time.Hour)
	if authToken.ExpiresAt != expectedExpiry {
		t.Errorf("Expected token to expire at %v, got %v", expectedExpiry, authToken.ExpiresAt)
	}
}

func TestAuthService_UpdateDisplayName(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create a test player
	player := &models.PlayerData{
		ID:         "test_player_id",
		Name:       "Old Name",
		Email:      stringPtr("test@example.com"),
		Provider:   stringPtr("google"),
		ProviderID: stringPtr("google_user_123"),
	}
	mockRepo.CreatePlayer(ctx, player)

	t.Run("Valid display name update", func(t *testing.T) {
		updatedPlayer, err := authService.UpdateDisplayName(ctx, "test_player_id", "New Name")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if updatedPlayer.Name != "New Name" {
			t.Errorf("Expected name 'New Name', got %s", updatedPlayer.Name)
		}

		// Verify player was updated in repository
		player, err := mockRepo.GetPlayerByID(ctx, "test_player_id")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if player.Name != "New Name" {
			t.Errorf("Expected updated name in repository 'New Name', got %s", player.Name)
		}
	})

	t.Run("Empty display name", func(t *testing.T) {
		_, err := authService.UpdateDisplayName(ctx, "test_player_id", "")
		if err == nil {
			t.Error("Expected error for empty display name")
		}
		if !strings.Contains(err.Error(), "display name cannot be empty") {
			t.Errorf("Expected error containing 'display name cannot be empty', got %s", err.Error())
		}
	})

	t.Run("Display name too short", func(t *testing.T) {
		_, err := authService.UpdateDisplayName(ctx, "test_player_id", "A")
		if err == nil {
			t.Error("Expected error for short display name")
		}
		if !strings.Contains(err.Error(), "display name must be at least 2 characters long") {
			t.Errorf("Expected error containing 'display name must be at least 2 characters long', got %s", err.Error())
		}
	})

	t.Run("Display name too long", func(t *testing.T) {
		longName := "ThisNameIsWayTooLongAndExceedsTwentyCharacters"
		_, err := authService.UpdateDisplayName(ctx, "test_player_id", longName)
		if err == nil {
			t.Error("Expected error for long display name")
		}
		if !strings.Contains(err.Error(), "display name must be 20 characters or less") {
			t.Errorf("Expected error containing 'display name must be 20 characters or less', got %s", err.Error())
		}
	})

	t.Run("Player not found", func(t *testing.T) {
		_, err := authService.UpdateDisplayName(ctx, "nonexistent_id", "New Name")
		if err == nil {
			t.Error("Expected error for nonexistent player")
		}
		if !strings.Contains(err.Error(), "player not found") {
			t.Errorf("Expected error containing 'player not found', got %s", err.Error())
		}
	})
}

// TestRefreshToken tests the RefreshToken functionality
func TestRefreshToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create a test player
	player := &models.PlayerData{
		ID:   "test_player_id",
		Name: "Test Player",
	}
	mockRepo.CreatePlayer(ctx, player)

	// Create an expired token
	expiredToken := &models.AuthToken{
		PlayerID:  player.ID,
		Token:     "expired_token_123",
		ExpiresAt: frozenTime.Add(-25 * time.Hour), // Expired (more than 24 hours)
	}
	mockRepo.CreateAuthToken(ctx, expiredToken)

	// Test successful token refresh
	t.Run("Success", func(t *testing.T) {
		resp, err := authService.RefreshToken(ctx, "expired_token_123")
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, player.ID, resp.Player.ID)
		assert.Equal(t, player.Name, resp.Player.Name)
		assert.NotEmpty(t, resp.Token)
		assert.NotEqual(t, "expired_token_123", resp.Token) // Should be a new token
	})

	// Test refresh with invalid token
	t.Run("InvalidToken", func(t *testing.T) {
		resp, err := authService.RefreshToken(ctx, "invalid_token")
		assert.Error(t, err)
		assert.Nil(t, resp)

		var playerNotFoundErr *models.PlayerNotFoundError
		assert.True(t, errors.As(err, &playerNotFoundErr))
		assert.Contains(t, playerNotFoundErr.Reason, "invalid token for refresh")
	})

	// Test refresh with nil token
	t.Run("NilToken", func(t *testing.T) {
		resp, err := authService.RefreshToken(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, resp)

		var playerNotFoundErr *models.PlayerNotFoundError
		assert.True(t, errors.As(err, &playerNotFoundErr))
		assert.Contains(t, playerNotFoundErr.Reason, "invalid token for refresh")
	})
}

// TestRefreshTokenWithExpiredToken tests refresh when token is actually expired
func TestRefreshTokenWithExpiredToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create a test player
	player := &models.PlayerData{
		ID:   "test_player_id",
		Name: "Test Player",
	}
	mockRepo.CreatePlayer(ctx, player)

	// Create an expired token
	expiredToken := &models.AuthToken{
		PlayerID:  player.ID,
		Token:     "expired_token_123",
		ExpiresAt: frozenTime.Add(-25 * time.Hour), // Expired (more than 24 hours)
	}
	mockRepo.CreateAuthToken(ctx, expiredToken)

	// Test refresh with expired token - should work now
	resp, err := authService.RefreshToken(ctx, "expired_token_123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, player.ID, resp.Player.ID)
	assert.Equal(t, player.Name, resp.Player.Name)
	assert.NotEmpty(t, resp.Token)
	assert.NotEqual(t, "expired_token_123", resp.Token) // Should be a new token
}

// TestRefreshTokenWithValidToken tests refresh when token is still valid
func TestRefreshTokenWithValidToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create a test player
	player := &models.PlayerData{
		ID:   "test_player_id",
		Name: "Test Player",
	}
	mockRepo.CreatePlayer(ctx, player)

	// Create a valid token
	validToken := &models.AuthToken{
		PlayerID:  player.ID,
		Token:     "valid_token_123",
		ExpiresAt: frozenTime.Add(1 * time.Hour), // Valid
	}
	mockRepo.CreateAuthToken(ctx, validToken)

	// Test refresh with valid token
	resp, err := authService.RefreshToken(ctx, "valid_token_123")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, player.ID, resp.Player.ID)
	assert.Equal(t, player.Name, resp.Player.Name)
	assert.NotEmpty(t, resp.Token)
	assert.NotEqual(t, "valid_token_123", resp.Token) // Should be a new token
}

// TestRefreshTokenWithMissingPlayer tests refresh when player doesn't exist
func TestRefreshTokenWithMissingPlayer(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Create a token for a non-existent player
	orphanedToken := &models.AuthToken{
		PlayerID:  "non_existent_player",
		Token:     "orphaned_token_123",
		ExpiresAt: frozenTime.Add(1 * time.Hour), // Valid
	}
	mockRepo.CreateAuthToken(ctx, orphanedToken)

	// Test refresh with orphaned token
	resp, err := authService.RefreshToken(ctx, "orphaned_token_123")
	assert.Error(t, err)
	assert.Nil(t, resp)

	var playerNotFoundErr *models.PlayerNotFoundError
	assert.True(t, errors.As(err, &playerNotFoundErr))
	assert.Contains(t, playerNotFoundErr.Reason, "player not found for refresh")
}

// TestRefreshTokenRepositoryError tests refresh when repository operations fail
func TestRefreshTokenRepositoryError(t *testing.T) {
	// Create a mock repository that returns errors
	mockRepo := &MockPlayerRepositoryWithErrors{}
	authService := NewAuthServiceWithTimeFunc(mockRepo, NewMockOAuthVerifier(), func() time.Time { return frozenTime })

	ctx := context.Background()

	// Test refresh with repository error
	resp, err := authService.RefreshToken(ctx, "any_token")
	assert.Error(t, err)
	assert.Nil(t, resp)

	var generalAuthErr *models.GeneralAuthError
	assert.True(t, errors.As(err, &generalAuthErr))
	assert.Contains(t, generalAuthErr.Reason, "failed to get auth token for refresh")
}

// MockPlayerRepositoryWithErrors is a mock that returns errors for testing error scenarios
type MockPlayerRepositoryWithErrors struct{}

func (m *MockPlayerRepositoryWithErrors) GetPlayer(playerId string) (models.Player, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) GetPlayers(gameId string) ([]models.Player, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) CreatePlayer(ctx context.Context, player *models.PlayerData) error {
	return fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) GetPlayerByID(ctx context.Context, id string) (*models.PlayerData, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) GetPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) GetPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) GetPlayerByName(ctx context.Context, name string) (*models.PlayerData, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) UpdatePlayer(ctx context.Context, player *models.PlayerData) error {
	return fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	return fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) GetAuthToken(ctx context.Context, token string) (*models.AuthToken, error) {
	return nil, fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) UpdateAuthToken(ctx context.Context, token *models.AuthToken) error {
	return fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) DeleteAuthToken(ctx context.Context, token string) error {
	return fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) DeleteExpiredTokens(ctx context.Context) error {
	return fmt.Errorf("mock error")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
