package rules

import (
	"ligain/backend/models"
	"ligain/backend/utils"
)

// ScorerOriginal is a scorer that uses the original rules of the game, defined here: https://docs.google.com/document/d/1Gv2s6EqsL5583PT56y8efra2PkMNJAEFGBv8Al5jxfQ/edit?tab=t.0 (todo: link to an english version on the repository instead)
type ScorerOriginal struct{}

func (s *ScorerOriginal) Score(match models.Match, bets []*models.Bet) []int {
	scores := make([]int, len(bets))
	for i, bet := range bets {
		otherBets := utils.SliceWithoutElementAtIndex(bets, i)
		scores[i] = s.scoreBet(match, bet, otherBets)
	}
	return scores
}

func (s *ScorerOriginal) scoreBet(match models.Match, bet *models.Bet, otherBets []*models.Bet) int {
	if bet == nil {
		return -100
	}
	if !isBetCorrectAgainstMatch(bet, match) {
		return 0
	}

	score := scoreBaseAgainstMatch(bet, match)
	score = addBonusIfUnfavorableOdds(match, score)
	score = addBonusDependingOnOtherBets(bet, otherBets, score)
	return score
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

// addBonusIfUnfavorableOdds adds a bonus if the odds are unfavorable
// The odds are unfavorable if the difference between the home and away odds is greater or equal to 1.5
// The bonus is * 1.5 in case of draw, * 2 in case of a win for the underdog.
func addBonusIfUnfavorableOdds(match models.Match, score int) int {
	favorite := getFavorite(match)
	if favorite != "" {
		if match.IsDraw() {
			return int(float64(score) * 1.5)
		}
		if match.GetWinner() != favorite {
			return int(float64(score) * 2)
		}
	}
	return score
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

// addBonusDependingOnOtherBets adds a bonus depending on the other bets
// The bonus is 25% of the score if there is at maximum 25% of the other bets with the same predicted result
// The bonus is 10% of the score if there is at maximum 50% of the other bets with the same predicted result
// Predicted result is "overall", we don't look at the precision here
func addBonusDependingOnOtherBets(bet *models.Bet, otherBets []*models.Bet, score int) int {
	totalNumberOfBets := len(otherBets) + 1
	numberOfBetsWithSameResult := 0
	for _, otherBet := range otherBets {
		if otherBet != nil && otherBet.GetPredictedResult() == bet.GetPredictedResult() {
			numberOfBetsWithSameResult++
		}
	}
	portionOfBetsWithSameResult := float64(numberOfBetsWithSameResult+1) / float64(totalNumberOfBets)
	if portionOfBetsWithSameResult <= 0.25 {
		return int(float64(score) * 1.25)
	}
	if portionOfBetsWithSameResult <= 0.5 {
		return int(float64(score) * 1.1)
	}
	return score
}
