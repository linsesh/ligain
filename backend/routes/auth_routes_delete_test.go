package routes

import (
	"context"
	"ligain/backend/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthServiceForDelete implements AuthServiceInterface for testing delete account
type MockAuthServiceForDelete struct {
	mock.Mock
}

func (m *MockAuthServiceForDelete) Authenticate(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockAuthServiceForDelete) AuthenticateGuest(ctx context.Context, displayName string) (*models.AuthResponse, error) {
	args := m.Called(ctx, displayName)
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockAuthServiceForDelete) ValidateToken(ctx context.Context, token string) (*models.PlayerData, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockAuthServiceForDelete) RefreshToken(ctx context.Context, expiredToken string) (*models.AuthResponse, error) {
	args := m.Called(ctx, expiredToken)
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockAuthServiceForDelete) GetAuthTokenDirectly(ctx context.Context, token string) (*models.AuthToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*models.AuthToken), args.Error(1)
}

func (m *MockAuthServiceForDelete) RefreshTokenByPlayerID(ctx context.Context, playerID string) (*models.AuthResponse, error) {
	args := m.Called(ctx, playerID)
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockAuthServiceForDelete) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthServiceForDelete) CleanupExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAuthServiceForDelete) GetOrCreatePlayer(ctx context.Context, verifiedUser map[string]interface{}, provider string, displayName string) (*models.PlayerData, error) {
	args := m.Called(ctx, verifiedUser, provider, displayName)
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockAuthServiceForDelete) UpdateDisplayName(ctx context.Context, playerID string, newDisplayName string) (*models.PlayerData, error) {
	args := m.Called(ctx, playerID, newDisplayName)
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockAuthServiceForDelete) DeleteAccount(ctx context.Context, playerID string) error {
	args := m.Called(ctx, playerID)
	return args.Error(0)
}

func TestAuthHandler_DeleteAccount_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockAuthService := new(MockAuthServiceForDelete)
	handler := NewAuthHandler(mockAuthService)

	// Create test player
	testPlayer := &models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}

	// Setup router with middleware that sets player in context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("player", testPlayer)
		c.Next()
	})
	router.DELETE("/api/auth/account", handler.DeleteAccount)

	// Setup expectations
	mockAuthService.On("DeleteAccount", mock.Anything, testPlayer.ID).Return(nil)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/auth/account", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Account deleted successfully")
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_DeleteAccount_NoPlayerInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockAuthService := new(MockAuthServiceForDelete)
	handler := NewAuthHandler(mockAuthService)

	// Setup router without setting player in context
	router := gin.New()
	router.DELETE("/api/auth/account", handler.DeleteAccount)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/auth/account", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication required")
}

func TestAuthHandler_DeleteAccount_InvalidPlayerType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockAuthService := new(MockAuthServiceForDelete)
	handler := NewAuthHandler(mockAuthService)

	// Setup router with invalid player type in context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("player", "invalid-player-type")
		c.Next()
	})
	router.DELETE("/api/auth/account", handler.DeleteAccount)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/auth/account", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal server error")
}

func TestAuthHandler_DeleteAccount_PlayerNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockAuthService := new(MockAuthServiceForDelete)
	handler := NewAuthHandler(mockAuthService)

	// Create test player
	testPlayer := &models.PlayerData{
		ID:   "non-existent-player",
		Name: "Test Player",
	}

	// Setup router with middleware that sets player in context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("player", testPlayer)
		c.Next()
	})
	router.DELETE("/api/auth/account", handler.DeleteAccount)

	// Setup expectations - return PlayerNotFoundError
	playerNotFoundErr := &models.PlayerNotFoundError{Reason: "player not found"}
	mockAuthService.On("DeleteAccount", mock.Anything, testPlayer.ID).Return(playerNotFoundErr)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/auth/account", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Player not found")
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_DeleteAccount_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup
	mockAuthService := new(MockAuthServiceForDelete)
	handler := NewAuthHandler(mockAuthService)

	// Create test player
	testPlayer := &models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}

	// Setup router with middleware that sets player in context
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("player", testPlayer)
		c.Next()
	})
	router.DELETE("/api/auth/account", handler.DeleteAccount)

	// Setup expectations - return general error
	generalErr := &models.GeneralAuthError{Reason: "database error"}
	mockAuthService.On("DeleteAccount", mock.Anything, testPlayer.ID).Return(generalErr)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/auth/account", nil)
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to delete account")
	mockAuthService.AssertExpectations(t)
}
