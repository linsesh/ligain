package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"ligain/backend/middleware"
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
)

const removedPlayerGameID = "123e4567-e89b-12d3-a456-426614174000"

// removedPlayerFixture is a real (non-mock) game of 3 players in which everybody
// bet on an in-progress match and on a finished match of the same matchday.
type removedPlayerFixture struct {
	ctx            context.Context
	player1        *models.PlayerData
	leaver         *models.PlayerData
	inProgress     models.Match
	finished       models.Match
	playerRepo     repositories.PlayerRepository
	gamePlayerRepo repositories.GamePlayerRepository
	gameRepo       repositories.GameRepository
	betRepo        repositories.BetRepository
}

func newRemovedPlayerFixture(t *testing.T) *removedPlayerFixture {
	ctx := context.Background()
	playerRepo := repositories.NewInMemoryPlayerRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(playerRepo)

	player1 := &models.PlayerData{Name: "Player1"}
	player2 := &models.PlayerData{Name: "Player2"}
	leaver := &models.PlayerData{Name: "Leaver"}
	players := []models.Player{player1, player2, leaver}
	// CreatePlayer assigns the IDs, so persist players before using their IDs in the game
	for _, p := range players {
		assert.NoError(t, playerRepo.CreatePlayer(ctx, p.(*models.PlayerData)))
		assert.NoError(t, gamePlayerRepo.AddPlayerToGame(ctx, removedPlayerGameID, p.GetID()))
	}

	date := time.Now().Add(-2 * time.Hour)
	inProgress := models.NewSeasonMatchWithKnownOdds("A", "B", "2026/2027", "Ligue 1", date, 11, 2.0, 3.0, 3.5)
	inProgress.Start()
	finished := models.NewFinishedSeasonMatch("C", "D", 2, 1, "2026/2027", "Ligue 1", date, 11, 2.0, 3.0, 3.5)

	bets := map[string]map[string]*models.Bet{}
	for _, m := range []models.Match{inProgress, finished} {
		bets[m.Id()] = map[string]*models.Bet{}
		for _, p := range players {
			bets[m.Id()][p.GetID()] = models.NewBet(m, 2, 1)
		}
	}
	scores := map[string]map[string]int{finished.Id(): {}}
	for _, p := range players {
		scores[finished.Id()][p.GetID()] = 100
	}

	game := rules.NewStartedGame("2026/2027", "Ligue 1", "Test", players,
		[]models.Match{inProgress}, []models.Match{finished}, &rules.ScorerOriginal{}, bets, scores)
	gameRepo := repositories.NewInMemoryGameRepository()
	gameRepo.SaveWithId(removedPlayerGameID, game)

	return &removedPlayerFixture{
		ctx: ctx, player1: player1, leaver: leaver, inProgress: inProgress, finished: finished,
		playerRepo: playerRepo, gamePlayerRepo: gamePlayerRepo,
		gameRepo: gameRepo, betRepo: repositories.NewInMemoryBetRepository(),
	}
}

// router exposes the real auth + matches routes, wired like main.go.
func (f *removedPlayerFixture) router(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	registry, err := services.NewGameServiceRegistry(f.gameRepo, f.betRepo, f.gamePlayerRepo, nil)
	assert.NoError(t, err)
	membershipService := services.NewGameMembershipService(repositories.NewNoopUnitOfWork(), f.gamePlayerRepo, f.gameRepo, nil, registry, nil)
	authService := services.NewAuthService(f.playerRepo, services.WithPlayerGamesRemover(membershipService))
	for _, p := range []*models.PlayerData{f.player1, f.leaver} {
		assert.NoError(t, f.playerRepo.CreateAuthToken(f.ctx, &models.AuthToken{
			PlayerID: p.ID, Token: "token-" + p.ID, ExpiresAt: time.Now().Add(time.Hour),
		}))
	}
	gameService, _ := registry.Get(removedPlayerGameID)
	creationService := &MockGameCreationService{}
	creationService.On("GetGameService", removedPlayerGameID, mock.Anything).Return(gameService, nil)

	router := gin.New()
	NewMatchHandler(creationService, authService).SetupRoutes(router)
	router.DELETE("/api/auth/account", middleware.PlayerAuth(authService), NewAuthHandler(authService, nil).DeleteAccount)
	return router
}

