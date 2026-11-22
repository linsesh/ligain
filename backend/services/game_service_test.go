package services

import (
	"context"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/rules"
	"ligain/backend/testutils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var matchTime = time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

// Mock implementations
type gameRepositoryMock struct {
	mock.Mock
}

func (r *gameRepositoryMock) CreateGame(game models.Game) (string, error) {
	args := r.Called(game)
	return args.String(0), args.Error(1)
}

func (r *gameRepositoryMock) GetGame(gameId string) (models.Game, error) {
	args := r.Called(gameId)
	return args.Get(0).(models.Game), args.Error(1)
}

func (r *gameRepositoryMock) SaveWithId(gameId string, game models.Game) error {
	args := r.Called(gameId, game)
	return args.Error(0)
}

func (r *gameRepositoryMock) GetAllGames() (map[string]models.Game, error) {
	args := r.Called()
	return args.Get(0).(map[string]models.Game), args.Error(1)
}

type scorerMock struct{}

func (s *scorerMock) Score(match models.Match, bets []*models.Bet) []int {
	scores := make([]int, len(bets))
	for i, bet := range bets {
		if bet == nil {
			scores[i] = -100
			continue
		}
		// Create a new bet with the same predictions but the finished match
		finishedBet := models.NewBet(match, bet.PredictedHomeGoals, bet.PredictedAwayGoals)
		if finishedBet.IsBetCorrect(match) {
			scores[i] = 500
		} else {
			scores[i] = 0
		}
	}
	return scores
}

func (s *scorerMock) ScoreWithBreakdown(match models.Match, bets []*models.Bet) []rules.ScoreBreakdown {
	scores := s.Score(match, bets)
	breakdowns := make([]rules.ScoreBreakdown, len(scores))
	for i, score := range scores {
		breakdowns[i] = rules.ScoreBreakdown{BaseScore: score, RiskMultiplier: 1.0, ClairvoyantMultiplier: 1.0, Total: score}
	}
	return breakdowns
}

// testPlayer is a concrete implementation of models.Player for testing
type testPlayer struct {
	id   string
	name string
}

func (p *testPlayer) GetID() string   { return p.id }
func (p *testPlayer) GetName() string { return p.name }

// newTestPlayer creates a new test player
func newTestPlayer(name string) models.Player {
	return &testPlayer{
		id:   name, // Use name as ID for simplicity in tests
		name: name,
	}
}

func newTestSeasonMatchWithOdds(homeTeam, awayTeam string, date time.Time, matchday int) *models.SeasonMatch {
	return models.NewSeasonMatchWithKnownOdds(homeTeam, awayTeam, "2024", "Premier League", date, matchday, 1.0, 2.0, 3.0)
}

func setupTestGameService() (*GameServiceImpl, models.Match, []models.Player) {
	match := newTestSeasonMatchWithOdds("Team1", "Team2", matchTime, 1)
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
	matches := []models.Match{match}

	game := rules.NewFreshGame("2024", "Premier League", "Test Game", players, matches, &scorerMock{}, testutils.NewInMemoryMatchReaderOf(matches))
	gameRepo := &gameRepositoryMock{}
	gameRepo.On("GetGame", "test-game").Return(game, nil)
	gameRepo.On("SaveWithId", "test-game", mock.AnythingOfType("*rules.GameImpl")).Return(nil)
	betRepo := repositories.NewInMemoryBetRepository()

	// Create player repository and add players to it
	playerRepo := repositories.NewInMemoryPlayerRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(playerRepo)

	// Add players to the repositories - convert testPlayer to PlayerData
	// Note: We need to set the ID first, then use CreatePlayer which will override it,
	// so instead we'll directly add to the repository
	for _, player := range players {
		playerData := &models.PlayerData{
			ID:   player.GetID(),
			Name: player.GetName(),
		}
		// Directly save to avoid UUID generation in CreatePlayer
		playerRepo.UpdatePlayer(context.Background(), playerData)
		gamePlayerRepo.AddPlayerToGame(context.Background(), "test-game", player.GetID())
	}

	matchRepo := repositories.NewInMemoryMatchRepository()
	_ = matchRepo.SaveMatch(match)
	service := NewGameService("test-game", gameRepo, betRepo, gamePlayerRepo, NewMatchService(matchRepo))
	return service, match, players
}

func TestGameService_GetMatchResult(t *testing.T) {
	service, _, players := setupTestGameService()
	player1 := players[0]
	t.Run("returns nil for non-existent match", func(t *testing.T) {
		matchResults := service.GetMatchResults()
		require.NotNil(t, matchResults)
		result := matchResults["non-existent"]
		assert.Nil(t, result)
	})

	t.Run("returns empty map for player with no incoming matches", func(t *testing.T) {
		incomingMatches := service.GetIncomingMatches(player1)
		require.NotNil(t, incomingMatches)
		assert.Len(t, incomingMatches, 1) // Should have the test match
	})
}

func TestGameService_UpdatePlayerBet(t *testing.T) {
	service, match, players := setupTestGameService()
	player1 := players[0]

	t.Run("successfully updates player bet", func(t *testing.T) {
		bet := models.NewBet(match, 2, 1)
		err := service.UpdatePlayerBet(player1, bet, matchTime.Add(-1*time.Hour))
		require.NoError(t, err)

		// Verify bet was saved
		bets, err := service.GetPlayerBets(player1)
		require.NoError(t, err)
		require.Len(t, bets, 1)
		assert.Equal(t, 2, bets[0].PredictedHomeGoals)
		assert.Equal(t, 1, bets[0].PredictedAwayGoals)
	})

	t.Run("fails when betting after match time", func(t *testing.T) {
		bet := models.NewBet(match, 1, 1)
		err := service.UpdatePlayerBet(player1, bet, matchTime.Add(1*time.Hour))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "too late to bet")
	})
}

