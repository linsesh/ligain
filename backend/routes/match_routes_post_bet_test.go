package routes

import (
	"bytes"
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

const postBetGameID = "post-bet-game-id"

// postBetFixture is a game that has bets holding only match IDs
type postBetFixture struct {
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

func newPostBetFixture(t *testing.T) *postBetFixture {
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
		require.NoError(t, gamePlayerRepo.AddPlayerToGame(ctx, postBetGameID, p.GetID()))
	}

	// Create matches and add to matchRepo
	date := time.Now().Add(2 * time.Hour)
	incomingMatch := models.NewSeasonMatchWithKnownOdds("A", "B", "2026/2027", "Ligue 1", date, 11, 2.0, 3.0, 3.5)
	require.NoError(t, matchRepo.SaveMatch(incomingMatch))

	// Create MatchService first (before creating the game)
	matchService := services.NewMatchService(matchRepo)

	// Create game with incoming match, injecting MatchService as the match reader
	game := rules.NewFreshGame("2026/2027", "Ligue 1", "Test Game", players, []models.Match{incomingMatch}, &rules.ScorerOriginal{}, matchService)

	require.NoError(t, gameRepo.SaveWithId(postBetGameID, game))

	// Create GameServiceRegistry with MatchService
	registry, err := services.NewGameServiceRegistry(gameRepo, betRepo, gamePlayerRepo, matchService)
	require.NoError(t, err)

	return &postBetFixture{
		ctx:                 ctx,
		player1:             player1,
		player2:             player2,
		gameID:              postBetGameID,
		playerRepo:          playerRepo,
		gamePlayerRepo:      gamePlayerRepo,
		gameRepo:            gameRepo,
		betRepo:             betRepo,
		matchRepo:           matchRepo,
		matchService:        matchService,
		gameServiceRegistry: registry,
	}
}

func (f *postBetFixture) router(t *testing.T) *gin.Engine {
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

// TestPostBet_SavesBetWithOnlyMatchID tests that a POST bet saves a bet with only the match ID
// (not an embedded Match object)
func TestPostBet_SavesBetWithOnlyMatchID(t *testing.T) {
	f := newPostBetFixture(t)
	router := f.router(t)

	// Get the initial match to find a match ID
	getReq := httptest.NewRequest("GET", "/api/game/"+f.gameID+"/matches", nil)
	getReq.Header.Set("Authorization", "Bearer token-"+f.player1.ID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, getReq)
	require.Equal(t, http.StatusOK, w.Code)

	var initialResponse map[string]map[string]map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &initialResponse))
	require.Len(t, initialResponse["incomingMatches"], 1)

	var matchID string
	for id := range initialResponse["incomingMatches"] {
		matchID = id
		break
	}
	require.NotEmpty(t, matchID)

	// Now POST a bet
	betPayload := map[string]any{
		"matchId":            matchID,
		"predictedHomeGoals": 2,
		"predictedAwayGoals": 1,
	}
	betBody, _ := json.Marshal(betPayload)

	betReq := httptest.NewRequest("POST", "/api/game/"+f.gameID+"/bet", bytes.NewBuffer(betBody))
	betReq.Header.Set("Content-Type", "application/json")
	betReq.Header.Set("Authorization", "Bearer token-"+f.player1.ID)
	w2 := httptest.NewRecorder()

	router.ServeHTTP(w2, betReq)

	// This will fail on first run as expected (RED test)
	// We expect the bet response to only have matchId, predictedHomeGoals, predictedAwayGoals
	// and NOT contain any Match object
	require.Equal(t, http.StatusOK, w2.Code, "POST bet should succeed, got: %s", w2.Body.String())

	var betResponse map[string]any
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &betResponse), "Response body: %s", w2.Body.String())

	t.Logf("Response: %+v", betResponse)
	t.Logf("Bet data type: %T", betResponse["bet"])
	t.Logf("Bet data value: %+v", betResponse["bet"])

	betData, ok := betResponse["bet"].(map[string]any)
	require.True(t, ok, "bet response should be a map, but is %T", betResponse["bet"])

	assert.Equal(t, matchID, betData["matchId"])
	assert.Equal(t, float64(2), betData["predictedHomeGoals"])
	assert.Equal(t, float64(1), betData["predictedAwayGoals"])
	// The bet should NOT contain a Match object
	assert.NotContains(t, betData, "match", "bet response should not contain a Match object")
	assert.NotContains(t, betData, "homeTeam", "bet response should not contain team names")
}

