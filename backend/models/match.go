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
	MatchStatusStarted   MatchStatus = "in-progress"
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
	SetHomeTeamOdds(odds float64)
	SetAwayTeamOdds(odds float64)
	SetDrawOdds(odds float64)
	AbsoluteGoalDifference() int
	IsDraw() bool
	TotalGoals() int
	AbsoluteDifferenceOddsBetweenHomeAndAway() float64
	GetWinner() string
	// Start marks the match as started
	Start()
	IsFinished() bool
	Finish(homeGoals, awayGoals int)
	IsInProgress() bool
	GetStatus() MatchStatus
}

// SeasonMatch represents a football match within a championship, like Ligue 1, Serie A, etc.
type SeasonMatch struct {
	HomeTeam        string      `json:"homeTeam"`
	AwayTeam        string      `json:"awayTeam"`
	HomeGoals       int         `json:"homeGoals"`
	AwayGoals       int         `json:"awayGoals"`
	HomeTeamOdds    float64     `json:"homeTeamOdds"`
	AwayTeamOdds    float64     `json:"awayTeamOdds"`
	DrawOdds        float64     `json:"drawOdds"`
	Status          MatchStatus `json:"status"`
	SeasonCode      string      `json:"seasonCode"`
	CompetitionCode string      `json:"competitionCode"`
	Date            time.Time   `json:"date"`
	Matchday        int         `json:"matchday"`
}

// NewMatch creates a new Match instance
func NewSeasonMatch(homeTeam, awayTeam string, seasonCode, competitionCode string, date time.Time, matchday int) *SeasonMatch {
	return &SeasonMatch{
		HomeTeam:        homeTeam,
		AwayTeam:        awayTeam,
		Status:          MatchStatusScheduled,
		SeasonCode:      seasonCode,
		CompetitionCode: competitionCode,
		Date:            date,
		Matchday:        matchday,
	}
}

func NewSeasonMatchWithKnownOdds(homeTeam, awayTeam string, seasonCode, competitionCode string, date time.Time, matchday int, homeTeamOdds, awayTeamOdds, drawOdds float64) *SeasonMatch {
	return &SeasonMatch{
		HomeTeam:        homeTeam,
		AwayTeam:        awayTeam,
		HomeTeamOdds:    homeTeamOdds,
		AwayTeamOdds:    awayTeamOdds,
		DrawOdds:        drawOdds,
		Status:          MatchStatusScheduled,
		SeasonCode:      seasonCode,
		CompetitionCode: competitionCode,
		Date:            date,
		Matchday:        matchday,
	}
}

func NewFinishedSeasonMatch(homeTeam, awayTeam string, homeGoals, awayGoals int, seasonCode, competitionCode string, date time.Time, matchday int, homeTeamOdds, awayTeamOdds, drawOdds float64) *SeasonMatch {
	return &SeasonMatch{
		HomeTeam:        homeTeam,
		AwayTeam:        awayTeam,
		HomeTeamOdds:    homeTeamOdds,
		AwayTeamOdds:    awayTeamOdds,
		DrawOdds:        drawOdds,
		HomeGoals:       homeGoals,
		AwayGoals:       awayGoals,
		Status:          MatchStatusFinished,
		SeasonCode:      seasonCode,
		CompetitionCode: competitionCode,
		Date:            date,
		Matchday:        matchday,
	}
}

// GetWinner returns the winning team or "Draw" if the match is tied
func (m *SeasonMatch) GetWinner() string {
	if m.HomeGoals > m.AwayGoals {
		return m.HomeTeam
	}
	if m.AwayGoals > m.HomeGoals {
		return m.AwayTeam
	}
	return "Draw"
}

func (m *SeasonMatch) AbsoluteGoalDifference() int {
	return int(math.Abs(float64(m.HomeGoals - m.AwayGoals)))
}

func (m *SeasonMatch) IsDraw() bool {
	return m.HomeGoals == m.AwayGoals
}

func (m *SeasonMatch) TotalGoals() int {
	return m.HomeGoals + m.AwayGoals
}

func (m *SeasonMatch) AbsoluteDifferenceOddsBetweenHomeAndAway() float64 {
	return math.Abs(m.HomeTeamOdds - m.AwayTeamOdds)
}

func (m *SeasonMatch) IsFinished() bool {
	return m.Status == MatchStatusFinished
}

func (m *SeasonMatch) IsInProgress() bool {
	return m.Status == MatchStatusStarted
}

// Finish marks the match as finished and sets the final score
func (m *SeasonMatch) Finish(homeGoals, awayGoals int) {
	m.HomeGoals = homeGoals
	m.AwayGoals = awayGoals
	m.Status = MatchStatusFinished
}

func (m *SeasonMatch) Id() string {
	return fmt.Sprintf("%s-%s-%s-%s-%d", m.CompetitionCode, m.SeasonCode, m.HomeTeam, m.AwayTeam, m.Matchday)
}

func (m *SeasonMatch) GetSeasonCode() string {
	return m.SeasonCode
}

func (m *SeasonMatch) GetCompetitionCode() string {
	return m.CompetitionCode
}

func (m *SeasonMatch) GetDate() time.Time {
	return m.Date
}

func (m *SeasonMatch) GetHomeTeam() string {
	return m.HomeTeam
}

func (m *SeasonMatch) GetAwayTeam() string {
	return m.AwayTeam
}

func (m *SeasonMatch) GetHomeGoals() int {
	return m.HomeGoals
}

func (m *SeasonMatch) GetAwayGoals() int {
	return m.AwayGoals
}

func (m *SeasonMatch) GetHomeTeamOdds() float64 {
	return m.HomeTeamOdds
}

func (m *SeasonMatch) GetAwayTeamOdds() float64 {
	return m.AwayTeamOdds
}

func (m *SeasonMatch) GetDrawOdds() float64 {
	return m.DrawOdds
}

func (m *SeasonMatch) SetHomeTeamOdds(odds float64) {
	m.HomeTeamOdds = odds
}

func (m *SeasonMatch) SetAwayTeamOdds(odds float64) {
	m.AwayTeamOdds = odds
}

func (m *SeasonMatch) SetDrawOdds(odds float64) {
	m.DrawOdds = odds
}

func (m *SeasonMatch) Start() {
	m.Status = MatchStatusStarted
}

func (m *SeasonMatch) GetStatus() MatchStatus {
	return m.Status
}
