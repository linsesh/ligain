package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/services"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
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
		player.ID = "mock_id"
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

func (m *MockPlayerRepository) UpdateAuthToken(ctx context.Context, token *models.AuthToken) error {
	// In-memory implementation - update in a simple map
	// Since we're using a map with token as key, we need to delete and recreate
	delete(m.tokens, token.Token)
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
	return nil
}

func (m *MockPlayerRepository) DeletePlayer(ctx context.Context, playerID string) error {
	delete(m.players, playerID)
	// Also delete any tokens for this player
	for token, authToken := range m.tokens {
		if authToken.PlayerID == playerID {
			delete(m.tokens, token)
		}
	}
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

// MockAuthService implements AuthServiceInterface for testing
type MockAuthService struct {
	shouldFail bool
	player     *models.PlayerData
	tokens     map[string]*models.AuthToken
}

func NewMockAuthService() *MockAuthService {
	return &MockAuthService{
		tokens: make(map[string]*models.AuthToken),
	}
}

func (m *MockAuthService) Authenticate(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuthService) AuthenticateGuest(ctx context.Context, displayName string) (*models.AuthResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*models.PlayerData, error) {
	if m.shouldFail {
		return nil, &models.TokenExpiredError{Reason: "token expired"}
	}

	// Check if token exists and is expired
	if authToken, exists := m.tokens[token]; exists {
		if authToken.ExpiresAt.Before(frozenTime) {
			return nil, &models.TokenExpiredError{Reason: "token expired"}
		}
		return &models.PlayerData{ID: authToken.PlayerID, Name: "Test Player"}, nil
	}

	return nil, &models.PlayerNotFoundError{Reason: "invalid token"}
}

func (m *MockAuthService) RefreshToken(ctx context.Context, expiredToken string) (*models.AuthResponse, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock refresh failed")
	}

	// Get the token to extract player ID
	if authToken, exists := m.tokens[expiredToken]; exists {
		// Generate new token
		newToken := "new_refreshed_token_" + expiredToken

		// Delete old token
		delete(m.tokens, expiredToken)

		// Create new token entry
		m.tokens[newToken] = &models.AuthToken{
			PlayerID:  authToken.PlayerID,
			Token:     newToken,
			ExpiresAt: frozenTime.Add(24 * time.Hour), // Valid for 24 hours
		}

		return &models.AuthResponse{
			Player: models.PlayerData{ID: authToken.PlayerID, Name: "Test Player"},
			Token:  newToken,
		}, nil
	}

	return nil, fmt.Errorf("token not found for refresh")
}

func (m *MockAuthService) GetAuthTokenDirectly(ctx context.Context, token string) (*models.AuthToken, error) {
	if authToken, exists := m.tokens[token]; exists {
		return authToken, nil
	}
	return nil, nil
}

func (m *MockAuthService) RefreshTokenByPlayerID(ctx context.Context, playerID string) (*models.AuthResponse, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock refresh failed")
	}

	return &models.AuthResponse{
		Player: models.PlayerData{ID: playerID, Name: "Test Player"},
		Token:  "new_refreshed_token",
	}, nil
}

func (m *MockAuthService) Logout(ctx context.Context, token string) error {
	return nil
}

func (m *MockAuthService) CleanupExpiredTokens(ctx context.Context) error {
	return nil
}

func (m *MockAuthService) GetOrCreatePlayer(ctx context.Context, verifiedUser map[string]interface{}, provider string, displayName string) (*models.PlayerData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuthService) UpdateDisplayName(ctx context.Context, playerID string, newDisplayName string) (*models.PlayerData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuthService) DeleteAccount(ctx context.Context, playerID string) error {
	return nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestAPIKeyAuth_ValidKey(t *testing.T) {
	// Set up test API key
	testAPIKey := "test_api_key_123"
	os.Setenv("API_KEY", testAPIKey)
	defer os.Unsetenv("API_KEY")

	router := setupTestRouter()
	router.Use(APIKeyAuth())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", testAPIKey)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["message"])
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	router := setupTestRouter()
	router.Use(APIKeyAuth())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No X-API-Key header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "API key is required", response["error"])
}

func TestAPIKeyAuth_InvalidKey(t *testing.T) {
	// Set up test API key
	testAPIKey := "test_api_key_123"
	os.Setenv("API_KEY", testAPIKey)
	defer os.Unsetenv("API_KEY")

	router := setupTestRouter()
	router.Use(APIKeyAuth())

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "wrong_key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid API key", response["error"])
}

