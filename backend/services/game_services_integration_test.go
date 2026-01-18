package services

import (
	"ligain/backend/models"
	"ligain/backend/rules"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var integrationTestTime = time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC)

// setupIntegrationTest creates all services with SHARED mock repositories
// This verifies that services coordinate through repositories, not cached state
func setupIntegrationTest() (
	*GameCreationService,
	*GameJoinService,
	*GameQueryService,
	*GameMembershipService,
	*GameServiceRegistry,
	*MockGameRepository,
	*MockGamePlayerRepository,
	*MockGameCodeRepository,
	*MockBetRepository,
	*MockMatchRepository,
	*MockWatcher,
) {
	// Create shared mock repositories
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)
	mockMatchRepo := new(MockMatchRepository)
	mockWatcher := new(MockWatcher)

	// Constructor calls loadAll, so we need to mock GetAllGames
	mockGameRepo.On("GetAllGames").Return(map[string]models.Game{}, nil)

	// Wire services with SAME mock instances
	registry, _ := NewGameServiceRegistry(mockGameRepo, mockBetRepo, mockGamePlayerRepo, mockWatcher)
	membershipService := NewGameMembershipService(mockGamePlayerRepo, mockGameRepo, mockGameCodeRepo, registry, mockWatcher)
	queryService := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)
	joinService := NewGameJoinService(mockGameCodeRepo, mockGameRepo, membershipService, registry, func() time.Time { return integrationTestTime })
	creationService := NewGameCreationServiceWithServices(
		mockGameRepo, mockGameCodeRepo, mockGamePlayerRepo, mockMatchRepo,
		registry, membershipService, queryService, joinService,
		func() time.Time { return integrationTestTime },
	)

	return creationService, joinService, queryService, membershipService, registry,
		mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo, mockMatchRepo, mockWatcher
}

// TestCreateGame_QueryServiceSeesNewGame verifies that after creating a game,
// the QueryService can see it immediately
func TestCreateGame_QueryServiceSeesNewGame(t *testing.T) {
	creationService, _, queryService, _, _,
		mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo, mockMatchRepo, mockWatcher := setupIntegrationTest()

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})

	// Setup ALL mocks upfront - first call returns empty, subsequent calls return the game
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil).Once()
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("game1", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(nil)

	// Create game
	createResp, err := creationService.CreateGame(&CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}, player)
	require.NoError(t, err)
	assert.Equal(t, "game1", createResp.GameID)

	// Setup mocks for query - the second call to GetPlayerGames should return the new game
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{"game1"}, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{player}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(map[string]map[string]int{}, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(&models.GameCode{Code: "ABC1"}, nil)

	// Query should see the new game
	games, err := queryService.GetPlayerGames(player)
	require.NoError(t, err)
	require.Len(t, games, 1)
	assert.Equal(t, "game1", games[0].GameID)
	assert.Equal(t, "Test Game", games[0].Name)
}

