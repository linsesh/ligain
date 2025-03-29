package models

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

// NewMatch creates a new Match instance
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
