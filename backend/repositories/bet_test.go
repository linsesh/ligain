package repositories

import (
	"liguain/backend/models"
	"testing"
	"time"
)

var testTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

func TestInMemoryBetRepository_SaveAndGetBets(t *testing.T) {
	repo := NewInMemoryBetRepository()
	player := models.Player{Name: "TestPlayer"}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	bet := models.NewBet(match, 2, 1)

	// Test saving a bet
	betId, err := repo.SaveBet("test-id", bet, player)
	if err != nil {
		t.Errorf("Failed to save bet: %v", err)
	}
	if betId == "" {
		t.Error("Expected non-empty bet ID")
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
	player := models.Player{Name: "TestPlayer"}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)

	// Save initial bet
	initialBet := models.NewBet(match, 2, 1)
	_, err := repo.SaveBet("test-id", initialBet, player)
	if err != nil {
		t.Errorf("Failed to save initial bet: %v", err)
	}

	// Update the bet
	updatedBet := models.NewBet(match, 3, 2)
	_, err = repo.SaveBet("test-id", updatedBet, player)
	if err != nil {
		t.Errorf("Failed to update bet: %v", err)
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
	player1 := models.Player{Name: "Player1"}
	player2 := models.Player{Name: "Player2"}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)

	// Save bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	_, err := repo.SaveBet("test-id", bet1, player1)
	if err != nil {
		t.Errorf("Failed to save bet1: %v", err)
	}
	_, err = repo.SaveBet("test-id", bet2, player2)
	if err != nil {
		t.Errorf("Failed to save bet2: %v", err)
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
		playerBets[player.Name] = bets[i]
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
	player := models.Player{Name: "TestPlayer"}
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
