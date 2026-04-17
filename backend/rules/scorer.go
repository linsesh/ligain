package rules

import "ligain/backend/models"

type ScoreBreakdown struct {
	BaseScore              int     `json:"baseScore"`
	RiskMultiplier         float64 `json:"riskMultiplier"`
	ClairvoyantMultiplier  float64 `json:"clairvoyantMultiplier"`
	Total                  int     `json:"total"`
}

// Given a match and all the bets, returns the score of each bet
type Scorer interface {
	Score(match models.Match, bets []*models.Bet) []int
	ScoreWithBreakdown(match models.Match, bets []*models.Bet) []ScoreBreakdown
}
