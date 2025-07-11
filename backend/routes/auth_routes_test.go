package routes

import (
	"context"
	"encoding/json"
	"errors"
	"liguain/backend/middleware"
	"liguain/backend/models"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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
