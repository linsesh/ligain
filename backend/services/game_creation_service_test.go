package services

import (
	"context"
	"errors"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/rules"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC)

// MockGameRepository is a mock implementation of GameRepository
type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) GetGame(gameId string) (models.Game, error) {
	args := m.Called(gameId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.Game), args.Error(1)
}

func (m *MockGameRepository) CreateGame(game models.Game) (string, error) {
	args := m.Called(game)
	return args.String(0), args.Error(1)
}

func (m *MockGameRepository) SaveWithId(gameId string, game models.Game) error {
	args := m.Called(gameId, game)
	return args.Error(0)
}

func (m *MockGameRepository) GetAllGames() (map[string]models.Game, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]models.Game), args.Error(1)
}

// MockGameCodeRepository is a mock implementation of GameCodeRepository
type MockGameCodeRepository struct {
	mock.Mock
}

func (m *MockGameCodeRepository) CreateGameCode(gameCode *models.GameCode) error {
	args := m.Called(gameCode)
	return args.Error(0)
}

func (m *MockGameCodeRepository) GetGameCodeByCode(code string) (*models.GameCode, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameCode), args.Error(1)
}

func (m *MockGameCodeRepository) GetGameCodeByGameID(gameID string) (*models.GameCode, error) {
	args := m.Called(gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.GameCode), args.Error(1)
}

func (m *MockGameCodeRepository) DeleteExpiredCodes() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGameCodeRepository) CodeExists(code string) (bool, error) {
	args := m.Called(code)
	return args.Bool(0), args.Error(1)
}

func (m *MockGameCodeRepository) DeleteGameCode(code string) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockGameCodeRepository) DeleteGameCodeByGameID(gameID string) error {
	args := m.Called(gameID)
	return args.Error(0)
}

// MockGamePlayerRepository is a mock implementation of GamePlayerRepository
type MockGamePlayerRepository struct {
	mock.Mock
}

func (m *MockGamePlayerRepository) AddPlayerToGame(ctx context.Context, gameID string, playerID string) error {
	args := m.Called(ctx, gameID, playerID)
	return args.Error(0)
}

func (m *MockGamePlayerRepository) RemovePlayerFromGame(ctx context.Context, gameID string, playerID string) error {
	args := m.Called(ctx, gameID, playerID)
	return args.Error(0)
}

func (m *MockGamePlayerRepository) GetPlayersInGame(ctx context.Context, gameID string) ([]models.Player, error) {
	args := m.Called(ctx, gameID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Player), args.Error(1)
}

func (m *MockGamePlayerRepository) GetPlayerGames(ctx context.Context, playerID string) ([]string, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGamePlayerRepository) IsPlayerInGame(ctx context.Context, gameID string, playerID string) (bool, error) {
	args := m.Called(ctx, gameID, playerID)
	return args.Bool(0), args.Error(1)
}

// MockMatchRepository is a mock implementation of MatchRepository
type MockMatchRepository struct {
	mock.Mock
}

func (m *MockMatchRepository) SaveMatch(match models.Match) error {
	args := m.Called(match)
	return args.Error(0)
}

func (m *MockMatchRepository) GetMatch(matchId string) (models.Match, error) {
	args := m.Called(matchId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.Match), args.Error(1)
}

func (m *MockMatchRepository) GetMatches() (map[string]models.Match, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]models.Match), args.Error(1)
}

func (m *MockMatchRepository) GetMatchesByCompetitionAndSeason(competitionCode, seasonCode string) ([]models.Match, error) {
	args := m.Called(competitionCode, seasonCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Match), args.Error(1)
}

type MockWatcher struct {
	mock.Mock
}

func (m *MockWatcher) Subscribe(handler GameService) error {
	args := m.Called(handler)
	return args.Error(0)
}
func (m *MockWatcher) Unsubscribe(gameID string) error {
	fmt.Println("[MOCK] Unsubscribe called with gameID:", gameID)
	args := m.Called(gameID)
	return args.Error(0)
}
func (m *MockWatcher) Start(ctx context.Context) error { return nil }
func (m *MockWatcher) Stop() error                     { return nil }

