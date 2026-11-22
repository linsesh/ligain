package routes

import (
	"context"
	"encoding/json"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/rules"
	"ligain/backend/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const matchUpdatesGameID = "match-updates-game-id"

// matchUpdatesFixture is a game that uses MatchService for all match reads
type matchUpdatesFixture struct {
	ctx                 context.Context
	player1             *models.PlayerData
	player2             *models.PlayerData
	gameID              string
	playerRepo          repositories.PlayerRepository
	gamePlayerRepo      repositories.GamePlayerRepository
	gameRepo            repositories.GameRepository
	betRepo             repositories.BetRepository
	matchRepo           repositories.MatchRepository
	matchService        *services.MatchServiceImpl
	gameServiceRegistry *services.GameServiceRegistry
}

func newMatchUpdatesFixture(t *testing.T) *matchUpdatesFixture {
	ctx := context.Background()
	playerRepo := repositories.NewInMemoryPlayerRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(playerRepo)
	gameRepo := repositories.NewInMemoryGameRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	matchRepo := repositories.NewInMemoryMatchRepository()

	player1 := &models.PlayerData{Name: "Player1"}
	player2 := &models.PlayerData{Name: "Player2"}
	players := []models.Player{player1, player2}

	for _, p := range players {
		require.NoError(t, playerRepo.CreatePlayer(ctx, p.(*models.PlayerData)))
		require.NoError(t, gamePlayerRepo.AddPlayerToGame(ctx, matchUpdatesGameID, p.GetID()))
	}

	// Create matches and add to matchRepo
	date := time.Now().Add(2 * time.Hour)
	incomingMatch := models.NewSeasonMatchWithKnownOdds("A", "B", "2026/2027", "Ligue 1", date, 11, 2.0, 3.0, 3.5)
	require.NoError(t, matchRepo.SaveMatch(incomingMatch))

	// Create MatchService first (before creating the game)
	matchService := services.NewMatchService(matchRepo)

	// Create game with incoming match, injecting MatchService as the match reader
	game := rules.NewFreshGame("2026/2027", "Ligue 1", "Test Game", players, []models.Match{incomingMatch}, &rules.ScorerOriginal{}, matchService)

	require.NoError(t, gameRepo.SaveWithId(matchUpdatesGameID, game))

	// Create GameServiceRegistry with MatchService
	registry, err := services.NewGameServiceRegistry(gameRepo, betRepo, gamePlayerRepo, matchService)
	require.NoError(t, err)

	// Create watcher (stateless) - for now, just use the matchService directly
	// in production, the watcher would poll and call matchService.TryUpdateMatches

	return &matchUpdatesFixture{
		ctx:                 ctx,
		player1:             player1,
		player2:             player2,
		gameID:              matchUpdatesGameID,
		playerRepo:          playerRepo,
		gamePlayerRepo:      gamePlayerRepo,
		gameRepo:            gameRepo,
		betRepo:             betRepo,
		matchRepo:           matchRepo,
		matchService:        matchService,
		gameServiceRegistry: registry,
	}
}

func (f *matchUpdatesFixture) router(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	for _, p := range []*models.PlayerData{f.player1, f.player2} {
		require.NoError(t, f.playerRepo.CreateAuthToken(f.ctx, &models.AuthToken{
			PlayerID: p.ID, Token: "token-" + p.ID, ExpiresAt: time.Now().Add(time.Hour),
		}))
	}

	gameService, ok := f.gameServiceRegistry.Get(f.gameID)
	require.True(t, ok, "GameService should be found in registry")

	creationService := &MockGameCreationService{}
	creationService.On("GetGameService", f.gameID, mock.Anything).Return(gameService, nil)

	router := gin.New()
	NewMatchHandler(creationService, services.NewAuthService(f.playerRepo)).SetupRoutes(router)
	return router
}

func (f *matchUpdatesFixture) call(router *gin.Engine, method, path string, as *models.PlayerData) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer token-"+as.ID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestGetMatches_ReflectsAMatchUpdate tests that after MatchService
// receives a provider update, the GET /matches endpoint returns the new values without
// any game copy involved (the game reads through MatchService)
func TestGetMatches_ReflectsAMatchUpdate(t *testing.T) {
	f := newMatchUpdatesFixture(t)
	router := f.router(t)

	// Get initial matches
	w := f.call(router, "GET", "/api/game/"+f.gameID+"/matches", f.player1)
	require.Equal(t, http.StatusOK, w.Code)
	var initialResponse map[string]map[string]map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &initialResponse))
	require.Len(t, initialResponse["incomingMatches"], 1)

	// Get the match ID
	var matchID string
	for id := range initialResponse["incomingMatches"] {
		matchID = id
		break
	}

	initialMatch := initialResponse["incomingMatches"][matchID]
	initialOdds := initialMatch["match"].(map[string]any)["homeTeamOdds"].(float64)

	// Apply a provider update: change the odds
	// Create with same teams to get same ID
	updatedMatch := models.NewSeasonMatchWithKnownOdds("A", "B", "2026/2027", "Ligue 1", time.Now().Add(2*time.Hour), 11, 3.5, 4.0, 2.5)

	require.NoError(t, f.matchService.TryUpdateMatches(models.MatchUpdatesOf(map[string]models.Match{matchID: updatedMatch})))

	// Get matches again
	w = f.call(router, "GET", "/api/game/"+f.gameID+"/matches", f.player1)
	require.Equal(t, http.StatusOK, w.Code)
	var updatedResponse map[string]map[string]map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updatedResponse))

	// The odds should have changed
	newMatch := updatedResponse["incomingMatches"][matchID]
	newOdds := newMatch["match"].(map[string]any)["homeTeamOdds"].(float64)
	assert.NotEqual(t, initialOdds, newOdds)
	assert.Equal(t, 3.5, newOdds)
}

