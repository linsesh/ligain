package repositories

import (
	"liguain/backend/models"
	"liguain/backend/testutils"
	"testing"
	"time"
)

var testTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

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

func TestInMemoryBetRepository_SaveAndGetBets(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player := newTestPlayer("TestPlayer")
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	bet := models.NewBet(match, 2, 1)

	// Test saving a bet
	betId, savedBet, err := repo.SaveBet("test-id", bet, player)
	if err != nil {
		t.Errorf("Failed to save bet: %v", err)
	}
	if betId == "" {
		t.Error("Expected non-empty bet ID")
	}
	if savedBet == nil {
		t.Error("Expected non-nil saved bet")
	}

	// Test retrieving the bet
	bets, err := repo.GetBets("test-id", player)
	if err != nil {
		t.Errorf("Failed to get bets: %v", err)
	}
	if len(bets) != 1 {
		t.Errorf("Expected 1 bet, got %d", len(bets))
	}
	if bets[0].Match.Id() != match.Id() {
		t.Errorf("Expected bet for match %s, got %s", match.Id(), bets[0].Match.Id())
	}
}

func TestInMemoryBetRepository_UpdateBet(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player := newTestPlayer("TestPlayer")
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)

	// Save initial bet
	initialBet := models.NewBet(match, 2, 1)
	_, savedBet, err := repo.SaveBet("test-id", initialBet, player)
	if err != nil {
		t.Errorf("Failed to save initial bet: %v", err)
	}
	if savedBet == nil {
		t.Error("Expected non-nil saved bet")
	}

	// Update the bet
	updatedBet := models.NewBet(match, 3, 2)
	_, savedBet, err = repo.SaveBet("test-id", updatedBet, player)
	if err != nil {
		t.Errorf("Failed to update bet: %v", err)
	}
	if savedBet == nil {
		t.Error("Expected non-nil saved bet")
	}

	// Verify only one bet exists (the updated one)
	bets, err := repo.GetBets("test-id", player)
	if err != nil {
		t.Errorf("Failed to get bets: %v", err)
	}
	if len(bets) != 1 {
		t.Errorf("Expected 1 bet after update, got %d", len(bets))
	}
	if bets[0].PredictedHomeGoals != 3 || bets[0].PredictedAwayGoals != 2 {
		t.Errorf("Expected updated bet (3-2), got (%d-%d)",
			bets[0].PredictedHomeGoals, bets[0].PredictedAwayGoals)
	}
}

func TestInMemoryBetRepository_GetBetsForMatch(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player1 := newTestPlayer("Player1")
	player2 := newTestPlayer("Player2")
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)

	// Save bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	_, savedBet1, err := repo.SaveBet("test-id", bet1, player1)
	if err != nil {
		t.Errorf("Failed to save bet1: %v", err)
	}
	if savedBet1 == nil {
		t.Error("Expected non-nil saved bet1")
	}
	_, savedBet2, err := repo.SaveBet("test-id", bet2, player2)
	if err != nil {
		t.Errorf("Failed to save bet2: %v", err)
	}
	if savedBet2 == nil {
		t.Error("Expected non-nil saved bet2")
	}

	// Get all bets for the match
	bets, players, err := repo.GetBetsForMatch(match, "test-id")
	if err != nil {
		t.Errorf("Failed to get bets for match: %v", err)
	}
	if len(bets) != 2 {
		t.Errorf("Expected 2 bets, got %d", len(bets))
	}
	if len(players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(players))
	}

	// Verify the bets and players match
	playerBets := make(map[string]*models.Bet)
	for i, player := range players {
		playerBets[player.GetName()] = bets[i]
	}

	if bet, ok := playerBets["Player1"]; !ok || bet.PredictedHomeGoals != 2 || bet.PredictedAwayGoals != 1 {
		t.Error("Player1's bet not found or incorrect")
	}
	if bet, ok := playerBets["Player2"]; !ok || bet.PredictedHomeGoals != 1 || bet.PredictedAwayGoals != 1 {
		t.Error("Player2's bet not found or incorrect")
	}
}

func TestInMemoryBetRepository_EmptyResults(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player := newTestPlayer("TestPlayer")
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)

	// Test getting bets for non-existent game
	bets, err := repo.GetBets("test-id", player)
	if err != nil {
		t.Errorf("Failed to get bets: %v", err)
	}
	if len(bets) != 0 {
		t.Errorf("Expected 0 bets for non-existent game, got %d", len(bets))
	}

	// Test getting bets for non-existent match
	bets, players, err := repo.GetBetsForMatch(match, "test-id")
	if err != nil {
		t.Errorf("Failed to get bets for match: %v", err)
	}
	if len(bets) != 0 {
		t.Errorf("Expected 0 bets for non-existent match, got %d", len(bets))
	}
	if len(players) != 0 {
		t.Errorf("Expected 0 players for non-existent match, got %d", len(players))
	}
}