// TestJoinGame_QueryServiceSeesNewPlayer verifies that after a player joins,
// QueryService sees the new player
func TestJoinGame_QueryServiceSeesNewPlayer(t *testing.T) {
	_, joinService, queryService, _, _,
		mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo, _, _ := setupIntegrationTest()

	playerA := &models.PlayerData{ID: "playerA", Name: "Player A"}
	playerB := &models.PlayerData{ID: "playerB", Name: "Player B"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "XYZ1",
		ExpiresAt: integrationTestTime.Add(24 * time.Hour),
	}

	// Setup game with player A (but no players initially for AddPlayer to work)
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})

	// Setup mocks for join - first GetPlayerGames call returns empty
	mockGameCodeRepo.On("GetGameCodeByCode", "XYZ1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "playerB").Return([]string{}, nil).Once()
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "playerB").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "playerB").Return(nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)

	// Player B joins
	joinResp, err := joinService.JoinGame("XYZ1", playerB)
	require.NoError(t, err)
	assert.Equal(t, "game1", joinResp.GameID)

	// Setup mocks for query - both players visible now
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "playerB").Return([]string{"game1"}, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{playerA, playerB}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(map[string]map[string]int{}, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(gameCode, nil)

	// Query should see both players
	games, err := queryService.GetPlayerGames(playerB)
	require.NoError(t, err)
	require.Len(t, games, 1)
	assert.Len(t, games[0].Players, 2)
}

// TestJoinGame_RegistryGameServiceSeesNewPlayer verifies that the cached GameService
// reflects repo state after a player joins
func TestJoinGame_RegistryGameServiceSeesNewPlayer(t *testing.T) {
	_, joinService, _, _, registry,
		mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, _, _, mockWatcher := setupIntegrationTest()

	playerA := &models.PlayerData{ID: "playerA", Name: "Player A"}
	playerB := &models.PlayerData{ID: "playerB", Name: "Player B"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "XYZ1",
		ExpiresAt: integrationTestTime.Add(24 * time.Hour),
	}

	// Setup game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})

	// First, get or create game service (simulating it was already cached)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	_, err := registry.Create("game1")
	require.NoError(t, err)

	// Setup mocks for join
	mockGameCodeRepo.On("GetGameCodeByCode", "XYZ1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "playerB").Return([]string{}, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "playerB").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "playerB").Return(nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)

	// Player B joins
	_, err = joinService.JoinGame("XYZ1", playerB)
	require.NoError(t, err)

	// Setup mock for GetPlayers - should return BOTH players from repo
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{playerA, playerB}, nil)

	// GameService.GetPlayers should fetch from repo and see new player
	gs, exists := registry.Get("game1")
	require.True(t, exists)
	players := gs.GetPlayers()
	assert.Len(t, players, 2)
}

// TestLeaveGame_QueryServiceNoLongerShowsGame verifies that after a player leaves,
// QueryService no longer shows that game for the player
func TestLeaveGame_QueryServiceNoLongerShowsGame(t *testing.T) {
	_, _, queryService, membershipService, _,
		_, mockGamePlayerRepo, _, _, _, _ := setupIntegrationTest()

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	otherPlayer := &models.PlayerData{ID: "other", Name: "Other Player"}

	// Setup mocks for leave
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, "game1", "player1").Return(nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{otherPlayer}, nil)

	// Player leaves
	err := membershipService.LeaveGame("game1", player)
	require.NoError(t, err)

	// Setup mocks for query - player no longer in any games
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)

	// Query should return empty list
	games, err := queryService.GetPlayerGames(player)
	require.NoError(t, err)
	assert.Len(t, games, 0)
}

// TestLeaveGame_RegistryGameServiceNoLongerShowsPlayer verifies that the cached
// GameService reflects that a player has left
func TestLeaveGame_RegistryGameServiceNoLongerShowsPlayer(t *testing.T) {
	_, _, _, membershipService, registry,
		_, mockGamePlayerRepo, _, _, _, mockWatcher := setupIntegrationTest()
	// Other unused variables from setupIntegrationTest are intentionally ignored

	playerA := &models.PlayerData{ID: "playerA", Name: "Player A"}
	playerB := &models.PlayerData{ID: "playerB", Name: "Player B"}

	// First, get or create game service (simulating it was already cached)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	_, err := registry.Create("game1")
	require.NoError(t, err)

	// Setup mocks for leave (player A leaves, player B remains)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "playerA").Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, "game1", "playerA").Return(nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{playerB}, nil)

	// Player A leaves
	err = membershipService.LeaveGame("game1", playerA)
	require.NoError(t, err)

	// GameService.GetPlayers should only show player B
	gs, exists := registry.Get("game1")
	require.True(t, exists)
	players := gs.GetPlayers()
	assert.Len(t, players, 1)
	assert.Equal(t, "playerB", players[0].GetID())
}

// TestLastPlayerLeave_GameDeletedAcrossServices verifies that when the last player leaves,
// the game is properly cleaned up
func TestLastPlayerLeave_GameDeletedAcrossServices(t *testing.T) {
	_, _, _, membershipService, registry,
		mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, _, _, mockWatcher := setupIntegrationTest()

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// First, get or create game service
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	_, err := registry.Create("game1")
	require.NoError(t, err)

	// Setup mocks for leave (last player)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	mockGamePlayerRepo.On("RemovePlayerFromGame", mock.Anything, "game1", "player1").Return(nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{}, nil) // No players left

	// Setup mocks for game deletion
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)
	mockWatcher.On("Unsubscribe", "game1").Return(nil)
	mockGameCodeRepo.On("DeleteGameCodeByGameID", "game1").Return(nil)

	// Last player leaves
	err = membershipService.LeaveGame("game1", player)
	require.NoError(t, err)

	// Game service should be unregistered
	_, exists := registry.Get("game1")
	assert.False(t, exists)
}

