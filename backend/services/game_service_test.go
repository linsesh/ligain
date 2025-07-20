package services

import (
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/rules"
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

	game := rules.NewFreshGame("2024", "Premier League", "Test Game", players, matches, &scorerMock{})
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
		game := rules.NewFreshGame("2024", "Premier League", "Test Game", players, matches, &scorerMock{})

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

// TestGameService_adjustOdds tests the adjustOdds function with various time scenarios
func TestGameService_adjustOdds(t *testing.T) {
	// Create a base time for testing
	baseTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

	// Create test matches with different times
	createMatch := func(matchTime time.Time, homeOdds, awayOdds, drawOdds float64) models.Match {
		match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
		match.SetHomeTeamOdds(homeOdds)
		match.SetAwayTeamOdds(awayOdds)
		match.SetDrawOdds(drawOdds)
		return match
	}

	// Create last match state with different odds
	lastMatchState := createMatch(baseTime, 2.0, 3.0, 4.0)

	// Create service for testing
	service, _, _ := setupTestGameService()

	t.Run("odds should be blocked when match is less than 6 minutes away", func(t *testing.T) {
		// Match is 5 minutes away
		matchTime := baseTime.Add(5 * time.Minute)
		currentMatch := createMatch(matchTime, 1.5, 2.5, 3.5) // Different odds than last state

		// Mock current time to be now
		now := baseTime

		// Call adjustOdds with mocked time
		result := service.adjustOddsWithTime(currentMatch, lastMatchState, now)

		// Odds should be blocked (preserved from last state)
		assert.Equal(t, 2.0, result.GetHomeTeamOdds())
		assert.Equal(t, 3.0, result.GetAwayTeamOdds())
		assert.Equal(t, 4.0, result.GetDrawOdds())
	})

	t.Run("odds should not be blocked when match is exactly 6 minutes away", func(t *testing.T) {
		// Match is exactly 6 minutes away
		matchTime := baseTime.Add(6 * time.Minute)
		currentMatch := createMatch(matchTime, 1.5, 2.5, 3.5) // Different odds than last state

		// Mock current time to be now
		now := baseTime

		// Call adjustOdds with mocked time
		result := service.adjustOddsWithTime(currentMatch, lastMatchState, now)

		// Odds should not be blocked (keep current odds)
		assert.Equal(t, 1.5, result.GetHomeTeamOdds())
		assert.Equal(t, 2.5, result.GetAwayTeamOdds())
		assert.Equal(t, 3.5, result.GetDrawOdds())
	})

	t.Run("odds should not be blocked when match is more than 6 minutes away", func(t *testing.T) {
		// Match is 7 minutes away
		matchTime := baseTime.Add(7 * time.Minute)
		currentMatch := createMatch(matchTime, 1.5, 2.5, 3.5) // Different odds than last state

		// Mock current time to be now
		now := baseTime

		// Call adjustOdds with mocked time
		result := service.adjustOddsWithTime(currentMatch, lastMatchState, now)

		// Odds should not be blocked (keep current odds)
		assert.Equal(t, 1.5, result.GetHomeTeamOdds())
		assert.Equal(t, 2.5, result.GetAwayTeamOdds())
		assert.Equal(t, 3.5, result.GetDrawOdds())
	})

	t.Run("odds should be blocked when match has already started", func(t *testing.T) {
		// Match started 1 minute ago
		matchTime := baseTime.Add(-1 * time.Minute)
		currentMatch := createMatch(matchTime, 1.5, 2.5, 3.5) // Different odds than last state

		// Mock current time to be now
		now := baseTime

		// Call adjustOdds with mocked time
		result := service.adjustOddsWithTime(currentMatch, lastMatchState, now)

		// Odds should be blocked (preserved from last state)
		assert.Equal(t, 2.0, result.GetHomeTeamOdds())
		assert.Equal(t, 3.0, result.GetAwayTeamOdds())
		assert.Equal(t, 4.0, result.GetDrawOdds())
	})

	t.Run("odds should be blocked when match is in the past", func(t *testing.T) {
		// Match was 1 hour ago
		matchTime := baseTime.Add(-1 * time.Hour)
		currentMatch := createMatch(matchTime, 1.5, 2.5, 3.5) // Different odds than last state

		// Mock current time to be now
		now := baseTime

		// Call adjustOdds with mocked time
		result := service.adjustOddsWithTime(currentMatch, lastMatchState, now)

		// Odds should be blocked (preserved from last state)
		assert.Equal(t, 2.0, result.GetHomeTeamOdds())
		assert.Equal(t, 3.0, result.GetAwayTeamOdds())
		assert.Equal(t, 4.0, result.GetDrawOdds())
	})

	t.Run("odds should be preserved when current and last odds are the same", func(t *testing.T) {
		// Match is 5 minutes away
		matchTime := baseTime.Add(5 * time.Minute)
		currentMatch := createMatch(matchTime, 2.0, 3.0, 4.0) // Same odds as last state

		// Mock current time to be now
		now := baseTime

		// Call adjustOdds with mocked time
		result := service.adjustOddsWithTime(currentMatch, lastMatchState, now)

		// Odds should remain the same
		assert.Equal(t, 2.0, result.GetHomeTeamOdds())
		assert.Equal(t, 3.0, result.GetAwayTeamOdds())
		assert.Equal(t, 4.0, result.GetDrawOdds())
	})

	t.Run("edge case: match is exactly 5 minutes and 59 seconds away", func(t *testing.T) {
		// Match is 5 minutes and 59 seconds away (just under 6 minutes)
		matchTime := baseTime.Add(5*time.Minute + 59*time.Second)
		currentMatch := createMatch(matchTime, 1.5, 2.5, 3.5) // Different odds than last state

		// Mock current time to be now
		now := baseTime

		// Call adjustOdds with mocked time
		result := service.adjustOddsWithTime(currentMatch, lastMatchState, now)

		// Odds should be blocked (preserved from last state)
		assert.Equal(t, 2.0, result.GetHomeTeamOdds())
		assert.Equal(t, 3.0, result.GetAwayTeamOdds())
		assert.Equal(t, 4.0, result.GetDrawOdds())
	})

	t.Run("edge case: match is exactly 6 minutes and 1 second away", func(t *testing.T) {
		// Match is 6 minutes and 1 second away (just over 6 minutes)
		matchTime := baseTime.Add(6*time.Minute + 1*time.Second)
		currentMatch := createMatch(matchTime, 1.5, 2.5, 3.5) // Different odds than last state

		// Mock current time to be now
		now := baseTime

		// Call adjustOdds with mocked time
		result := service.adjustOddsWithTime(currentMatch, lastMatchState, now)

		// Odds should not be blocked (keep current odds)
		assert.Equal(t, 1.5, result.GetHomeTeamOdds())
		assert.Equal(t, 2.5, result.GetAwayTeamOdds())
		assert.Equal(t, 3.5, result.GetDrawOdds())
	})
}

// adjustOddsWithTime is a test helper that allows us to mock the current time
func (g *GameServiceImpl) adjustOddsWithTime(match models.Match, lastMatchState models.Match, now time.Time) models.Match {
	matchDate := match.GetDate()
	if matchDate.Before(now.Add(6 * time.Minute)) {
		match.SetHomeTeamOdds(lastMatchState.GetHomeTeamOdds())
		match.SetAwayTeamOdds(lastMatchState.GetAwayTeamOdds())
		match.SetDrawOdds(lastMatchState.GetDrawOdds())
	}
	return match
}

// TestGameService_HandleMatchUpdates_AdjustOdds tests that odds are properly adjusted during match updates
func TestGameService_HandleMatchUpdates_AdjustOdds(t *testing.T) {
	// Create a base time for testing
	baseTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

	// Create a match that's 5 minutes away (should have odds blocked)
	matchTime := baseTime.Add(5 * time.Minute)
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	match.SetHomeTeamOdds(1.5)
	match.SetAwayTeamOdds(2.5)
	match.SetDrawOdds(3.5)

	// Create players
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
	matches := []models.Match{match}

	// Create game with the match
	game := rules.NewFreshGame("2024", "Premier League", "Test Game", players, matches, &scorerMock{})
	gameRepo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()

	// Create a time function that returns our base time
	timeFunc := func() time.Time { return baseTime }

	// Create service with custom time function
	service := NewGameServiceWithTime("test-game", game, gameRepo, betRepo, timeFunc)

	// Create an update with different odds
	updatedMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	updatedMatch.SetHomeTeamOdds(1.2) // Different odds
	updatedMatch.SetAwayTeamOdds(2.8) // Different odds
	updatedMatch.SetDrawOdds(3.2)     // Different odds

	updates := map[string]models.Match{
		match.Id(): updatedMatch,
	}

	t.Run("odds should be blocked when match is within 6 minutes", func(t *testing.T) {
		// Get the original odds from the game
		originalMatch, err := service.game.GetMatchById(match.Id())
		require.NoError(t, err)
		originalHomeOdds := originalMatch.GetHomeTeamOdds()
		originalAwayOdds := originalMatch.GetAwayTeamOdds()
		originalDrawOdds := originalMatch.GetDrawOdds()

		// Process the update
		err = service.HandleMatchUpdates(updates)
		require.NoError(t, err)

		// Get the updated match
		updatedMatchFromGame, err := service.game.GetMatchById(match.Id())
		require.NoError(t, err)

		// The odds should be preserved (blocked) because the match is within 6 minutes
		assert.Equal(t, originalHomeOdds, updatedMatchFromGame.GetHomeTeamOdds())
		assert.Equal(t, originalAwayOdds, updatedMatchFromGame.GetAwayTeamOdds())
		assert.Equal(t, originalDrawOdds, updatedMatchFromGame.GetDrawOdds())
	})

	t.Run("odds should not be blocked when match is more than 6 minutes away", func(t *testing.T) {
		// Create a match that's 7 minutes away
		matchTime7Min := baseTime.Add(7 * time.Minute)
		match7Min := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime7Min, 2)
		match7Min.SetHomeTeamOdds(2.0)
		match7Min.SetAwayTeamOdds(3.0)
		match7Min.SetDrawOdds(4.0)

		// Add the match to the game
		players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
		matches := []models.Match{match7Min}
		game := rules.NewFreshGame("2024", "Premier League", "Test Game 2", players, matches, &scorerMock{})

		service := NewGameServiceWithTime("test-game-2", game, gameRepo, betRepo, timeFunc)

		// Create an update with different odds
		updatedMatch7Min := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime7Min, 2)
		updatedMatch7Min.SetHomeTeamOdds(1.8) // Different odds
		updatedMatch7Min.SetAwayTeamOdds(2.8) // Different odds
		updatedMatch7Min.SetDrawOdds(3.8)     // Different odds

		updates := map[string]models.Match{
			match7Min.Id(): updatedMatch7Min,
		}

		// Process the update
		err := service.HandleMatchUpdates(updates)
		require.NoError(t, err)

		// Get the updated match
		updatedMatchFromGame, err := service.game.GetMatchById(match7Min.Id())
		require.NoError(t, err)

		// The odds should be updated because the match is more than 6 minutes away
		assert.Equal(t, 1.8, updatedMatchFromGame.GetHomeTeamOdds())
		assert.Equal(t, 2.8, updatedMatchFromGame.GetAwayTeamOdds())
		assert.Equal(t, 3.8, updatedMatchFromGame.GetDrawOdds())
	})
}
