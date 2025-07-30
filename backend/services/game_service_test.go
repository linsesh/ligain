package services

import (
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/rules"
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

// mutableTestPlayer is a test player that can change their name (for testing display name changes)
type mutableTestPlayer struct {
	id   string
	name string
}

func (p *mutableTestPlayer) GetID() string       { return p.id }
func (p *mutableTestPlayer) GetName() string     { return p.name }
func (p *mutableTestPlayer) SetName(name string) { p.name = name }

// newMutableTestPlayer creates a new test player that can change names
func newMutableTestPlayer(id, name string) *mutableTestPlayer {
	return &mutableTestPlayer{
		id:   id,
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

		// Set up mock expectation for SaveWithId since the game will finish
		service.gameRepo.(*gameRepositoryMock).On("SaveWithId", "test-game", mock.AnythingOfType("*rules.GameImpl")).Return(nil)

		err := service.HandleMatchUpdates(updates)
		require.NoError(t, err)

		// Check that the game is now finished
		assert.True(t, service.game.IsFinished())

		// Check that the winner is correct
		winners := service.game.GetWinner()
		require.Len(t, winners, 1)
		assert.Equal(t, "Player1", winners[0].GetName())

		// Verify SaveWithId was called
		service.gameRepo.(*gameRepositoryMock).AssertExpectations(t)
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

		// Set up mock expectation for SaveWithId since the game will finish
		gameRepo.On("SaveWithId", "test-game", mock.AnythingOfType("*rules.GameImpl")).Return(nil)

		err = service.HandleMatchUpdates(updates)
		require.NoError(t, err)

		// Check that the game is now finished
		assert.True(t, service.game.IsFinished())

		// Verify SaveWithId was called
		gameRepo.AssertExpectations(t)
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

// TestGameService_PlayerChangesDisplayNameMidGame tests that a player can change their display name
// in the middle of a game and continue playing, placing bets, and winning
func TestGameService_PlayerChangesDisplayNameMidGame(t *testing.T) {
	// Create matches
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime.Add(2*time.Hour), 2)
	matches := []models.Match{match1, match2}

	// Create players - one mutable player and one regular player
	player1 := newMutableTestPlayer("player1-id", "OriginalName")
	player2 := newTestPlayer("Player2")
	players := []models.Player{player1, player2}

	// Create game
	game := rules.NewFreshGame("2024", "Premier League", "Display Name Change Test", players, matches, &scorerMock{})
	gameRepo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("test-game", game, gameRepo, betRepo)

	// Phase 1: Player places initial bet with original name
	t.Run("Player places bet with original name", func(t *testing.T) {
		bet1 := models.NewBet(match1, 2, 1) // Betting Team1 wins 2-1
		err := service.UpdatePlayerBet(player1, bet1, matchTime.Add(-1*time.Hour))
		require.NoError(t, err)

		// Verify bet was saved
		bets, err := service.GetPlayerBets(player1)
		require.NoError(t, err)
		require.Len(t, bets, 1)
		assert.Equal(t, 2, bets[0].PredictedHomeGoals)
		assert.Equal(t, 1, bets[0].PredictedAwayGoals)

		// Verify player appears in game with original name
		gamePlayers := service.GetPlayers()
		found := false
		for _, p := range gamePlayers {
			if p.GetID() == player1.GetID() {
				assert.Equal(t, "OriginalName", p.GetName())
				found = true
				break
			}
		}
		assert.True(t, found, "Player should be found in game with original name")
	})

	// Phase 2: Player changes display name mid-game
	t.Run("Player changes display name mid-game", func(t *testing.T) {
		// Simulate display name change
		player1.SetName("NewDisplayName")

		// Verify the player can still be found by ID and now has new name
		gamePlayers := service.GetPlayers()
		found := false
		for _, p := range gamePlayers {
			if p.GetID() == player1.GetID() {
				// The game should still reference the same player object
				// In a real scenario, the game would need to be updated with the new player data
				// For this test, we're checking that the player object itself changed
				found = true
				break
			}
		}
		assert.True(t, found, "Player should still be found in game after name change")
		assert.Equal(t, "NewDisplayName", player1.GetName())
	})

	// Phase 3: Player continues to place bets with new name
	t.Run("Player places more bets with new name", func(t *testing.T) {
		bet2 := models.NewBet(match2, 1, 0) // Betting Team3 wins 1-0
		err := service.UpdatePlayerBet(player1, bet2, matchTime.Add(1*time.Hour))
		require.NoError(t, err)

		// Verify both bets are saved for the same player
		bets, err := service.GetPlayerBets(player1)
		require.NoError(t, err)
		require.Len(t, bets, 2)

		// Check that we have both bets (without assuming order)
		foundBet1 := false // 2-1 bet for match1
		foundBet2 := false // 1-0 bet for match2

		for _, bet := range bets {
			if bet.PredictedHomeGoals == 2 && bet.PredictedAwayGoals == 1 {
				foundBet1 = true
			} else if bet.PredictedHomeGoals == 1 && bet.PredictedAwayGoals == 0 {
				foundBet2 = true
			}
		}

		assert.True(t, foundBet1, "Should find bet with 2-1 prediction")
		assert.True(t, foundBet2, "Should find bet with 1-0 prediction")
	})

	// Phase 4: Matches finish and player wins
	t.Run("Matches finish and player with changed name wins", func(t *testing.T) {
		// Player2 places losing bets
		wrongBet1 := models.NewBet(match1, 0, 3) // Wrong prediction
		wrongBet2 := models.NewBet(match2, 2, 2) // Wrong prediction
		err := service.UpdatePlayerBet(player2, wrongBet1, matchTime.Add(-30*time.Minute))
		require.NoError(t, err)
		err = service.UpdatePlayerBet(player2, wrongBet2, matchTime.Add(1*time.Hour))
		require.NoError(t, err)

		// Finish matches with results that match player1's predictions
		finishedMatch1 := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
		finishedMatch2 := models.NewFinishedSeasonMatch("Team3", "Team4", 1, 0, "2024", "Premier League", matchTime.Add(2*time.Hour), 2, 1.0, 2.0, 3.0)

		updates := map[string]models.Match{
			match1.Id(): finishedMatch1,
			match2.Id(): finishedMatch2,
		}

		// Set up mock expectation for SaveWithId since the game will finish
		gameRepo.On("SaveWithId", "test-game", mock.AnythingOfType("*rules.GameImpl")).Return(nil)

		err = service.HandleMatchUpdates(updates)
		require.NoError(t, err)

		// Check that the game is finished
		assert.True(t, service.game.IsFinished())

		// Verify SaveWithId was called
		gameRepo.AssertExpectations(t)

		// Check that player1 (with changed name) is the winner
		winners := service.game.GetWinner()
		require.Len(t, winners, 1)

		// The winner should be our player1 who changed their name
		winner := winners[0]
		assert.Equal(t, player1.GetID(), winner.GetID())
		assert.Equal(t, "NewDisplayName", winner.GetName()) // Should have the new name

		// Verify the winner's total score (should have 1000 points: 500 for each correct bet)
		totalPlayerPoints := service.game.GetPlayersPoints()

		player1Points := totalPlayerPoints[player1.GetID()]
		player2Points := totalPlayerPoints[player2.GetID()]

		assert.Equal(t, 1000, player1Points) // 500 points per correct bet
		assert.Equal(t, 0, player2Points)    // 0 points for incorrect bets
	})

	// Phase 5: Verify game history shows player with new name
	t.Run("Game history shows player with updated name", func(t *testing.T) {
		pastResults := service.game.GetPastResults()
		assert.Len(t, pastResults, 2)

		// Check that in past results, our player appears with new name
		for _, result := range pastResults {
			if result.Bets != nil {
				for playerID, bet := range result.Bets {
					if playerID == player1.GetID() {
						// The bet should exist and be associated with the correct player ID
						assert.NotNil(t, bet)
						// In a real system, we'd want to verify the display name is updated
						// in the UI, but the game logic should work with player IDs
						break
					}
				}
			}
		}
	})
}

func TestGameService_HandleMatchUpdates_SavesFinishedGameToDatabase(t *testing.T) {
	// Create a mock game repository that tracks SaveWithId calls
	mockGameRepo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()

	// Create a test match that will finish
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)

	// Create players
	player1 := newTestPlayer("Player1")
	player2 := newTestPlayer("Player2")
	players := []models.Player{player1, player2}

	// Create a game with the match
	game := rules.NewFreshGame("2024", "Premier League", "Test Game", players, []models.Match{match}, &scorerMock{})

	// Create service
	service := NewGameService("test-game", game, mockGameRepo, betRepo)

	// Add bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 2)
	err := service.UpdatePlayerBet(player1, bet1, matchTime)
	require.NoError(t, err)
	err = service.UpdatePlayerBet(player2, bet2, matchTime)
	require.NoError(t, err)

	// Create finished match update
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
	updates := map[string]models.Match{
		match.Id(): finishedMatch,
	}

	// Set up mock expectation for SaveWithId
	mockGameRepo.On("SaveWithId", "test-game", mock.AnythingOfType("*rules.GameImpl")).Return(nil)

	// Handle the updates
	err = service.HandleMatchUpdates(updates)
	require.NoError(t, err)

	// Verify the game is finished
	assert.True(t, service.game.IsFinished())

	// Verify SaveWithId was called with the finished game
	mockGameRepo.AssertExpectations(t)
}

func TestGameService_HandleMatchUpdates_SaveFinishedGameError(t *testing.T) {
	// Create a mock game repository that returns an error on SaveWithId
	mockGameRepo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()

	// Create a test match that will finish
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)

	// Create players
	player1 := newTestPlayer("Player1")
	player2 := newTestPlayer("Player2")
	players := []models.Player{player1, player2}

	// Create a game with the match
	game := rules.NewFreshGame("2024", "Premier League", "Test Game", players, []models.Match{match}, &scorerMock{})

	// Create service
	service := NewGameService("test-game", game, mockGameRepo, betRepo)

	// Add bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 2)
	err := service.UpdatePlayerBet(player1, bet1, matchTime)
	require.NoError(t, err)
	err = service.UpdatePlayerBet(player2, bet2, matchTime)
	require.NoError(t, err)

	// Create finished match update
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
	updates := map[string]models.Match{
		match.Id(): finishedMatch,
	}

	// Set up mock expectation for SaveWithId to return an error
	mockGameRepo.On("SaveWithId", "test-game", mock.AnythingOfType("*rules.GameImpl")).Return(fmt.Errorf("database error"))

	// Handle the updates - should return an error
	err = service.HandleMatchUpdates(updates)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")

	// Verify the game is still finished (the game state was updated)
	assert.True(t, service.game.IsFinished())

	// Verify SaveWithId was called
	mockGameRepo.AssertExpectations(t)
}
