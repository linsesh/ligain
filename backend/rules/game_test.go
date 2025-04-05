package rules

import (
	"liguain/backend/models"
	"testing"
	"time"
)

type ScorerTest struct{}

func (s *ScorerTest) Score(match models.Match, bets []*models.Bet) []int {
	scores := make([]int, len(bets))
	for i, bet := range bets {
		if bet.IsBetCorrect() {
			scores[i] = 500
		} else {
			scores[i] = 0
		}
	}
	return scores
}

// Reference time for all tests
var testTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

func TestNewGame(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{
		models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1),
		models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", testTime, 1),
	}
	scorer := &ScorerTest{}

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
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	game.AddPlayerBet(&bettingPlayer, bet, testTime)

	bet_map, err := game.GetMatchBets(match)
	if err != nil {
		t.Errorf("Error retrieving match bets: %v", err)
	}
	playerBet := bet_map[bettingPlayer]
	if playerBet != bet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet, playerBet)
	}
}

func TestAddPlayerBetUpdateBet(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	game.AddPlayerBet(&bettingPlayer, bet, testTime)
	bet_map, err := game.GetMatchBets(match)
	if err != nil {
		t.Errorf("Error retrieving match bets: %v", err)
	}
	playerBet := bet_map[bettingPlayer]
	if playerBet != bet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet, playerBet)
	}
	updatedBet := models.NewBet(match, 1, 2)
	game.AddPlayerBet(&bettingPlayer, updatedBet, testTime)

	bet_map, err = game.GetMatchBets(match)
	if err != nil {
		t.Errorf("Error retrieving match bets: %v", err)
	}
	playerBet = bet_map[bettingPlayer]
	if playerBet != updatedBet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", updatedBet, playerBet)
	}
}

func TestAddPlayerBetNonExistingPlayer(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := models.Player{Name: "Player3"}
	err := game.AddPlayerBet(&bettingPlayer, bet, testTime)
	if err == nil {
		t.Errorf("Expected error for non-existing player")
	}
}

func TestAddPlayerBetNonExistingMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	fake_match := models.NewSeasonMatch("Sardinia", "Corsica", "2024", "Corsica Cup", testTime, 1)
	bet := models.NewBet(fake_match, 2, 1)
	bettingPlayer := players[0]
	err := game.AddPlayerBet(&bettingPlayer, bet, testTime)
	if err == nil {
		t.Errorf("Expected error for non-existing match")
	}
}

func TestAddSeveralsPlayerBetsGetMatchBets(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	bettingPlayer1 := players[0]
	bettingPlayer2 := players[1]
	game.AddPlayerBet(&bettingPlayer1, bet1, testTime)
	game.AddPlayerBet(&bettingPlayer2, bet2, testTime)

	bet_map, err := game.GetMatchBets(match)
	if err != nil {
		t.Errorf("Error retrieving match bets: %v", err)
	}
	playerBet1 := bet_map[bettingPlayer1]
	playerBet2 := bet_map[bettingPlayer2]
	if playerBet1 != bet1 {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet1, playerBet1)
	}
	if playerBet2 != bet2 {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet2, playerBet2)
	}
}

func TestAddPlayerBetMatchStarted(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime.Add(-1*time.Hour), 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	err := game.AddPlayerBet(&bettingPlayer, bet, testTime)
	if err == nil {
		t.Errorf("Expected error for match started")
	}
}

func TestGetMatchBetsNonExistingMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	game.AddPlayerBet(&bettingPlayer, bet, testTime)

	fake_match := models.NewSeasonMatch("Sardinia", "Corsica", "2024", "Corsica Cup", testTime, 1)
	_, err := game.GetMatchBets(fake_match)
	if err == nil {
		t.Errorf("Expected error for non-existing match")
	}
}

func TestAddFinishedMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Add bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)

	// Finish the match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	matchScores, totalPoints, err := game.AddFinishedMatch(finishedMatch)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check match scores
	if matchScores[players[0]] != 500 { // Good bet
		t.Errorf("Expected 500 points for Player1 in match, got %d", matchScores[players[0]])
	}
	if matchScores[players[1]] != 0 { // Wrong bet
		t.Errorf("Expected 0 points for Player2 in match, got %d", matchScores[players[1]])
	}

	// Check total points
	if totalPoints[players[0]] != 500 {
		t.Errorf("Expected 500 total points for Player1, got %d", totalPoints[players[0]])
	}
	if totalPoints[players[1]] != 0 {
		t.Errorf("Expected 0 total points for Player2, got %d", totalPoints[players[1]])
	}
}

func TestAddFinishedMatch_NonExistingMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Test with non-existing match
	nonExistingMatch := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", testTime, 1)
	_, _, err := game.AddFinishedMatch(nonExistingMatch)
	if err == nil {
		t.Error("Expected error for non-existing match")
	}

	// Test with unfinished match
	_, _, err = game.AddFinishedMatch(match)
	if err == nil {
		t.Error("Expected error for unfinished match")
	}
}

