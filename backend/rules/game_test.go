package rules

import (
	"liguain/backend/models"
	"liguain/backend/repositories"
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

	if game.GetSeasonYear() != "2024" {
		t.Errorf("Expected season code '2024', got %s", game.GetSeasonYear())
	}
	if game.GetCompetitionName() != "Premier League" {
		t.Errorf("Expected competition code 'Premier League', got %s", game.GetCompetitionName())
	}
	if game.GetGameStatus() != GameStatusNotStarted {
		t.Errorf("Expected game status 'not started', got %s", game.GetGameStatus())
	}
	if len(game.GetIncomingMatches()) != 2 {
		t.Errorf("Expected 2 incoming matches, got %d", len(game.GetIncomingMatches()))
	}

	// Finish the first match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	bets := make(map[models.Player]*models.Bet)
	scores, err := game.CalculateMatchScores(finishedMatch, bets)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	game.ApplyMatchScores(finishedMatch, scores)

	// Verify game is finished after all matches are played
	if game.IsFinished() {
		t.Errorf("Expected game to not be finished after only a fraction of matches were played")
	}
}

func TestAddPlayerBetGetMatchBets(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}
	betRepo := repositories.NewInMemoryBetRepository()

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	err := game.CheckPlayerBetValidity(&bettingPlayer, bet, testTime)
	if err != nil {
		t.Errorf("Error checking bet validity: %v", err)
	}
	_, err = betRepo.SaveBet("2024", bet, bettingPlayer)
	if err != nil {
		t.Errorf("Error saving bet: %v", err)
	}

	bets, players, err := betRepo.GetBetsForMatch(match, "2024")
	if err != nil {
		t.Errorf("Error retrieving match bets: %v", err)
	}
	if len(bets) != 1 {
		t.Errorf("Expected 1 bet, got %d", len(bets))
	}
	if bets[0] != bet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet, bets[0])
	}

	// Finish the match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)

	// Create bets map
	betsMap := make(map[models.Player]*models.Bet)
	for i, player := range players {
		betsMap[player] = bets[i]
	}

	scores, err := game.CalculateMatchScores(finishedMatch, betsMap)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	game.ApplyMatchScores(finishedMatch, scores)

	if !game.IsFinished() {
		t.Errorf("Expected game to be finished after all matches are played")
	}
}

func TestAddPlayerBetUpdateBet(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}
	betRepo := repositories.NewInMemoryBetRepository()

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	err := game.CheckPlayerBetValidity(&bettingPlayer, bet, testTime)
	if err != nil {
		t.Errorf("Error checking bet validity: %v", err)
	}
	_, err = betRepo.SaveBet("2024", bet, bettingPlayer)
	if err != nil {
		t.Errorf("Error saving bet: %v", err)
	}

	bets, players, err := betRepo.GetBetsForMatch(match, "2024")
	if err != nil {
		t.Errorf("Error retrieving match bets: %v", err)
	}
	if len(bets) != 1 {
		t.Errorf("Expected 1 bet, got %d", len(bets))
	}
	if bets[0] != bet {
		t.Errorf("Retrieved bet is not the same as the one added, expected %v, got %v", bet, bets[0])
	}

	updatedBet := models.NewBet(match, 1, 2)
	err = game.CheckPlayerBetValidity(&bettingPlayer, updatedBet, testTime)
	if err != nil {
		t.Errorf("Error checking updated bet validity: %v", err)
	}
	_, err = betRepo.SaveBet("2024", updatedBet, bettingPlayer)
	if err != nil {
		t.Errorf("Error saving updated bet: %v", err)
	}

	bets, players, err = betRepo.GetBetsForMatch(match, "2024")
	if err != nil {
		t.Errorf("Error retrieving match bets: %v", err)
	}
	if len(bets) != 1 {
		t.Errorf("Expected 1 bet after update, got %d", len(bets))
	}
	if bets[0] != updatedBet {
		t.Errorf("Retrieved bet is not the same as the updated one, expected %v, got %v", updatedBet, bets[0])
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
	err := game.CheckPlayerBetValidity(&bettingPlayer, bet, testTime)
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
	err := game.CheckPlayerBetValidity(&bettingPlayer, bet, testTime)
	if err == nil {
		t.Errorf("Expected error for non-existing match")
	}
}

