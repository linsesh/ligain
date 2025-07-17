package services

import (
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/rules"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var matchTime = time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

// Mock implementations
type gameRepositoryMock struct{}

func (r *gameRepositoryMock) CreateGame(game models.Game) (string, error) {
	return "test-game-id", nil
}

func (r *gameRepositoryMock) GetGame(gameId string) (models.Game, error) {
	return nil, nil // Not used in tests
}

func (r *gameRepositoryMock) SaveWithId(gameId string, game models.Game) error {
	return nil
}

type scorerMock struct{}

func (s *scorerMock) Score(match models.Match, bets []*models.Bet) []int {
	scores := make([]int, len(bets))
	for i, bet := range bets {
		// Create a new bet with the same predictions but the finished match
		finishedBet := models.NewBet(match, bet.PredictedHomeGoals, bet.PredictedAwayGoals)
		if finishedBet.IsBetCorrect() {
			scores[i] = 500
		} else {
			scores[i] = 0
		}
	}
	return scores
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

func setupTestGameService() (*GameServiceImpl, models.Match, []models.Player) {
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
	matches := []models.Match{match}

	game := rules.NewFreshGame("2024", "Premier League", players, matches, &scorerMock{})
	gameRepo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()

	service := NewGameService("test-game", game, gameRepo, betRepo)
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

func TestGameService_HandleMatchUpdates(t *testing.T) {
	service, match, players := setupTestGameService()
	player1 := players[0]

	// Add a bet first
	bet := models.NewBet(match, 2, 1)
	err := service.UpdatePlayerBet(player1, bet, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)

	t.Run("handles match updates correctly", func(t *testing.T) {
		// Create a finished match update
		finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)

		updates := map[string]models.Match{
			match.Id(): finishedMatch,
		}

		err := service.HandleMatchUpdates(updates)
		require.NoError(t, err)

		// Check that the game is now finished
		assert.True(t, service.game.IsFinished())

		// Check that the winner is correct
		winners := service.game.GetWinner()
		require.Len(t, winners, 1)
		assert.Equal(t, "Player1", winners[0].GetName())
	})

	t.Run("handles multiple match updates", func(t *testing.T) {
		// Create another match
		match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime.Add(1*time.Hour), 2)

		// Add the match to the game by creating a new game with both matches
		players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
		matches := []models.Match{match, match2}
		game := rules.NewFreshGame("2024", "Premier League", players, matches, &scorerMock{})

		// Create a new service with the updated game
		gameRepo := &gameRepositoryMock{}
		betRepo := repositories.NewInMemoryBetRepository()
		service := NewGameService("test-game", game, gameRepo, betRepo)

		// Add bet for second match
		bet2 := models.NewBet(match2, 1, 0)
		err := service.UpdatePlayerBet(player1, bet2, matchTime)
		require.NoError(t, err)

		// Create updates for both matches
		finishedMatch1 := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
		finishedMatch2 := models.NewFinishedSeasonMatch("Team3", "Team4", 1, 0, "2024", "Premier League", matchTime.Add(1*time.Hour), 2, 1.0, 2.0, 3.0)

		updates := map[string]models.Match{
			match.Id():  finishedMatch1,
			match2.Id(): finishedMatch2,
		}

		err = service.HandleMatchUpdates(updates)
		require.NoError(t, err)

		// Check that the game is now finished
		assert.True(t, service.game.IsFinished())
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
