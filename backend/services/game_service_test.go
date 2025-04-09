package services

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/rules"
	"testing"
	"time"
)

var testTime = time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)

// Mock implementations
type GameRepositoryMock struct{}

func (r *GameRepositoryMock) SaveGame(game rules.Game) (string, error) {
	return "test-game-id", nil
}

func (r *GameRepositoryMock) UpdateScores(match models.Match, scores map[models.Player]int) error {
	return nil
}

func (r *GameRepositoryMock) GetGame(gameId string) (rules.Game, error) {
	return nil, nil // Not used in tests
}

type BetRepositoryMock struct{}

func (r *BetRepositoryMock) GetBets(gameId string, player models.Player) ([]models.Bet, error) {
	return nil, nil // Not used in tests
}

func (r *BetRepositoryMock) SaveBet(bet models.Bet) (string, error) {
	return "test-bet-id", nil
}

func (r *BetRepositoryMock) GetBetsForMatch(match models.Match) ([]models.Bet, error) {
	return nil, nil // Not used in tests
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
	if m.index >= len(m.updates) {
		done <- MatchWatcherServiceResult{Value: make(map[string]models.Match), Err: nil}
		return
	}
	update := m.updates[m.index]
	m.index++
	done <- MatchWatcherServiceResult{Value: update, Err: nil}
}

func (m *MatchWatcherServiceMock) WatchMatches(matches []models.Match) {
	// Not used in tests
}

type ScorerMock struct{}

func (s *ScorerMock) Score(match models.Match, bets []*models.Bet) []int {
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

// Test cases
func TestGameService_Play_SingleMatch(t *testing.T) {
	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	// Create a game
	game := rules.NewGame("2024", "Premier League", players, matches, &ScorerMock{})

	// Setup mock updates
	updates := []map[string]models.Match{
		{
			match.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0),
		},
	}

	// Create service with mocks
	repo := &GameRepositoryMock{}
	betRepo := &BetRepositoryMock{}
	service, err := NewGameService(game, repo, betRepo, NewMatchWatcherServiceMock(updates), 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create game service: %v", err)
	}
	service.watcher = NewMatchWatcherServiceMock(updates)

	// Add some bets
	bet1 := models.NewBet(match, 2, 1) // Correct good result
	bet2 := models.NewBet(match, 1, 1) // Wrong result
	service.updateBet(bet1, players[0])
	service.updateBet(bet2, players[1])

	// Play the game
	winners, err := service.Play()
	if err != nil {
		t.Fatalf("Failed to play game: %v", err)
	}

	// Verify results
	if len(winners) != 1 {
		t.Errorf("Expected 1 winner, got %d", len(winners))
	}
	if winners[0].Name != "Player1" {
		t.Errorf("Expected Player1 to win, got %s", winners[0].Name)
	}
}

func TestGameService_Play_MultipleMatches(t *testing.T) {
	// Setup test data
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", testTime.Add(time.Hour), 2)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match1, match2}

	// Create a game
	game := rules.NewGame("2024", "Premier League", players, matches, &ScorerMock{})

	// Setup mock updates
	updates := []map[string]models.Match{
		{
			match1.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0),
		},
		{
			match2.Id(): models.NewFinishedSeasonMatch("Team3", "Team4", 2, 1, "2024", "Premier League", testTime.Add(time.Hour), 2, 1.0, 2.0, 3.0),
		},
	}

	// Create service with mocks
	repo := &GameRepositoryMock{}
	betRepo := &BetRepositoryMock{}
	service, err := NewGameService(game, repo, betRepo, NewMatchWatcherServiceMock(updates), 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create game service: %v", err)
	}
	service.watcher = NewMatchWatcherServiceMock(updates)

	// Add some bets
	bet1 := models.NewBet(match1, 2, 1) // Correct bet for match 1
	bet2 := models.NewBet(match1, 1, 1) // Wrong bet for match 1
	bet3 := models.NewBet(match2, 2, 1) // Correct bet for match 2
	bet4 := models.NewBet(match2, 1, 1) // Wrong bet for match 2
	service.updateBet(bet1, players[0])
	service.updateBet(bet2, players[1])
	service.updateBet(bet3, players[0])
	service.updateBet(bet4, players[1])

	// Play the game
	winners, err := service.Play()
	if err != nil {
		t.Fatalf("Failed to play game: %v", err)
	}

	// Verify results
	if len(winners) != 1 {
		t.Errorf("Expected 1 winner, got %d", len(winners))
	}
	if winners[0].Name != "Player1" {
		t.Errorf("Expected Player1 to win, got %s", winners[0].Name)
	}
}

func TestGameService_Play_NoWinner(t *testing.T) {
	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	// Create a game
	game := rules.NewGame("2024", "Premier League", players, matches, &ScorerMock{})

	// Setup mock updates
	updates := []map[string]models.Match{
		{
			match.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0),
		},
	}

	// Create service with mocks
	repo := &GameRepositoryMock{}
	betRepo := &BetRepositoryMock{}
	service, err := NewGameService(game, repo, betRepo, NewMatchWatcherServiceMock(updates), 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create game service: %v", err)
	}
	service.watcher = NewMatchWatcherServiceMock(updates)

	// Add wrong bets for both players
	bet1 := models.NewBet(match, 1, 1) // Wrong bet
	bet2 := models.NewBet(match, 0, 2) // Wrong bet
	service.updateBet(bet1, players[0])
	service.updateBet(bet2, players[1])

	// Play the game
	winners, err := service.Play()
	if err != nil {
		t.Fatalf("Failed to play game: %v", err)
	}

	// Verify results - both players should be winners with 0 points
	if len(winners) != 2 {
		t.Errorf("Expected 2 winners (tie with 0 points), got %d", len(winners))
	}
	winnerNames := make(map[string]bool)
	for _, winner := range winners {
		winnerNames[winner.Name] = true
	}
	if !winnerNames["Player1"] {
		t.Errorf("Expected Player1 to be a winner")
	}
	if !winnerNames["Player2"] {
		t.Errorf("Expected Player2 to be a winner")
	}
}