func TestAddFinishedMatch_AlreadyFinishedMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}}
	match := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	_, _, err := game.AddFinishedMatch(match)
	if err == nil {
		t.Error("Expected error for already finished match")
	}
}

func TestGetIncomingMatches(t *testing.T) {
	players := []models.Player{{Name: "Player1"}}
	matches := []models.Match{
		models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1),
		models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", testTime, 1),
		models.NewFinishedSeasonMatch("Team5", "Team6", 2, 1, "2024", "Premier League", testTime.Add(-24*time.Hour), 1, 1.0, 2.0, 3.0),
	}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	incomingMatches := game.GetIncomingMatches()
	if len(incomingMatches) != 2 {
		t.Errorf("Expected 2 incoming matches, got %d", len(incomingMatches))
	}
}

func TestGetPlayersPoints(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

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
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)

	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	matchScores, totalPoints, err := game.AddFinishedMatch(finishedMatch)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check match scores
	if matchScores[players[0]] != 500 {
		t.Errorf("Expected 500 points for Player1 in match, got %d", matchScores[players[0]])
	}
	if matchScores[players[1]] != 0 {
		t.Errorf("Expected 0 points for Player2 in match, got %d", matchScores[players[1]])
	}

	// Check total points
	if totalPoints[players[0]] != 500 {
		t.Errorf("Expected 500 total points for Player1, got %d", totalPoints[players[0]])
	}
	if totalPoints[players[1]] != 0 {
		t.Errorf("Expected 0 total points for Player2, got %d", totalPoints[players[1]])
	}
}

func TestGetWinner_SingleWinner(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}, {Name: "Player3"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Add bets for all players
	bet1 := models.NewBet(match, 2, 1) // Correct bet
	bet2 := models.NewBet(match, 1, 1) // Wrong bet
	bet3 := models.NewBet(match, 0, 2) // Wrong bet
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)
	game.AddPlayerBet(&players[2], bet3, testTime)

	// Finish the match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	matchScores, totalPoints, err := game.AddFinishedMatch(finishedMatch)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check match scores
	if matchScores[players[0]] != 500 {
		t.Errorf("Expected 500 points for Player1 in match, got %d", matchScores[players[0]])
	}
	if matchScores[players[1]] != 0 {
		t.Errorf("Expected 0 points for Player2 in match, got %d", matchScores[players[1]])
	}
	if matchScores[players[2]] != 0 {
		t.Errorf("Expected 0 points for Player3 in match, got %d", matchScores[players[2]])
	}

	// Check total points
	if totalPoints[players[0]] != 500 {
		t.Errorf("Expected 500 total points for Player1, got %d", totalPoints[players[0]])
	}
	if totalPoints[players[1]] != 0 {
		t.Errorf("Expected 0 total points for Player2, got %d", totalPoints[players[1]])
	}
	if totalPoints[players[2]] != 0 {
		t.Errorf("Expected 0 total points for Player3, got %d", totalPoints[players[2]])
	}

	winners := game.GetWinner()
	if len(winners) != 1 {
		t.Errorf("Expected 1 winner, got %d", len(winners))
	}
	if winners[0] != players[0] {
		t.Errorf("Expected Player1 to be the winner, got %v", winners[0])
	}
}

func TestGetWinner_MultipleWinners(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}, {Name: "Player3"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Add bets for all players
	bet1 := models.NewBet(match, 2, 1) // Correct bet
	bet2 := models.NewBet(match, 2, 1) // Correct bet
	bet3 := models.NewBet(match, 0, 2) // Wrong bet
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)
	game.AddPlayerBet(&players[2], bet3, testTime)

	// Finish the match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	matchScores, totalPoints, err := game.AddFinishedMatch(finishedMatch)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check match scores
	if matchScores[players[0]] != 500 {
		t.Errorf("Expected 500 points for Player1 in match, got %d", matchScores[players[0]])
	}
	if matchScores[players[1]] != 500 {
		t.Errorf("Expected 500 points for Player2 in match, got %d", matchScores[players[1]])
	}
	if matchScores[players[2]] != 0 {
		t.Errorf("Expected 0 points for Player3 in match, got %d", matchScores[players[2]])
	}

	// Check total points
	if totalPoints[players[0]] != 500 {
		t.Errorf("Expected 500 total points for Player1, got %d", totalPoints[players[0]])
	}
	if totalPoints[players[1]] != 500 {
		t.Errorf("Expected 500 total points for Player2, got %d", totalPoints[players[1]])
	}
	if totalPoints[players[2]] != 0 {
		t.Errorf("Expected 0 total points for Player3, got %d", totalPoints[players[2]])
	}

	winners := game.GetWinner()
	if len(winners) != 2 {
		t.Errorf("Expected 2 winners, got %d", len(winners))
	}
	// Check that both winners are in the slice
	winner1Found := false
	winner2Found := false
	for _, winner := range winners {
		if winner == players[0] {
			winner1Found = true
		}
		if winner == players[1] {
			winner2Found = true
		}
	}
	if !winner1Found || !winner2Found {
		t.Errorf("Expected both Player1 and Player2 to be winners")
	}
}

