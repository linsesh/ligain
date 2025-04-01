package rules

import (
	"liguain/backend/models"
	"testing"
	"time"
)

func TestNewGame(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []*models.Match{
		models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now()),
		models.NewMatch("Team3", "Team4", "2024", "Premier League", time.Now()),
	}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	if game.GetSeasonCode() != "2024" {
		t.Errorf("Expected season code '2024', got %s", game.GetSeasonCode())
	}
	if game.GetCompetitionCode() != "Premier League" {
		t.Errorf("Expected competition code 'Premier League', got %s", game.GetCompetitionCode())
	}
	if game.GetGameStatus() != GameStatusNotStarted {
		t.Errorf("Expected game status 'not started', got %s", game.GetGameStatus())
	}
	if len(game.GetIncomingMatches()) != 2 {
		t.Errorf("Expected 2 incoming matches, got %d", len(game.GetIncomingMatches()))
	}
}

func TestAddPlayerBetGetMatchBets(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now())
	matches := []*models.Match{match}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	game.AddPlayerBet(&bettingPlayer, bet)

	playerBet := game.GetMatchBets(match)[bettingPlayer]
	if playerBet != bet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet, playerBet)
	}
}

func TestAddPlayerBetNonExistingMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now())
	matches := []*models.Match{match}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	game.AddPlayerBet(&bettingPlayer, bet)

	playerBet := game.GetMatchBets(match)[bettingPlayer]
	if playerBet != bet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet, playerBet)
	}
}

func TestAddPlayerBetGetMatchBets(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now())
	matches := []*models.Match{match}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	game.AddPlayerBet(&bettingPlayer, bet)

	playerBet := game.GetMatchBets(match)[bettingPlayer]
	if playerBet != bet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet, playerBet)
	}
}

func TestGetMatchBetsNonExistingMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now())
	matches := []*models.Match{match}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	game.AddPlayerBet(&bettingPlayer, bet)

	playerBet := game.GetMatchBets(match)[bettingPlayer]
	if playerBet != bet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet, playerBet)
	}
}

func TestAddFinishedMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now())
	matches := []*models.Match{match}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Add bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	game.AddPlayerBet(&players[0], bet1)
	game.AddPlayerBet(&players[1], bet2)

	// Finish the match
	finishedMatch := models.NewFinishedMatch("Team1", "Team2", 2, 1, "2024", "Premier League", time.Now())
	points, err := game.AddFinishedMatch(finishedMatch)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check points
	if points[players[0]] != 500 { // Perfect bet
		t.Errorf("Expected 500 points for Player1, got %d", points[players[0]])
	}
	if points[players[1]] != 0 { // Wrong bet
		t.Errorf("Expected 0 points for Player2, got %d", points[players[1]])
	}
}

func TestAddFinishedMatch_Errors(t *testing.T) {
	players := []models.Player{{Name: "Player1"}}
	match := models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now())
	matches := []*models.Match{match}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Test with non-existent match
	nonExistentMatch := models.NewMatch("Team3", "Team4", "2024", "Premier League", time.Now())
	_, err := game.AddFinishedMatch(nonExistentMatch)
	if err == nil {
		t.Error("Expected error for non-existent match")
	}

	// Test with unfinished match
	_, err = game.AddFinishedMatch(match)
	if err == nil {
		t.Error("Expected error for unfinished match")
	}
}

func TestGetIncomingMatches(t *testing.T) {
	players := []models.Player{{Name: "Player1"}}
	matches := []*models.Match{
		models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now()),
		models.NewMatch("Team3", "Team4", "2024", "Premier League", time.Now()),
		models.NewFinishedMatch("Team5", "Team6", 2, 1, "2024", "Premier League", time.Now()),
	}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	incomingMatches := game.GetIncomingMatches()
	if len(incomingMatches) != 2 {
		t.Errorf("Expected 2 incoming matches, got %d", len(incomingMatches))
	}

	// Check that only unfinished matches are returned
	for _, match := range incomingMatches {
		if match.IsFinished() {
			t.Error("Incoming matches should not include finished matches")
		}
	}
}

func TestGetPlayersPoints(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewMatch("Team1", "Team2", "2024", "Premier League", time.Now())
	matches := []*models.Match{match}
	scorer := &ScorerOriginal{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Initial points should be 0
	points := game.GetPlayersPoints()
	if points[players[0]] != 0 {
		t.Errorf("Expected 0 points for Player1, got %d", points[players[0]])
	}
	if points[players[1]] != 0 {
		t.Errorf("Expected 0 points for Player2, got %d", points[players[1]])
	}

	// Add bets and finish match
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	game.AddPlayerBet(&players[0], bet1)
	game.AddPlayerBet(&players[1], bet2)

	finishedMatch := models.NewFinishedMatch("Team1", "Team2", 2, 1, "2024", "Premier League", time.Now())
	_, err := game.AddFinishedMatch(finishedMatch)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check updated points
	points = game.GetPlayersPoints()
	if points[players[0]] != 500 {
		t.Errorf("Expected 500 points for Player1, got %d", points[players[0]])
	}
	if points[players[1]] != 0 {
		t.Errorf("Expected 0 points for Player2, got %d", points[players[1]])
	}
}

func TestGame_AddBet(t *testing.T) {
	game := NewGame("Premier League", "2024", []*Player{}, []*Match{})
	match := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Now())
	player := NewPlayer("John Doe")
	bet := NewBet(match, 3, 1)

	game.AddMatch(match)
	game.AddPlayer(player)
	game.AddBet(player, bet)

	if len(game.Bets) != 1 {
		t.Errorf("Expected 1 bet in game, got %d", len(game.Bets))
	}
	if game.Bets[0] != bet {
		t.Error("Added bet is not the same as the one in game")
	}
}