func (f *removedPlayerFixture) call(router *gin.Engine, method, path string, as *models.PlayerData) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer token-"+as.ID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestGetMatches_AfterPlayerDeletedTheirAccount(t *testing.T) {
	f := newRemovedPlayerFixture(t)
	router := f.router(t)

	w := f.call(router, "DELETE", "/api/auth/account", f.leaver)
	assert.Equal(t, http.StatusOK, w.Code)

	w = f.call(router, "GET", "/api/game/"+removedPlayerGameID+"/matches", f.player1)
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]map[string]map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

	cases := map[string]map[string]any{
		"incoming in-progress match": response["incomingMatches"][f.inProgress.Id()],
		"past finished match":        response["pastMatches"][f.finished.Id()],
	}
	for name, match := range cases {
		t.Run(name, func(t *testing.T) {
			if !assert.NotNil(t, match) {
				return
			}
			assert.Len(t, match["bets"], 2, "only remaining players' bets expected")
			assert.NotContains(t, match["bets"], f.leaver.ID)
			if scores, ok := match["scores"].(map[string]any); ok {
				assert.NotContains(t, scores, f.leaver.ID)
			}
		})
	}
}

// foreignKeyBetRepo behaves like the database for scores: SaveScore fails for a
// player that does not exist anymore (score.player_id foreign key), or for the
// players listed in failFor.
type foreignKeyBetRepo struct {
	repositories.BetRepository
	playerRepo repositories.PlayerRepository
	failFor    map[string]bool
}

func (r *foreignKeyBetRepo) SaveScore(gameID string, match models.Match, player models.Player, points int) error {
	if _, err := r.playerRepo.GetPlayerByID(context.Background(), player.GetID()); err != nil || r.failFor[player.GetID()] {
		return fmt.Errorf("error saving score: violates foreign key constraint for player %s", player.GetID())
	}
	return r.BetRepository.SaveScore(gameID, match, player, points)
}

// finishInProgressMatch is what the match watcher does in production when the
// provider reports the final score: it hands the finished match to the game.
func (f *removedPlayerFixture) finishInProgressMatch(t *testing.T, router *gin.Engine) error {
	registry, err := services.NewGameServiceRegistry(f.gameRepo, f.betRepo, f.gamePlayerRepo, nil)
	assert.NoError(t, err)
	gs, ok := registry.Get(removedPlayerGameID)
	assert.True(t, ok)
	finishedNow := models.NewFinishedSeasonMatch("A", "B", 2, 1, "2026/2027", "Ligue 1", time.Now().Add(-2*time.Hour), 11, 2.0, 3.0, 3.5)
	return gs.HandleMatchUpdates(map[string]models.Match{finishedNow.Id(): finishedNow})
}

func (f *removedPlayerFixture) getPastMatch(t *testing.T, router *gin.Engine, matchID string) map[string]any {
	w := f.call(router, "GET", "/api/game/"+removedPlayerGameID+"/matches", f.player1)
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]map[string]map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
	assert.NotContains(t, response["incomingMatches"], matchID, "a finished match must not stay in incomingMatches")
	return response["pastMatches"][matchID]
}

// A player deletes their account, then a match finishes: the remaining players
// must still be scored and see the finished match with bets and scores.
func TestGetMatches_MatchFinishesAfterPlayerDeletedTheirAccount(t *testing.T) {
	f := newRemovedPlayerFixture(t)
	f.betRepo = &foreignKeyBetRepo{BetRepository: f.betRepo, playerRepo: f.playerRepo}
	router := f.router(t)

	assert.Equal(t, http.StatusOK, f.call(router, "DELETE", "/api/auth/account", f.leaver).Code)
	assert.NoError(t, f.finishInProgressMatch(t, router))

	match := f.getPastMatch(t, router, f.inProgress.Id())
	if assert.NotNil(t, match, "finished match must be in pastMatches") {
		assert.Len(t, match["bets"], 2)
		assert.Len(t, match["scores"], 2)
	}
}

// One player whose score cannot be saved must not prevent the others from being
// scored, nor keep the finished match in incomingMatches.
func TestGetMatches_MatchFinishesWhenOnePlayerScoreCannotBeSaved(t *testing.T) {
	f := newRemovedPlayerFixture(t)
	f.betRepo = &foreignKeyBetRepo{BetRepository: f.betRepo, playerRepo: f.playerRepo, failFor: map[string]bool{f.leaver.ID: true}}
	router := f.router(t)

	err := f.finishInProgressMatch(t, router)
	assert.Error(t, err, "the failure must still be reported")

	match := f.getPastMatch(t, router, f.inProgress.Id())
	if assert.NotNil(t, match, "finished match must be in pastMatches") {
		scores, _ := match["scores"].(map[string]any)
		assert.Contains(t, scores, f.player1.ID)
		assert.Len(t, scores, 3-1, "the 2 players whose score can be saved must be scored")
	}
}
