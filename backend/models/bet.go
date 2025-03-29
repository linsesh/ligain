package models

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
