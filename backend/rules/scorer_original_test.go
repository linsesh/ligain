package rules

import (
	"ligain/backend/models"
	"slices"
	"testing"
	"time"
)

func TestScorerOriginal_ScoreWithBreakdown(t *testing.T) {
	scorer := &ScorerOriginal{}

	t.Run("Perfect score, no risk, no clairvoyant", func(t *testing.T) {
		match := models.NewFinishedSeasonMatch("Team A", "Team B", 2, 1, "2024", "L1", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 2.0, 3.0, 2.5)
		bets := []*models.Bet{
			{Match: match, PredictedHomeGoals: 2, PredictedAwayGoals: 1},
			{Match: match, PredictedHomeGoals: 2, PredictedAwayGoals: 1},
			{Match: match, PredictedHomeGoals: 2, PredictedAwayGoals: 1},
			{Match: match, PredictedHomeGoals: 2, PredictedAwayGoals: 1},
		}
		breakdowns := scorer.ScoreWithBreakdown(match, bets)
		bd := breakdowns[0]
		if bd.BaseScore != 500 {
			t.Errorf("Expected base 500, got %d", bd.BaseScore)
		}
		if bd.RiskMultiplier != 1.0 {
			t.Errorf("Expected risk 1.0, got %f", bd.RiskMultiplier)
		}
		if bd.ClairvoyantMultiplier != 1.0 {
			t.Errorf("Expected clairvoyant 1.0, got %f", bd.ClairvoyantMultiplier)
		}
		if bd.Total != 500 {
			t.Errorf("Expected total 500, got %d", bd.Total)
		}
	})

	t.Run("Perfect score with underdog win (risk x2) and clairvoyant x1.25", func(t *testing.T) {
		// Bastia (8.0) vs Real Madrid (1.1) — Real Madrid is favorite, Bastia is underdog
		match := models.NewFinishedSeasonMatch("Bastia", "Real Madrid", 2, 0, "2024", "L1", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 8.0, 1.1, 6.0)
		bets := []*models.Bet{
			{Match: match, PredictedHomeGoals: 2, PredictedAwayGoals: 0},
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 4},
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 5},
			{Match: match, PredictedHomeGoals: 1, PredictedAwayGoals: 4},
		}
		breakdowns := scorer.ScoreWithBreakdown(match, bets)
		bd := breakdowns[0]
		if bd.BaseScore != 500 {
			t.Errorf("Expected base 500, got %d", bd.BaseScore)
		}
		if bd.RiskMultiplier != 2.0 {
			t.Errorf("Expected risk 2.0, got %f", bd.RiskMultiplier)
		}
		if bd.ClairvoyantMultiplier != 1.25 {
			t.Errorf("Expected clairvoyant 1.25, got %f", bd.ClairvoyantMultiplier)
		}
		if bd.Total != 1250 {
			t.Errorf("Expected total 1250, got %d", bd.Total)
		}
	})

	t.Run("Close score with draw bonus (risk x1.5)", func(t *testing.T) {
		match := models.NewFinishedSeasonMatch("Bastia", "Real Madrid", 2, 2, "2024", "L1", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 8.0, 1.1, 6.0)
		bets := []*models.Bet{
			{Match: match, PredictedHomeGoals: 1, PredictedAwayGoals: 1},
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 4},
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 5},
			{Match: match, PredictedHomeGoals: 1, PredictedAwayGoals: 4},
		}
		breakdowns := scorer.ScoreWithBreakdown(match, bets)
		bd := breakdowns[0]
		if bd.BaseScore != 400 {
			t.Errorf("Expected base 400, got %d", bd.BaseScore)
		}
		if bd.RiskMultiplier != 1.5 {
			t.Errorf("Expected risk 1.5, got %f", bd.RiskMultiplier)
		}
		if bd.ClairvoyantMultiplier != 1.25 {
			t.Errorf("Expected clairvoyant 1.25, got %f", bd.ClairvoyantMultiplier)
		}
		if bd.Total != 750 {
			t.Errorf("Expected total 750, got %d", bd.Total)
		}
	})

	t.Run("Wrong prediction returns zero breakdown", func(t *testing.T) {
		match := models.NewFinishedSeasonMatch("Team A", "Team B", 2, 1, "2024", "L1", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 2.0, 3.0, 2.5)
		bets := []*models.Bet{
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 2},
		}
		breakdowns := scorer.ScoreWithBreakdown(match, bets)
		bd := breakdowns[0]
		if bd.BaseScore != 0 || bd.RiskMultiplier != 1.0 || bd.ClairvoyantMultiplier != 1.0 || bd.Total != 0 {
			t.Errorf("Expected zero breakdown, got %+v", bd)
		}
	})

	t.Run("Nil bet returns missed bet breakdown", func(t *testing.T) {
		match := models.NewFinishedSeasonMatch("Team A", "Team B", 2, 1, "2024", "L1", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 2.0, 3.0, 2.5)
		bets := []*models.Bet{nil}
		breakdowns := scorer.ScoreWithBreakdown(match, bets)
		bd := breakdowns[0]
		if bd.BaseScore != -100 || bd.Total != -100 {
			t.Errorf("Expected -100 base and total for nil bet, got %+v", bd)
		}
		if bd.RiskMultiplier != 1.0 || bd.ClairvoyantMultiplier != 1.0 {
			t.Errorf("Expected multipliers 1.0 for nil bet, got %+v", bd)
		}
	})

	t.Run("Backward compatibility - Score still returns same ints", func(t *testing.T) {
		match := models.NewFinishedSeasonMatch("Bastia", "Real Madrid", 2, 0, "2024", "L1", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 8.0, 1.1, 6.0)
		bets := []*models.Bet{
			{Match: match, PredictedHomeGoals: 2, PredictedAwayGoals: 0},
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 4},
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 5},
			{Match: match, PredictedHomeGoals: 1, PredictedAwayGoals: 4},
		}
		scores := scorer.Score(match, bets)
		expected := []int{1250, 0, 0, 0}
		if !slices.Equal(scores, expected) {
			t.Errorf("Score() backward compat: expected %v, got %v", expected, scores)
		}
	})

	t.Run("Good result with clairvoyant x1.1 (50 percent or less same result)", func(t *testing.T) {
		match := models.NewFinishedSeasonMatch("Team A", "Team B", 2, 1, "2024", "L1", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 2.0, 3.0, 2.5)
		bets := []*models.Bet{
			{Match: match, PredictedHomeGoals: 2, PredictedAwayGoals: 1},
			{Match: match, PredictedHomeGoals: 2, PredictedAwayGoals: 1},
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 1},
			{Match: match, PredictedHomeGoals: 0, PredictedAwayGoals: 0},
		}
		breakdowns := scorer.ScoreWithBreakdown(match, bets)
		bd := breakdowns[0]
		if bd.BaseScore != 500 {
			t.Errorf("Expected base 500, got %d", bd.BaseScore)
		}
		if bd.RiskMultiplier != 1.0 {
			t.Errorf("Expected risk 1.0, got %f", bd.RiskMultiplier)
		}
		if bd.ClairvoyantMultiplier != 1.1 {
			t.Errorf("Expected clairvoyant 1.1, got %f", bd.ClairvoyantMultiplier)
		}
		if bd.Total != 550 {
			t.Errorf("Expected total 550, got %d", bd.Total)
		}
	})
}

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