func TestGameService_GetGameID(t *testing.T) {
	service, _, _ := setupTestGameService()

	gameID := service.GetGameID()
	assert.Equal(t, "test-game", gameID)
}

func TestGameService_GetPlayers(t *testing.T) {
	service, _, _ := setupTestGameService()

	servicePlayers := service.GetPlayers()
	require.Len(t, servicePlayers, 2)
	assert.Equal(t, "Player1", servicePlayers[0].GetName())
	assert.Equal(t, "Player2", servicePlayers[1].GetName())
}

func TestGameService_OnMatchesUpdated_IgnoresMatchesNotInIncomingList(t *testing.T) {
	// Setup: game with one incoming match
	incomingMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)
	otherMatch := models.NewSeasonMatchWithKnownOdds("Team3", "Team4", "2024", "PL", matchTime, 2, 1.0, 2.0, 3.0)

	player := &models.PlayerData{ID: "player1", Name: "Player 1"}
	game := rules.NewFreshGame("2024", "PL", "Test Game", []models.Player{player}, []models.Match{incomingMatch}, &scorerMock{}, testutils.NewInMemoryMatchReaderOf([]models.Match{incomingMatch}))
	game.AddPlayerBet(player, models.NewBet(incomingMatch, 1, 1))

	gameRepo := &gameRepositoryMock{}
	gameRepo.On("GetGame", "game1").Return(game, nil)
	gameRepo.On("SaveWithId", "game1", game).Return(nil)
	betRepo := repositories.NewInMemoryBetRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(repositories.NewInMemoryPlayerRepository())
	matchRepo := repositories.NewInMemoryMatchRepository()
	matchRepo.SaveMatch(incomingMatch)
	matchRepo.SaveMatch(otherMatch)
	matchService := NewMatchService(matchRepo)

	gameService := NewGameService("game1", gameRepo, betRepo, gamePlayerRepo, matchService)

	// Finish the other match (not in this game)
	finishedOther := models.NewFinishedSeasonMatch("Team3", "Team4", 2, 1, "2024", "PL", matchTime, 2, 1.0, 2.0, 3.0)
	matchRepo.SaveMatch(finishedOther)

	// Apply update with the other match's ID
	err := gameService.OnMatchesUpdated([]string{otherMatch.Id()})

	// Should succeed but not score anything (match not in incoming list)
	assert.NoError(t, err)

	// Verify game was not saved (no scoring occurred)
	gameRepo.AssertNotCalled(t, "SaveWithId")
}