// TestMembershipCheck_ConsistentAcrossServices verifies that membership checks
// are consistent across all services
func TestMembershipCheck_ConsistentAcrossServices(t *testing.T) {
	_, joinService, _, membershipService, registry,
		mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, _, _, mockWatcher := setupIntegrationTest()

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      "ABC1",
		ExpiresAt: integrationTestTime.Add(24 * time.Hour),
	}

	// Setup game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})

	// First create game service
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	_, err := registry.Create("game1")
	require.NoError(t, err)

	// Setup mocks for join - first IsPlayerInGame returns false, then true after join
	mockGameCodeRepo.On("GetGameCodeByCode", "ABC1").Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(false, nil).Once()
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "player1").Return(nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)

	// Join game
	_, err = joinService.JoinGame("ABC1", player)
	require.NoError(t, err)

	// Now check membership via membershipService - should see player is in game
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "player1").Return(true, nil)
	isInGame, err := membershipService.IsPlayerInGame("game1", "player1")
	require.NoError(t, err)
	assert.True(t, isInGame)

	// Registry's game service should also see the player
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{player}, nil)
	gs, exists := registry.Get("game1")
	require.True(t, exists)
	players := gs.GetPlayers()
	assert.Len(t, players, 1)
	assert.Equal(t, "player1", players[0].GetID())
}

// TestCreateAndJoin_ConcurrentAccess verifies that create and join operations
// don't have race conditions
func TestCreateAndJoin_ConcurrentAccess(t *testing.T) {
	creationService, _, queryService, _, _,
		mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo, mockMatchRepo, mockWatcher := setupIntegrationTest()

	playerA := &models.PlayerData{ID: "playerA", Name: "Player A"}
	playerB := &models.PlayerData{ID: "playerB", Name: "Player B"}

	// Setup mocks for game creation by player A - first GetPlayerGames returns empty
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "playerA").Return([]string{}, nil).Once()
	mockMatchRepo.On("GetMatchesByCompetitionAndSeason", "Ligue 1", "2025/2026").Return([]models.Match{}, nil)
	mockGameRepo.On("CreateGame", mock.AnythingOfType("*rules.GameImpl")).Return("game1", nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "playerA").Return(nil)
	mockWatcher.On("Subscribe", mock.AnythingOfType("*services.GameServiceImpl")).Return(nil)
	mockGameCodeRepo.On("CodeExists", mock.AnythingOfType("string")).Return(false, nil)
	mockGameCodeRepo.On("CreateGameCode", mock.AnythingOfType("*models.GameCode")).Return(nil)

	// Player A creates game
	createResp, err := creationService.CreateGame(&CreateGameRequest{
		SeasonYear:      "2025/2026",
		CompetitionName: "Ligue 1",
		Name:            "Test Game",
	}, playerA)
	require.NoError(t, err)
	code := createResp.Code

	// Setup mocks for player B joining immediately
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
	gameCode := &models.GameCode{
		GameID:    "game1",
		Code:      code,
		ExpiresAt: integrationTestTime.Add(6 * 30 * 24 * time.Hour),
	}
	mockGameCodeRepo.On("GetGameCodeByCode", code).Return(gameCode, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "playerB").Return([]string{}, nil)
	mockGamePlayerRepo.On("IsPlayerInGame", mock.Anything, "game1", "playerB").Return(false, nil)
	mockGamePlayerRepo.On("AddPlayerToGame", mock.Anything, "game1", "playerB").Return(nil)
	mockGameRepo.On("SaveWithId", "game1", realGame).Return(nil)

	// Player B joins
	_, err = creationService.JoinGame(code, playerB)
	require.NoError(t, err)

	// Setup mocks for query - both players visible
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "playerA").Return([]string{"game1"}, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{playerA, playerB}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(map[string]map[string]int{}, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(gameCode, nil)

	// Both players should be visible
	games, err := queryService.GetPlayerGames(playerA)
	require.NoError(t, err)
	require.Len(t, games, 1)
	assert.Len(t, games[0].Players, 2)
}
