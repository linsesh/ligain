package rules

import (
	"testing"
	"liguain/backend/models"
	"slices"
)

func TestScorerOriginal_ScoreClassAllPerfect(t *testing.T) {
	scorer := &ScorerOriginal{}

	match := &models.Match{
		HomeTeam:  "Manchester United",
		AwayTeam:  "Liverpool",
		HomeGoals: 2,
		AwayGoals: 1,
	}

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

	match := &models.Match{
		HomeTeam:  "Manchester United",
		AwayTeam:  "Liverpool",
		HomeGoals: 2,
		AwayGoals: 1,
	}

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

	match := &models.Match{
		HomeTeam:  "Manchester United",
		AwayTeam:  "Liverpool",
		HomeGoals: 2,
		AwayGoals: 1,
	}

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
	expectedScores := []int{625, 0, 0, 0}

	if !slices.Equal(scores, expectedScores) {
		t.Errorf("Expected scores to be %v, got %v", expectedScores, scores)
	}
}
