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

// setupEmptyRegistry creates a registry with no pre-loaded games
func setupEmptyRegistry(t *testing.T, watcher MatchWatcherService) (*GameServiceRegistry, *MockGameRepository, *MockBetRepository, *MockGamePlayerRepository) {
	mockGameRepo := new(MockGameRepository)
	mockBetRepo := new(MockBetRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)

	// Constructor calls loadAll, so we need to mock GetAllGames
	mockGameRepo.On("GetAllGames").Return(map[string]models.Game{}, nil)

	registry, err := NewGameServiceRegistry(mockGameRepo, mockBetRepo, mockGamePlayerRepo, watcher)
	require.NoError(t, err)

	return registry, mockGameRepo, mockBetRepo, mockGamePlayerRepo
}

func TestGameServiceRegistry_Create_NewGame(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations for watcher subscription
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	// Execute
	gameService, err := registry.Create("game1")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, gameService)
	assert.Equal(t, "game1", gameService.GetGameID())

	// Verify mock
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_Create_ThenGet_ReturnsSameInstance(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations - Subscribe should only be called once (Create)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil).Once()

	// Execute - Create registers the service
	gameService1, err := registry.Create("game1")
	require.NoError(t, err)

	// Execute - Get should return the same instance
	gameService2, exists := registry.Get("game1")
	require.True(t, exists)

	// Assert - same instance returned
	assert.Equal(t, gameService1, gameService2)
	assert.Equal(t, "game1", gameService2.GetGameID())

	// Verify Subscribe was called only once
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_Create_SubscriptionFails(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations - subscription fails
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(errors.New("subscription error"))

	// Execute
	gameService, err := registry.Create("game1")

	// Assert - should fail when subscription fails
	assert.Error(t, err)
	assert.Nil(t, gameService)
	assert.Contains(t, err.Error(), "failed to subscribe game to watcher")

	// Verify mock
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_Create_NoWatcher(t *testing.T) {
	// Setup - no watcher
	registry, _, _, _ := setupEmptyRegistry(t, nil)

	// Execute
	gameService, err := registry.Create("game1")

	// Assert - should succeed without watcher
	assert.NoError(t, err)
	assert.NotNil(t, gameService)
	assert.Equal(t, "game1", gameService.GetGameID())
}

func TestGameServiceRegistry_Register(t *testing.T) {
	registry, mockGameRepo, mockBetRepo, mockGamePlayerRepo := setupEmptyRegistry(t, nil)

	// Create a game service to register
	gameService := NewGameService("game1", mockGameRepo, mockBetRepo, mockGamePlayerRepo)

	// Execute
	registry.Register("game1", gameService)

	// Get should return the registered service
	retrievedService, exists := registry.Get("game1")
	assert.True(t, exists)
	assert.Equal(t, gameService, retrievedService)
}

func TestGameServiceRegistry_Unregister(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// First create a game
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	gameService1, err := registry.Create("game1")
	require.NoError(t, err)
	require.NotNil(t, gameService1)

	// Verify it exists
	_, exists := registry.Get("game1")
	require.True(t, exists)

	// Execute - unregister the game
	registry.Unregister("game1")

	// After unregister, Get should return false
	_, exists = registry.Get("game1")
	assert.False(t, exists)

	// Verify Subscribe was called once
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_LoadsGamesOnConstruction(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockBetRepo := new(MockBetRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockWatcher := new(MockWatcher)

	// Create real games
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

	// Execute - constructor loads games
	registry, err := NewGameServiceRegistry(mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)

	// Assert
	require.NoError(t, err)

	// Verify services are loaded (Get returns them without creating new ones)
	gameService1, exists := registry.Get("game1")
	assert.True(t, exists)
	assert.NotNil(t, gameService1)
	assert.Equal(t, "game1", gameService1.GetGameID())

	gameService2, exists := registry.Get("game2")
	assert.True(t, exists)
	assert.NotNil(t, gameService2)
	assert.Equal(t, "game2", gameService2.GetGameID())

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_ConstructorFailsOnRepositoryError(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockBetRepo := new(MockBetRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockWatcher := new(MockWatcher)

	// Mock expectations - repository error
	mockGameRepo.On("GetAllGames").Return(nil, errors.New("database error"))

	// Execute
	registry, err := NewGameServiceRegistry(mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, registry)
	assert.Contains(t, err.Error(), "failed to load games from repository")

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
}

func TestGameServiceRegistry_ConstructorContinuesOnSubscriptionError(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockBetRepo := new(MockBetRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockWatcher := new(MockWatcher)

	// Create real games
	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	realGame1 := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game 1", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})
	games := map[string]models.Game{
		"game1": realGame1,
	}

	// Mock expectations - subscription fails but should continue
	mockGameRepo.On("GetAllGames").Return(games, nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(errors.New("subscription error"))

	// Execute - should succeed even if subscription fails (just logs warning)
	registry, err := NewGameServiceRegistry(mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)

	// Assert - constructor doesn't fail on subscription errors
	assert.NoError(t, err)
	assert.NotNil(t, registry)

	// Verify mocks
	mockGameRepo.AssertExpectations(t)
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_ConcurrentGet(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// First create the game
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil).Once()
	originalService, err := registry.Create("game1")
	require.NoError(t, err)

	// Execute - concurrent Get access
	const numGoroutines = 10
	done := make(chan bool)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			gs, exists := registry.Get("game1")
			assert.True(t, exists)
			assert.Equal(t, originalService, gs)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_Get_Exists(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	// Create a game first
	_, err := registry.Create("game1")
	require.NoError(t, err)

	// Execute - Get existing game
	gameService, exists := registry.Get("game1")

	// Assert
	assert.True(t, exists)
	assert.NotNil(t, gameService)
	assert.Equal(t, "game1", gameService.GetGameID())
}

func TestGameServiceRegistry_Get_NotExists(t *testing.T) {
	registry, _, _, _ := setupEmptyRegistry(t, nil)

	// Execute - Get non-existent game
	gameService, exists := registry.Get("nonexistent")

	// Assert
	assert.False(t, exists)
	assert.Nil(t, gameService)
}
