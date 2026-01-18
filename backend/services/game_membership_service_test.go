package services

import (
	"errors"
	"ligain/backend/models"
	"ligain/backend/rules"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupMembershipTestRegistry creates a registry for membership tests
func setupMembershipTestRegistry(t *testing.T, mockGameRepo *MockGameRepository, mockBetRepo *MockBetRepository, mockGamePlayerRepo *MockGamePlayerRepository, watcher MatchWatcherService) *GameServiceRegistry {
	// Constructor calls loadAll, so we need to mock GetAllGames
	mockGameRepo.On("GetAllGames").Return(map[string]models.Game{}, nil)

	registry, err := NewGameServiceRegistry(mockGameRepo, mockBetRepo, mockGamePlayerRepo, watcher)
	require.NoError(t, err)
	return registry
}

func TestGameMembershipService_AddPlayerToGame_Success(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - player not already in game
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)

	// Execute
	err := service.AddPlayerToGame("game1", player)

	// Assert
	assert.NoError(t, err)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_AddPlayerToGame_AlreadyInGame(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - player already in game
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)

	// Execute - should return nil (idempotent)
	err := service.AddPlayerToGame("game1", player)

	// Assert
	assert.NoError(t, err)
	mockGamePlayerRepo.AssertExpectations(t)
	// Should not call AddPlayerToGame
	mockGamePlayerRepo.AssertNotCalled(t, "AddPlayerToGame", mock.Anything, mock.Anything, mock.Anything)
}

func TestGameMembershipService_AddPlayerToGame_RepositoryError(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - error checking membership
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, errors.New("database error"))

	// Execute
	err := service.AddPlayerToGame("game1", player)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error checking if player is in game")
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_RemovePlayerFromGame_Success(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	otherPlayer := &models.PlayerData{ID: "player2", Name: "Other Player"}

	// Mock expectations
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, "game1", "player1").Return(nil)
	// Other players remain
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{otherPlayer}, nil)

	// Execute
	err := service.RemovePlayerFromGame("game1", player)

	// Assert
	assert.NoError(t, err)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_RemovePlayerFromGame_NotInGame(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations - player not in game
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)

	// Execute
	err := service.RemovePlayerFromGame("game1", player)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrPlayerNotInGame, err)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_RemovePlayerFromGame_LastPlayerDeletesGame(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Create a real game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, "game1", "player1").Return(nil)
	// No players left
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{}, nil)

	// Game deletion expectations
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)
	mockWatcher.On("Unsubscribe", "game1").Return(nil)
	mockGameCodeRepo.On("DeleteGameCodeByGameID", "game1").Return(nil)

	// Execute
	err := service.RemovePlayerFromGame("game1", player)

	// Assert
	assert.NoError(t, err)
	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameMembershipService_IsPlayerInGame_True(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)

	// Mock expectations
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)

	// Execute
	isInGame, err := service.IsPlayerInGame("game1", "player1")

	// Assert
	assert.NoError(t, err)
	assert.True(t, isInGame)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_IsPlayerInGame_False(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)

	// Mock expectations
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)

	// Execute
	isInGame, err := service.IsPlayerInGame("game1", "player1")

	// Assert
	assert.NoError(t, err)
	assert.False(t, isInGame)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_GetPlayersInGame_Success(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)

	player1 := &models.PlayerData{ID: "player1", Name: "Player 1"}
	player2 := &models.PlayerData{ID: "player2", Name: "Player 2"}
	expectedPlayers := []models.Player{player1, player2}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return(expectedPlayers, nil)

	// Execute
	players, err := service.GetPlayersInGame("game1")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedPlayers, players)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_GetPlayersInGame_RepositoryError(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, nil)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, nil)

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return(nil, errors.New("database error"))

	// Execute
	players, err := service.GetPlayersInGame("game1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, players)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_LeaveGame_Success(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	otherPlayer := &models.PlayerData{ID: "player2", Name: "Other Player"}

	// Mock expectations
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, "game1", "player1").Return(nil)
	// Other players remain
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{otherPlayer}, nil)

	// Execute
	err := service.LeaveGame("game1", player)

	// Assert
	assert.NoError(t, err)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_LeaveGame_UnregistersFromRegistry(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// First register a game service
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	_, err := registry.Create("game1")
	require.NoError(t, err)

	// Mock expectations for leaving
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, "game1", "player1").Return(nil)
	// No players left - should trigger game deletion
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{}, nil)

	// Game deletion expectations
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)
	mockWatcher.On("Unsubscribe", "game1").Return(nil)
	mockGameCodeRepo.On("DeleteGameCodeByGameID", "game1").Return(nil)

	// Execute
	err = service.LeaveGame("game1", player)

	// Assert
	assert.NoError(t, err)

	// Verify game service is unregistered
	_, exists := registry.Get("game1")
	assert.False(t, exists)

	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameMembershipService_AddPlayerToGame_UpdatesCachedGameService(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// First register a game service
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	_, err := registry.Create("game1")
	require.NoError(t, err)

	// Mock AddPlayer expectations
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)

	// Create a real game for the UpdateCachedGameService call
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)

	// Execute
	err = service.AddPlayerToGame("game1", player)

	// Assert
	assert.NoError(t, err)
	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameMembershipService_RemovePlayerFromGame_UpdatesCachedGameService(t *testing.T) {
	// Setup
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameRepo := new(MockGameRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)
	mockWatcher := new(MockWatcher)

	registry := setupMembershipTestRegistry(t, mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	service := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)

	player1 := &models.PlayerData{ID: "player1", Name: "Player 1"}
	player2 := &models.PlayerData{ID: "player2", Name: "Player 2"}

	// Create a real game with both players
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player1, player2}, []models.Match{}, &rules.ScorerOriginal{})

	// First register a game service with the game
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)
	_, err := registry.Create("game1")
	require.NoError(t, err)

	// Verify the underlying game object has both players initially
	// This is what scoring uses (game.GetPlayers(), not GameService.GetPlayers())
	assert.Len(t, realGame.GetPlayers(), 2)

	// Mock RemovePlayer expectations
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, "game1", "player1").Return(nil)
	// Other player remains
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{player2}, nil)

	// Execute - remove player1
	err = service.RemovePlayerFromGame("game1", player1)

	// Assert
	assert.NoError(t, err)

	// BUG: The cached game service's underlying game object still has player1
	// This means scoring (handleScoreUpdate) would still attribute points to player1
	// because it calls game.GetPlayers(), not gamePlayerRepo.GetPlayersInGame()
	cachedGs, exists := registry.Get("game1")
	require.True(t, exists)
	_ = cachedGs // We need to check the underlying game object

	// The underlying game object should only have player2 now
	// If this fails, removed players continue to get points attributed
	assert.Len(t, realGame.GetPlayers(), 1, "Underlying game object should have player removed - otherwise scoring continues for departed players")
	if len(realGame.GetPlayers()) > 0 {
		assert.Equal(t, "player2", realGame.GetPlayers()[0].GetID())
	}

	mockGamePlayerRepo.AssertExpectations(t)
}
