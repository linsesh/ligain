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

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	if game.GetSeasonYear() != "2024" {
		t.Errorf("Expected season code '2024', got %s", game.GetSeasonYear())
	}
	if game.GetCompetitionName() != "Premier League" {
		t.Errorf("Expected competition code 'Premier League', got %s", game.GetCompetitionName())
	}
	if game.GetGameStatus() != models.GameStatusNotStarted {
		t.Errorf("Expected game status 'not started', got %s", game.GetGameStatus())
	}
	if len(game.GetIncomingMatches()) != 2 {
		t.Errorf("Expected 2 incoming matches, got %d", len(game.GetIncomingMatches()))
	}

	// Finish the first match
	matches[0].Finish(2, 1)
	scores, err := game.CalculateMatchScores(matches[0])
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	game.ApplyMatchScores(matches[0], scores)

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

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	err := game.CheckPlayerBetValidity(bettingPlayer, bet, testTime)
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

	match.Finish(2, 1)

	for i, player := range players {
		game.AddPlayerBet(player, bets[i])
	}

	scores, err := game.CalculateMatchScores(match)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	game.ApplyMatchScores(match, scores)

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

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	err := game.CheckPlayerBetValidity(bettingPlayer, bet, testTime)
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
	err = game.CheckPlayerBetValidity(bettingPlayer, updatedBet, testTime)
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

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := models.Player{Name: "Player3"}
	err := game.CheckPlayerBetValidity(bettingPlayer, bet, testTime)
	if err == nil {
		t.Errorf("Expected error for non-existing player")
	}
}

func TestAddPlayerBetNonExistingMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	fake_match := models.NewSeasonMatch("Sardinia", "Corsica", "2024", "Corsica Cup", testTime, 1)
	bet := models.NewBet(fake_match, 2, 1)
	bettingPlayer := players[0]
	err := game.CheckPlayerBetValidity(bettingPlayer, bet, testTime)
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

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	bettingPlayer1 := players[0]
	bettingPlayer2 := players[1]

	err := game.CheckPlayerBetValidity(bettingPlayer1, bet1, testTime)
	if err != nil {
		t.Errorf("Error checking bet1 validity: %v", err)
	}
	_, err = betRepo.SaveBet("2024", bet1, bettingPlayer1)
	if err != nil {
		t.Errorf("Error saving bet1: %v", err)
	}

	err = game.CheckPlayerBetValidity(bettingPlayer2, bet2, testTime)
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

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	bet := models.NewBet(match, 2, 1)
	bettingPlayer := players[0]
	err := game.CheckPlayerBetValidity(bettingPlayer, bet, testTime)
	if err == nil {
		t.Errorf("Expected error for match started")
	}
}

