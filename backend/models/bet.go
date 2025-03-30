package models

import "math"

// Match represents a bet made by one player
type Bet struct {
	Match              *Match
	PredictedHomeGoals int
	PredictedAwayGoals int
}

// NewBet creates a new Bet instance
func NewBet(match *Match, predictedHomeGoals int, predictedAwayGoals int) *Bet {
	return &Bet{
		Match:              match,
		PredictedHomeGoals: predictedHomeGoals,
		PredictedAwayGoals: predictedAwayGoals,
	}
}

// IsBetCorrect only checks if the predicted team won when the bet isn't a draw, or checks the draw the draw if it was predicted
func (b *Bet) IsBetCorrect() bool {
	if b.Match.HomeGoals > b.Match.AwayGoals {
		return b.PredictedHomeGoals > b.PredictedAwayGoals
	}
	if b.Match.HomeGoals < b.Match.AwayGoals {
		return b.PredictedHomeGoals < b.PredictedAwayGoals
	}
	return b.PredictedHomeGoals == b.PredictedAwayGoals
}

func (b *Bet) IsBetPerfect() bool {
	return b.PredictedHomeGoals == b.Match.HomeGoals && b.PredictedAwayGoals == b.Match.AwayGoals
}

func (b *Bet) AbsoluteDifferenceGoalDifferenceWithMatch() int {
	return int(math.Abs(float64(b.Match.AbsoluteGoalDifference() - b.AbsoluteGoalDifference())))
}

func (b *Bet) IsGoalDifferenceTheSameAsMatch() bool {
	return b.AbsoluteDifferenceGoalDifferenceWithMatch() == 0
}

func (b *Bet) AbsoluteGoalDifference() int {
	return int(math.Abs(float64(b.PredictedHomeGoals - b.PredictedAwayGoals)))
}

func (b *Bet) TotalPredictedGoals() int {
	return b.PredictedHomeGoals + b.PredictedAwayGoals
}

func (b *Bet) AbsoluteDifferenceTotalGoalsWithMatch() int {
	return int(math.Abs(float64(b.TotalPredictedGoals() - b.Match.TotalGoals())))
}

func (b *Bet) GetPredictedResult() string {
	if b.PredictedHomeGoals > b.PredictedAwayGoals {
		return b.Match.HomeTeam
	}
	if b.PredictedHomeGoals < b.PredictedAwayGoals {
		return b.Match.AwayTeam
	}
	return "Draw"
}
