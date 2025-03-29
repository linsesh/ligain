package models

import (
	"testing"
)

func TestMatch_HomeTeamWins(t *testing.T) {
	match := NewMatch("Manchester United", "Liverpool")
	match.HomeGoals = 3
	match.AwayGoals = 1
	match.HomeTeamOdds = 1.5
	match.AwayTeamOdds = 4.0
	match.DrawOdds = 3.5

	winner := match.GetWinner()
	if winner != "Manchester United" {
		t.Errorf("Expected winner to be Manchester United, got %s", winner)
	}

	if match.HomeTeamOdds != 1.5 {
		t.Errorf("Expected home team odds to be 1.5, got %f", match.HomeTeamOdds)
	}
}

func TestMatch_AwayTeamWins(t *testing.T) {
	match := NewMatch("Arsenal", "Chelsea")
	match.HomeGoals = 0
	match.AwayGoals = 2
	match.HomeTeamOdds = 2.0
	match.AwayTeamOdds = 1.8
	match.DrawOdds = 3.2

	winner := match.GetWinner()
	if winner != "Chelsea" {
		t.Errorf("Expected winner to be Chelsea, got %s", winner)
	}

	if match.AwayTeamOdds != 1.8 {
		t.Errorf("Expected away team odds to be 1.8, got %f", match.AwayTeamOdds)
	}
}

func TestMatch_Draw(t *testing.T) {
	match := NewMatch("Tottenham", "West Ham")
	match.HomeGoals = 1
	match.AwayGoals = 1
	match.HomeTeamOdds = 1.9
	match.AwayTeamOdds = 3.5
	match.DrawOdds = 3.8

	winner := match.GetWinner()
	if winner != "Draw" {
		t.Errorf("Expected result to be Draw, got %s", winner)
	}

	if match.DrawOdds != 3.8 {
		t.Errorf("Expected draw odds to be 3.8, got %f", match.DrawOdds)
	}
} 