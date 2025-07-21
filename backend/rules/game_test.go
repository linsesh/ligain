package rules

import (
	"ligain/backend/models"
	"testing"
	"time"
)

type ScorerTest struct{}

func (s *ScorerTest) Score(match models.Match, bets []*models.Bet) []int {
	scores := make([]int, len(bets))
	for i, bet := range bets {
		if bet.IsBetCorrect() {
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

// Reference time for all tests
var testTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

func TestGameBasicProperties(t *testing.T) {
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
	matches := []models.Match{
		models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1),
	}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", "Test Game", players, matches, scorer)

	// Test basic properties
	if game.GetSeasonYear() != "2024" {
		t.Errorf("Expected season year '2024', got %s", game.GetSeasonYear())
	}
	if game.GetCompetitionName() != "Premier League" {
		t.Errorf("Expected competition name 'Premier League', got %s", game.GetCompetitionName())
	}
	if game.GetName() != "Test Game" {
		t.Errorf("Expected game name 'Test Game', got %s", game.GetName())
	}
	if game.GetGameStatus() != models.GameStatusNotStarted {
		t.Errorf("Expected game status 'not started', got %s", game.GetGameStatus())
	}
}

func TestCheckPlayerBetValidity(t *testing.T) {
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", "Test Game", players, matches, scorer)

	tests := []struct {
		name          string
		player        models.Player
		bet           *models.Bet
		checkTime     time.Time
		expectedError bool
	}{
		{
			name:          "Valid bet before match",
			player:        players[0],
			bet:           models.NewBet(match, 2, 1),
			checkTime:     testTime.Add(-1 * time.Hour),
			expectedError: false,
		},
		{
			name:          "Invalid bet after match start",
			player:        players[0],
			bet:           models.NewBet(match, 2, 1),
			checkTime:     testTime.Add(1 * time.Hour),
			expectedError: true,
		},
		{
			name:          "Invalid player",
			player:        newTestPlayer("NonExistentPlayer"),
			bet:           models.NewBet(match, 2, 1),
			checkTime:     testTime.Add(-1 * time.Hour),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := game.CheckPlayerBetValidity(tt.player, tt.bet, tt.checkTime)
			if (err != nil) != tt.expectedError {
				t.Errorf("CheckPlayerBetValidity() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}

func TestCheckPlayerBetValidityWithMixedPlayerSources(t *testing.T) {
	// This test specifically verifies that players loaded from different sources
	// (bet table vs game_player table) are all properly validated

	// Create players
	player1 := newTestPlayer("Player1")     // Will have bets
	player2 := newTestPlayer("Player2")     // Will have bets
	player3 := newTestPlayer("Player3")     // Will be in game but no bets yet
	nonMember := newTestPlayer("NonMember") // Not in game at all

	players := []models.Player{player1, player2, player3}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", "Test Game", players, matches, scorer)

	// Add bets for Player1 and Player2 only
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)

	err := game.AddPlayerBet(player1, bet1)
	if err != nil {
		t.Fatalf("Failed to add bet for Player1: %v", err)
	}

	err = game.AddPlayerBet(player2, bet2)
	if err != nil {
		t.Fatalf("Failed to add bet for Player2: %v", err)
	}

	// Test cases
	tests := []struct {
		name          string
		player        models.Player
		expectedError bool
		description   string
	}{
		{
			name:          "Player with bet should be valid",
			player:        player1,
			expectedError: false,
			description:   "Player1 has made a bet and should be able to place another bet",
		},
		{
			name:          "Player with bet should be valid",
			player:        player2,
			expectedError: false,
			description:   "Player2 has made a bet and should be able to place another bet",
		},
		{
			name:          "Player without bet but in game should be valid",
			player:        player3,
			expectedError: false,
			description:   "Player3 is in the game but hasn't bet yet, should still be able to place a bet",
		},
		{
			name:          "Player not in game should be invalid",
			player:        nonMember,
			expectedError: true,
			description:   "NonMember is not in the game and should not be able to place a bet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new bet for testing
			testBet := models.NewBet(match, 0, 0)

			err := game.CheckPlayerBetValidity(tt.player, testBet, testTime.Add(-1*time.Hour))

			if (err != nil) != tt.expectedError {
				t.Errorf("%s: CheckPlayerBetValidity() error = %v, expectedError %v",
					tt.description, err, tt.expectedError)
			}

			if tt.expectedError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			}

			if !tt.expectedError && err != nil {
				t.Errorf("%s: Expected no error but got: %v", tt.description, err)
			}
		})
	}
}

func TestAddPlayerBetAndGetIncomingMatches(t *testing.T) {
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", "Test Game", players, matches, scorer)

	// Add a valid bet
	bet := models.NewBet(match, 2, 1)
	err := game.AddPlayerBet(players[0], bet)
	if err != nil {
		t.Errorf("Expected no error when adding valid bet, got %v", err)
	}

	// Verify bet is in incoming matches
	incomingMatches := game.GetIncomingMatches(players[0])
	if len(incomingMatches) != 1 {
		t.Errorf("Expected 1 incoming match, got %d", len(incomingMatches))
	}

	matchResult := incomingMatches[match.Id()]
	if matchResult == nil {
		t.Fatal("Expected match result not to be nil")
	}

	if len(matchResult.Bets) != 1 {
		t.Errorf("Expected 1 bet for match, got %d", len(matchResult.Bets))
	}

	if matchResult.Bets[players[0].GetID()] != bet {
		t.Errorf("Expected bet %v for player, got %v", bet, matchResult.Bets[players[0].GetID()])
	}
}

func TestCalculateAndApplyMatchScores(t *testing.T) {
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", "Test Game", players, matches, scorer)

	// Add bets for both players
	correctBet := models.NewBet(match, 2, 1)
	wrongBet := models.NewBet(match, 1, 1)

	game.AddPlayerBet(players[0], correctBet)
	game.AddPlayerBet(players[1], wrongBet)

	// Finish match with result matching Player1's bet
	match.Finish(2, 1)

	// Calculate scores
	scores, err := game.CalculateMatchScores(match)
	if err != nil {
		t.Errorf("Expected no error calculating scores, got %v", err)
	}

	// Verify scores
	if scores[players[0].GetID()] != 500 {
		t.Errorf("Expected 500 points for correct bet, got %d", scores[players[0].GetID()])
	}
	if scores[players[1].GetID()] != 0 {
		t.Errorf("Expected 0 points for wrong bet, got %d", scores[players[1].GetID()])
	}

	// Apply scores
	game.ApplyMatchScores(match, scores)

	// Verify match moved to past results
	pastResults := game.GetPastResults()
	if len(pastResults) != 1 {
		t.Errorf("Expected 1 past result, got %d", len(pastResults))
	}

	// Verify total points
	totalPoints := game.GetPlayersPoints()
	if totalPoints[players[0].GetID()] != 500 {
		t.Errorf("Expected 500 total points for Player1, got %d", totalPoints[players[0].GetID()])
	}
	if totalPoints[players[1].GetID()] != 0 {
		t.Errorf("Expected 0 total points for Player2, got %d", totalPoints[players[1].GetID()])
	}
}

func TestGameFinishAndWinner(t *testing.T) {
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2")}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", "Test Game", players, matches, scorer)

	// Add bets
	correctBet := models.NewBet(match, 2, 1)
	wrongBet := models.NewBet(match, 1, 1)

	game.AddPlayerBet(players[0], correctBet)
	game.AddPlayerBet(players[1], wrongBet)

	// Initially game should not be finished
	if game.IsFinished() {
		t.Error("Expected game to not be finished initially")
	}

	// Finish match and apply scores
	match.Finish(2, 1)
	scores, _ := game.CalculateMatchScores(match)
	game.ApplyMatchScores(match, scores)

	// Now game should be finished as all matches are done
	if !game.IsFinished() {
		t.Error("Expected game to be finished after all matches completed")
	}

	// Check winner
	winners := game.GetWinner()
	if len(winners) != 1 {
		t.Errorf("Expected 1 winner, got %d", len(winners))
	}
	if winners[0] != players[0] {
		t.Errorf("Expected Player1 to be winner, got %v", winners[0])
	}
}

func TestUpdateMatch(t *testing.T) {
	players := []models.Player{newTestPlayer("Player1")}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", "Test Game", players, matches, scorer)

	// Update match with new time
	updatedMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
	err := game.UpdateMatch(updatedMatch)
	if err != nil {
		t.Errorf("Expected no error updating match, got %v", err)
	}

	// Verify update in incoming matches
	incomingMatches := game.GetIncomingMatches(players[0])
	updatedResult := incomingMatches[match.Id()]
	if updatedResult.Match.GetDate() != testTime.Add(24*time.Hour) {
		t.Error("Match time was not updated correctly")
	}

	// Try to update non-existent match
	nonExistentMatch := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", testTime, 2)
	err = game.UpdateMatch(nonExistentMatch)
	if err == nil {
		t.Error("Expected error updating non-existent match, got nil")
	}
}

func TestGame_CalculateMatchScores_ForgottenBets(t *testing.T) {
	players := []models.Player{newTestPlayer("Player1"), newTestPlayer("Player2"), newTestPlayer("Player3")}
	match := models.NewSeasonMatch("TeamA", "TeamB", "2024", "Test League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerOriginal{}
	game := NewFreshGame("2024", "Test League", "Forgotten Game", players, matches, scorer)

	// Only Player1 bets
	bet := models.NewBet(match, 1, 0)
	_ = game.AddPlayerBet(players[0], bet)
	match.Finish(1, 0)

	scores, err := game.CalculateMatchScores(match)
	if err != nil {
		t.Fatalf("Expected no error calculating scores, got %v", err)
	}
	if scores[players[0].GetID()] != 550 {
		t.Errorf("Expected 550 for Player1, got %d", scores[players[0].GetID()])
	}
	if scores[players[1].GetID()] != -100 {
		t.Errorf("Expected -100 for Player2 (forgot bet), got %d", scores[players[1].GetID()])
	}
	if scores[players[2].GetID()] != -100 {
		t.Errorf("Expected -100 for Player3 (forgot bet), got %d", scores[players[2].GetID()])
	}

	// All forgot
	players2 := []models.Player{newTestPlayer("P1"), newTestPlayer("P2")}
	match2 := models.NewSeasonMatch("TeamC", "TeamD", "2024", "Test League", testTime, 2)
	matches2 := []models.Match{match2}
	game2 := NewFreshGame("2024", "Test League", "Forgotten Game 2", players2, matches2, scorer)
	match2.Finish(0, 0)

	scores2, err := game2.CalculateMatchScores(match2)
	if err != nil {
		t.Fatalf("Expected no error calculating scores, got %v", err)
	}
	for _, p := range players2 {
		if scores2[p.GetID()] != -100 {
			t.Errorf("Expected -100 for %s (forgot bet), got %d", p.GetID(), scores2[p.GetID()])
		}
	}
}
