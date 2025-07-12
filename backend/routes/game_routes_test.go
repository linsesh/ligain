package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"liguain/backend/middleware"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/services"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGameCreationService is a mock implementation of GameCreationServiceInterface
type MockGameCreationService struct {
	mock.Mock
}

func (m *MockGameCreationService) CreateGame(req *services.CreateGameRequest) (*services.CreateGameResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.CreateGameResponse), args.Error(1)
}

func (m *MockGameCreationService) CleanupExpiredCodes() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGameCreationService) JoinGame(code string, player models.Player) (*services.JoinGameResponse, error) {
	args := m.Called(code, player)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.JoinGameResponse), args.Error(1)
}

// MockAuthService is a mock implementation of AuthServiceInterface
type MockGameAuthService struct {
	mock.Mock
}

func (m *MockGameAuthService) Authenticate(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockGameAuthService) ValidateToken(ctx context.Context, token string) (*models.PlayerData, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockGameAuthService) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockGameAuthService) GetOrCreatePlayer(ctx context.Context, verifiedUser map[string]interface{}, provider string, displayName string) (*models.PlayerData, error) {
	args := m.Called(ctx, verifiedUser, provider, displayName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockGameAuthService) CleanupExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupGameTestRouter() (*gin.Engine, *MockGameCreationService, *MockGameAuthService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockGameCreationService := new(MockGameCreationService)
	mockAuthService := new(MockGameAuthService)

	handler := NewGameHandler(mockGameCreationService, mockAuthService)

	// Add middleware to routes manually for testing
	router.POST("/api/games", middleware.PlayerAuth(mockAuthService), handler.createGame)

	return router, mockGameCreationService, mockAuthService
}

func TestCreateGame_Success(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data
	requestBody := services.CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	response := &services.CreateGameResponse{
		GameID: "test-game-id",
		Code:   "ABC1",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("CreateGame", &requestBody).Return(response, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)

	var responseBody services.CreateGameResponse
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "test-game-id", responseBody.GameID)
	assert.Equal(t, "ABC1", responseBody.Code)

	// Verify mocks
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestCreateGame_InvalidJSON(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Mock expectations for authentication middleware
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBufferString("invalid json"))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid request format")

	// Verify no service calls were made
	mockGameCreationService.AssertNotCalled(t, "CreateGame")
	mockAuthService.AssertExpectations(t)
}

func TestCreateGame_MissingSeasonYear(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data with missing seasonYear
	requestBody := map[string]string{
		"competitionName": "Premier League",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "SeasonYear")

	// Verify no service calls were made
	mockGameCreationService.AssertNotCalled(t, "CreateGame")
	mockAuthService.AssertExpectations(t)
}

func TestCreateGame_MissingCompetitionName(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data with missing competitionName
	requestBody := map[string]string{
		"seasonYear": "2025/2026",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "CompetitionName")

	// Verify no service calls were made
	mockGameCreationService.AssertNotCalled(t, "CreateGame")
	mockAuthService.AssertExpectations(t)
}

func TestCreateGame_ServiceError(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data
	requestBody := services.CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("CreateGame", &requestBody).Return(nil, repositories.ErrGameCodeNotFound)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to create game", response["error"])

	// Verify mocks
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestCreateGame_Unauthorized(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data
	requestBody := services.CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations - authentication fails
	mockAuthService.On("ValidateToken", mock.Anything, "invalidtoken").Return(nil, repositories.ErrGameCodeNotFound)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer invalidtoken")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid or expired token", response["error"])

	// Verify no service calls were made
	mockGameCreationService.AssertNotCalled(t, "CreateGame")
	mockAuthService.AssertExpectations(t)
}

func TestCreateGame_InvalidCompetitionName(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data with invalid competition name
	requestBody := services.CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Premier League",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("CreateGame", &requestBody).Return(nil, services.ErrInvalidCompetition)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "only 'Ligue 1' is supported as competition name", response["error"])

	// Verify mocks
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestCreateGame_InvalidSeasonYear(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data with invalid season year
	requestBody := services.CreateGameRequest{
		SeasonYear:      "2024/2025",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("CreateGame", &requestBody).Return(nil, services.ErrInvalidSeasonYear)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer testtoken")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "only '2025/2026' is supported as season year", response["error"])

	// Verify mocks
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}