// TestGameService_OnMatchesUpdated_IsIdempotentForPastMatches tests that OnMatchesUpdated
// is idempotent for a match already in past (no rescoring)
func TestGameService_OnMatchesUpdated_IsIdempotentForPastMatches(t *testing.T) {
	// Setup: game with two matches, one already in past
	incomingMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)
	pastMatch := models.NewFinishedSeasonMatch("Team3", "Team4", 1, 0, "2024", "PL", matchTime.Add(1*time.Hour), 2, 1.0, 2.0, 3.0)

	player := &models.PlayerData{ID: "player1", Name: "Player 1"}
	game := rules.NewFreshGame("2024", "PL", "Test Game", []models.Player{player}, []models.Match{incomingMatch, pastMatch}, &scorerMock{}, testutils.NewInMemoryMatchReaderOf([]models.Match{incomingMatch, pastMatch}))
	game.AddPlayerBet(player, models.NewBet(incomingMatch, 2, 1))
	game.AddPlayerBet(player, models.NewBet(pastMatch, 1, 0))

	// Score the past match (this moves it to past list in the game)
	game.ApplyMatchScores(pastMatch, map[string]int{player.ID: 500})

	gameRepo := &gameRepositoryMock{}
	gameRepo.On("GetGame", "game1").Return(game, nil)
	betRepo := repositories.NewInMemoryBetRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(repositories.NewInMemoryPlayerRepository())
	matchRepo := repositories.NewInMemoryMatchRepository()
	matchRepo.SaveMatch(incomingMatch)
	matchRepo.SaveMatch(pastMatch)
	matchService := NewMatchService(matchRepo)

	gameService := NewGameService("game1", gameRepo, betRepo, gamePlayerRepo, matchService)

	// Apply update with the already-past match ID - should be ignored
	err := gameService.OnMatchesUpdated([]string{pastMatch.Id()})

	// Should succeed but not rescore
	assert.NoError(t, err)

	// Verify game was not saved (no scoring occurred)
	gameRepo.AssertNotCalled(t, "SaveWithId")
}

// TestGameService_OnMatchesUpdated_ContinuesAfterMatchServiceError tests that one match's
// MatchService.GetMatch error doesn't stop others from being processed
func TestGameService_OnMatchesUpdated_ContinuesAfterMatchServiceError(t *testing.T) {
	// Setup: game with two incoming matches
	match1 := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)
	match2 := models.NewSeasonMatchWithKnownOdds("Team3", "Team4", "2024", "PL", matchTime.Add(1*time.Hour), 2, 1.0, 2.0, 3.0)

	player := &models.PlayerData{ID: "player1", Name: "Player 1"}
	game := rules.NewFreshGame("2024", "PL", "Test Game", []models.Player{player}, []models.Match{match1, match2}, &scorerMock{}, testutils.NewInMemoryMatchReaderOf([]models.Match{match1, match2}))
	game.AddPlayerBet(player, models.NewBet(match1, 1, 1))
	game.AddPlayerBet(player, models.NewBet(match2, 0, 2))

	gameRepo := &gameRepositoryMock{}
	gameRepo.On("GetGame", "game1").Return(game, nil)
	gameRepo.On("SaveWithId", "game1", game).Return(nil)
	betRepo := repositories.NewInMemoryBetRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(repositories.NewInMemoryPlayerRepository())
	matchRepo := repositories.NewInMemoryMatchRepository()
	// Don't save match1, only match2
	matchService := NewMatchService(matchRepo)

	gameService := NewGameService("game1", gameRepo, betRepo, gamePlayerRepo, matchService)

	// Finish match2
	finishedMatch2 := models.NewFinishedSeasonMatch("Team3", "Team4", 0, 2, "2024", "PL", matchTime.Add(1*time.Hour), 2, 1.0, 2.0, 3.0)
	matchRepo.SaveMatch(finishedMatch2)

	// Apply update with both match IDs - match1 will fail to load (not in repo), match2 should still be scored
	err := gameService.OnMatchesUpdated([]string{match1.Id(), match2.Id()})

	// Should return error (from match1) but match2 should have been scored anyway
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get match")

	// Verify game was saved (match2 was scored and game is now finished)
	gameRepo.AssertCalled(t, "SaveWithId", "game1", mock.AnythingOfType("*rules.GameImpl"))
}