// TestGetMatches_FinishEventScoresPlayersAndMovesMatch tests that when a match finishes
// (via MatchService event), all players are scored and the match moves to pastMatches
func TestGetMatches_FinishEventScoresPlayersAndMovesMatch(t *testing.T) {
	f := newMatchUpdatesFixture(t)
	router := f.router(t)

	// Get initial matches
	w := f.call(router, "GET", "/api/game/"+f.gameID+"/matches", f.player1)
	require.Equal(t, http.StatusOK, w.Code)
	var initialResponse map[string]map[string]map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &initialResponse))

	var matchID string
	for id := range initialResponse["incomingMatches"] {
		matchID = id
		break
	}

	// Finish the match by applying a provider update
	// Create with same teams to get same ID
	finishedMatch := models.NewFinishedSeasonMatch("A", "B", 2, 1, "2026/2027", "Ligue 1", time.Now().Add(2*time.Hour), 11, 2.0, 3.0, 3.5)

	require.NoError(t, f.matchService.TryUpdateMatches(models.MatchUpdatesOf(map[string]models.Match{matchID: finishedMatch})))

	// Wait for async updates to be processed using polling
	var afterFinishResponse map[string]map[string]map[string]any
	require.Eventually(t, func() bool {
		w = f.call(router, "GET", "/api/game/"+f.gameID+"/matches", f.player1)
		if w.Code != http.StatusOK {
			return false
		}
		afterFinishResponse = nil // Unmarshal into an existing map would keep the entries of the previous attempts
		if err := json.Unmarshal(w.Body.Bytes(), &afterFinishResponse); err != nil {
			return false
		}
		// The response is built in two steps (incoming matches, then past matches): while the match is moving
		// it can be in both lists, or in neither. Wait for the settled state.
		_, inPast := afterFinishResponse["pastMatches"][matchID]
		_, inIncoming := afterFinishResponse["incomingMatches"][matchID]
		return inPast && !inIncoming
	}, 2*time.Second, 50*time.Millisecond, "match should move to pastMatches after finish")

	// Match should have moved to pastMatches
	assert.NotContains(t, afterFinishResponse["incomingMatches"], matchID)
	assert.Contains(t, afterFinishResponse["pastMatches"], matchID)

	// Match should have scores
	pastMatch := afterFinishResponse["pastMatches"][matchID]
	scores := pastMatch["scores"].(map[string]any)
	assert.Len(t, scores, 2, "both players should be scored")
}

// TestGetMatches_LoadedGameListsMatchesFromMatchService tests that a game loaded from
// the database lists its matches from MatchService (integration test)
func TestGetMatches_LoadedGameListsMatchesFromMatchService(t *testing.T) {
	f := newMatchUpdatesFixture(t)
	router := f.router(t)

	// Get matches - they should come from MatchService, not a game copy
	w := f.call(router, "GET", "/api/game/"+f.gameID+"/matches", f.player1)
	require.Equal(t, http.StatusOK, w.Code)
	var response map[string]map[string]map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	// Should have 1 incoming match
	assert.Len(t, response["incomingMatches"], 1)
	assert.Len(t, response["pastMatches"], 0)
}

// TestGetMatches_StoppedOddsAfterKickoffDontChangeReturned tests the matchday-11 case:
// stopped bookmaker odds reported after kickoff do not change the odds returned
func TestGetMatches_StoppedOddsAfterKickoffDontChangeReturned(t *testing.T) {
	f := newMatchUpdatesFixture(t)

	// Setup: create a match with original odds
	kickoffTime := time.Now()
	originalMatch := models.NewSeasonMatchWithKnownOdds("A", "B", "2026/2027", "Ligue 1", kickoffTime, 11, 2.0, 3.0, 3.5)
	require.NoError(t, f.matchRepo.SaveMatch(originalMatch))

	// Start the match (kickoff happened)
	originalMatch.Start()
	require.NoError(t, f.matchRepo.SaveMatch(originalMatch))

	// Apply a provider update with new odds after kickoff
	// (this simulates the matchday-11 bug where stopped odds were reported)
	stoppedOddsMatch := models.NewSeasonMatchWithKnownOdds("A", "B", "2026/2027", "Ligue 1", kickoffTime, 11, 1.02, 100.0, 1.02)
	stoppedOddsMatch.Start() // Mark as started (to simulate in-match situation)

	require.NoError(t, f.matchService.TryUpdateMatches(models.MatchUpdatesOf(map[string]models.Match{originalMatch.Id(): stoppedOddsMatch})))

	// Read the match from MatchService
	readMatch, err := f.matchService.GetMatch(originalMatch.Id())
	require.NoError(t, err)

	// The odds should remain the original ones (the freeze should have prevented the update)
	assert.Equal(t, 2.0, readMatch.GetHomeTeamOdds())
	assert.Equal(t, 3.0, readMatch.GetAwayTeamOdds())
	assert.Equal(t, 3.5, readMatch.GetDrawOdds())
}
