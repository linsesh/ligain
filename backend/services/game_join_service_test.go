package services

import (
	"errors"
	"ligain/backend/models"
	"ligain/backend/rules"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var joinTestTime = time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC)

// setupJoinTestRegistry creates a registry for join tests
func setupJoinTestRegistry(t *testing.T, mockGameRepo *MockGameRepository, mockBetRepo *MockBetRepository, mockGamePlayerRepo *MockGamePlayerRepository, watcher MatchWatcherService) *GameServiceRegistry {
	// Constructor calls loadAll, so we need to mock GetAllGames
	mockGameRepo.On("GetAllGames").Return(map[string]models.Game{}, nil)

	registry, err := NewGameServiceRegistry(mockGameRepo, mockBetRepo, mockGamePlayerRepo, watcher)
	require.NoError(t, err)
	return registry
}

func TestGameJoinService_JoinGame_Success(t *testing.T) {
	// Setup
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupJoinTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)
	service := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return joinTestTime })

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: joinTestTime.Add(24 * time.Hour),
	}

	// Create a real game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "game1", response.GameID)
	assert.Equal(t, "2025/2026", response.SeasonYear)
	assert.Equal(t, "Ligue 1", response.CompetitionName)
	assert.Equal(t, "Successfully joined the game", response.Message)

	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameJoinService_JoinGame_InvalidCode(t *testing.T) {
	// Setup
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupJoinTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)
	service := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return joinTestTime })

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", "INVALID").Return(nil, errors.New("game code not found"))

	// Execute
	response, err := service.JoinGame("INVALID", player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "invalid game code")

	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameJoinService_JoinGame_ExpiredCode(t *testing.T) {
	// Setup
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupJoinTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)
	service := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return joinTestTime })

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: joinTestTime.Add(-24 * time.Hour), // Expired
	}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "game code has expired")

	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameJoinService_JoinGame_FinishedGame(t *testing.T) {
	// Setup
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupJoinTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)
	service := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return joinTestTime })

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: joinTestTime.Add(24 * time.Hour),
	}

	// Create a finished game
	finishedGame := &SimpleMockGameFinished{}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(finishedGame, nil)

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "cannot join a finished game")

	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

func TestGameJoinService_JoinGame_PlayerGameLimit(t *testing.T) {
	// Setup
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupJoinTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)
	service := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return joinTestTime })

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: joinTestTime.Add(24 * time.Hour),
	}

	// Create a real game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations - player already has 5 games
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{"g1", "g2", "g3", "g4", "g5"}, nil)

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Equal(t, ErrPlayerGameLimit, err)

	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameJoinService_JoinGame_GameNotFound(t *testing.T) {
	// Setup
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupJoinTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)
	service := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return joinTestTime })

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: joinTestTime.Add(24 * time.Hour),
	}

	// Mock expectations
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(nil, errors.New("game not found"))

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get game")

	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

func TestGameJoinService_JoinGame_UpdatesCachedGameService(t *testing.T) {
	// Setup
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupJoinTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)
	service := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return joinTestTime })

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: joinTestTime.Add(24 * time.Hour),
	}

	// Create a real game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})

	// First create a cached game service
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	_, err := registry.GetOrCreate("game1")
	assert.NoError(t, err)

	// Mock expectations for joining
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)

	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
}

func TestGameJoinService_JoinGame_PlayerCanJoinFifthGame(t *testing.T) {
	// Setup
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupJoinTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)
	service := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return joinTestTime })

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: joinTestTime.Add(24 * time.Hour),
	}

	// Create a real game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations - player has 4 games (can join 5th)
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{"g1", "g2", "g3", "g4"}, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)

	// Execute
	response, err := service.JoinGame("ABC1", player)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, response)

	mockGameCodeRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockGamePlayerRepo.AssertExpectations(t)
}