// TestGameService_OnMatchesUpdated_SavesFinishedGameToDatabase tests that finishing game is persisted
func TestGameService_OnMatchesUpdated_SavesFinishedGameToDatabase(t *testing.T) {
	// Setup: game with one incoming match
	incomingMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)

	player := &models.PlayerData{ID: "player1", Name: "Player 1"}
	game := rules.NewFreshGame("2024", "PL", "Test Game", []models.Player{player}, []models.Match{incomingMatch}, &scorerMock{}, testutils.NewInMemoryMatchReaderOf([]models.Match{incomingMatch}))
	game.AddPlayerBet(player, models.NewBet(incomingMatch, 2, 1))

	gameRepo := &gameRepositoryMock{}
	gameRepo.On("GetGame", "game1").Return(game, nil)
	gameRepo.On("SaveWithId", "game1", game).Return(nil)
	betRepo := repositories.NewInMemoryBetRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(repositories.NewInMemoryPlayerRepository())
	matchRepo := repositories.NewInMemoryMatchRepository()
	matchRepo.SaveMatch(incomingMatch)
	matchService := NewMatchService(matchRepo)

	gameService := NewGameService("game1", gameRepo, betRepo, gamePlayerRepo, matchService)

	// Finish the match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)
	matchRepo.SaveMatch(finishedMatch)

	// Apply update
	err := gameService.OnMatchesUpdated([]string{incomingMatch.Id()})

	// Should succeed and game should be finished and saved
	assert.NoError(t, err)
	gameRepo.AssertCalled(t, "SaveWithId", "game1", mock.AnythingOfType("*rules.GameImpl"))
}

// TestGameService_OnMatchesUpdated_NonDrawResultScoresAllPlayers tests that non-draw finished match scores all players
func TestGameService_OnMatchesUpdated_NonDrawResultScoresAllPlayers(t *testing.T) {
	// Setup: game with two players and one match with non-draw result
	incomingMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)

	player1 := &models.PlayerData{ID: "p1", Name: "Player 1"}
	player2 := &models.PlayerData{ID: "p2", Name: "Player 2"}
	game := rules.NewFreshGame("2024", "PL", "Test Game", []models.Player{player1, player2}, []models.Match{incomingMatch}, &scorerMock{}, testutils.NewInMemoryMatchReaderOf([]models.Match{incomingMatch}))

	// Both players bet on Team1 to win (2-1)
	game.AddPlayerBet(player1, models.NewBet(incomingMatch, 2, 1))
	game.AddPlayerBet(player2, models.NewBet(incomingMatch, 2, 1))

	gameRepo := &gameRepositoryMock{}
	gameRepo.On("GetGame", "game1").Return(game, nil)
	gameRepo.On("SaveWithId", "game1", game).Return(nil)
	betRepo := repositories.NewInMemoryBetRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(repositories.NewInMemoryPlayerRepository())
	matchRepo := repositories.NewInMemoryMatchRepository()
	matchRepo.SaveMatch(incomingMatch)
	matchService := NewMatchService(matchRepo)

	gameService := NewGameService("game1", gameRepo, betRepo, gamePlayerRepo, matchService)

	// Finish the match with 2-1 (matches both players' predictions)
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)
	matchRepo.SaveMatch(finishedMatch)

	// Apply update
	err := gameService.OnMatchesUpdated([]string{incomingMatch.Id()})

	// Should succeed
	assert.NoError(t, err)

	// Verify both players got points
	points := game.GetPlayersPoints()
	assert.Equal(t, 500, points[player1.ID])
	assert.Equal(t, 500, points[player2.ID])
}