func TestGame_GetBetsForPlayer(t *testing.T) {
	game := NewGame("Premier League", "2024")
	match := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Now())
	player := NewPlayer("John Doe")
	bet := NewBet(match, 3, 1)

	game.AddMatch(match)
	game.AddPlayer(player)
	game.AddBet(player, bet)

	playerBets := game.GetBetsForPlayer(player)
	if len(playerBets) != 1 {
		t.Errorf("Expected 1 bet for player, got %d", len(playerBets))
	}
	if playerBets[0] != bet {
		t.Error("Retrieved bet is not the same as the one added")
	}
}

func TestGame_GetBetsForMatch(t *testing.T) {
	game := NewGame("Premier League", "2024")
	match := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Now())
	player := NewPlayer("John Doe")
	bet := NewBet(match, 3, 1)

	game.AddMatch(match)
	game.AddPlayer(player)
	game.AddBet(player, bet)

	matchBets := game.GetBetsForMatch(match)
	if len(matchBets) != 1 {
		t.Errorf("Expected 1 bet for match, got %d", len(matchBets))
	}
	if matchBets[0] != bet {
		t.Error("Retrieved bet is not the same as the one added")
	}
}

func TestGame_GetPlayerScoreWithMultipleBets(t *testing.T) {
	game := NewGame("Premier League", "2024")

	// First match
	match1 := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Now())
	player := NewPlayer("John Doe")
	bet1 := NewBet(match1, 3, 1) // Perfect bet (500 points)

	// Second match
	match2 := NewFinishedMatch("Arsenal", "Chelsea", 2, 2, "2024", "Premier League", time.Now())
	bet2 := NewBet(match2, 2, 2) // Perfect bet (500 points)

	game.AddMatch(match1)
	game.AddMatch(match2)
	game.AddPlayer(player)
	game.AddBet(player, bet1)
	game.AddBet(player, bet2)

	score := game.GetPlayerScore(player)
	if score != 1000 { // Two perfect bets should give 1000 points
		t.Errorf("Expected score of 1000, got %d", score)
	}
}

func TestGame_AddPlayerBet(t *testing.T) {
	player := &Player{Name: "John Doe"}
	match := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Now())
	bet := NewBet(match, 3, 1)

	game := NewGame("2024", "Premier League", []*Player{player}, []*Match{match})

	game.AddPlayerBet(player, bet)

	playerBets := game.GetPlayerBets(player)
	if len(playerBets) != 1 {
		t.Errorf("Expected 1 bet for player, got %d", len(playerBets))
	}
	if playerBets[0] != bet {
		t.Error("Retrieved bet is not the same as the one added")
	}
}

func TestGame_GetIncomingMatches(t *testing.T) {
	finishedMatch := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Now())
	scheduledMatch := NewScheduledMatch("Arsenal", "Chelsea", "2024", "Premier League", time.Now().Add(24*time.Hour))

	game := NewGame("2024", "Premier League", []*Player{}, []*Match{finishedMatch, scheduledMatch})

	incomingMatches := game.GetIncomingMatches()
	if len(incomingMatches) != 1 {
		t.Errorf("Expected 1 incoming match, got %d", len(incomingMatches))
	}
	if incomingMatches[0] != scheduledMatch {
		t.Error("Retrieved match is not the same as the scheduled match")
	}
}

func TestGame_GetPlayersPoints(t *testing.T) {
	player := &Player{Name: "John Doe"}
	match := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Now())
	bet := NewBet(match, 3, 1)

	game := NewGame("2024", "Premier League", []*Player{player}, []*Match{match})

	game.AddPlayerBet(player, bet)
	points := game.GetPlayersPoints()

	if points[*player] != 500 { // Perfect bet should give 500 points
		t.Errorf("Expected 500 points for player, got %d", points[*player])
	}
}

func TestGame_GetSeasonAndCompetitionCodes(t *testing.T) {
	game := NewGame("2024", "Premier League", []*Player{}, []*Match{})

	if game.GetSeasonCode() != "2024" {
		t.Errorf("Expected season code '2024', got %s", game.GetSeasonCode())
	}

	if game.GetCompetitionCode() != "Premier League" {
		t.Errorf("Expected competition code 'Premier League', got %s", game.GetCompetitionCode())
	}
}

func TestGame_InitialStatus(t *testing.T) {
	game := NewGame("2024", "Premier League", []*Player{}, []*Match{})

	if game.GetGameStatus() != GameStatusNotStarted {
		t.Errorf("Expected game status 'not started', got %s", game.GetGameStatus())
	}
}

func TestGame_AddFinishedMatchAndGetPlayerScore(t *testing.T) {
	game := NewGame("2024", "Premier League", []*Player{}, []*Match{})
	match := NewFinishedMatch("Manchester United", "Liverpool", 3, 1, "2024", "Premier League", time.Now())
	player := NewPlayer("John Doe")
	bet := NewBet(match, 3, 1)

	game.AddFinishedMatch(match)
	game.AddPlayer(player)
	game.AddBet(player, bet)

	score := game.GetPlayerScore(player)
	if score != 500 {
		t.Errorf("Expected score of 500, got %d", score)
	}
}
