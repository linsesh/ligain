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
type Match interface {
	Id() string
	GetSeasonCode() string
	GetCompetitionCode() string
	GetDate() time.Time
	GetHomeTeam() string
	GetAwayTeam() string
	GetHomeGoals() int
	GetAwayGoals() int
	GetHomeTeamOdds() float64
	GetAwayTeamOdds() float64
	GetDrawOdds() float64
	AbsoluteGoalDifference() int
	IsDraw() bool
	TotalGoals() int
	AbsoluteDifferenceOddsBetweenHomeAndAway() float64
	IsFinished() bool
	Finish(homeGoals, awayGoals int)
	GetWinner() string
}

// SeasonMatch represents a football match within a championship, like Ligue 1, Serie A, etc.
type SeasonMatch struct {
	homeTeam        string
	awayTeam        string
	homeGoals       int
	awayGoals       int
	homeTeamOdds    float64
	awayTeamOdds    float64
	drawOdds        float64
	status          MatchStatus
	seasonCode      string
	competitionCode string
	date            time.Time // time of the match could be modified
	matchday        int
}

// NewMatch creates a new Match instance
func NewSeasonMatch(homeTeam, awayTeam string, seasonCode, competitionCode string, date time.Time, matchday int) *SeasonMatch {
	return &SeasonMatch{
		homeTeam:        homeTeam,
		awayTeam:        awayTeam,
		status:          MatchStatusScheduled,
		seasonCode:      seasonCode,
		competitionCode: competitionCode,
		date:            date,
		matchday:        matchday,
	}
}

func NewSeasonMatchWithKnownOdds(homeTeam, awayTeam string, seasonCode, competitionCode string, date time.Time, matchday int, homeTeamOdds, awayTeamOdds, drawOdds float64) *SeasonMatch {
	return &SeasonMatch{
		homeTeam:        homeTeam,
		awayTeam:        awayTeam,
		homeTeamOdds:    homeTeamOdds,
		awayTeamOdds:    awayTeamOdds,
		drawOdds:        drawOdds,
		status:          MatchStatusScheduled,
		seasonCode:      seasonCode,
		competitionCode: competitionCode,
		date:            date,
		matchday:        matchday,
	}
}

func NewFinishedSeasonMatch(homeTeam, awayTeam string, homeGoals, awayGoals int, seasonCode, competitionCode string, date time.Time, matchday int, homeTeamOdds, awayTeamOdds, drawOdds float64) *SeasonMatch {
	return &SeasonMatch{
		homeTeam:        homeTeam,
		awayTeam:        awayTeam,
		homeTeamOdds:    homeTeamOdds,
		awayTeamOdds:    awayTeamOdds,
		drawOdds:        drawOdds,
		homeGoals:       homeGoals,
		awayGoals:       awayGoals,
		status:          MatchStatusFinished,
		seasonCode:      seasonCode,
		competitionCode: competitionCode,
		date:            date,
		matchday:        matchday,
	}
}

// GetWinner returns the winning team or "Draw" if the match is tied
func (m *SeasonMatch) GetWinner() string {
	if m.homeGoals > m.awayGoals {
		return m.homeTeam
	}
	if m.awayGoals > m.homeGoals {
		return m.awayTeam
	}
	return "Draw"
}

func (m *SeasonMatch) AbsoluteGoalDifference() int {
	return int(math.Abs(float64(m.homeGoals - m.awayGoals)))
}

func (m *SeasonMatch) IsDraw() bool {
	return m.homeGoals == m.awayGoals
}

func (m *SeasonMatch) TotalGoals() int {
	return m.homeGoals + m.awayGoals
}

func (m *SeasonMatch) AbsoluteDifferenceOddsBetweenHomeAndAway() float64 {
	return math.Abs(m.homeTeamOdds - m.awayTeamOdds)
}

// IsFinished returns true if the match has been played
func (m *SeasonMatch) IsFinished() bool {
	return m.status == MatchStatusFinished
}

// Finish marks the match as finished and sets the final score
func (m *SeasonMatch) Finish(homeGoals, awayGoals int) {
	m.homeGoals = homeGoals
	m.awayGoals = awayGoals
	m.status = MatchStatusFinished
}

func (m *SeasonMatch) Id() string {
	return fmt.Sprintf("%s-%s-%s-%s-%d", m.competitionCode, m.seasonCode, m.homeTeam, m.awayTeam, m.matchday)
}

func (m *SeasonMatch) GetSeasonCode() string {
	return m.seasonCode
}

func (m *SeasonMatch) GetCompetitionCode() string {
	return m.competitionCode
}

func (m *SeasonMatch) GetDate() time.Time {
	return m.date
}

func (m *SeasonMatch) GetHomeTeam() string {
	return m.homeTeam
}

func (m *SeasonMatch) GetAwayTeam() string {
	return m.awayTeam
}

func (m *SeasonMatch) GetHomeGoals() int {
	return m.homeGoals
}

func (m *SeasonMatch) GetAwayGoals() int {
	return m.awayGoals
}

func (m *SeasonMatch) GetHomeTeamOdds() float64 {
	return m.homeTeamOdds
}

func (m *SeasonMatch) GetAwayTeamOdds() float64 {
	return m.awayTeamOdds
}

func (m *SeasonMatch) GetDrawOdds() float64 {
	return m.drawOdds
}
