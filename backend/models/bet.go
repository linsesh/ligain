package models

import (
	"time"
)

// Bet represents a bet made by one player on a match
// It holds only the match ID + predicted goals, not an embedded copy of the match
type Bet struct {
	MatchID            string `json:"matchId"`
	PredictedHomeGoals int    `json:"predictedHomeGoals"`
	PredictedAwayGoals int    `json:"predictedAwayGoals"`
}

// NewBet creates a new Bet instance from a match and predicted goals
// For convenience, it extracts the match ID from the passed match
func NewBet(match Match, predictedHomeGoals int, predictedAwayGoals int) *Bet {
	return &Bet{
		MatchID:            match.Id(),
		PredictedHomeGoals: predictedHomeGoals,
		PredictedAwayGoals: predictedAwayGoals,
	}
}

// IsBetCorrect checks if the predicted result matches the actual match result
// The match must be passed because bets no longer embed a copy
func (b *Bet) IsBetCorrect(match Match) bool {
	if match.GetHomeGoals() > match.GetAwayGoals() {
		return b.PredictedHomeGoals > b.PredictedAwayGoals
	}
	if match.GetHomeGoals() < match.GetAwayGoals() {
		return b.PredictedHomeGoals < b.PredictedAwayGoals
	}
	return b.PredictedHomeGoals == b.PredictedAwayGoals
}

// GetPredictedResult returns "home", "away", or "draw" based on predicted goals
// This is a stable value that doesn't depend on the match state
func (b *Bet) GetPredictedResult() string {
	if b.PredictedHomeGoals > b.PredictedAwayGoals {
		return "home"
	}
	if b.PredictedHomeGoals < b.PredictedAwayGoals {
		return "away"
	}
	return "draw"
}

// IsModifiable checks if the bet can be modified based on the match's state and time
// The match must be passed because bets no longer embed a copy
func (b *Bet) IsModifiable(match Match, now time.Time) bool {
	return !match.IsFinished() && now.Before(match.GetDate())
}
