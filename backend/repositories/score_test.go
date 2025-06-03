package repositories

import (
	"testing"
)

func TestInMemoryScoreRepository_SaveScore(t *testing.T) {
	repo := NewInMemoryScoreRepository()
	gameId := "game1"
	betId := "bet1"
	points := 3

	err := repo.SaveScore(gameId, betId, points)
	if err != nil {
		t.Errorf("SaveScore failed: %v", err)
	}

	// Verify the score was saved
	if got, err := repo.GetScore(gameId, betId); err != nil || got != points {
		t.Errorf("GetScore = %d, %v; want %d, nil", got, err, points)
	}
}

func TestInMemoryScoreRepository_GetScore_NonExistent(t *testing.T) {
	repo := NewInMemoryScoreRepository()
	gameId := "game1"
	betId := "non-existent-bet"

	// Should return 0 for non-existent bet
	if got, err := repo.GetScore(gameId, betId); err != nil || got != 0 {
		t.Errorf("GetScore = %d, %v; want 0, nil", got, err)
	}
}

func TestInMemoryScoreRepository_GetScore_DifferentGame(t *testing.T) {
	repo := NewInMemoryScoreRepository()
	gameId1 := "game1"
	gameId2 := "game2"
	betId := "bet1"
	points := 3

	// Save score for game1
	err := repo.SaveScore(gameId1, betId, points)
	if err != nil {
		t.Errorf("SaveScore failed: %v", err)
	}

	// Should return 0 for same betId but different game
	if got, err := repo.GetScore(gameId2, betId); err != nil || got != 0 {
		t.Errorf("GetScore = %d, %v; want 0, nil", got, err)
	}
}

func TestInMemoryScoreRepository_GetScoresByGame_Empty(t *testing.T) {
	repo := NewInMemoryScoreRepository()
	gameId := "game1"

	// Should return empty map for non-existent game
	scores, err := repo.GetScores(gameId)
	if err != nil {
		t.Errorf("GetScores failed: %v", err)
	}
	if len(scores) != 0 {
		t.Errorf("GetScores returned %d scores; want 0", len(scores))
	}
}

func TestInMemoryScoreRepository_GetScoresByGame_MultipleScores(t *testing.T) {
	repo := NewInMemoryScoreRepository()
	gameId := "game1"
	scores := map[string]int{
		"bet1": 3,
		"bet2": 1,
		"bet3": 0,
	}

	// Save multiple scores
	for betId, points := range scores {
		err := repo.SaveScore(gameId, betId, points)
		if err != nil {
			t.Errorf("SaveScore failed for %s: %v", betId, err)
		}
	}

	// Get all scores for the game
	got, err := repo.GetScores(gameId)
	if err != nil {
		t.Errorf("GetScores failed: %v", err)
	}

	// Verify all scores are present
	for betId, want := range scores {
		if got[betId] != want {
			t.Errorf("GetScores[%s] = %d; want %d", betId, got[betId], want)
		}
	}
}

func TestInMemoryScoreRepository_GetScoresByGame_IsolatedGames(t *testing.T) {
	repo := NewInMemoryScoreRepository()
	gameId1 := "game1"
	gameId2 := "game2"

	// Save scores for different games
	err := repo.SaveScore(gameId1, "bet1", 3)
	if err != nil {
		t.Errorf("SaveScore failed for game1: %v", err)
	}
	err = repo.SaveScore(gameId2, "bet2", 1)
	if err != nil {
		t.Errorf("SaveScore failed for game2: %v", err)
	}

	// Get scores for game1
	scores1, err := repo.GetScores(gameId1)
	if err != nil {
		t.Errorf("GetScores failed for game1: %v", err)
	}
	if len(scores1) != 1 || scores1["bet1"] != 3 {
		t.Errorf("GetScores for game1 = %v; want map[bet1:3]", scores1)
	}

	// Get scores for game2
	scores2, err := repo.GetScores(gameId2)
	if err != nil {
		t.Errorf("GetScores failed for game2: %v", err)
	}
	if len(scores2) != 1 || scores2["bet2"] != 1 {
		t.Errorf("GetScores for game2 = %v; want map[bet2:1]", scores2)
	}
}
