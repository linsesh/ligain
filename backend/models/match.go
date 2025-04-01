package models

import (
	"fmt"
	"math"
	"time"
)

// MatchStatus represents the current status of a match
type MatchStatus string

const (
	MatchStatusScheduled MatchStatus = "scheduled"
	MatchStatusFinished  MatchStatus = "finished"
)

// Match represents a football match with teams, scores and odds
type Match struct {
	HomeTeam        string
	AwayTeam        string
	HomeGoals       int
	AwayGoals       int
	HomeTeamOdds    float64
	AwayTeamOdds    float64
	DrawOdds        float64
	Status          MatchStatus
	SeasonCode      string
	CompetitionCode string
	Date            time.Time
}

// NewMatch creates a new Match instance
func NewMatch(homeTeam, awayTeam string, seasonCode, competitionCode string, date time.Time) *Match {
	return &Match{
		HomeTeam:        homeTeam,
		AwayTeam:        awayTeam,
		Status:          MatchStatusScheduled,
		SeasonCode:      seasonCode,
		CompetitionCode: competitionCode,
		Date:            date,
	}
}

func NewFinishedMatch(homeTeam, awayTeam string, homeGoals, awayGoals int, seasonCode, competitionCode string, date time.Time) *Match {
	return &Match{
		HomeTeam:        homeTeam,
		AwayTeam:        awayTeam,
		HomeGoals:       homeGoals,
		AwayGoals:       awayGoals,
		Status:          MatchStatusFinished,
		SeasonCode:      seasonCode,
		CompetitionCode: competitionCode,
		Date:            date,
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

// IsFinished returns true if the match has been played
func (m *Match) IsFinished() bool {
	return m.Status == MatchStatusFinished
}

// Finish marks the match as finished and sets the final score
func (m *Match) Finish(homeGoals, awayGoals int) {
	m.HomeGoals = homeGoals
	m.AwayGoals = awayGoals
	m.Status = MatchStatusFinished
}

func (m *Match) Id() string {
	return fmt.Sprintf("%s-%s-%s,%d-%d", m.SeasonCode, m.CompetitionCode, m.Date.Format("Y-m-d"), m.HomeGoals, m.AwayGoals)
}