func TestAddFinishedMatch(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	// Add bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)

	err := game.CheckPlayerBetValidity(players[0], bet1, testTime)
	if err != nil {
		t.Errorf("Error checking bet1 validity: %v", err)
	}
	err = game.AddPlayerBet(players[0], bet1)
	if err != nil {
		t.Errorf("Error adding bet1: %v", err)
	}

	err = game.CheckPlayerBetValidity(players[1], bet2, testTime)
	if err != nil {
		t.Errorf("Error checking bet2 validity: %v", err)
	}
	err = game.AddPlayerBet(players[1], bet2)
	if err != nil {
		t.Errorf("Error adding bet2: %v", err)
	}

	match.Finish(2, 1)
	scores, err := game.CalculateMatchScores(match)
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
	game.ApplyMatchScores(match, scores)

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

func TestNewStartedGame(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}

	// Create incoming match
	incomingMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	incomingMatches := []models.Match{incomingMatch}

	// Create past matches
	pastMatch1 := models.NewFinishedSeasonMatch("Team3", "Team4", 2, 1, "2024", "Premier League", testTime.Add(-24*time.Hour), 2, 1.0, 2.0, 3.0)
	pastMatch2 := models.NewFinishedSeasonMatch("Team5", "Team6", 0, 0, "2024", "Premier League", testTime.Add(-48*time.Hour), 3, 1.0, 2.0, 3.0)
	pastMatches := []models.Match{pastMatch1, pastMatch2}

	scorer := &ScorerTest{}

	// Initialize bets map
	bets := make(map[string]map[models.Player]*models.Bet)
	bets[incomingMatch.Id()] = make(map[models.Player]*models.Bet)
	bets[pastMatch1.Id()] = make(map[models.Player]*models.Bet)
	bets[pastMatch2.Id()] = make(map[models.Player]*models.Bet)

	// Add bets for incoming match
	incomingBet1 := models.NewBet(incomingMatch, 2, 1)
	incomingBet2 := models.NewBet(incomingMatch, 1, 1)
	bets[incomingMatch.Id()][players[0]] = incomingBet1
	bets[incomingMatch.Id()][players[1]] = incomingBet2

	// Add bets for past matches
	pastBet1 := models.NewBet(pastMatch1, 2, 1) // Correct prediction
	pastBet2 := models.NewBet(pastMatch1, 1, 1) // Wrong prediction
	pastBet3 := models.NewBet(pastMatch2, 0, 0) // Correct prediction
	pastBet4 := models.NewBet(pastMatch2, 1, 1) // Wrong prediction

	bets[pastMatch1.Id()][players[0]] = pastBet1
	bets[pastMatch1.Id()][players[1]] = pastBet2
	bets[pastMatch2.Id()][players[0]] = pastBet3
	bets[pastMatch2.Id()][players[1]] = pastBet4

	// Initialize scores map
	scores := make(map[string]map[models.Player]int)
	scores[pastMatch1.Id()] = make(map[models.Player]int)
	scores[pastMatch2.Id()] = make(map[models.Player]int)

	// Add scores for past matches
	scores[pastMatch1.Id()][players[0]] = 500 // Correct prediction
	scores[pastMatch1.Id()][players[1]] = 0   // Wrong prediction
	scores[pastMatch2.Id()][players[0]] = 500 // Correct prediction
	scores[pastMatch2.Id()][players[1]] = 0   // Wrong prediction

	game := NewStartedGame("2024", "Premier League", players, incomingMatches, pastMatches, scorer, bets, scores)

	// Test basic game properties
	if game.GetSeasonYear() != "2024" {
		t.Errorf("Expected season code '2024', got %s", game.GetSeasonYear())
	}
	if game.GetCompetitionName() != "Premier League" {
		t.Errorf("Expected competition code 'Premier League', got %s", game.GetCompetitionName())
	}

	// Test incoming matches
	incomingMatchesResult := game.GetIncomingMatches()
	if len(incomingMatchesResult) != 1 {
		t.Errorf("Expected 1 incoming match, got %d", len(incomingMatchesResult))
	}
	incomingMatchResult := incomingMatchesResult[incomingMatch.Id()]
	if incomingMatchResult.Match != incomingMatch {
		t.Errorf("Expected incoming match %v, got %v", incomingMatch, incomingMatchResult.Match)
	}
	if len(incomingMatchResult.Bets) != 2 {
		t.Errorf("Expected 2 bets for incoming match, got %d", len(incomingMatchResult.Bets))
	}

	// Test past results
	pastResults := game.GetPastResults()
	if len(pastResults) != 2 {
		t.Errorf("Expected 2 past matches, got %d", len(pastResults))
	}

	// Verify past match 1
	pastResult1 := pastResults[pastMatch1.Id()]
	if pastResult1.Match != pastMatch1 {
		t.Errorf("Expected past match 1 %v, got %v", pastMatch1, pastResult1.Match)
	}
	if len(pastResult1.Bets) != 2 {
		t.Errorf("Expected 2 bets for past match 1, got %d", len(pastResult1.Bets))
	}

	// Verify past match 2
	pastResult2 := pastResults[pastMatch2.Id()]
	if pastResult2.Match != pastMatch2 {
		t.Errorf("Expected past match 2 %v, got %v", pastMatch2, pastResult2.Match)
	}
	if len(pastResult2.Bets) != 2 {
		t.Errorf("Expected 2 bets for past match 2, got %d", len(pastResult2.Bets))
	}

	// Test total points
	points := game.GetPlayersPoints()
	if len(points) != 2 {
		t.Errorf("Expected points for 2 players, got %d", len(points))
	}
	if points[players[0]] != 1000 {
		t.Errorf("Expected 1000 points for Player1 (500 for each correct prediction), got %d", points[players[0]])
	}
	if points[players[1]] != 0 {
		t.Errorf("Expected 0 points for Player2 (both predictions wrong), got %d", points[players[1]])
	}
}