func TestScorerOriginal_ScoreForgottenBets(t *testing.T) {
	scorer := &ScorerOriginal{}
	match := models.NewFinishedSeasonMatch("Forgotten FC", "Absent United", 1, 0, "2024", "Test League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 2.0, 2.0, 2.0)

	t.Run("All players forgot to bet", func(t *testing.T) {
		bets := []*models.Bet{nil, nil, nil}
		scores := scorer.Score(match, bets)
		expected := []int{-100, -100, -100}
		if !slices.Equal(scores, expected) {
			t.Errorf("Expected all -100 for forgotten bets, got %v", scores)
		}
	})

	t.Run("One player forgot to bet", func(t *testing.T) {
		bets := []*models.Bet{
			{
				Match:              match,
				PredictedHomeGoals: 1,
				PredictedAwayGoals: 0,
			},
			nil,
			{
				Match:              match,
				PredictedHomeGoals: 0,
				PredictedAwayGoals: 2,
			},
		}
		scores := scorer.Score(match, bets)
		expected := []int{550, -100, 0}
		if !slices.Equal(scores, expected) {
			t.Errorf("Expected [550 -100 0], got %v", scores)
		}
	})

	t.Run("Mixed forgotten, correct, and wrong", func(t *testing.T) {
		bets := []*models.Bet{
			nil,
			{
				Match:              match,
				PredictedHomeGoals: 1,
				PredictedAwayGoals: 0,
			},
			{
				Match:              match,
				PredictedHomeGoals: 0,
				PredictedAwayGoals: 1,
			},
		}
		scores := scorer.Score(match, bets)
		expected := []int{-100, 550, 0}
		if !slices.Equal(scores, expected) {
			t.Errorf("Expected [-100 550 0], got %v", scores)
		}
	})
}