func TestGetWinner_NoWinners(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Add bets for both players
	bet1 := models.NewBet(match, 1, 1) // Wrong bet
	bet2 := models.NewBet(match, 0, 2) // Wrong bet
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)

	// Finish the match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	matchScores, totalPoints, err := game.AddFinishedMatch(finishedMatch)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check match scores
	if matchScores[players[0]] != 0 {
		t.Errorf("Expected 0 points for Player1 in match, got %d", matchScores[players[0]])
	}
	if matchScores[players[1]] != 0 {
		t.Errorf("Expected 0 points for Player2 in match, got %d", matchScores[players[1]])
	}

	// Check total points
	if totalPoints[players[0]] != 0 {
		t.Errorf("Expected 0 total points for Player1, got %d", totalPoints[players[0]])
	}
	if totalPoints[players[1]] != 0 {
		t.Errorf("Expected 0 total points for Player2, got %d", totalPoints[players[1]])
	}

	winners := game.GetWinner()
	if len(winners) != 2 {
		t.Errorf("Expected both players to be winners with 0 points, got %d winners", len(winners))
	}
	// Check that both players are in the slice
	player1Found := false
	player2Found := false
	for _, winner := range winners {
		if winner == players[0] {
			player1Found = true
		}
		if winner == players[1] {
			player2Found = true
		}
	}
	if !player1Found || !player2Found {
		t.Errorf("Expected both players to be winners with 0 points")
	}
}

func TestGetWinner_MultipleMatches(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}, {Name: "Player3"}}
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", testTime.Add(24*time.Hour), 2)
	matches := []models.Match{match1, match2}
	scorer := &ScorerTest{}

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Add bets for first match
	bet1 := models.NewBet(match1, 2, 1) // Correct bet
	bet2 := models.NewBet(match1, 1, 1) // Wrong bet
	bet3 := models.NewBet(match1, 0, 2) // Wrong bet
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)
	game.AddPlayerBet(&players[2], bet3, testTime)

	// Finish first match
	finishedMatch1 := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	matchScores1, _, err := game.AddFinishedMatch(finishedMatch1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check first match scores
	if matchScores1[players[0]] != 500 {
		t.Errorf("Expected 500 points for Player1 in first match, got %d", matchScores1[players[0]])
	}
	if matchScores1[players[1]] != 0 {
		t.Errorf("Expected 0 points for Player2 in first match, got %d", matchScores1[players[1]])
	}
	if matchScores1[players[2]] != 0 {
		t.Errorf("Expected 0 points for Player3 in first match, got %d", matchScores1[players[2]])
	}

	// Add bets for second match
	bet4 := models.NewBet(match2, 1, 1) // Wrong bet
	bet5 := models.NewBet(match2, 2, 1) // Correct bet
	bet6 := models.NewBet(match2, 2, 1) // Correct bet
	game.AddPlayerBet(&players[0], bet4, testTime)
	game.AddPlayerBet(&players[1], bet5, testTime)
	game.AddPlayerBet(&players[2], bet6, testTime)

	// Finish second match
	finishedMatch2 := models.NewFinishedSeasonMatch("Team3", "Team4", 2, 1, "2024", "Premier League", testTime.Add(24*time.Hour), 2, 1.0, 2.0, 3.0)
	matchScores2, totalPoints2, err := game.AddFinishedMatch(finishedMatch2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check second match scores
	if matchScores2[players[0]] != 0 {
		t.Errorf("Expected 0 points for Player1 in second match, got %d", matchScores2[players[0]])
	}
	if matchScores2[players[1]] != 500 {
		t.Errorf("Expected 500 points for Player2 in second match, got %d", matchScores2[players[1]])
	}
	if matchScores2[players[2]] != 500 {
		t.Errorf("Expected 500 points for Player3 in second match, got %d", matchScores2[players[2]])
	}

	// Check final total points
	if totalPoints2[players[0]] != 500 {
		t.Errorf("Expected 500 total points for Player1, got %d", totalPoints2[players[0]])
	}
	if totalPoints2[players[1]] != 500 {
		t.Errorf("Expected 500 total points for Player2, got %d", totalPoints2[players[1]])
	}
	if totalPoints2[players[2]] != 500 {
		t.Errorf("Expected 500 total points for Player3, got %d", totalPoints2[players[2]])
	}

	winners := game.GetWinner()
	if len(winners) != 2 {
		t.Errorf("Expected 2 winners, got %d", len(winners))
	}
	// Check that both winners are in the slice
	winner1Found := false
	winner2Found := false
	for _, winner := range winners {
		if winner == players[1] {
			winner1Found = true
		}
		if winner == players[2] {
			winner2Found = true
		}
	}
	if !winner1Found || !winner2Found {
		t.Errorf("Expected both Player2 and Player3 to be winners with 500 points each")
	}
}
