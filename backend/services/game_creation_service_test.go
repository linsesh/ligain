package services

import (
	"errors"
	"liguain/backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestGameCreationService_CreateGame_Success(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(nil)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test-game-id", response.GameID)
	assert.Len(t, response.Code, 4)
	assert.Regexp(t, "^[A-Z0-9]{4}$", response.Code)

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_GameCreationFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("", errors.New("database error"))
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request)

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
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations - all codes exist (simulating collision)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(true, nil).Times(10)
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to generate unique code after 10 attempts")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_CodeExistsCheckFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, errors.New("database error"))
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "error checking code existence")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_CreateGame_GameCodeCreationFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("test-game-id", nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(errors.New("database error"))
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)

	// Execute
	response, err := service.CreateGame(request)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to create game code")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameCreationService_CleanupExpiredCodes(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

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
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

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
	mockMatchRepo := new(MockMatchRepository)
	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Premier League",
	}
	response, err := service.CreateGame(request)
	assert.ErrorIs(t, err, ErrInvalidCompetition)
	assert.Nil(t, response)
}

func TestGameCreationService_CreateGame_InvalidSeasonYear(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockMatchRepo := new(MockMatchRepository)
	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

	request := &CreateGameRequest{
		SeasonYear:      "2024/2025",
		CompetitionName: "Ligue 1",
	}
	response, err := service.CreateGame(request)
	assert.ErrorIs(t, err, ErrInvalidSeasonYear)
	assert.Nil(t, response)
}

func TestGameCreationService_CreateGame_MatchLoadingFails(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

	request := &CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
	}

	// Mock expectations - match loading fails
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return(nil, errors.New("database error"))

	// Execute
	response, err := service.CreateGame(request)

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
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

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
	return models.GameStatusNotStarted
}

func (m *SimpleMockGame) CheckPlayerBetValidity(player models.Player, bet *models.Bet, datetime time.Time) error {
	return nil
}

func (m *SimpleMockGame) AddPlayerBet(player models.Player, bet *models.Bet) error {
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

func TestGameCreationService_JoinGame_InvalidCode(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

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
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

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
	mockMatchRepo := new(MockMatchRepository)

	service := NewGameCreationService(mockGameRepo, mockGameCodeRepo, mockMatchRepo)

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