// setupCreationTestService creates a GameCreationService for tests
func setupCreationTestService(t *testing.T, mockGameRepo *MockGameRepository, mockGameCodeRepo *MockGameCodeRepository, mockGamePlayerRepo *MockGamePlayerRepository, mockBetRepo *MockBetRepository, mockMatchRepo *MockMatchRepository, watcher MatchWatcherService) GameCreationServiceInterface {
	// Constructor calls loadAll, so we need to mock GetAllGames
	mockGameRepo.On("GetAllGames").Return(map[string]models.Game{}, nil)

	service, err := NewGameCreationServiceWithTimeFunc(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, watcher, func() time.Time { return testTime })
	require.NoError(t, err)
	return service
}

func TestGameCreationService_CreateGame_Success(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "test-game-id", "player1").Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(nil)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	// Execute
	response, err := service.CreateGame(request, player)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test-game-id", response.GameID)
	assert.Len(t, response.Code, 4)
	assert.Regexp(t, "^[A-Z0-9]{4}$", response.Code)
	mockWatcher.AssertCalled(t, "Subscribe", mock.AnythingOfType("*services.GameServiceImpl"))

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_GameCreationFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("", errors.New("database error"))
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request, player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create game")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_CodeGenerationFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - all codes exist (simulating collision)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "test-game-id", "player1").Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(true, nil).Times(10)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request, player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to generate unique code after 10 attempts")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_CodeExistsCheckFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "test-game-id", "player1").Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, errors.New("database error"))
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request, player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "error checking code existence")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_GameCodeCreationFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "test-game-id", "player1").Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(errors.New("database error"))
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request, player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create game code")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_CleanupExpiredCodes(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	// Mock expectations
	mockGameCodeRepo.On("DeleteExpiredCodes").Return(nil)

	// Execute
	err := service.CleanupExpiredCodes()

	// Assert
	assert.NoError(t, err)

	// Verify mocks
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_CleanupExpiredCodes_Fails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	// Mock expectations
	mockGameCodeRepo.On("DeleteExpiredCodes").Return(errors.New("database error"))

	// Execute
	err := service.CleanupExpiredCodes()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	// Verify mocks
	mockGameCodeRepo.AssertExpectations(t)
}

// New tests for validation
func TestGameCreationService_CreateGame_InvalidCompetition(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Premier League",
		Name:            "Test Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	response, err := service.CreateGame(request, player)
	assert.ErrorIs(t, err, ErrInvalidCompetition)
	assert.Nil(t, response)
}

func TestGameCreationService_CreateGame_InvalidSeasonYear(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2024/2025",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	response, err := service.CreateGame(request, player)
	assert.ErrorIs(t, err, ErrInvalidSeasonYear)
	assert.Nil(t, response)
}

func TestGameCreationService_CreateGame_MatchLoadingFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - match loading fails
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return(nil, errors.New("database error"))

	// Execute
	response, err := service.CreateGame(request, player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to load matches")

	// Verify mocks
	mockMatchRepo.AssertExpectations(t)
}

func TestGameCreationService_JoinGame_Success(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	code := "ABC1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      code,
		ExpiresAt: testTime.Add(24 * time.Hour),
	}

	// Create a real game instead of using a mock
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", code).Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)

	// Mock IsPlayerInGame to return true after joining (for GetGameService call)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)

	// Execute
	response, err := service.JoinGame(code, player)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "game1", response.GameID)
	assert.Equal(t, "2025/2026", response.SeasonYear)
	assert.Equal(t, "Ligue 1", response.CompetitionName)
	assert.Equal(t, "Successfully joined the game", response.Message)

	// Verify mocks
	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

// SimpleMockGame is a simple mock implementation for testing
type SimpleMockGame struct{}

func (m *SimpleMockGame) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}

func (m *SimpleMockGame) GetPastResults() map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}

func (m *SimpleMockGame) GetSeasonYear() string {
	return "2025/2026"
}

func (m *SimpleMockGame) GetCompetitionName() string {
	return "Ligue 1"
}

func (m *SimpleMockGame) GetGameStatus() models.GameStatus {
	return models.GameStatusScheduled
}

func (m *SimpleMockGame) GetName() string {
	return "Test Game"
}

func (m *SimpleMockGame) CheckPlayerBetValidity(player models.Player, bet *models.Bet, datetime time.Time) error {
	return nil
}