// TestGameService_OnMatchesUpdated_SaveFinishedGameError tests that SaveWithId error is returned
func TestGameService_OnMatchesUpdated_SaveFinishedGameError(t *testing.T) {
	// Setup: game with one match
	incomingMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)

	player := &models.PlayerData{ID: "player1", Name: "Player 1"}
	game := rules.NewFreshGame("2024", "PL", "Test Game", []models.Player{player}, []models.Match{incomingMatch}, &scorerMock{}, testutils.NewInMemoryMatchReaderOf([]models.Match{incomingMatch}))
	game.AddPlayerBet(player, models.NewBet(incomingMatch, 2, 1))

	gameRepo := &gameRepositoryMock{}
	gameRepo.On("GetGame", "game1").Return(game, nil)
	gameRepo.On("SaveWithId", "game1", game).Return(fmt.Errorf("database error"))
	betRepo := repositories.NewInMemoryBetRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(repositories.NewInMemoryPlayerRepository())
	matchRepo := repositories.NewInMemoryMatchRepository()
	matchRepo.SaveMatch(incomingMatch)
	matchService := NewMatchService(matchRepo)

	gameService := NewGameService("game1", gameRepo, betRepo, gamePlayerRepo, matchService)

	// Finish the match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "PL", matchTime, 1, 1.0, 2.0, 3.0)
	matchRepo.SaveMatch(finishedMatch)

	// Apply update
	err := gameService.OnMatchesUpdated([]string{incomingMatch.Id()})

	// Should return the save error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// failingMatchService makes reading one match fail, like a repository error would
type failingMatchService struct {
	MatchService
	failID string
}

func (m *failingMatchService) GetMatch(id string) (models.Match, error) {
	if id == m.failID {
		return nil, fmt.Errorf("cannot read match %s", id)
	}
	return m.MatchService.GetMatch(id)
}

// finishedMatchFixture is a game of the given players over the given matches, with a real MatchService
// over an in-memory match repository. The game service is NOT subscribed: tests deliver the event by calling
// OnMatchesUpdated, which is what the MatchService does for a subscribed game.
type finishedMatchFixture struct {
	gameService  *GameServiceImpl
	matchService *MatchServiceImpl
	gameRepo     repositories.GameRepository
}

func newFinishedMatchFixture(t *testing.T, scorer rules.Scorer, players []models.Player, matches ...models.Match) *finishedMatchFixture {
	game := rules.NewFreshGame("2024", "Premier League", "Test Game", players, matches, scorer, testutils.NewInMemoryMatchReaderOf(matches))
	gameRepo := repositories.NewInMemoryGameRepository()
	require.NoError(t, gameRepo.SaveWithId("test-game", game))
	matchRepo := repositories.NewInMemoryMatchRepository()
	for _, m := range matches {
		require.NoError(t, matchRepo.SaveMatch(m))
	}
	matchService := NewMatchService(matchRepo)
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(repositories.NewInMemoryPlayerRepository())
	return &finishedMatchFixture{
		gameService:  NewGameService("test-game", gameRepo, repositories.NewInMemoryBetRepository(), gamePlayerRepo, matchService),
		matchService: matchService,
		gameRepo:     gameRepo,
	}
}

func (f *finishedMatchFixture) provider(t *testing.T, finished ...models.Match) {
	updates := make(map[string]models.Match)
	for _, m := range finished {
		updates[m.Id()] = m
	}
	require.NoError(t, f.matchService.TryUpdateMatches(models.MatchUpdatesOf(updates)))
}

func TestGameService_FinishedMatchEvent_FinishesAndSavesTheGame(t *testing.T) {
	match := newTestSeasonMatchWithOdds("Team1", "Team2", matchTime, 1)
	player1, player2 := newTestPlayer("Player1"), newTestPlayer("Player2")
	f := newFinishedMatchFixture(t, &scorerMock{}, []models.Player{player1, player2}, match)
	require.NoError(t, f.gameService.UpdatePlayerBet(player1, models.NewBet(match, 2, 1), matchTime))
	require.NoError(t, f.gameService.UpdatePlayerBet(player2, models.NewBet(match, 1, 2), matchTime))

	f.provider(t, models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0))
	require.NoError(t, f.gameService.OnMatchesUpdated([]string{match.Id()}))

	saved, err := f.gameRepo.GetGame("test-game")
	require.NoError(t, err)
	assert.True(t, saved.IsFinished())
}

