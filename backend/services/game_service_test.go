package services

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/rules"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

var matchTime = time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

// Mock implementations
type GameRepositoryMock struct{}

func (r *GameRepositoryMock) CreateGame(game models.Game) (string, error) {
	return "test-game-id", nil
}

func (r *GameRepositoryMock) GetGame(gameId string) (models.Game, error) {
	return nil, nil // Not used in tests
}

func (r *GameRepositoryMock) SaveWithId(gameId string, game models.Game) error {
	return nil
}

type MatchWatcherServiceMock struct {
	updates []map[string]models.Match
	index   int
}

func NewMatchWatcherServiceMock(updates []map[string]models.Match) *MatchWatcherServiceMock {
	return &MatchWatcherServiceMock{
		updates: updates,
		index:   0,
	}
}

func (m *MatchWatcherServiceMock) GetUpdates(ctx context.Context, done chan MatchWatcherServiceResult) {
	var result MatchWatcherServiceResult
	//log.Infof("Getting updates for match %v", m.index)
	if m.index >= len(m.updates) {
		result = MatchWatcherServiceResult{Value: make(map[string]models.Match), Err: nil}
	} else {
		update := m.updates[m.index]
		m.index++
		log.Infof("Sending updates for match %v", m.index-1)
		result = MatchWatcherServiceResult{Value: update, Err: nil}
	}
	select {
	case <-ctx.Done():
		log.Errorf("The GetUpdates function failed to send the result")
	case done <- result:
	}
}

func (m *MatchWatcherServiceMock) WatchMatches(matches []models.Match) {
	// Not used in tests
}

type ScorerMock struct{}

func (s *ScorerMock) Score(match models.Match, bets []*models.Bet) []int {
	scores := make([]int, len(bets))
	log.Infof("Given match: %v", match)
	for i, bet := range bets {
		// Create a new bet with the same predictions but the finished match
		finishedBet := models.NewBet(match, bet.PredictedHomeGoals, bet.PredictedAwayGoals)
		if finishedBet.IsBetCorrect() {
			log.Infof("Correct bet: %v", finishedBet)
			scores[i] = 500
		} else {
			log.Infof("Wrong bet: %v", finishedBet)
			scores[i] = 0
		}
	}
	return scores
}

type MockGame struct {
	models.Game
}

func (g *MockGame) CalculateMatchScores(match models.Match, bets map[models.Player]*models.Bet) (map[models.Player]int, error) {
	scores := make(map[models.Player]int)
	for player, bet := range bets {
		if bet.PredictedHomeGoals == 2 && bet.PredictedAwayGoals == 1 {
			scores[player] = 3 // Correct prediction
		} else {
			scores[player] = 0 // Wrong prediction
		}
	}
	return scores, nil
}

// Test cases
func TestGameService_Play_SingleMatch(t *testing.T) {
	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	// Create a game
	game := rules.NewFreshGame("2024", "Premier League", players, matches, &ScorerMock{})

	// Setup mock updates
	updates := []map[string]models.Match{
		{
			match.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0),
		},
	}

	// Create service with mocks
	repo := &GameRepositoryMock{}
	scoresRepo := repositories.NewInMemoryScoresRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, scoresRepo, betRepo, NewMatchWatcherServiceMock(updates), 10*time.Millisecond)

	// Add some bets
	goodBet := models.NewBet(match, 2, 1)  // Correct good result
	wrongBet := models.NewBet(match, 1, 1) // Wrong result
	service.UpdatePlayerBet(players[0], goodBet, matchTime.Add(-1*time.Second))
	service.UpdatePlayerBet(players[1], wrongBet, matchTime.Add(-1*time.Second))

	// Play the game with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	done := make(chan struct{})
	var winners []models.Player
	var playErr error

	go func() {
		winners, playErr = service.Play()
		close(done)
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Play function timed out after 1 second")
	case <-done:
		if playErr != nil {
			t.Fatalf("Failed to play game: %v", playErr)
		}
	}

	if len(winners) != 1 {
		t.Errorf("Expected 1 winner, got %d", len(winners))
	}
	if winners[0].Name != "Player1" {
		t.Errorf("Expected Player1 to win, got %s", winners[0].Name)
	}

	if !game.IsFinished() {
		t.Errorf("Expected game to be finished after all matches are played")
	}
}