func (m *SimpleMockGame) AddPlayerBet(player models.Player, bet *models.Bet) error {
	return nil
}

func (m *SimpleMockGame) AddPlayer(player models.Player) error {
	return nil
}

func (m *SimpleMockGame) RemovePlayer(player models.Player) error {
	return nil
}

func (m *SimpleMockGame) CalculateMatchScores(match models.Match) (map[string]int, error) {
	return make(map[string]int), nil
}

func (m *SimpleMockGame) ApplyMatchScores(match models.Match, scores map[string]int) {
}

func (m *SimpleMockGame) UpdateMatch(match models.Match) error {
	return nil
}

func (m *SimpleMockGame) GetPlayersPoints() map[string]int {
	return make(map[string]int)
}

func (m *SimpleMockGame) GetPlayers() []models.Player {
	return []models.Player{}
}

func (m *SimpleMockGame) IsFinished() bool {
	return false
}

func (m *SimpleMockGame) GetWinner() []models.Player {
	return []models.Player{}
}

func (m *SimpleMockGame) GetIncomingMatchesForTesting() map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}

func (m *SimpleMockGame) GetMatchById(matchId string) (models.Match, error) {
	return nil, errors.New("match not found")
}

func (m *SimpleMockGame) Finish() {}

func TestGameCreationService_JoinGame_InvalidCode(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	code := "INVALID"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", code).Return(nil, errors.New("game code not found"))

	// Execute
	response, err := service.JoinGame(code, player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid game code")

	// Verify mocks
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_JoinGame_ExpiredCode(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	code := "ABC1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      code,
		ExpiresAt: testTime.Add(-24 * time.Hour), // Actually expired
	}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", code).Return(gameCode, nil)

	// Execute
	response, err := service.JoinGame(code, player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "game code has expired")

	// Verify mocks
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_JoinGame_GameNotFound(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	code := "ABC1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      code,
		ExpiresAt: testTime.Add(24 * time.Hour),
	}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", code).Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(nil, errors.New("game not found"))

	// Execute
	response, err := service.JoinGame(code, player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get game")

	// Verify mocks
	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

func TestGameCreationService_GetPlayerGames_Success(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameIDs := []string{"game1", "game2"}

	// Create real games instead of using mocks
	realGame1 := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game 1", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})
	realGame2 := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game 2", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame1, nil)
	mockGameRepo.On("GetGame", "game2").Return(realGame2, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{player}, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game2").Return([]models.Player{player}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(map[string]map[string]int{}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game2").Return(map[string]map[string]int{}, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(nil, errors.New("code not found"))
	mockGameCodeRepo.On("GetGameCodeByGameID", "game2").Return(nil, errors.New("code not found"))

	// Execute
	playerGames, err := service.GetPlayerGames(player)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, playerGames, 2)
	assert.Equal(t, "game1", playerGames[0].GameID)
	assert.Equal(t, "game2", playerGames[1].GameID)
	assert.Equal(t, "2025/2026", playerGames[0].SeasonYear)
	assert.Equal(t, "Ligue 1", playerGames[0].CompetitionName)

	// Verify mocks
	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_GetPlayerGames_EmptyList(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)

	// Execute
	playerGames, err := service.GetPlayerGames(player)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, playerGames, 0)

	// Verify mocks
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_GetPlayerGames_RepositoryError(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(nil, errors.New("database error"))

	// Execute
	playerGames, err := service.GetPlayerGames(player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, playerGames)
	assert.Contains(t, err.Error(), "error getting player games")

	// Verify mocks
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_GetPlayerGames_WithPlayersAndScores(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)
	mockMatchRepo := new(MockMatchRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	player1 := &models.PlayerData{ID: "player1", Name: "Test Player"}
	player2 := &models.PlayerData{ID: "player2", Name: "Other Player"}
	gameIDs := []string{"game1"}

	// Create a real game instead of using a mock
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player1, player2}, []models.Match{}, &rules.ScorerOriginal{})

	mockPlayers := []models.Player{player1, player2}
	mockScores := map[string]map[string]int{
		"match1": {"player1": 10, "player2": 5},
		"match2": {"player1": 20, "player2": 15},
	}

	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return(mockPlayers, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(mockScores, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(nil, errors.New("code not found"))

	playerGames, err := service.GetPlayerGames(player1)
	assert.NoError(t, err)
	assert.Len(t, playerGames, 1)
	game := playerGames[0]
	assert.Equal(t, "game1", game.GameID)
	assert.Len(t, game.Players, 2)
	for _, p := range game.Players {
		if p.ID == "player1" {
			assert.Equal(t, 30, p.TotalScore)
			assert.Equal(t, 10, p.ScoresByMatch["match1"])
			assert.Equal(t, 20, p.ScoresByMatch["match2"])
		}
		if p.ID == "player2" {
			assert.Equal(t, 20, p.TotalScore)
			assert.Equal(t, 5, p.ScoresByMatch["match1"])
			assert.Equal(t, 15, p.ScoresByMatch["match2"])
		}
	}

	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

// MockBetRepository is a mock implementation of BetRepository
type MockBetRepository struct {
	mock.Mock
}

func (m *MockBetRepository) GetScoresByMatchAndPlayer(gameId string) (map[string]map[string]int, error) {
	args := m.Called(gameId)
	return args.Get(0).(map[string]map[string]int), args.Error(1)
}
func (m *MockBetRepository) GetBets(gameId string, player models.Player) ([]*models.Bet, error) {
	return nil, nil
}
func (m *MockBetRepository) SaveBet(gameId string, bet *models.Bet, player models.Player) (string, *models.Bet, error) {
	return "", nil, nil
}
func (m *MockBetRepository) GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error) {
	return nil, nil, nil
}
func (m *MockBetRepository) SaveWithId(gameId string, betId string, bet *models.Bet, player models.Player) error {
	return nil
}
func (m *MockBetRepository) SaveScore(gameId string, match models.Match, player models.Player, points int) error {
	return nil
}
func (m *MockBetRepository) GetScore(gameId string, betId string) (int, error) { return 0, nil }
func (m *MockBetRepository) GetScores(gameId string) (map[string]map[string]int, error) {
	return nil, nil
}

func TestGameCreationService_PlayerJoinCacheIssue(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Create a game first
	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations for game creation
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "test-game-id", "player1").Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(nil)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	response, err := service.CreateGame(request, player)
	require.NoError(t, err)
	require.NotNil(t, response)

	// Now simulate a second player joining
	secondPlayer := &models.PlayerData{ID: "player2", Name: "Second Player"}

	// Mock expectations for joining
	mockGameCodeRepo.On("GetGameCodeByCode", "TEST").Return(&models.GameCode{
		ID:        "code-id",
		Code:      "TEST",
		GameID:    "test-game-id",
		CreatedAt: testTime,
		ExpiresAt: testTime.Add(24 * time.Hour), // Set to expire in 24 hours
	}, nil)

	// Create a real game instead of using a mock
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player, secondPlayer}, []models.Match{}, &rules.ScorerOriginal{})
	mockGameRepo.On("GetGame", "test-game-id").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player2").Return([]string{}, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "test-game-id", "player2").Return(nil)

	// Mock IsPlayerInGame to return false (player2 is not in the game yet)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "test-game-id", "player2").Return(false, nil).Once()

	// Join the game
	joinedGame, err := service.JoinGame("TEST", secondPlayer)
	require.NoError(t, err)
	require.NotNil(t, joinedGame)

	// Mock IsPlayerInGame to return true for player2 after joining (for GetGameService call)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "test-game-id", "player2").Return(true, nil)

	// Mock GetPlayersInGame to return the players (needed by GetPlayers() which now fetches from repository)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "test-game-id").Return([]models.Player{player, secondPlayer}, nil)

	// Now try to get the game service directly (this should use the cached version)
	gameService, err := service.GetGameService("test-game-id", secondPlayer)
	require.NoError(t, err)
	require.NotNil(t, gameService)

	// This should fail because the cached game service doesn't include the new player
	// The player should be in the game's players list
	players := gameService.GetPlayers()
	found := false
	for _, p := range players {
		if p.GetID() == "player2" {
			found = true
			break
		}
	}

	// This assertion should fail, proving the caching issue
	assert.True(t, found, "Player should be in the cached game service")
}

func TestGameCreationService_LeaveGame_Success(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	gameID := "game1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, gameID, player.ID).Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, gameID, player.ID).Return(nil)
	// Simulate other players remaining
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, gameID).Return([]models.Player{&models.PlayerData{ID: "other", Name: "Other Player"}}, nil)

	err := service.LeaveGame(gameID, player)
	assert.NoError(t, err)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_LeaveGame_NotInGame(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	gameID := "game1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, gameID, player.ID).Return(false, nil)

	err := service.LeaveGame(gameID, player)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrPlayerNotInGame))
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_LeaveGame_RepoError(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	gameID := "game1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	repoErr := errors.New("db error")
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, gameID, player.ID).Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, gameID, player.ID).Return(repoErr)

	err := service.LeaveGame(gameID, player)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error removing player from game")
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_LeaveGame_LastPlayerFinishesGame(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	gameID := "game1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Player is in game
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, gameID, player.ID).Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, gameID, player.ID).Return(nil)
	// No players left after removal
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, gameID).Return([]models.Player{}, nil)
	// Game loaded for finishing
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
	mockGameRepo.On("GetGame", gameID).Return(realGame, nil)
	// SaveWithId called to persist finished status
	mockGameRepo.On("SaveWithId", gameID, realGame).Return(nil)
	// Watcher unsubscribed
	mockWatcher.On("Unsubscribe", mock.MatchedBy(func(arg string) bool { return arg == gameID })).Return(nil)
	// Game code deleted
	mockGameCodeRepo.On("DeleteGameCodeByGameID", gameID).Return(nil)

	err := service.LeaveGame(gameID, player)
	assert.NoError(t, err)
	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_JoinGame_FinishedGame(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	code := "ABC1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      code,
		ExpiresAt: testTime.Add(24 * time.Hour),
	}
	// Game is finished
	mockGame := &SimpleMockGameFinished{}

	mockGameCodeRepo.On("GetGameCodeByCode", code).Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(mockGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)

	response, err := service.JoinGame(code, player)
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "cannot join a finished game")
	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

