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

func TestGameServiceRegistry_GetOrCreate_NewGame(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations for watcher subscription
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	// Execute
	gameService, err := registry.GetOrCreate("game1")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, gameService)
	assert.Equal(t, "game1", gameService.GetGameID())

	// Verify mock
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_GetOrCreate_CachedGame(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations - Subscribe should only be called once (first GetOrCreate)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil).Once()

	// Execute - first call creates
	gameService1, err := registry.GetOrCreate("game1")
	require.NoError(t, err)

	// Execute - second call should return cached instance
	gameService2, err := registry.GetOrCreate("game1")
	require.NoError(t, err)

	// Assert - same instance returned
	assert.Equal(t, gameService1, gameService2)
	assert.Equal(t, "game1", gameService2.GetGameID())

	// Verify Subscribe was called only once
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_GetOrCreate_SubscriptionFails(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations - subscription fails
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(errors.New("subscription error"))

	// Execute
	gameService, err := registry.GetOrCreate("game1")

	// Assert - should fail when subscription fails
	assert.Error(t, err)
	assert.Nil(t, gameService)
	assert.Contains(t, err.Error(), "failed to subscribe game to watcher")

	// Verify mock
	mockWatcher.AssertExpectations(t)
}

func TestGameServiceRegistry_GetOrCreate_NoWatcher(t *testing.T) {
	// Setup - no watcher
	registry, _, _, _ := setupEmptyRegistry(t, nil)

	// Execute
	gameService, err := registry.GetOrCreate("game1")

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
	retrievedService, err := registry.GetOrCreate("game1")
	assert.NoError(t, err)
	assert.Equal(t, gameService, retrievedService)
}

func TestGameServiceRegistry_Unregister(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// First create a game
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	gameService1, err := registry.GetOrCreate("game1")
	require.NoError(t, err)
	require.NotNil(t, gameService1)

	// Execute - unregister the game
	registry.Unregister("game1")

	// After unregister, GetOrCreate should create a new service (not return cached)
	gameService2, err := registry.GetOrCreate("game1")
	require.NoError(t, err)

	// Should be different instance
	assert.NotEqual(t, gameService1, gameService2)

	// Verify Subscribe was called twice (once for each GetOrCreate)
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

func TestGameServiceRegistry_ConcurrentAccess(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations - may be called multiple times due to concurrency
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	// Execute - concurrent access
	const numGoroutines = 10
	done := make(chan bool)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := registry.GetOrCreate("game1")
			assert.NoError(t, err)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify that only one service instance exists
	gameService1, _ := registry.GetOrCreate("game1")
	gameService2, _ := registry.GetOrCreate("game1")
	assert.Equal(t, gameService1, gameService2)
}

func TestGameServiceRegistry_Get_Exists(t *testing.T) {
	mockWatcher := new(MockWatcher)
	registry, _, _, _ := setupEmptyRegistry(t, mockWatcher)

	// Mock expectations
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)

	// Create a game first
	_, err := registry.GetOrCreate("game1")
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
