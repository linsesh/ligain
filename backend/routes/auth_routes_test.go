package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// frozenTime is a fixed time for consistent testing
var frozenTime = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

// MockPlayerRepository is a mock implementation for testing
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
	return nil, nil
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

func (m *MockPlayerRepository) UpdateAuthToken(ctx context.Context, token *models.AuthToken) error {
	// In-memory implementation - update in a simple map
	// Since we're using a map with token as key, we need to delete and recreate
	delete(m.tokens, token.Token)
	m.tokens[token.Token] = token
	return nil
}

func (m *MockPlayerRepository) DeleteAuthToken(ctx context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func (m *MockPlayerRepository) DeleteExpiredTokens(ctx context.Context) error {
	// In-memory implementation
	// This would typically iterate through tokens and delete expired ones
	// For simplicity in testing, we'll just return nil
	return nil
}

func (m *MockPlayerRepository) DeletePlayer(ctx context.Context, playerID string) error {
	delete(m.players, playerID)
	return nil
}

func (m *MockPlayerRepository) UpdateAvatar(ctx context.Context, playerID string, objectKey string, signedURL string, expiresAt time.Time) error {
	return nil
}

func (m *MockPlayerRepository) UpdateAvatarSignedURL(ctx context.Context, playerID string, signedURL string, expiresAt time.Time) error {
	return nil
}

func (m *MockPlayerRepository) ClearAvatar(ctx context.Context, playerID string) error {
	return nil
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

func (m *MockPlayerRepositoryWithErrors) DeletePlayer(ctx context.Context, playerID string) error {
	return fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) UpdateAvatar(ctx context.Context, playerID string, objectKey string, signedURL string, expiresAt time.Time) error {
	return fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) UpdateAvatarSignedURL(ctx context.Context, playerID string, signedURL string, expiresAt time.Time) error {
	return fmt.Errorf("mock error")
}

func (m *MockPlayerRepositoryWithErrors) ClearAvatar(ctx context.Context, playerID string) error {
	return fmt.Errorf("mock error")
}

// MockAuthService implements AuthServiceInterface for testing
type MockAuthService struct {
	shouldFail bool
	player     *models.PlayerData
}

func (m *MockAuthService) Authenticate(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error) {
	if m.shouldFail {
		return nil, errors.New("mock authentication failed")
	}

	if m.player == nil {
		m.player = &models.PlayerData{
			ID:       "test-player-id",
			Name:     "Test Player",
			Email:    &req.Email,
			Provider: &req.Provider,
		}
	}

	return &models.AuthResponse{
		Player: *m.player,
		Token:  "mock-token",
	}, nil
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*models.PlayerData, error) {
	if m.shouldFail {
		return nil, errors.New("mock validation failed")
	}
	if m.player == nil {
		return &models.PlayerData{
			ID:   "test-player-id",
			Name: "Test Player",
		}, nil
	}
	return m.player, nil
}

func (m *MockAuthService) Logout(ctx context.Context, token string) error {
	if m.shouldFail {
		return errors.New("mock logout failed")
	}
	return nil
}

func (m *MockAuthService) CleanupExpiredTokens(ctx context.Context) error {
	if m.shouldFail {
		return errors.New("mock cleanup failed")
	}
	return nil
}

func (m *MockAuthService) GetOrCreatePlayer(ctx context.Context, verifiedUser map[string]interface{}, provider string, displayName string) (*models.PlayerData, error) {
	if m.shouldFail {
		return nil, errors.New("mock get or create player failed")
	}

	if m.player == nil {
		m.player = &models.PlayerData{
			ID:       "test-player-id",
			Name:     displayName,
			Provider: &provider,
		}
	}

	return m.player, nil
}

func (m *MockAuthService) AuthenticateGuest(ctx context.Context, displayName string) (*models.AuthResponse, error) {
	if m.shouldFail {
		return nil, errors.New("mock guest authentication failed")
	}
	if displayName == "" {
		return nil, errors.New("display name is required")
	}
	m.player = &models.PlayerData{
		ID:   "guest-id",
		Name: displayName,
	}

	return &models.AuthResponse{
		Player: *m.player,
		Token:  "mock-guest-token",
	}, nil
}

func (m *MockAuthService) RefreshToken(ctx context.Context, expiredToken string) (*models.AuthResponse, error) {
	if m.shouldFail {
		return nil, errors.New("mock refresh failed")
	}
	if expiredToken == "" {
		return nil, errors.New("token is required")
	}
	if m.player == nil {
		return nil, errors.New("player not found")
	}

	return &models.AuthResponse{
		Player: *m.player,
		Token:  "mock-refreshed-token",
	}, nil
}

func (m *MockAuthService) GetAuthTokenDirectly(ctx context.Context, token string) (*models.AuthToken, error) {
	if m.shouldFail {
		return nil, errors.New("mock get token failed")
	}
	if token == "" {
		return nil, nil
	}
	return &models.AuthToken{
		PlayerID:  "test-player-id",
		Token:     token,
		ExpiresAt: frozenTime.Add(1 * time.Hour),
	}, nil
}

func (m *MockAuthService) RefreshTokenByPlayerID(ctx context.Context, playerID string) (*models.AuthResponse, error) {
	if m.shouldFail {
		return nil, errors.New("mock refresh by player ID failed")
	}
	if playerID == "" {
		return nil, errors.New("player ID is required")
	}
	if m.player == nil {
		m.player = &models.PlayerData{
			ID:   playerID,
			Name: "Test Player",
		}
	}

	return &models.AuthResponse{
		Player: *m.player,
		Token:  "mock-refreshed-token",
	}, nil
}

func (m *MockAuthService) UpdateDisplayName(ctx context.Context, playerID string, displayName string) (*models.PlayerData, error) {
	if m.shouldFail {
		return nil, errors.New("mock update display name failed")
	}
	if displayName == "" {
		return nil, errors.New("display name cannot be empty")
	}
	if m.player == nil || m.player.ID != playerID {
		return nil, errors.New("player not found")
	}
	m.player.Name = displayName
	return m.player, nil
}

func (m *MockAuthService) DeleteAccount(ctx context.Context, playerID string) error {
	if m.shouldFail {
		return errors.New("mock delete account failed")
	}
	if m.player == nil || m.player.ID != playerID {
		return &models.PlayerNotFoundError{Reason: "player not found"}
	}
	return nil
}

func TestSignInHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  bool
		authService    *MockAuthService
		includeAPIKey  bool
		apiKey         string
	}{
		{
			name: "successful sign in",
			requestBody: `{
				"provider": "google",
				"token": "mock_token",
				"name": "Test User",
				"email": "test@example.com"
			}`,
			expectedStatus: http.StatusOK,
			expectedError:  false,
			authService:    &MockAuthService{},
			includeAPIKey:  false, // Test without API key middleware
		},
		{
			name:           "invalid JSON",
			requestBody:    `invalid json`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			authService:    &MockAuthService{},
			includeAPIKey:  false,
		},
		{
			name: "authentication failure",
			requestBody: `{
				"provider": "google",
				"token": "mock_token",
				"name": "Test User",
				"email": "test@example.com"
			}`,
			expectedStatus: http.StatusUnauthorized, // was InternalServerError, now Unauthorized
			expectedError:  true,
			authService:    &MockAuthService{shouldFail: true},
			includeAPIKey:  false,
		},
		{
			name: "missing API key",
			requestBody: `{
				"provider": "google",
				"token": "mock_token",
				"name": "Test User",
				"email": "test@example.com"
			}`,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			authService:    &MockAuthService{},
			includeAPIKey:  true,
			apiKey:         "", // No API key
		},
		{
			name: "invalid API key",
			requestBody: `{
				"provider": "google",
				"token": "mock_token",
				"name": "Test User",
				"email": "test@example.com"
			}`,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			authService:    &MockAuthService{},
			includeAPIKey:  true,
			apiKey:         "wrong_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()
			authHandler := NewAuthHandler(tt.authService)

			// Add API key middleware if requested
			if tt.includeAPIKey {
				// Set environment variable for API key
				originalAPIKey := os.Getenv("API_KEY")
				defer os.Setenv("API_KEY", originalAPIKey)

				if tt.apiKey != "" {
					os.Setenv("API_KEY", "test_api_key")
				} else {
					os.Unsetenv("API_KEY")
				}

				router.Use(middleware.APIKeyAuth())
			}

			router.POST("/signin", authHandler.SignIn)

			// Create request
			req, err := http.NewRequest("POST", "/signin", strings.NewReader(tt.requestBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Add API key header if testing with API key
			if tt.includeAPIKey && tt.apiKey != "" {
				req.Header.Set("X-API-Key", tt.apiKey)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				// Parse response
				var response models.AuthResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Assert response fields
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, "test-player-id", response.Player.ID)
				assert.Equal(t, "Test Player", response.Player.Name)
				assert.Equal(t, "test@example.com", *response.Player.Email)
				assert.Equal(t, "google", *response.Player.Provider)
			}
		})
	}
}

func TestSignOutHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  bool
		authService    *MockAuthService
	}{
		{
			name:           "successful sign out",
			authHeader:     "Bearer mock-token",
			expectedStatus: http.StatusOK,
			expectedError:  false,
			authService:    &MockAuthService{},
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusBadRequest, // was Unauthorized, now BadRequest
			expectedError:  true,
			authService:    &MockAuthService{},
		},
		{
			name:           "logout failure",
			authHeader:     "Bearer mock-token",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
			authService:    &MockAuthService{shouldFail: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()
			authHandler := NewAuthHandler(tt.authService)
			router.POST("/signout", authHandler.SignOut)

			// Create request
			req, err := http.NewRequest("POST", "/signout", nil)
			assert.NoError(t, err)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// stringPtr is a helper function to create a pointer to a string
func stringPtr(s string) *string {
	return &s
}

func TestCurrentUserHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  bool
		authService    *MockAuthService
	}{
		{
			name:           "successful get current user",
			authHeader:     "Bearer mock-token",
			expectedStatus: http.StatusOK,
			expectedError:  false,
			authService:    &MockAuthService{false, &models.PlayerData{ID: "test-player-id", Name: "Test Player"}},
		},
		{
			name:           "successful get current user with avatar",
			authHeader:     "Bearer mock-token",
			expectedStatus: http.StatusOK,
			expectedError:  false,
			authService: &MockAuthService{false, &models.PlayerData{
				ID:              "test-player-id",
				Name:            "Test Player",
				AvatarSignedURL: stringPtr("https://storage.example.com/avatar.jpg"),
			}},
		},
		{
			name:           "missing auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized, // was BadRequest, now Unauthorized
			expectedError:  true,
			authService:    &MockAuthService{},
		},
		{
			name:           "validation failure",
			authHeader:     "Bearer mock-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			authService:    &MockAuthService{shouldFail: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Gin router
			router := gin.New()
			authHandler := NewAuthHandler(tt.authService)
			router.GET("/me", middleware.PlayerAuth(tt.authService), authHandler.GetCurrentPlayer)

			// Create request
			req, err := http.NewRequest("GET", "/me", nil)
			assert.NoError(t, err)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				// Parse response as raw JSON to check field names
				var rawResponse map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &rawResponse)
				assert.NoError(t, err)

				player := rawResponse["player"].(map[string]interface{})

				// Assert response fields
				assert.Equal(t, "test-player-id", player["id"])
				assert.Equal(t, "Test Player", player["name"])

				// For the avatar test case, verify avatar_url field is used (not avatar_signed_url)
				if tt.name == "successful get current user with avatar" {
					assert.Equal(t, "https://storage.example.com/avatar.jpg", player["avatar_url"])
					assert.Nil(t, player["avatar_signed_url"], "avatar_signed_url should NOT exist in response")
				}
			}
		})
	}
}

func TestSignInGuestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedError  bool
		authService    *MockAuthService
	}{
		{
			name:           "successful guest sign in",
			requestBody:    `{"name": "GuestUser"}`,
			expectedStatus: http.StatusOK,
			expectedError:  false,
			authService:    &MockAuthService{},
		},
		{
			name:           "missing name",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			authService:    &MockAuthService{},
		},
		{
			name:           "auth service failure",
			requestBody:    `{"name": "GuestUser"}`,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			authService:    &MockAuthService{shouldFail: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			authHandler := NewAuthHandler(tt.authService)
			router.POST("/signin/guest", authHandler.SignInGuest)

			req, err := http.NewRequest("POST", "/signin/guest", strings.NewReader(tt.requestBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				_, hasError := resp["error"]
				assert.True(t, hasError)
			} else {
				var resp struct {
					Player models.PlayerData `json:"player"`
					Token  string            `json:"token"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "GuestUser", resp.Player.Name)
				assert.Equal(t, "mock-guest-token", resp.Token)
			}
		})
	}
}