func TestPlayerAuth_ValidToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := services.NewAuthServiceWithTimeFunc(mockRepo, nil, func() time.Time { return frozenTime })

	// Create a test player and token
	ctx := context.Background()
	player := &models.PlayerData{
		ID:   "test_player_id",
		Name: "Test Player",
	}
	mockRepo.CreatePlayer(ctx, player)

	authToken := &models.AuthToken{
		PlayerID:  player.ID,
		Token:     "valid_token_123",
		ExpiresAt: frozenTime.Add(24 * time.Hour), // Valid for 24 hours
	}
	mockRepo.CreateAuthToken(ctx, authToken)

	router := setupTestRouter()
	router.Use(PlayerAuth(authService))

	router.GET("/test", func(c *gin.Context) {
		playerFromContext, exists := c.Get("player")
		if exists && playerFromContext != nil {
			c.JSON(http.StatusOK, gin.H{"player": playerFromContext})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "player not found in context"})
		}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token_123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["player"])
}

func TestPlayerAuth_MissingHeader(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := services.NewAuthServiceWithTimeFunc(mockRepo, nil, func() time.Time { return frozenTime })

	router := setupTestRouter()
	router.Use(PlayerAuth(authService))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No Authorization header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Authorization header is required", response["error"])
}

func TestPlayerAuth_InvalidFormat(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := services.NewAuthServiceWithTimeFunc(mockRepo, nil, func() time.Time { return frozenTime })

	router := setupTestRouter()
	router.Use(PlayerAuth(authService))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token123") // Wrong format
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid authorization header format", response["error"])
}

func TestPlayerAuth_InvalidToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := services.NewAuthServiceWithTimeFunc(mockRepo, nil, func() time.Time { return frozenTime })

	router := setupTestRouter()
	router.Use(PlayerAuth(authService))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid or expired token", response["error"])
}

func TestPlayerAuth_ExpiredToken(t *testing.T) {
	mockRepo := NewMockPlayerRepository()
	authService := services.NewAuthServiceWithTimeFunc(mockRepo, nil, func() time.Time { return frozenTime })

	// Create a test player and expired token
	ctx := context.Background()
	player := &models.PlayerData{
		ID:   "test_player_id",
		Name: "Test Player",
	}
	mockRepo.CreatePlayer(ctx, player)

	authToken := &models.AuthToken{
		PlayerID:  player.ID,
		Token:     "expired_token_123",
		ExpiresAt: frozenTime.Add(-1 * time.Hour), // Expired
	}
	mockRepo.CreateAuthToken(ctx, authToken)

	router := setupTestRouter()
	router.Use(PlayerAuth(authService))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer expired_token_123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid or expired token", response["error"])
}

// TestPlayerAuthWithRefresh tests the PlayerAuthWithRefresh middleware
func TestPlayerAuthWithRefresh(t *testing.T) {
	mockAuthService := NewMockAuthService()

	// Create test tokens in the mock auth service
	expiredToken := &models.AuthToken{
		PlayerID:  "test_player_id",
		Token:     "expired_token_123",
		ExpiresAt: frozenTime.Add(-25 * time.Hour), // Expired (more than 24 hours)
	}
	mockAuthService.tokens["expired_token_123"] = expiredToken

	validToken := &models.AuthToken{
		PlayerID:  "test_player_id",
		Token:     "valid_token_123",
		ExpiresAt: frozenTime.Add(1 * time.Hour), // Valid
	}
	mockAuthService.tokens["valid_token_123"] = validToken

	router := setupTestRouter()
	router.Use(PlayerAuthWithRefresh(mockAuthService))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("ValidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid_token_123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})

	t.Run("ExpiredTokenWithRefresh", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer expired_token_123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check that the response headers indicate token refresh
		assert.Equal(t, "true", w.Header().Get("X-Token-Refreshed"))
		assert.NotEmpty(t, w.Header().Get("X-New-Token"))

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["message"])
	})

	t.Run("InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid_token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid token", response["error"])
	})

	t.Run("MissingAuthorizationHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Authorization header is required", response["error"])
	})

	t.Run("InvalidAuthorizationFormat", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat valid_token_123")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid authorization header format", response["error"])
	})
}

// TestPlayerAuthWithRefreshFailedRefresh tests when token refresh fails
func TestPlayerAuthWithRefreshFailedRefresh(t *testing.T) {
	// Create a mock auth service that fails refresh operations
	mockAuthService := NewMockAuthService()
	mockAuthService.shouldFail = true

	router := setupTestRouter()
	router.Use(PlayerAuthWithRefresh(mockAuthService))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer any_token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Token expired and refresh failed", response["error"])
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
