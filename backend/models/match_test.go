package models

import (
	"testing"
	"time"
)

func TestMatch_HomeTeamWins(t *testing.T) {
	match := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC))

	winner := match.GetWinner()
	if winner != "Manchester United" {
		t.Errorf("Expected winner to be Manchester United, got %s", winner)
	}
}

func TestMatch_AwayTeamWins(t *testing.T) {
	match := NewFinishedMatch("Arsenal", "Chelsea", 0, 2, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC))

	winner := match.GetWinner()
	if winner != "Chelsea" {
		t.Errorf("Expected winner to be Chelsea, got %s", winner)
	}
}

func TestMatch_Draw(t *testing.T) {
	match := NewFinishedMatch("Tottenham", "West Ham", 1, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC))

	winner := match.GetWinner()
	if winner != "Draw" {
		t.Errorf("Expected result to be Draw, got %s", winner)
	}
}