// Helper for finished game

type SimpleMockGameFinished struct{}

func (m *SimpleMockGameFinished) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}
func (m *SimpleMockGameFinished) GetPastResults() map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}
func (m *SimpleMockGameFinished) GetSeasonYear() string            { return "2025/2026" }
func (m *SimpleMockGameFinished) GetCompetitionName() string       { return "Ligue 1" }
func (m *SimpleMockGameFinished) GetGameStatus() models.GameStatus { return models.GameStatusFinished }
func (m *SimpleMockGameFinished) GetName() string                  { return "Test Game" }
func (m *SimpleMockGameFinished) CheckPlayerBetValidity(player models.Player, bet *models.Bet, datetime time.Time) error {
	return nil
}
func (m *SimpleMockGameFinished) AddPlayerBet(player models.Player, bet *models.Bet) error {
	return nil
}
func (m *SimpleMockGameFinished) AddPlayer(player models.Player) error    { return nil }
func (m *SimpleMockGameFinished) RemovePlayer(player models.Player) error { return nil }
func (m *SimpleMockGameFinished) CalculateMatchScores(match models.Match) (map[string]int, error) {
	return make(map[string]int), nil
}
func (m *SimpleMockGameFinished) ApplyMatchScores(match models.Match, scores map[string]int) {}
func (m *SimpleMockGameFinished) UpdateMatch(match models.Match) error                       { return nil }
func (m *SimpleMockGameFinished) GetPlayersPoints() map[string]int                           { return make(map[string]int) }
func (m *SimpleMockGameFinished) GetPlayers() []models.Player                                { return []models.Player{} }
func (m *SimpleMockGameFinished) IsFinished() bool                                           { return true }
func (m *SimpleMockGameFinished) GetWinner() []models.Player                                 { return []models.Player{} }
func (m *SimpleMockGameFinished) GetIncomingMatchesForTesting() map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}
func (m *SimpleMockGameFinished) GetMatchById(matchId string) (models.Match, error) {
	return nil, errors.New("match not found")
}

