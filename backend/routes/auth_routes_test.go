package routes

import (
	"context"
	"encoding/json"
	"errors"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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
		Token:  "guest-token",
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
				// Parse response
				type playerResponse struct {
					Player models.PlayerData `json:"player"`
				}
				var response playerResponse
				err = json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Assert response fields
				assert.Equal(t, "test-player-id", response.Player.ID)
				assert.Equal(t, "Test Player", response.Player.Name)
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
				assert.Equal(t, "guest-token", resp.Token)
			}
		})
	}
}

func TestAuthHandler_UpdateDisplayName(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	// Create mock auth service
	mockAuthService := &MockAuthService{}
	handler := NewAuthHandler(mockAuthService)

	// Create router and setup routes
	router := gin.New()
	handler.SetupRoutes(router)

	// Create a test player
	testPlayer := &models.PlayerData{
		ID:   "test_player_id",
		Name: "Test Player",
	}

	t.Run("Valid display name update", func(t *testing.T) {
		// Setup mock player
		mockAuthService.player = &models.PlayerData{
			ID:   "test_player_id",
			Name: "Updated Name",
		}

		// Create request body
		requestBody := map[string]string{
			"displayName": "Updated Name",
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Create request
		req, _ := http.NewRequest("PUT", "/api/auth/profile/display-name", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Add player to context (simulating middleware)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("player", testPlayer)

		// Call handler
		handler.UpdateDisplayName(c)

		// Check response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		player := response["player"].(map[string]interface{})
		assert.Equal(t, "Updated Name", player["name"])
	})

	t.Run("Invalid request body", func(t *testing.T) {
		// Create invalid request body
		req, _ := http.NewRequest("PUT", "/api/auth/profile/display-name", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Add player to context (simulating middleware)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("player", testPlayer)

		// Call handler
		handler.UpdateDisplayName(c)

		// Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response["error"])
	})

	t.Run("Auth service error", func(t *testing.T) {
		// Setup mock to fail
		mockAuthService.shouldFail = true

		// Create request body
		requestBody := map[string]string{
			"displayName": "Invalid Name",
		}
		jsonBody, _ := json.Marshal(requestBody)

		// Create request
		req, _ := http.NewRequest("PUT", "/api/auth/profile/display-name", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Add player to context (simulating middleware)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("player", testPlayer)

		// Call handler
		handler.UpdateDisplayName(c)

		// Reset shouldFail
		mockAuthService.shouldFail = false

		// Check response
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "mock update display name failed")
	})

	t.Run("No player in context", func(t *testing.T) {
		// Create request
		req, _ := http.NewRequest("PUT", "/api/auth/profile/display-name", bytes.NewBuffer([]byte("{\"displayName\":\"Test\"}")))
		req.Header.Set("Content-Type", "application/json")

		// Create response recorder
		w := httptest.NewRecorder()

		// Create context without player (simulating missing middleware)
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		// Call handler
		handler.UpdateDisplayName(c)

		// Check response
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Player not found in context", response["error"])
	})
}