func TestGameService_Play_MultipleMatches(t *testing.T) {
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Corsica Championship", matchTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Corsica Championship", matchTime.Add(time.Hour), 1)
	match3 := models.NewSeasonMatch("Team4", "Team5", "2024", "Corsica Championship", matchTime.Add(time.Hour), 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match1, match2, match3}

	game := rules.NewFreshGame("2024", "Corsica Championship", players, matches, &ScorerMock{})

	updates := []map[string]models.Match{
		{
			match1.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Corsica Championship", matchTime, 1, 1.0, 2.0, 3.0),
		},
		{
			match2.Id(): models.NewFinishedSeasonMatch("Team3", "Team4", 2, 1, "2024", "Corsica Championship", matchTime.Add(time.Hour), 1, 1.0, 2.0, 3.0),
		},
		{
			match3.Id(): models.NewFinishedSeasonMatch("Team4", "Team5", 1, 1, "2024", "Corsica Championship", matchTime.Add(time.Hour), 1, 1.0, 2.0, 3.0),
		},
	}

	repo := &GameRepositoryMock{}
	scoresRepo := repositories.NewInMemoryScoresRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, scoresRepo, betRepo, NewMatchWatcherServiceMock(updates), 10*time.Millisecond)

	// Add some bets
	good_bet_match_1 := models.NewBet(match1, 2, 1)
	wrong_bet_match_1 := models.NewBet(match1, 1, 1)
	good_bet_match_2 := models.NewBet(match2, 2, 1)
	wrong_bet_match_2 := models.NewBet(match2, 1, 1)
	good_bet_match_3 := models.NewBet(match3, 1, 1)
	wrong_bet_match_3 := models.NewBet(match3, 2, 1)
	service.UpdatePlayerBet(players[0], good_bet_match_1, matchTime.Add(-1*time.Second))
	service.UpdatePlayerBet(players[1], wrong_bet_match_1, matchTime.Add(-1*time.Second))
	service.UpdatePlayerBet(players[1], good_bet_match_2, matchTime)
	service.UpdatePlayerBet(players[0], wrong_bet_match_2, matchTime)
	service.UpdatePlayerBet(players[1], good_bet_match_3, matchTime)
	service.UpdatePlayerBet(players[0], wrong_bet_match_3, matchTime)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	done := make(chan struct{})
	var winners []models.Player
	var playErr error

	go func() {
		winners, playErr = service.Play()
		close(done)
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Play function timed out after 1 second")
	case <-done:
		if playErr != nil {
			t.Fatalf("Failed to play game: %v", playErr)
		}
	}

	if len(winners) != 1 {
		t.Errorf("Expected 1 winner, got %d", len(winners))
	}
	if winners[0].Name != "Player2" {
		t.Errorf("Expected Player2 to win, got %s", winners[0].Name)
	}

	if !game.IsFinished() {
		t.Errorf("Expected game to be finished after all matches are played")
	}
}

