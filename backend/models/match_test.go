package models

import (
	"testing"
)

func TestMatch_HomeTeamWins(t *testing.T) {
	match := NewMatch("Manchester United", "Liverpool")
	match.HomeGoals = 3
	match.AwayGoals = 1

	winner := match.GetWinner()
	if winner != "Manchester United" {
		t.Errorf("Expected winner to be Manchester United, got %s", winner)
	}
}

func TestMatch_AwayTeamWins(t *testing.T) {
	match := NewMatch("Arsenal", "Chelsea")
	match.HomeGoals = 0
	match.AwayGoals = 2

	winner := match.GetWinner()
	if winner != "Chelsea" {
		t.Errorf("Expected winner to be Chelsea, got %s", winner)
	}
}

func TestMatch_Draw(t *testing.T) {
	match := NewMatch("Tottenham", "West Ham")
	match.HomeGoals = 1
	match.AwayGoals = 1

	winner := match.GetWinner()
	if winner != "Draw" {
		t.Errorf("Expected result to be Draw, got %s", winner)
	}
} 