package services

import (
	"errors"
	"liguain/backend/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGameRepository is a mock implementation of GameRepository
type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) GetGame(gameId string) (models.Game, error) {
	args := m.Called(gameId)
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
