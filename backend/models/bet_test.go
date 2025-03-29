package models

import "testing"

func TestBet_HomeTeamWins(t *testing.T) {
	match := NewMatch("Manchester United", "Liverpool")
	match.HomeGoals = 3
	match.AwayGoals = 1

	bet := NewBet(match, 1, 0)
	if !bet.IsBetCorrect() {
		t.Errorf("Expected bet to be correct")
	}
}

func TestBet_AwayTeamWins(t *testing.T) {
	match := NewMatch("Arsenal", "Chelsea")
	match.HomeGoals = 0
	match.AwayGoals = 2

	bet := NewBet(match, 0, 2)
	if !bet.IsBetCorrect() {
		t.Errorf("Expected bet to be correct")
	}
}

func TestBet_Draw(t *testing.T) {
	match := NewMatch("Tottenham", "West Ham")
	match.HomeGoals = 1
	match.AwayGoals = 1

	bet := NewBet(match, 0, 0)
	if !bet.IsBetCorrect() {
		t.Errorf("Expected bet to be correct")
	}
}

func TestBet_HomeTeamWinsButPredictedWrong(t *testing.T) {
	match := NewMatch("Manchester United", "Liverpool")
	match.HomeGoals = 3
	match.AwayGoals = 1

	bet := NewBet(match, 0, 2)
	if bet.IsBetCorrect() {
		t.Errorf("Expected bet to be incorrect")
	}
}

func TestBet_AwayTeamWinsButPredictedWrong(t *testing.T) {
	match := NewMatch("Arsenal", "Chelsea")
	match.HomeGoals = 0
	match.AwayGoals = 2

	bet := NewBet(match, 2, 0)
	if bet.IsBetCorrect() {
		t.Errorf("Expected bet to be incorrect")
	}
}

func TestBet_DrawButPredictedWrong(t *testing.T) {
	match := NewMatch("Tottenham", "West Ham")
	match.HomeGoals = 1
	match.AwayGoals = 1

	bet := NewBet(match, 2, 0)
	if bet.IsBetCorrect() {
		t.Errorf("Expected bet to be incorrect")
	}
}
