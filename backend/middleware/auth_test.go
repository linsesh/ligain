package middleware

import (
	"context"
	"encoding/json"
	"liguain/backend/models"
	"liguain/backend/services"
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

func (m *MockPlayerRepository) SavePlayer(player models.Player) (string, error) {
	if pd, ok := player.(*models.PlayerData); ok {
		if pd.ID == "" {
			pd.ID = "mock_id"
		}
		m.players[pd.ID] = pd
		return pd.ID, nil
	}
	return "", nil
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