func TestAddSeveralsPlayerBetsGetMatchBets(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}
	betRepo := repositories.NewInMemoryBetRepository()

	game := NewGame("2024", "Premier League", players, matches, scorer)

	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	bettingPlayer1 := players[0]
	bettingPlayer2 := players[1]

	err := game.CheckPlayerBetValidity(&bettingPlayer1, bet1, testTime)
	if err != nil {
		t.Errorf("Error checking bet1 validity: %v", err)
	}
	_, err = betRepo.SaveBet("2024", bet1, bettingPlayer1)
	if err != nil {
		t.Errorf("Error saving bet1: %v", err)
	}

	err = game.CheckPlayerBetValidity(&bettingPlayer2, bet2, testTime)
	if err != nil {
		t.Errorf("Error checking bet2 validity: %v", err)
	}
	_, err = betRepo.SaveBet("2024", bet2, bettingPlayer2)
	if err != nil {
		t.Errorf("Error saving bet2: %v", err)
	}

	bets, players, err := betRepo.GetBetsForMatch(match, "2024")
	if err != nil {
		t.Errorf("Error retrieving match bets: %v", err)
	}
	if len(bets) != 2 {
		t.Errorf("Expected 2 bets, got %d", len(bets))
	}

	// Verify both bets are present
	betMap := make(map[string]*models.Bet)
	for i, player := range players {
		betMap[player.Name] = bets[i]
	}

	if betMap["Player1"] != bet1 {
		t.Errorf("Retrieved bet1 is not the same as the one added, expected %v, got %v", bet1, betMap["Player1"])
	}
	if betMap["Player2"] != bet2 {
		t.Errorf("Retrieved bet2 is not the same as the one added, expected %v, got %v", bet2, betMap["Player2"])
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
	err := game.CheckPlayerBetValidity(&bettingPlayer, bet, testTime)
	if err == nil {
		t.Errorf("Expected error for match started")
	}
}

func TestAddFinishedMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}
	betRepo := repositories.NewInMemoryBetRepository()

	game := NewGame("2024", "Premier League", players, matches, scorer)

	// Add bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)

	err := game.CheckPlayerBetValidity(&players[0], bet1, testTime)
	if err != nil {
		t.Errorf("Error checking bet1 validity: %v", err)
	}
	_, err = betRepo.SaveBet("2024", bet1, players[0])
	if err != nil {
		t.Errorf("Error saving bet1: %v", err)
	}

	err = game.CheckPlayerBetValidity(&players[1], bet2, testTime)
	if err != nil {
		t.Errorf("Error checking bet2 validity: %v", err)
	}
	_, err = betRepo.SaveBet("2024", bet2, players[1])
	if err != nil {
		t.Errorf("Error saving bet2: %v", err)
	}

	// Finish the match
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0)
	bets, players, err := betRepo.GetBetsForMatch(finishedMatch, "2024")
	if err != nil {
		t.Errorf("Error getting bets for match: %v", err)
	}

	// Create bets map
	betsMap := make(map[models.Player]*models.Bet)
	for i, player := range players {
		betsMap[player] = bets[i]
	}

	scores, err := game.CalculateMatchScores(finishedMatch, betsMap)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check match scores
	if scores[players[0]] != 500 { // Good bet
		t.Errorf("Expected 500 points for Player1 in match, got %d", scores[players[0]])
	}
	if scores[players[1]] != 0 { // Wrong bet
		t.Errorf("Expected 0 points for Player2 in match, got %d", scores[players[1]])
	}

	// Apply scores
	game.ApplyMatchScores(finishedMatch, scores)

	// Check total points
	points := game.GetPlayersPoints()
	if points[players[0]] != 500 {
		t.Errorf("Expected 500 total points for Player1, got %d", points[players[0]])
	}
	if points[players[1]] != 0 {
		t.Errorf("Expected 0 total points for Player2, got %d", points[players[1]])
	}

	// Verify game is finished after all matches are played
	if !game.IsFinished() {
		t.Errorf("Expected game to be finished after all matches are played")
	}
}
