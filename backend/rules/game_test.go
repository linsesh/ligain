package rules

import (
	"liguain/backend/models"
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

// Reference time for all tests
var testTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

func TestGameBasicProperties(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{
		models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1),
	}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	// Test basic properties
	if game.GetSeasonYear() != "2024" {
		t.Errorf("Expected season year '2024', got %s", game.GetSeasonYear())
	}
	if game.GetCompetitionName() != "Premier League" {
		t.Errorf("Expected competition name 'Premier League', got %s", game.GetCompetitionName())
	}
	if game.GetGameStatus() != models.GameStatusNotStarted {
		t.Errorf("Expected game status 'not started', got %s", game.GetGameStatus())
	}
}

func TestCheckPlayerBetValidity(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

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
			player:        models.Player{Name: "NonExistentPlayer"},
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

func TestAddPlayerBetAndGetIncomingMatches(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	// Add a valid bet
	bet := models.NewBet(match, 2, 1)
	err := game.AddPlayerBet(players[0], bet)
	if err != nil {
		t.Errorf("Expected no error when adding valid bet, got %v", err)
	}

	// Verify bet is in incoming matches
	incomingMatches := game.GetIncomingMatches()
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

	if matchResult.Bets[players[0]] != bet {
		t.Errorf("Expected bet %v for player, got %v", bet, matchResult.Bets[players[0]])
	}
}

func TestCalculateAndApplyMatchScores(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

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
	if scores[players[0]] != 500 {
		t.Errorf("Expected 500 points for correct bet, got %d", scores[players[0]])
	}
	if scores[players[1]] != 0 {
		t.Errorf("Expected 0 points for wrong bet, got %d", scores[players[1]])
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
	if totalPoints[players[0]] != 500 {
		t.Errorf("Expected 500 total points for Player1, got %d", totalPoints[players[0]])
	}
	if totalPoints[players[1]] != 0 {
		t.Errorf("Expected 0 total points for Player2, got %d", totalPoints[players[1]])
	}
}

func TestGameFinishAndWinner(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

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
	players := []models.Player{{Name: "Player1"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	// Update match with new time
	updatedMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
	err := game.UpdateMatch(updatedMatch)
	if err != nil {
		t.Errorf("Expected no error updating match, got %v", err)
	}

	// Verify update in incoming matches
	incomingMatches := game.GetIncomingMatches()
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
