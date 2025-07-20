package services

import (
	"context"
	"errors"
	"ligain/backend/models"
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
func (m *MockWatcher) Unsubscribe(gameID string) error { return nil }
func (m *MockWatcher) Start(ctx context.Context) error { return nil }
func (m *MockWatcher) Stop() error                     { return nil }

func TestGameCreationService_CreateGame_Success(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - all codes exist (simulating collision)
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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - match loading fails
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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	code := "ABC1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      code,
		ExpiresAt: testTime.Add(24 * time.Hour),
	}
	mockGame := &SimpleMockGame{}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", code).Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(mockGame, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)

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

func TestGameCreationService_JoinGame_InvalidCode(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	code := "ABC1"
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      code,
		ExpiresAt: time.Now().Add(-24 * time.Hour), // Actually expired
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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameIDs := []string{"game1", "game2"}
	mockGame1 := &SimpleMockGame{}
	mockGame2 := &SimpleMockGame{}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", "game1").Return(mockGame1, nil)
	mockGameRepo.On("GetGame", "game2").Return(mockGame2, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{}, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game2").Return([]models.Player{}, nil)
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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

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

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameIDs := []string{"game1"}
	mockGame := &SimpleMockGame{}
	mockPlayers := []models.Player{
		&models.PlayerData{ID: "player1", Name: "Test Player"},
		&models.PlayerData{ID: "player2", Name: "Other Player"},
	}
	mockScores := map[string]map[string]int{
		"match1": {"player1": 10, "player2": 5},
		"match2": {"player1": 20, "player2": 15},
	}

	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", "game1").Return(mockGame, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return(mockPlayers, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(mockScores, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(nil, errors.New("code not found"))

	playerGames, err := service.GetPlayerGames(player)
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
func (m *MockBetRepository) GetScores(gameId string) (map[string]int, error)   { return nil, nil }

func TestGameCreationService_PlayerJoinCacheIssue(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockBetRepo, mockMatchRepo, mockWatcher)

	// Create a game first
	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations for game creation
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
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour), // Set to expire in 24 hours
	}, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "test-game-id", "player2").Return(nil)

	// Mock the GetGame call that happens during JoinGame
	mockGameRepo.On("GetGame", "test-game-id").Return(&SimpleMockGame{}, nil)

	// Mock IsPlayerInGame to return false (player2 is not in the game yet)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "test-game-id", "player2").Return(false, nil)

	// Join the game
	joinedGame, err := service.JoinGame("TEST", secondPlayer)
	require.NoError(t, err)
	require.NotNil(t, joinedGame)

	// Now try to get the game service directly (this should use the cached version)
	gameService, err := service.GetGameService("test-game-id")
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
