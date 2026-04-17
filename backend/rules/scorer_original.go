package rules

import (
	"ligain/backend/models"
)

// ScorerOriginal is a scorer that uses the original rules of the game, defined here: https://docs.google.com/document/d/1Gv2s6EqsL5583PT56y8efra2PkMNJAEFGBv8Al5jxfQ/edit?tab=t.0 (todo: link to an english version on the repository instead)
type ScorerOriginal struct{}

func (s *ScorerOriginal) Score(match models.Match, bets []*models.Bet) []int {
	breakdowns := s.ScoreWithBreakdown(match, bets)
	scores := make([]int, len(breakdowns))
	for i, bd := range breakdowns {
		scores[i] = bd.Total
	}
	return scores
}

func (s *ScorerOriginal) ScoreWithBreakdown(match models.Match, bets []*models.Bet) []ScoreBreakdown {
	breakdowns := make([]ScoreBreakdown, len(bets))
	for i, bet := range bets {
		breakdowns[i] = s.scoreBetWithBreakdown(match, bet, betsWithout(bets, i))
	}
	return breakdowns
}

func betsWithout(bets []*models.Bet, index int) []*models.Bet {
	result := make([]*models.Bet, 0, len(bets)-1)
	for i, b := range bets {
		if i != index {
			result = append(result, b)
		}
	}
	return result
}

func (s *ScorerOriginal) scoreBetWithBreakdown(match models.Match, bet *models.Bet, otherBets []*models.Bet) ScoreBreakdown {
	if bet == nil {
		return ScoreBreakdown{BaseScore: -100, RiskMultiplier: 1.0, ClairvoyantMultiplier: 1.0, Total: -100}
	}
	if !isBetCorrectAgainstMatch(bet, match) {
		return ScoreBreakdown{BaseScore: 0, RiskMultiplier: 1.0, ClairvoyantMultiplier: 1.0, Total: 0}
	}

	baseScore := scoreBaseAgainstMatch(bet, match)
	riskMultiplier := getRiskMultiplier(match)
	clairvoyantMultiplier := getClairvoyantMultiplier(bet, otherBets)
	total := int(float64(baseScore) * riskMultiplier * clairvoyantMultiplier)

	return ScoreBreakdown{
		BaseScore:             baseScore,
		RiskMultiplier:        riskMultiplier,
		ClairvoyantMultiplier: clairvoyantMultiplier,
		Total:                 total,
	}
}

// isBetCorrectAgainstMatch checks if a bet is correct against a specific match
// This is needed because bet.IsBetCorrect() uses the bet's embedded match,
// but we need to check against the actual finished match being scored
func isBetCorrectAgainstMatch(bet *models.Bet, match models.Match) bool {
	if match.GetHomeGoals() > match.GetAwayGoals() {
		return bet.PredictedHomeGoals > bet.PredictedAwayGoals
	}
	if match.GetHomeGoals() < match.GetAwayGoals() {
		return bet.PredictedHomeGoals < bet.PredictedAwayGoals
	}
	return bet.PredictedHomeGoals == bet.PredictedAwayGoals
}

// scoreBaseAgainstMatch calculates base score against a specific match
func scoreBaseAgainstMatch(bet *models.Bet, match models.Match) int {
	if isBetPerfectAgainstMatch(bet, match) {
		return 500
	}
	if isBetCloseAgainstMatch(bet, match) {
		return 400
	}
	return 300
}

// isBetPerfectAgainstMatch checks if bet prediction exactly matches the match result
func isBetPerfectAgainstMatch(bet *models.Bet, match models.Match) bool {
	return bet.PredictedHomeGoals == match.GetHomeGoals() && bet.PredictedAwayGoals == match.GetAwayGoals()
}

// isBetCloseAgainstMatch checks if bet is close to the match result
func isBetCloseAgainstMatch(bet *models.Bet, match models.Match) bool {
	return isGoalDifferenceTheSameAsMatch(bet, match) && absoluteDifferenceTotalGoalsWithMatch(bet, match) <= 2
}

// isGoalDifferenceTheSameAsMatch checks if goal difference matches
func isGoalDifferenceTheSameAsMatch(bet *models.Bet, match models.Match) bool {
	betGoalDiff := absoluteGoalDifference(bet.PredictedHomeGoals, bet.PredictedAwayGoals)
	matchGoalDiff := absoluteGoalDifference(match.GetHomeGoals(), match.GetAwayGoals())
	return betGoalDiff == matchGoalDiff
}

// absoluteDifferenceTotalGoalsWithMatch calculates difference in total goals
func absoluteDifferenceTotalGoalsWithMatch(bet *models.Bet, match models.Match) int {
	betTotal := bet.PredictedHomeGoals + bet.PredictedAwayGoals
	matchTotal := match.GetHomeGoals() + match.GetAwayGoals()
	if betTotal > matchTotal {
		return betTotal - matchTotal
	}
	return matchTotal - betTotal
}

// absoluteGoalDifference calculates absolute goal difference
func absoluteGoalDifference(home, away int) int {
	if home > away {
		return home - away
	}
	return away - home
}

func getRiskMultiplier(match models.Match) float64 {
	favorite := getFavorite(match)
	if favorite == "" {
		return 1.0
	}
	if match.IsDraw() {
		return 1.5
	}
	if match.GetWinner() != favorite {
		return 2.0
	}
	return 1.0
}

// getFavorite returns the favorite team of the match
// The function is not implemented in the match file because the favorite is defined by the scorer, and could be different for each scorer
func getFavorite(match models.Match) string {
	if match.AbsoluteDifferenceOddsBetweenHomeAndAway() < 1.5 {
		return ""
	}
	if match.GetHomeTeamOdds() < match.GetAwayTeamOdds() {
		return match.GetHomeTeam()
	}
	return match.GetAwayTeam()
}

func getClairvoyantMultiplier(bet *models.Bet, otherBets []*models.Bet) float64 {
	totalNumberOfBets := len(otherBets) + 1
	numberOfBetsWithSameResult := 0
	for _, otherBet := range otherBets {
		if otherBet != nil && otherBet.GetPredictedResult() == bet.GetPredictedResult() {
			numberOfBetsWithSameResult++
		}
	}
	portionOfBetsWithSameResult := float64(numberOfBetsWithSameResult+1) / float64(totalNumberOfBets)
	if portionOfBetsWithSameResult <= 0.25 {
		return 1.25
	}
	if portionOfBetsWithSameResult <= 0.5 {
		return 1.1
	}
	return 1.0
}
