package rules

import "ligain/backend/models"

// Given a match and all the bets, returns the score of each bet
type Scorer interface {
	Score(match models.Match, bets []*models.Bet) []int
}