func TestGameService_Play_NoWinner(t *testing.T) {
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Corsica Championship", matchTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Corsica Championship", matchTime.Add(time.Hour), 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match1, match2}

	game := rules.NewFreshGame("2024", "Corsica Championship", players, matches, &ScorerMock{})

	updates := []map[string]models.Match{
		{
			match1.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Corsica Championship", matchTime, 1, 1.0, 2.0, 3.0),
		},
		{
			match2.Id(): models.NewFinishedSeasonMatch("Team3", "Team4", 1, 2, "2024", "Corsica Championship", matchTime.Add(time.Hour), 1, 1.0, 2.0, 3.0),
		},
	}

	repo := &GameRepositoryMock{}
	scoresRepo := repositories.NewInMemoryScoresRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, scoresRepo, betRepo, NewMatchWatcherServiceMock(updates), 10*time.Millisecond)

	// Add some bets
	good_bet_match_1 := models.NewBet(match1, 2, 1)
	good_bet_match_2 := models.NewBet(match2, 1, 2)
	service.UpdatePlayerBet(players[0], good_bet_match_1, matchTime.Add(-1*time.Second))
	service.UpdatePlayerBet(players[1], good_bet_match_2, matchTime)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	done := make(chan struct{})
	var winners []models.Player
	var playErr error

	go func() {
		winners, playErr = service.Play()
		close(done)
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Play function timed out after 1 second")
	case <-done:
		if playErr != nil {
			t.Fatalf("Failed to play game: %v", playErr)
		}
	}

	if len(winners) != 2 {
		t.Errorf("Expected 2 winners, got %d", len(winners))
	}

	if !game.IsFinished() {
		t.Errorf("Expected game to be finished after all matches are played")
	}
}

func TestGameService_UpdatePlayerBet(t *testing.T) {
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{{Name: "Player1"}}
	matches := []models.Match{match}

	game := rules.NewFreshGame("2024", "Premier League", players, matches, &ScorerMock{})
	repo := &GameRepositoryMock{}
	scoresRepo := repositories.NewInMemoryScoresRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, scoresRepo, betRepo, nil, 10*time.Millisecond)

	bet := models.NewBet(match, 2, 1)
	err := service.UpdatePlayerBet(players[0], bet, matchTime.Add(-1*time.Second))
	if err != nil {
		t.Errorf("Failed to update player bet: %v", err)
	}

	// Verify bet was saved
	bets, _, err := betRepo.GetBetsForMatch(match, "1")
	if err != nil {
		t.Errorf("Failed to get bets for match: %v", err)
	}
	if len(bets) != 1 {
		t.Errorf("Expected 1 bet, got %d", len(bets))
	}
	if bets[0] != bet {
		t.Errorf("Retrieved bet is not the same as the one added")
	}

	// Verify bet was saved with the correct game ID
	gameBets, err := betRepo.GetBets("1", players[0])
	if err != nil {
		t.Errorf("Failed to get bets for game: %v", err)
	}
	if len(gameBets) != 1 {
		t.Errorf("Expected 1 bet for game, got %d", len(gameBets))
	}
	if gameBets[0] != bet {
		t.Errorf("Retrieved bet for game is not the same as the one added")
	}
}

func TestGameService_HandleScoreUpdate(t *testing.T) {
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	game := rules.NewFreshGame("2024", "Premier League", players, matches, &ScorerMock{})
	repo := &GameRepositoryMock{}
	scoresRepo := repositories.NewInMemoryScoresRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, scoresRepo, betRepo, nil, 10*time.Millisecond)

	// Add some bets
	bet1 := models.NewBet(match, 2, 1) // Correct prediction
	bet2 := models.NewBet(match, 1, 1) // Wrong prediction
	service.UpdatePlayerBet(players[0], bet1, matchTime.Add(-1*time.Second))
	service.UpdatePlayerBet(players[1], bet2, matchTime.Add(-1*time.Second))

	// Handle score update
	finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
	service.handleUpdates(map[string]models.Match{finishedMatch.Id(): finishedMatch})

	// Verify game state
	if !game.IsFinished() {
		t.Errorf("Expected game to be finished after score update")
	}

	points := game.GetPlayersPoints()
	if points[players[0]] != 500 {
		t.Errorf("Expected Player1 to have 500 points, got %d", points[players[0]])
	}
	if points[players[1]] != 0 {
		t.Errorf("Expected Player2 to have 0 points, got %d", points[players[1]])
	}
}