func TestGameService_FinishedMatchEvent_ReportsTheErrorWhenTheGameCannotBeSaved(t *testing.T) {
	match := newTestSeasonMatchWithOdds("Team1", "Team2", matchTime, 1)
	player := newTestPlayer("Player1")
	game := rules.NewFreshGame("2024", "Premier League", "Test Game", []models.Player{player}, []models.Match{match}, &scorerMock{}, testutils.NewInMemoryMatchReaderOf([]models.Match{match}))
	gameRepo := &gameRepositoryMock{}
	gameRepo.On("GetGame", "test-game").Return(game, nil)
	gameRepo.On("SaveWithId", "test-game", mock.AnythingOfType("*rules.GameImpl")).Return(nil)
	matchRepo := repositories.NewInMemoryMatchRepository()
	require.NoError(t, matchRepo.SaveMatch(match))
	matchService := NewMatchService(matchRepo)
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(repositories.NewInMemoryPlayerRepository())
	gameService := NewGameService("test-game", gameRepo, repositories.NewInMemoryBetRepository(), gamePlayerRepo, matchService)
	require.NoError(t, gameService.UpdatePlayerBet(player, models.NewBet(match, 2, 1), matchTime))
	require.NoError(t, matchService.TryUpdateMatches(models.MatchUpdatesOf(map[string]models.Match{
		match.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0),
	})))

	gameRepo.ExpectedCalls = nil
	gameRepo.On("GetGame", "test-game").Return(game, nil)
	gameRepo.On("SaveWithId", "test-game", mock.AnythingOfType("*rules.GameImpl")).Return(fmt.Errorf("database error"))
	err := gameService.OnMatchesUpdated([]string{match.Id()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestGameService_FinishedMatchEvent_ReadFailureOnOneMatchDoesNotStopTheOthers(t *testing.T) {
	match1 := newTestSeasonMatchWithOdds("Team1", "Team2", matchTime, 1)
	match2 := newTestSeasonMatchWithOdds("Team3", "Team4", matchTime.Add(time.Hour), 2)
	player := newTestPlayer("Player1")
	f := newFinishedMatchFixture(t, &rules.ScorerOriginal{}, []models.Player{player}, match1, match2)
	require.NoError(t, f.gameService.UpdatePlayerBet(player, models.NewBet(match1, 2, 1), matchTime))
	require.NoError(t, f.gameService.UpdatePlayerBet(player, models.NewBet(match2, 0, 2), matchTime))
	f.provider(t,
		models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0),
		models.NewFinishedSeasonMatch("Team3", "Team4", 0, 2, "2024", "Premier League", matchTime.Add(time.Hour), 2, 1.0, 2.0, 3.0),
	)
	f.gameService.matchService = &failingMatchService{MatchService: f.matchService, failID: match1.Id()}

	err := f.gameService.OnMatchesUpdated([]string{match1.Id(), match2.Id()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), match1.Id())
	results := f.gameService.GetMatchResults()
	assert.Contains(t, results, match2.Id(), "the readable match is still scored and moved to past")
	assert.NotContains(t, results, match1.Id())
}

func TestGameService_FinishedMatchEvent_NonDrawResultScoresEveryCorrectBet(t *testing.T) {
	match := newTestSeasonMatchWithOdds("Team1", "Team2", matchTime, 1)
	winner1, winner2, loser := newTestPlayer("Winner1"), newTestPlayer("Winner2"), newTestPlayer("Loser")
	f := newFinishedMatchFixture(t, &rules.ScorerOriginal{}, []models.Player{winner1, winner2, loser}, match)
	require.NoError(t, f.gameService.UpdatePlayerBet(winner1, models.NewBet(match, 2, 1), matchTime))
	require.NoError(t, f.gameService.UpdatePlayerBet(winner2, models.NewBet(match, 3, 0), matchTime))
	require.NoError(t, f.gameService.UpdatePlayerBet(loser, models.NewBet(match, 1, 1), matchTime))

	f.provider(t, models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0))
	require.NoError(t, f.gameService.OnMatchesUpdated([]string{match.Id()}))

	points := f.gameService.GetMatchResults()[match.Id()].Scores
	assert.Greater(t, points[winner1.GetID()], 0)
	assert.Greater(t, points[winner2.GetID()], 0)
	assert.Equal(t, 0, points[loser.GetID()])
}

// TestGameService_PlayerChangesDisplayNameMidGame tests that a player can change their display name
// in the middle of a game and continue playing, placing bets, and winning