func TestInMemoryBetRepository_SaveAndGetScore(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player := newTestPlayer("TestPlayer")
	match := &testutils.MockMatch{} // Use the mock match from testutils
	bet := models.NewBet(match, 2, 1)

	// Save bet first
	betId, savedBet, err := repo.SaveBet("test-id", bet, player)
	if err != nil {
		t.Errorf("Failed to save bet: %v", err)
	}
	if savedBet == nil {
		t.Error("Expected non-nil saved bet")
	}

	// Test saving a score
	err = repo.SaveScore("test-id", match, player, 3)
	if err != nil {
		t.Errorf("Failed to save score: %v", err)
	}

	// Test retrieving the score
	score, err := repo.GetScore("test-id", betId)
	if err != nil {
		t.Errorf("Failed to get score: %v", err)
	}
	if score != 3 {
		t.Errorf("Expected score 3, got %d", score)
	}
}

func TestInMemoryBetRepository_GetScoreNotFound(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player := newTestPlayer("TestPlayer")
	match := &testutils.MockMatch{}
	bet := models.NewBet(match, 2, 1)

	// Save bet without score
	betId, savedBet, err := repo.SaveBet("test-id", bet, player)
	if err != nil {
		t.Errorf("Failed to save bet: %v", err)
	}
	if savedBet == nil {
		t.Error("Expected non-nil saved bet")
	}

	// Try to get non-existent score
	_, err = repo.GetScore("test-id", betId)
	if err != ErrScoreNotFound {
		t.Errorf("Expected ErrScoreNotFound, got %v", err)
	}
}

func TestInMemoryBetRepository_GetScores(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player1 := newTestPlayer("Player1")
	player2 := newTestPlayer("Player2")
	match := &testutils.MockMatch{}

	// Save bets and scores
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	betId1, savedBet1, err := repo.SaveBet("test-id", bet1, player1)
	if err != nil {
		t.Errorf("Failed to save bet1: %v", err)
	}
	if savedBet1 == nil {
		t.Error("Expected non-nil saved bet1")
	}
	betId2, savedBet2, err := repo.SaveBet("test-id", bet2, player2)
	if err != nil {
		t.Errorf("Failed to save bet2: %v", err)
	}
	if savedBet2 == nil {
		t.Error("Expected non-nil saved bet2")
	}

	// Save scores
	err = repo.SaveScore("test-id", match, player1, 3)
	if err != nil {
		t.Errorf("Failed to save score1: %v", err)
	}
	err = repo.SaveScore("test-id", match, player2, 1)
	if err != nil {
		t.Errorf("Failed to save score2: %v", err)
	}

	// Get all scores
	scores, err := repo.GetScores("test-id")
	if err != nil {
		t.Errorf("Failed to get scores: %v", err)
	}
	if len(scores) != 2 {
		t.Errorf("Expected 2 scores, got %d", len(scores))
	}
	if scores[betId1] != 3 {
		t.Errorf("Expected score 3 for bet1, got %d", scores[betId1])
	}
	if scores[betId2] != 1 {
		t.Errorf("Expected score 1 for bet2, got %d", scores[betId2])
	}
}

func TestInMemoryBetRepository_GetScoresByMatchAndPlayer(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player1 := newTestPlayer("Player1")
	player2 := newTestPlayer("Player2")
	match := &testutils.MockMatch{}

	// Save bets and scores
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	_, savedBet1, err := repo.SaveBet("test-id", bet1, player1)
	if err != nil {
		t.Errorf("Failed to save bet1: %v", err)
	}
	if savedBet1 == nil {
		t.Error("Expected non-nil saved bet1")
	}
	_, savedBet2, err := repo.SaveBet("test-id", bet2, player2)
	if err != nil {
		t.Errorf("Failed to save bet2: %v", err)
	}
	if savedBet2 == nil {
		t.Error("Expected non-nil saved bet2")
	}

	// Save scores
	err = repo.SaveScore("test-id", match, player1, 3)
	if err != nil {
		t.Errorf("Failed to save score1: %v", err)
	}
	err = repo.SaveScore("test-id", match, player2, 1)
	if err != nil {
		t.Errorf("Failed to save score2: %v", err)
	}

	// Get scores by match and player
	scores, err := repo.GetScoresByMatchAndPlayer("test-id")
	if err != nil {
		t.Errorf("Failed to get scores by match and player: %v", err)
	}
	if len(scores) != 1 {
		t.Errorf("Expected scores for 1 match, got %d", len(scores))
	}
	matchScores := scores[match.Id()]
	if len(matchScores) != 2 {
		t.Errorf("Expected scores for 2 players, got %d", len(matchScores))
	}
	if matchScores[player1.GetID()] != 3 {
		t.Errorf("Expected score 3 for player1, got %d", matchScores[player1.GetID()])
	}
	if matchScores[player2.GetID()] != 1 {
		t.Errorf("Expected score 1 for player2, got %d", matchScores[player2.GetID()])
	}
}
