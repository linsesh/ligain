package models

import (
	"testing"
	"time"
)

func TestBet_HomeTeamWins(t *testing.T) {
	match := NewFinishedSeasonMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bet := NewBet(match, 1, 0)
	if !bet.IsBetCorrect() {
		t.Errorf("Expected bet to be correct")
	}
}

func TestBet_AwayTeamWins(t *testing.T) {
	match := NewFinishedSeasonMatch("Arsenal", "Chelsea", 0, 2, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bet := NewBet(match, 0, 2)
	if !bet.IsBetCorrect() {
		t.Errorf("Expected bet to be correct")
	}
}

func TestBet_Draw(t *testing.T) {
	match := NewFinishedSeasonMatch("Tottenham", "West Ham", 1, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bet := NewBet(match, 0, 0)
	if !bet.IsBetCorrect() {
		t.Errorf("Expected bet to be correct")
	}
}

func TestBet_HomeTeamWinsButPredictedWrong(t *testing.T) {
	match := NewFinishedSeasonMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bet := NewBet(match, 0, 2)
	if bet.IsBetCorrect() {
		t.Errorf("Expected bet to be incorrect")
	}
}

func TestBet_AwayTeamWinsButPredictedWrong(t *testing.T) {
	match := NewFinishedSeasonMatch("Arsenal", "Chelsea", 0, 2, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bet := NewBet(match, 2, 0)
	if bet.IsBetCorrect() {
		t.Errorf("Expected bet to be incorrect")
	}
}

func TestBet_DrawButPredictedWrong(t *testing.T) {
	match := NewFinishedSeasonMatch("Tottenham", "West Ham", 1, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	bet := NewBet(match, 2, 0)
	if bet.IsBetCorrect() {
		t.Errorf("Expected bet to be incorrect")
	}
}

func TestBetIsModifiable(t *testing.T) {
	// Use a fixed reference time
	referenceTime := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)

	// Create a match that starts 1 hour after reference time
	matchTime := referenceTime.Add(time.Hour)
	match := NewSeasonMatch("Home", "Away", "2024", "Premier League", matchTime, 1)

	// Create a bet for this match
	bet := NewBet(match, 2, 1)

	// Test cases
	tests := []struct {
		name     string
		now      time.Time
		expected bool
	}{
		{
			name:     "Before match start",
			now:      referenceTime,
			expected: true,
		},
		{
			name:     "After match start",
			now:      matchTime.Add(time.Minute),
			expected: false,
		},
		{
			name:     "At match start",
			now:      matchTime,
			expected: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bet.IsModifiable(tt.now); got != tt.expected {
				t.Errorf("Bet.IsModifiable() = %v, want %v", got, tt.expected)
			}
		})
	}

	// Test with in-progress match
	match = NewSeasonMatch("Home", "Away", "2024", "Premier League", referenceTime.Add(-time.Hour), 1)
	match.Start()
	bet = NewBet(match, 2, 1)
	if bet.IsModifiable(referenceTime) {
		t.Error("Bet should not be modifiable when match is in progress")
	}
}