// TestPostBet_ThenGetMatches_BetIsReturned tests that after POSTing a bet, GET /matches returns it
func TestPostBet_ThenGetMatches_BetIsReturned(t *testing.T) {
	f := newPostBetFixture(t)
	router := f.router(t)

	// Get the initial match to find a match ID
	getReq := httptest.NewRequest("GET", "/api/game/"+f.gameID+"/matches", nil)
	getReq.Header.Set("Authorization", "Bearer token-"+f.player1.ID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, getReq)
	require.Equal(t, http.StatusOK, w.Code)

	var initialResponse map[string]map[string]map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &initialResponse))

	var matchID string
	for id := range initialResponse["incomingMatches"] {
		matchID = id
		break
	}

	// POST a bet
	betPayload := map[string]any{
		"matchId":            matchID,
		"predictedHomeGoals": 2,
		"predictedAwayGoals": 1,
	}
	betBody, _ := json.Marshal(betPayload)

	betReq := httptest.NewRequest("POST", "/api/game/"+f.gameID+"/bet", bytes.NewBuffer(betBody))
	betReq.Header.Set("Content-Type", "application/json")
	betReq.Header.Set("Authorization", "Bearer token-"+f.player1.ID)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, betReq)
	require.Equal(t, http.StatusOK, w2.Code)

	// Now GET /matches and verify the bet is there
	getReq2 := httptest.NewRequest("GET", "/api/game/"+f.gameID+"/matches", nil)
	getReq2.Header.Set("Authorization", "Bearer token-"+f.player1.ID)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, getReq2)
	require.Equal(t, http.StatusOK, w3.Code)

	var matchesResponse map[string]map[string]map[string]any
	require.NoError(t, json.Unmarshal(w3.Body.Bytes(), &matchesResponse))

	matchResult := matchesResponse["incomingMatches"][matchID]
	bets := matchResult["bets"].(map[string]any)
	assert.Len(t, bets, 1, "player1 should have a bet on this match")

	betFromMatches := bets[f.player1.ID].(map[string]any)
	assert.Equal(t, float64(2), betFromMatches["predictedHomeGoals"])
	assert.Equal(t, float64(1), betFromMatches["predictedAwayGoals"])
}

// TestBet_MatchFinished_BetIsScored tests that after a match finishes through MatchService,
// the bet is scored against the current match (not a stale copy in the bet)
func TestBet_MatchFinished_BetIsScored(t *testing.T) {
	f := newPostBetFixture(t)
	router := f.router(t)

	// Get the initial match
	getReq := httptest.NewRequest("GET", "/api/game/"+f.gameID+"/matches", nil)
	getReq.Header.Set("Authorization", "Bearer token-"+f.player1.ID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, getReq)
	require.Equal(t, http.StatusOK, w.Code)

	var initialResponse map[string]map[string]map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &initialResponse))

	var matchID string
	for id := range initialResponse["incomingMatches"] {
		matchID = id
		break
	}

	// POST a bet: predict away team wins 1-2
	betPayload := map[string]any{
		"matchId":            matchID,
		"predictedHomeGoals": 1,
		"predictedAwayGoals": 2,
	}
	betBody, _ := json.Marshal(betPayload)

	betReq := httptest.NewRequest("POST", "/api/game/"+f.gameID+"/bet", bytes.NewBuffer(betBody))
	betReq.Header.Set("Content-Type", "application/json")
	betReq.Header.Set("Authorization", "Bearer token-"+f.player1.ID)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, betReq)
	require.Equal(t, http.StatusOK, w2.Code)

	// Now finish the match with result 1-2 (matches the bet prediction)
	finishedMatch := models.NewFinishedSeasonMatch("A", "B", 1, 2, "2026/2027", "Ligue 1", time.Now().Add(2*time.Hour), 11, 2.0, 3.0, 3.5)
	require.NoError(t, f.matchService.TryUpdateMatches(models.MatchUpdatesOf(map[string]models.Match{matchID: finishedMatch})))

	// Wait for scoring to complete
	var afterFinishResponse map[string]map[string]map[string]any
	require.Eventually(t, func() bool {
		getReq2 := httptest.NewRequest("GET", "/api/game/"+f.gameID+"/matches", nil)
		getReq2.Header.Set("Authorization", "Bearer token-"+f.player1.ID)
		w3 := httptest.NewRecorder()
		router.ServeHTTP(w3, getReq2)
		if w3.Code != http.StatusOK {
			return false
		}
		afterFinishResponse = nil // Unmarshal into an existing map would keep the entries of the previous attempts
		if err := json.Unmarshal(w3.Body.Bytes(), &afterFinishResponse); err != nil {
			return false
		}
		// The response is built in two steps (incoming matches, then past matches): wait for the settled state
		if _, inIncoming := afterFinishResponse["incomingMatches"][matchID]; inIncoming {
			return false
		}
		// Check if match has moved to past and has scores
		if pastMatch, exists := afterFinishResponse["pastMatches"][matchID]; exists {
			if scores, ok := pastMatch["scores"].(map[string]any); ok {
				return len(scores) > 0
			}
		}
		return false
	}, 2*time.Second, 50*time.Millisecond, "match should be scored and moved to pastMatches")

	// Verify the bet was scored correctly
	pastMatch := afterFinishResponse["pastMatches"][matchID]
	scores := pastMatch["scores"].(map[string]any)
	// Both players should be scored (player1 bet correctly, player2 didn't place a bet)
	assert.True(t, len(scores) >= 1, "at least player1 should be scored")

	player1Score := scores[f.player1.ID].(map[string]any)
	assert.Greater(t, int(player1Score["points"].(float64)), 0, "player1 should have positive points for correct bet")
}