func (m *SimpleMockGameFinished) Finish() {}

// SimpleMockGameService is a simple mock that implements GameService without making repository calls
type SimpleMockGameService struct {
	gameID string
}

func (m *SimpleMockGameService) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}

func (m *SimpleMockGameService) GetMatchResults() map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}

func (m *SimpleMockGameService) UpdatePlayerBet(player models.Player, bet *models.Bet, now time.Time) error {
	return nil
}

func (m *SimpleMockGameService) GetPlayerBets(player models.Player) ([]*models.Bet, error) {
	return nil, nil
}

func (m *SimpleMockGameService) GetPlayers() []models.Player {
	return nil
}

func (m *SimpleMockGameService) HandleMatchUpdates(updates map[string]models.Match) error {
	return nil
}

func (m *SimpleMockGameService) GetGameID() string {
	return m.gameID
}

func (m *SimpleMockGameService) AddPlayer(player models.Player) error {
	return nil
}

func (m *SimpleMockGameService) RemovePlayer(player models.Player) error {
	return nil
}

func TestGameCreationService_Join_Leave_Rejoin_Pattern(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Step 1: Create game and join as player1
	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "test-game-id", "player1").Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(nil)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	createResp, err := service.CreateGame(request, player)
	assert.NoError(t, err)
	assert.NotNil(t, createResp)
	gameID := createResp.GameID
	code := createResp.Code

	// Step 2: Leave the game as player1
	// Create a real game with the player for the cached game service to use when removing player
	otherPlayer := &models.PlayerData{ID: "other", Name: "Other Player"}
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player, otherPlayer}, []models.Match{}, &rules.ScorerOriginal{})
	// GetGame is called multiple times during leave and rejoin operations
	mockGameRepo.On("GetGame", gameID).Return(realGame, nil)
	mockGameRepo.On("SaveWithId", gameID, realGame).Return(nil)

	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, gameID, player.ID).Return(true, nil).Once()
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, gameID, player.ID).Return(nil)
	// Simulate other players remaining after leave
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, gameID).Return([]models.Player{otherPlayer}, nil)

	err = service.LeaveGame(gameID, player)
	assert.NoError(t, err)

	// Step 3: Re-join the same game as player1
	mockGameCodeRepo.On("GetGameCodeByCode", code).Return(&models.GameCode{
		GameID:    gameID,
		Code:      code,
		ExpiresAt: testTime.Add(24 * time.Hour),
	}, nil)

	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, gameID, player.ID).Return(false, nil).Once()
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, gameID, player.ID).Return(nil)

	joinResp, err := service.JoinGame(code, player)
	assert.NoError(t, err)
	assert.NotNil(t, joinResp)
	assert.Equal(t, gameID, joinResp.GameID)

	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
	mockMatchRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
}