func TestGetPastResults(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	// Add bets for both players
	bet1 := models.NewBet(match, 2, 1)
	bet2 := models.NewBet(match, 1, 1)
	game.AddPlayerBet(players[0], bet1)
	game.AddPlayerBet(players[1], bet2)

	// Calculate and apply scores
	match.Finish(2, 1)
	scores, err := game.CalculateMatchScores(match)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	game.ApplyMatchScores(match, scores)

	// Get past results
	pastResults := game.GetPastResults()
	if len(pastResults) != 1 {
		t.Errorf("Expected 1 past result, got %d", len(pastResults))
	}

	result, exists := pastResults[match.Id()]
	if !exists {
		t.Errorf("Expected to find match %s in past results", match.Id())
	}

	// Verify the result contains the correct match and bets
	if result.Match != match {
		t.Errorf("Expected match %v, got %v", match, result.Match)
	}
	if len(result.Bets) != 2 {
		t.Errorf("Expected 2 bets, got %d", len(result.Bets))
	}
}

func TestGetPlayersPoints(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", testTime.Add(24*time.Hour), 2)
	matches := []models.Match{match1, match2}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	// Add bets for both players on both matches
	bet1 := models.NewBet(match1, 2, 1) // Correct prediction
	bet2 := models.NewBet(match1, 1, 1) // Wrong prediction
	bet3 := models.NewBet(match2, 1, 1) // Correct prediction
	bet4 := models.NewBet(match2, 2, 0) // Wrong prediction

	game.AddPlayerBet(players[0], bet1)
	game.AddPlayerBet(players[1], bet2)
	game.AddPlayerBet(players[0], bet3)
	game.AddPlayerBet(players[1], bet4)

	// Calculate and apply scores for both matches
	match1.Finish(2, 1)
	scores1, _ := game.CalculateMatchScores(match1)
	game.ApplyMatchScores(match1, scores1)
	match2.Finish(1, 1)
	scores2, _ := game.CalculateMatchScores(match2)
	game.ApplyMatchScores(match2, scores2)

	// Get total points
	points := game.GetPlayersPoints()
	if len(points) != 2 {
		t.Errorf("Expected points for 2 players, got %d", len(points))
	}

	// Player1 should have 1000 points (500 for each correct prediction)
	if points[players[0]] != 1000 {
		t.Errorf("Expected 1000 points for Player1, got %d", points[players[0]])
	}
	// Player2 should have 0 points (both predictions wrong)
	if points[players[1]] != 0 {
		t.Errorf("Expected 0 points for Player2, got %d", points[players[1]])
	}
}

func TestGetWinner(t *testing.T) {
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}, {Name: "Player3"}}
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matches := []models.Match{match}
	scorer := &ScorerTest{}

	game := NewFreshGame("2024", "Premier League", players, matches, scorer)

	// Add bets for all players
	bet1 := models.NewBet(match, 2, 1) // Correct prediction
	bet2 := models.NewBet(match, 2, 1) // Correct prediction
	bet3 := models.NewBet(match, 1, 1) // Wrong prediction

	game.AddPlayerBet(players[0], bet1)
	game.AddPlayerBet(players[1], bet2)
	game.AddPlayerBet(players[2], bet3)

	// Calculate and apply scores
	match.Finish(2, 1)
	scores, _ := game.CalculateMatchScores(match)
	game.ApplyMatchScores(match, scores)

	// Get winners
	winners := game.GetWinner()
	if len(winners) != 2 {
		t.Errorf("Expected 2 winners, got %d", len(winners))
	}

	// Verify the correct players are winners
	winnerNames := make(map[string]bool)
	for _, winner := range winners {
		winnerNames[winner.Name] = true
	}
	if !winnerNames["Player1"] || !winnerNames["Player2"] {
		t.Errorf("Expected Player1 and Player2 to be winners")
	}
	if winnerNames["Player3"] {
		t.Errorf("Expected Player3 to not be a winner")
	}
}
