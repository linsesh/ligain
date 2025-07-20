package models

import (
	"testing"
	"time"
)

func TestMatch_HomeTeamWins(t *testing.T) {
	match := NewFinishedSeasonMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	winner := match.GetWinner()
	if winner != "Manchester United" {
		t.Errorf("Expected winner to be Manchester United, got %s", winner)
	}
}

func TestMatch_AwayTeamWins(t *testing.T) {
	match := NewFinishedSeasonMatch("Arsenal", "Chelsea", 0, 2, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	winner := match.GetWinner()
	if winner != "Chelsea" {
		t.Errorf("Expected winner to be Chelsea, got %s", winner)
	}
}

func TestMatch_Draw(t *testing.T) {
	match := NewFinishedSeasonMatch("Tottenham", "West Ham", 1, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 2.0, 3.0)

	winner := match.GetWinner()
	if winner != "Draw" {
		t.Errorf("Expected result to be Draw, got %s", winner)
	}
}

func TestMatch_HasClearFavorite(t *testing.T) {
	// Test with clear favorite (odds difference > 1.5)
	match := NewFinishedSeasonMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 3.0, 2.0)
	if !match.HasClearFavorite() {
		t.Errorf("Expected match to have clear favorite with odds difference of 2.0")
	}

	// Test without clear favorite (odds difference <= 1.5)
	match2 := NewFinishedSeasonMatch("Arsenal", "Chelsea", 0, 2, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.5, 2.0, 3.0)
	if match2.HasClearFavorite() {
		t.Errorf("Expected match to not have clear favorite with odds difference of 0.5")
	}
}

func TestMatch_GetFavoriteTeam(t *testing.T) {
	// Test with home team as favorite
	match := NewFinishedSeasonMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.0, 3.0, 2.0)
	favorite := match.GetFavoriteTeam()
	if favorite != "Manchester United" {
		t.Errorf("Expected favorite to be Manchester United, got %s", favorite)
	}

	// Test with away team as favorite
	match2 := NewFinishedSeasonMatch("Arsenal", "Chelsea", 0, 2, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 3.0, 1.0, 2.0)
	favorite2 := match2.GetFavoriteTeam()
	if favorite2 != "Chelsea" {
		t.Errorf("Expected favorite to be Chelsea, got %s", favorite2)
	}

	// Test without clear favorite
	match3 := NewFinishedSeasonMatch("Tottenham", "West Ham", 1, 1, "2024", "Premier League", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC), 1, 1.5, 2.0, 3.0)
	favorite3 := match3.GetFavoriteTeam()
	if favorite3 != "" {
		t.Errorf("Expected no favorite, got %s", favorite3)
	}
}