func TestNewGameCreationServiceWithLoadedGames_Success(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	// Create real games instead of mocks
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	realGame1 := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game 1", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})
	realGame2 := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game 2", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})
	games := map[string]models.Game{
		"game1": realGame1,
		"game2": realGame2,
	}

	// Mock expectations
	mockGameRepo.On("GetAllGames").Return(games, nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil).Times(2)

	// Execute
	service, err := NewGameCreationServiceWithLoadedGames(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, service)

	// Verify that the service can access the loaded games
	player = &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock that player is in both games
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game2", "player1").Return(true, nil)

	// Test GetGameService for both games
	gameService1, err := service.GetGameService("game1", player)
	assert.NoError(t, err)
	assert.NotNil(t, gameService1)
	assert.Equal(t, "game1", gameService1.GetGameID())

	gameService2, err := service.GetGameService("game2", player)
	assert.NoError(t, err)
	assert.NotNil(t, gameService2)
	assert.Equal(t, "game2", gameService2.GetGameID())

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestNewGameCreationServiceWithLoadedGames_RepositoryError(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	// Mock expectations - repository error
	mockGameRepo.On("GetAllGames").Return(nil, errors.New("database error"))

	// Execute
	service, err := NewGameCreationServiceWithLoadedGames(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "failed to load games from repository")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
}

func TestNewGameCreationServiceWithLoadedGames_EmptyRepository(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	// Mock expectations - empty repository
	mockGameRepo.On("GetAllGames").Return(map[string]models.Game{}, nil)

	// Execute
	service, err := NewGameCreationServiceWithLoadedGames(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, service)

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
}

func TestNewGameCreationServiceWithLoadedGames_WatcherSubscriptionError(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	// Create mock games
	mockGame1 := &SimpleMockGame{}
	mockGame2 := &SimpleMockGame{}
	games := map[string]models.Game{
		"game1": mockGame1,
		"game2": mockGame2,
	}

	// Mock expectations - watcher subscription fails for one game
	mockGameRepo.On("GetAllGames").Return(games, nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil).Once()
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(errors.New("subscription error")).Once()

	// Execute
	service, err := NewGameCreationServiceWithLoadedGames(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Assert - should still succeed even if some subscriptions fail
	assert.NoError(t, err)
	assert.NotNil(t, service)

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
}

func TestNewGameCreationServiceWithLoadedGames_NoWatcher(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	// Create mock games
	mockGame1 := &SimpleMockGame{}
	mockGame2 := &SimpleMockGame{}
	games := map[string]models.Game{
		"game1": mockGame1,
		"game2": mockGame2,
	}

	// Mock expectations
	mockGameRepo.On("GetAllGames").Return(games, nil)

	// Execute with nil watcher
	service, err := NewGameCreationServiceWithLoadedGames(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, service)

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
}

func TestNewGameCreationServiceWithLoadedGames_IntegrationWithExistingMethods(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	// Create mock games
	mockGame1 := &SimpleMockGame{}
	games := map[string]models.Game{
		"game1": mockGame1,
	}

	// Mock expectations for loading games
	mockGameRepo.On("GetAllGames").Return(games, nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	// Create service with loaded games
	service, err := NewGameCreationServiceWithLoadedGamesAndTimeFunc(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher, func() time.Time { return testTime })
	assert.NoError(t, err)
	assert.NotNil(t, service)

	// Test that we can still create new games
	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "New Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations for creating a new game
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("new-game-id", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "new-game-id", "player1").Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(nil)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	response, err := service.CreateGame(request, player)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "new-game-id", response.GameID)

	// Test that we can access the newly created game
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "new-game-id", "player1").Return(true, nil)
	gameService, err := service.GetGameService("new-game-id", player)
	assert.NoError(t, err)
	assert.NotNil(t, gameService)
	assert.Equal(t, "new-game-id", gameService.GetGameID())

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
	mockMatchRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_PlayerGameLimitReached(t *testing.T) {
	// Setup
	mockGameRepo := &MockGameRepository{}
	mockGameCodeRepo := &MockGameCodeRepository{}
	mockGamePlayerRepo := &MockGamePlayerRepository{}
	mockBetRepo := &MockBetRepository{}
	mockMatchRepo := &MockMatchRepository{}
	mockWatcher := &MockWatcher{}

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Test data
	req := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - player already has 5 games
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{"game1", "game2", "game3", "game4", "game5"}, nil)

	// Execute
	response, err := service.CreateGame(req, player)

	// Assert
	assert.Nil(t, response)
	assert.Error(t, err)
	assert.Equal(t, ErrPlayerGameLimit, err)

	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_PlayerGameLimitCheckFails(t *testing.T) {
	// Setup
	mockGameRepo := &MockGameRepository{}
	mockGameCodeRepo := &MockGameCodeRepository{}
	mockGamePlayerRepo := &MockGamePlayerRepository{}
	mockBetRepo := &MockBetRepository{}
	mockMatchRepo := &MockMatchRepository{}
	mockWatcher := &MockWatcher{}

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Test data
	req := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - database error when checking player games
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(nil, errors.New("database error"))

	// Execute
	response, err := service.CreateGame(req, player)

	// Assert
	assert.Nil(t, response)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check player games")

	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_JoinGame_PlayerGameLimitReached(t *testing.T) {
	// Setup
	mockGameRepo := &MockGameRepository{}
	mockGameCodeRepo := &MockGameCodeRepository{}
	mockGamePlayerRepo := &MockGamePlayerRepository{}
	mockBetRepo := &MockBetRepository{}
	mockMatchRepo := &MockMatchRepository{}
	mockWatcher := &MockWatcher{}

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Test data
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	mockGame := &SimpleMockGame{}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(mockGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{"game1", "game2", "game3", "game4", "game5"}, nil)

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.Nil(t, response)
	assert.Error(t, err)
	assert.Equal(t, ErrPlayerGameLimit, err)

	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_JoinGame_PlayerGameLimitCheckFails(t *testing.T) {
	// Setup
	mockGameRepo := &MockGameRepository{}
	mockGameCodeRepo := &MockGameCodeRepository{}
	mockGamePlayerRepo := &MockGamePlayerRepository{}
	mockBetRepo := &MockBetRepository{}
	mockMatchRepo := &MockMatchRepository{}
	mockWatcher := &MockWatcher{}

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Test data
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	mockGame := &SimpleMockGame{}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(mockGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(nil, errors.New("database error"))

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.Nil(t, response)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check player games")

	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_PlayerCanCreateFifthGame(t *testing.T) {
	// Setup
	mockGameRepo := &MockGameRepository{}
	mockGameCodeRepo := &MockGameCodeRepository{}
	mockGamePlayerRepo := &MockGamePlayerRepository{}
	mockBetRepo := &MockBetRepository{}
	mockMatchRepo := &MockMatchRepository{}
	mockWatcher := &MockWatcher{}

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Test data
	req := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - player has exactly 4 games (should allow 5th)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{"game1", "game2", "game3", "game4"}, nil)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("new-game-id", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "new-game-id", "player1").Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	// Execute
	response, err := service.CreateGame(req, player)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "new-game-id", response.GameID)

	mockGamePlayerRepo.AssertExpectations(t)
	mockMatchRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
}

func TestGameCreationService_GetPlayerGames_FinishedGameStatus(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := setupCreationTestService(t, mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameID := "game1"
	gameIDs := []string{gameID}

	// Create a real game with matches that will be finished
	match1 := models.NewSeasonMatch("Team1", "Team2", "2025/2026", "Ligue 1", testTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2025/2026", "Ligue 1", testTime.Add(1*time.Hour), 2)
	matches := []models.Match{match1, match2}

	// Create a real game with the matches
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player}, matches, &rules.ScorerOriginal{})

	// Create a game service to interact with the game
	gameService := NewGameService(gameID, mockGameRepo, mockBetRepo, &MockGamePlayerRepository{})

	// Set up mock to return the real game when GetGame is called
	mockGameRepo.On("GetGame", gameID).Return(realGame, nil)

	// Add bets directly to the game (simpler for testing)
	bet1 := models.NewBet(match1, 2, 1)
	bet2 := models.NewBet(match2, 1, 0)
	realGame.AddPlayerBet(player, bet1)
	realGame.AddPlayerBet(player, bet2)

	// Create finished matches
	finishedMatch1 := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2025/2026", "Ligue 1", testTime, 1, 1.0, 2.0, 3.0)
	finishedMatch2 := models.NewFinishedSeasonMatch("Team3", "Team4", 1, 0, "2025/2026", "Ligue 1", testTime.Add(1*time.Hour), 2, 1.0, 2.0, 3.0)

	// Handle match updates through the game service (this is what happens in production)
	updates := map[string]models.Match{
		match1.Id(): finishedMatch1,
		match2.Id(): finishedMatch2,
	}

	// Set up mock expectation for SaveWithId since the game will finish
	mockGameRepo.On("SaveWithId", gameID, mock.AnythingOfType("*rules.GameImpl")).Return(nil)

	err := gameService.HandleMatchUpdates(updates)
	require.NoError(t, err)

	// Verify the game is actually finished
	require.True(t, realGame.IsFinished(), "Game should be finished after all matches are completed")

	gameCode := &models.GameCode{
		GameID:    gameID,
		Code:      "ABC1",
		ExpiresAt: testTime.Add(24 * time.Hour),
	}

	// Set up mock expectations for GetPlayerGames
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", gameID).Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, gameID).Return([]models.Player{player}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", gameID).Return(make(map[string]map[string]int), nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", gameID).Return(gameCode, nil)

	// Get player games
	playerGames, err := service.GetPlayerGames(player)
	require.NoError(t, err)
	require.Len(t, playerGames, 1)

	// Verify the game status is "finished"
	game := playerGames[0]
	assert.Equal(t, "finished", game.Status)
	assert.Equal(t, gameID, game.GameID)
	assert.Equal(t, "Test Game", game.Name)

	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}
