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
	// Clear favorite methods
	HasClearFavorite() bool
	GetFavoriteTeam() string
	// Clone returns an independent copy of the match
	Clone() Match
	// TryUpdate returns a copy of the match with the update applied, and whether anything changed.
	// Status, score and date are always applied. Odds are refused from OddsFreezeDelay before kickoff on.
	TryUpdate(update MatchUpdate, now time.Time) (Match, bool)
}

// MatchUpdate is a new state reported for a match. TryUpdate decides which parts of it are applied.
type MatchUpdate struct {
	Status       MatchStatus
	Date         time.Time
	HomeGoals    int
	AwayGoals    int
	HomeTeamOdds float64
	AwayTeamOdds float64
	DrawOdds     float64
}

// MatchUpdateOf reads the state of a match as an update
func MatchUpdateOf(match Match) MatchUpdate {
	return MatchUpdate{
		Status:       match.GetStatus(),
		Date:         match.GetDate(),
		HomeGoals:    match.GetHomeGoals(),
		AwayGoals:    match.GetAwayGoals(),
		HomeTeamOdds: match.GetHomeTeamOdds(),
		AwayTeamOdds: match.GetAwayTeamOdds(),
		DrawOdds:     match.GetDrawOdds(),
	}
}

// MatchUpdatesOf reads the state of several matches, keyed by match id, as updates
func MatchUpdatesOf(matches map[string]Match) map[string]MatchUpdate {
	updates := make(map[string]MatchUpdate, len(matches))
	for id, match := range matches {
		updates[id] = MatchUpdateOf(match)
	}
	return updates
}

// OddsFreezeDelay is how long before kickoff the odds stop following the provider.
// We use 6 minutes to be sure the odds are always blocked at least 5 minutes before the match starts.
const OddsFreezeDelay = 6 * time.Minute

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

// HasClearFavorite returns true if there is a clear favorite (odds difference > 1.5)
func (m *SeasonMatch) HasClearFavorite() bool {
	return m.AbsoluteDifferenceOddsBetweenHomeAndAway() > 1.5
}

// GetFavoriteTeam returns the favorite team or empty string if no clear favorite
func (m *SeasonMatch) GetFavoriteTeam() string {
	if !m.HasClearFavorite() {
		return ""
	}
	if m.HomeTeamOdds < m.AwayTeamOdds {
		return m.HomeTeam
	}
	return m.AwayTeam
}

// Clone returns an independent copy of the match
func (m *SeasonMatch) Clone() Match {
	copied := *m
	return &copied
}

func (m *SeasonMatch) hasUsableOdds() bool {
	return m.HomeTeamOdds > 0 && m.AwayTeamOdds > 0 && m.DrawOdds > 0
}

// TryUpdate returns a copy of the match with the update applied, and whether anything changed.
// Status, score and date are always applied (cancellations and postponements must go through).
// The odds are refused from OddsFreezeDelay before kickoff on, once we have usable ones: scores are computed
// from them, and bookmakers settle them at extreme prices once the match is over. The freeze is decided on the
// date of the update, so a postponed match is no longer about to start and its odds follow again.
func (m *SeasonMatch) TryUpdate(update MatchUpdate, now time.Time) (Match, bool) {
	updated := *m
	updated.Date = update.Date
	updated.Status = update.Status
	updated.HomeGoals = update.HomeGoals
	updated.AwayGoals = update.AwayGoals
	if !m.hasUsableOdds() || !update.Date.Before(now.Add(OddsFreezeDelay)) {
		updated.HomeTeamOdds = update.HomeTeamOdds
		updated.AwayTeamOdds = update.AwayTeamOdds
		updated.DrawOdds = update.DrawOdds
	}
	return &updated, !m.sameStateAs(&updated)
}

func (m *SeasonMatch) sameStateAs(other *SeasonMatch) bool {
	return m.Status == other.Status && m.Date.Equal(other.Date) &&
		m.HomeGoals == other.HomeGoals && m.AwayGoals == other.AwayGoals &&
		m.HomeTeamOdds == other.HomeTeamOdds && m.AwayTeamOdds == other.AwayTeamOdds && m.DrawOdds == other.DrawOdds
}
