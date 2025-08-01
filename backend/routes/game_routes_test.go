package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/services"
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

func (m *MockGameCreationService) CreateGame(req *services.CreateGameRequest, player models.Player) (*services.CreateGameResponse, error) {
	args := m.Called(req, player)
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

func (m *MockGameCreationService) GetPlayerGames(player models.Player) ([]services.PlayerGame, error) {
	args := m.Called(player)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]services.PlayerGame), args.Error(1)
}

func (m *MockGameCreationService) GetGameService(gameID string, player models.Player) (services.GameService, error) {
	args := m.Called(gameID, player)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(services.GameService), args.Error(1)
}

func (m *MockGameCreationService) LeaveGame(gameID string, player models.Player) error {
	args := m.Called(gameID, player)
	return args.Error(0)
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

// Remove defaultGuestAuth and put the logic directly in the method
func (m *MockGameAuthService) AuthenticateGuest(ctx context.Context, displayName string) (*models.AuthResponse, error) {
	if displayName == "" {
		return nil, errors.New("display name is required")
	}
	return &models.AuthResponse{
		Player: models.PlayerData{
			ID:   "guest-id",
			Name: displayName,
		},
		Token: "guest-token",
	}, nil
}

func (m *MockGameAuthService) UpdateDisplayName(ctx context.Context, playerID string, displayName string) (*models.PlayerData, error) {
	args := m.Called(ctx, playerID, displayName)
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func setupGameTestRouter() (*gin.Engine, *MockGameCreationService, *MockGameAuthService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockGameCreationService := new(MockGameCreationService)
	mockAuthService := new(MockGameAuthService)

	handler := NewGameHandler(mockGameCreationService, mockAuthService)

	// Add middleware to routes manually for testing
	router.POST("/api/games", middleware.PlayerAuth(mockAuthService), handler.createGame)
	router.GET("/api/games", middleware.PlayerAuth(mockAuthService), handler.getPlayerGames)

	return router, mockGameCreationService, mockAuthService
}

func setupGameTestRouterWithLeave() (*gin.Engine, *MockGameCreationService, *MockGameAuthService) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockGameCreationService := new(MockGameCreationService)
	mockAuthService := new(MockGameAuthService)

	handler := NewGameHandler(mockGameCreationService, mockAuthService)

	router.POST("/api/games", middleware.PlayerAuth(mockAuthService), handler.createGame)
	router.GET("/api/games", middleware.PlayerAuth(mockAuthService), handler.getPlayerGames)
	router.DELETE("/api/games/:gameId/leave", middleware.PlayerAuth(mockAuthService), handler.leaveGame)

	return router, mockGameCreationService, mockAuthService
}

func TestCreateGame_Success(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data
	requestBody := services.CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
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
	mockGameCreationService.On("CreateGame", &requestBody, mock.AnythingOfType("*models.PlayerData")).Return(response, nil)

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
		Name:            "Test Game",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("CreateGame", &requestBody, mock.AnythingOfType("*models.PlayerData")).Return(nil, repositories.ErrGameCodeNotFound)

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
		Name:            "Test Game",
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
		Name:            "Test Game",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("CreateGame", &requestBody, mock.AnythingOfType("*models.PlayerData")).Return(nil, services.ErrInvalidCompetition)

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
		Name:            "Test Game",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("CreateGame", &requestBody, mock.AnythingOfType("*models.PlayerData")).Return(nil, services.ErrInvalidSeasonYear)

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

func TestGetPlayerGames_Success(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data
	player := &models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}

	expectedGames := []services.PlayerGame{
		{
			GameID:          "game-1",
			SeasonYear:      "2025/2026",
			CompetitionName: "Ligue 1",
			Status:          "active",
		},
		{
			GameID:          "game-2",
			SeasonYear:      "2025/2026",
			CompetitionName: "Ligue 1",
			Status:          "in progress",
		},
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("GetPlayerGames", player).Return(expectedGames, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/games", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "games")

	games := response["games"].([]interface{})
	assert.Len(t, games, 2)

	// Verify first game
	game1 := games[0].(map[string]interface{})
	assert.Equal(t, "game-1", game1["gameId"])
	assert.Equal(t, "2025/2026", game1["seasonYear"])
	assert.Equal(t, "Ligue 1", game1["competitionName"])
	assert.Equal(t, "active", game1["status"])

	// Verify second game
	game2 := games[1].(map[string]interface{})
	assert.Equal(t, "game-2", game2["gameId"])
	assert.Equal(t, "2025/2026", game2["seasonYear"])
	assert.Equal(t, "Ligue 1", game2["competitionName"])
	assert.Equal(t, "in progress", game2["status"])

	// Verify mocks
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestGetPlayerGames_EmptyList(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data
	player := &models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("GetPlayerGames", player).Return([]services.PlayerGame{}, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/games", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "games")

	games := response["games"].([]interface{})
	assert.Len(t, games, 0)

	// Verify mocks
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestGetPlayerGames_ServiceError(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data
	player := &models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}

	// Mock expectations
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(&models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}, nil)
	mockGameCreationService.On("GetPlayerGames", player).Return(nil, repositories.ErrGameCodeNotFound)

	// Create request
	req := httptest.NewRequest("GET", "/api/games", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to get player games", response["error"])

	// Verify mocks
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestGetPlayerGames_Unauthorized(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Mock expectations - authentication fails
	mockAuthService.On("ValidateToken", mock.Anything, "invalidtoken").Return(nil, repositories.ErrGameCodeNotFound)

	// Create request
	req := httptest.NewRequest("GET", "/api/games", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
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
	mockGameCreationService.AssertNotCalled(t, "GetPlayerGames")
	mockAuthService.AssertExpectations(t)
}

func TestGetPlayerGames_NoAuthHeader(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Create request without authorization header
	req := httptest.NewRequest("GET", "/api/games", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Authorization header is required", response["error"])

	// Verify no service calls were made
	mockGameCreationService.AssertNotCalled(t, "GetPlayerGames")
	mockAuthService.AssertNotCalled(t, "ValidateToken")
}

func TestLeaveGame_Success(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouterWithLeave()
	gameID := "test-game-id"
	player := &models.PlayerData{ID: "test-player-id", Name: "Test Player"}

	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(player, nil)
	mockGameCreationService.On("LeaveGame", gameID, player).Return(nil)

	req := httptest.NewRequest("DELETE", "/api/games/"+gameID+"/leave", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Successfully left the game", response["message"])
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestLeaveGame_NotInGame(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouterWithLeave()
	gameID := "test-game-id"
	player := &models.PlayerData{ID: "test-player-id", Name: "Test Player"}

	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(player, nil)
	mockGameCreationService.On("LeaveGame", gameID, player).Return(errors.New("player is not in the game"))

	req := httptest.NewRequest("DELETE", "/api/games/"+gameID+"/leave", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "player is not in the game", response["error"])
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestLeaveGame_Error(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouterWithLeave()
	gameID := "test-game-id"
	player := &models.PlayerData{ID: "test-player-id", Name: "Test Player"}

	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(player, nil)
	mockGameCreationService.On("LeaveGame", gameID, player).Return(errors.New("some db error"))

	req := httptest.NewRequest("DELETE", "/api/games/"+gameID+"/leave", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to leave game", response["error"])
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

// TestGetPlayerGamesAPIStatusInconsistency tests the actual API endpoint for game status consistency
func TestGetPlayerGamesAPIStatusInconsistency(t *testing.T) {
	router, mockGameCreationService, mockAuthService := setupGameTestRouter()

	// Setup test data
	player := &models.PlayerData{
		ID:   "test-player-id",
		Name: "Test Player",
	}

	// Initial games with "in progress" status
	initialGames := []services.PlayerGame{
		{
			GameID:          "game-1",
			SeasonYear:      "2025/2026",
			CompetitionName: "Ligue 1",
			Name:            "Test Game",
			Status:          "in progress",
			Players:         []services.PlayerGameInfo{},
			Code:            "ABC1",
		},
	}

	// Updated games with "finished" status
	updatedGames := []services.PlayerGame{
		{
			GameID:          "game-1",
			SeasonYear:      "2025/2026",
			CompetitionName: "Ligue 1",
			Name:            "Test Game",
			Status:          "finished",
			Players:         []services.PlayerGameInfo{},
			Code:            "ABC1",
		},
	}

	// Mock expectations for initial call
	mockAuthService.On("ValidateToken", mock.Anything, "testtoken").Return(player, nil).Times(2)
	mockGameCreationService.On("GetPlayerGames", player).Return(initialGames, nil).Once()
	mockGameCreationService.On("GetPlayerGames", player).Return(updatedGames, nil).Once()

	// Create request
	req := httptest.NewRequest("GET", "/api/games", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert initial response
	assert.Equal(t, http.StatusOK, w.Code)

	var initialResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &initialResponse)
	assert.NoError(t, err)
	assert.Contains(t, initialResponse, "games")

	initialGamesResponse := initialResponse["games"].([]interface{})
	assert.Len(t, initialGamesResponse, 1)

	// Verify initial game status
	initialGame := initialGamesResponse[0].(map[string]interface{})
	assert.Equal(t, "in progress", initialGame["status"], "Initial game status should be 'in progress'")

	// Create second request
	req2 := httptest.NewRequest("GET", "/api/games", nil)
	req2.Header.Set("Authorization", "Bearer testtoken")
	w2 := httptest.NewRecorder()

	// Perform second request
	router.ServeHTTP(w2, req2)

	// Assert updated response
	assert.Equal(t, http.StatusOK, w2.Code)

	var updatedResponse map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &updatedResponse)
	assert.NoError(t, err)
	assert.Contains(t, updatedResponse, "games")

	updatedGamesResponse := updatedResponse["games"].([]interface{})
	assert.Len(t, updatedGamesResponse, 1)

	// Verify updated game status
	updatedGame := updatedGamesResponse[0].(map[string]interface{})
	assert.Equal(t, "finished", updatedGame["status"], "Updated game status should be 'finished'")

	// Verify mocks
	mockGameCreationService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}
