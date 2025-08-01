package testutils

import (
	"ligain/backend/models"
	"math"
	"time"
)

// MockMatch implements the models.Match interface for testing
type MockMatch struct {
	id string
}

func NewMockMatch(id string) *MockMatch {
	return &MockMatch{id: id}
}

func (m *MockMatch) Id() string {
	return m.id
}

func (m *MockMatch) GetSeasonCode() string {
	return "2024"
}

func (m *MockMatch) GetCompetitionCode() string {
	return "TEST"
}

func (m *MockMatch) GetDate() time.Time {
	return time.Now()
}

func (m *MockMatch) GetHomeTeam() string {
	return "home"
}

func (m *MockMatch) GetAwayTeam() string {
	return "away"
}

func (m *MockMatch) GetHomeGoals() int {
	return 0
}

func (m *MockMatch) GetAwayGoals() int {
	return 0
}

func (m *MockMatch) GetHomeTeamOdds() float64 {
	return 1.5
}

func (m *MockMatch) GetAwayTeamOdds() float64 {
	return 2.5
}

func (m *MockMatch) GetDrawOdds() float64 {
	return 3.0
}

func (m *MockMatch) AbsoluteGoalDifference() int {
	return 0
}

func (m *MockMatch) IsDraw() bool {
	return true
}

func (m *MockMatch) TotalGoals() int {
	return 0
}

func (m *MockMatch) AbsoluteDifferenceOddsBetweenHomeAndAway() float64 {
	return math.Abs(m.GetHomeTeamOdds() - m.GetAwayTeamOdds())
}

func (m *MockMatch) GetWinner() string {
	return "Draw"
}

func (m *MockMatch) Start() {
	// no-op for mock
}

func (m *MockMatch) Finish(homeGoals, awayGoals int) {
	// no-op for mock
}

func (m *MockMatch) IsFinished() bool {
	return true
}

func (m *MockMatch) IsInProgress() bool {
	return false
}

func (m *MockMatch) GetStatus() models.MatchStatus {
	return models.MatchStatusFinished
}

func (m *MockMatch) SetHomeTeamOdds(odds float64) {
	// no-op for mock
}

func (m *MockMatch) SetAwayTeamOdds(odds float64) {
	// no-op for mock
}

func (m *MockMatch) SetDrawOdds(odds float64) {
	// no-op for mock
}

func (m *MockMatch) HasClearFavorite() bool {
	return m.AbsoluteDifferenceOddsBetweenHomeAndAway() > 1.5
}

func (m *MockMatch) GetFavoriteTeam() string {
	if !m.HasClearFavorite() {
		return ""
	}
	if m.GetHomeTeamOdds() < m.GetAwayTeamOdds() {
		return m.GetHomeTeam()
	}
	return m.GetAwayTeam()
}
