package rules

import "liguain/backend/models"

// ScorerOriginal is a scorer that uses the original rules of the game, defined here: https://docs.google.com/document/d/1Gv2s6EqsL5583PT56y8efra2PkMNJAEFGBv8Al5jxfQ/edit?tab=t.0 (todo: link to an english version on the repository instead)
type ScorerOriginal struct{}

func (s *ScorerOriginal) Score(match *models.Match, bets []*models.Bet) []int {
	return []int{0, 0, 0, 0}
}
