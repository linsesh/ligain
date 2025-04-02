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
