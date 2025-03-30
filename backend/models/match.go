package models

import "math"

// Match represents a football match with teams, scores and odds
type Match struct {
	HomeTeam     string
	AwayTeam     string
	HomeGoals    int
	AwayGoals    int
	HomeTeamOdds float64
	AwayTeamOdds float64
	DrawOdds     float64
}

func NewMatch(homeTeam, awayTeam string) *Match {
	return &Match{
		HomeTeam: homeTeam,
		AwayTeam: awayTeam,
	}
}

// GetWinner returns the winning team or "Draw" if the match is tied
func (m *Match) GetWinner() string {
	if m.HomeGoals > m.AwayGoals {
		return m.HomeTeam
	}
	if m.AwayGoals > m.HomeGoals {
		return m.AwayTeam
	}
	return "Draw"
}

func (m *Match) AbsoluteGoalDifference() int {
	return int(math.Abs(float64(m.HomeGoals - m.AwayGoals)))
}

func (m *Match) IsDraw() bool {
	return m.HomeGoals == m.AwayGoals
}

func (m *Match) TotalGoals() int {
	return m.HomeGoals + m.AwayGoals
}

func (m *Match) AbsoluteDifferenceOddsBetweenHomeAndAway() float64 {
	return math.Abs(m.HomeTeamOdds - m.AwayTeamOdds)
}
