package rules

import (
	"liguain/backend/models"
	"slices"
	"testing"
	"time"
)

func TestScorerOriginal_ScoreClassAllPerfect(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Manchester United", "Liverpool", 2, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 1,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{500, 500, 500, 500}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected all scores to be 500, got %v", scores)
	}
}

func TestScorerOriginal_ScoreClassHalfPerfectHalfWrong(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Manchester United", "Liverpool", 2, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 0,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{550, 550, 0, 0}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}

func TestScorerOriginal_ScoreClassOneQuarterPerfectOthersWrong(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Manchester United", "Liverpool", 2, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 0,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{625, 0, 0, 0}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}

func TestScorerOriginal_ScoreClassAllPerfectWithFavoriteBeaten(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Bastia", "Real Madrid", 2, 0, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 8.0, 1.1, 6.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 0,
		},
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 0,
		},
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 0,
		},
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 0,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{1000, 1000, 1000, 1000}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}

func TestScorerOriginal_ScoreClassOnePerfectOthersWrongWithFavoriteLosing(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Bastia", "Real Madrid", 2, 0, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 8.0, 1.1, 6.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 0,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 4,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 5,
		},
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 4,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{1250, 0, 0, 0}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}

func TestScorerOriginal_ScoreClassOneCloseScoretOthersWrongWithFavoriteDraw(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Bastia", "Real Madrid", 2, 2, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 8.0, 1.1, 6.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 4,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 5,
		},
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 4,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{750, 0, 0, 0}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}

func TestScorerOriginal_ScoreClassTwoPerfectOneCloseScoreOthersWrongWithFavoriteWinning(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Bastia", "Real Madrid", 1, 3, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 8.0, 1.1, 6.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 2,
			PredictedAwayGoals: 4,
		},
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 3,
		},
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 3,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{0, 400, 500, 500}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}

func TestScorerOriginal_ScoreClassAllPossibleOutcomes(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Manchester United", "Liverpool", 1, 3, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 1.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 1,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 2,
		},
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 3,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 6,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{0, 400, 500, 300}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}

func TestScorerOriginal_ScoreClassTwoCloseTwoOk(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := models.NewFinishedSeasonMatch("Manchester United", "Liverpool", 2, 4, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 1.0)

	bets := []*models.Bet{
		{
			Match:              match,
			PredictedHomeGoals: 1,
			PredictedAwayGoals: 3,
		},
		{
			Match:              match,
			PredictedHomeGoals: 3,
			PredictedAwayGoals: 5,
		},
		{
			Match:              match,
			PredictedHomeGoals: 0,
			PredictedAwayGoals: 2,
		},
		{
			Match:              match,
			PredictedHomeGoals: 3,
			PredictedAwayGoals: 4,
		},
	}

	scores := scorer.Score(match, bets)
	expectedScores := []int{400, 400, 300, 300}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}
