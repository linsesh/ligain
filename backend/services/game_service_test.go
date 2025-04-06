package services

import (
	"context"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/rules"
	"liguain/backend/utils"
	"testing"
	"time"
)

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

// ScorerMock for testing
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

// GameMock implements the Game interface for testing
type GameMock struct {
	players []models.Player
	matches []models.Match
	scorer  rules.Scorer
	bets    map[string]map[models.Player]*models.Bet
	points  map[models.Player]int
}

func NewGameMock(players []models.Player, matches []models.Match, scorer rules.Scorer) *GameMock {
	g := &GameMock{
		players: players,
		matches: matches,
		scorer:  scorer,
		bets:    make(map[string]map[models.Player]*models.Bet),
		points:  make(map[models.Player]int),
	}
	for _, match := range matches {
		g.bets[match.Id()] = make(map[models.Player]*models.Bet)
	}
	return g
}

func (g *GameMock) GetPlayers() []models.Player {
	return g.players
}

func (g *GameMock) GetMatches() []models.Match {
	return g.matches
}

func (g *GameMock) GetScorer() rules.Scorer {
	return g.scorer
}

func (g *GameMock) GetSeasonYear() string {
	return "2024"
}

func (g *GameMock) GetCompetitionName() string {
	return "Premier League"
}

func (g *GameMock) UpdateMatch(match models.Match) error {
	for i, m := range g.matches {
		if m.Id() == match.Id() {
			g.matches[i] = match
			return nil
		}
	}
	return fmt.Errorf("match %v not found", match.Id())
}

func (g *GameMock) CalculateMatchScores(match models.Match) (map[models.Player]int, error) {
	for i, m := range g.matches {
		if m.Id() == match.Id() {
			if !match.IsFinished() {
				return nil, fmt.Errorf("match is not finished")
			}
			g.matches[i] = match
			matchBets := g.bets[match.Id()]
			players, bets := utils.MapKeysValues(matchBets)
			scores := g.scorer.Score(match, bets)
			scoresMap := make(map[models.Player]int)
			for i, score := range scores {
				scoresMap[players[i]] = score
			}
			return scoresMap, nil
		}
	}
	return nil, fmt.Errorf("match %v not found", match.Id())
}

func (g *GameMock) ApplyMatchScores(scores map[models.Player]int) error {
	for player, score := range scores {
		g.points[player] += score
	}
	return nil
}

func (g *GameMock) IsFinished() bool {
	return g.GetGameStatus() == rules.GameStatusFinished
}

func (g *GameMock) GetWinner() []models.Player {
	bestScore := 0
	winners := make([]models.Player, 0)
	for player, points := range g.points {
		if points > bestScore {
			bestScore = points
			winners = []models.Player{player}
		} else if points == bestScore {
			winners = append(winners, player)
		}
	}
	return winners
}

func (g *GameMock) GetGameStatus() rules.GameStatus {
	return rules.GameStatusScheduled // For testing purposes, always return "in progress"
}

func (g *GameMock) GetIncomingMatches() []models.Match {
	incomingMatches := make([]models.Match, 0)
	for _, match := range g.matches {
		if !match.IsFinished() {
			incomingMatches = append(incomingMatches, match)
		}
	}
	return incomingMatches
}

func (g *GameMock) GetMatchBets(match models.Match) (map[models.Player]*models.Bet, error) {
	if g.bets[match.Id()] == nil {
		return nil, fmt.Errorf("match %v not found", match.Id())
	}
	return g.bets[match.Id()], nil
}

func (g *GameMock) AddPlayerBet(player *models.Player, bet *models.Bet, datetime time.Time) error {
	if g.bets[bet.Match.Id()] == nil {
		return fmt.Errorf("match %v not found", bet.Match.Id())
	}
	if datetime.After(bet.Match.GetDate()) {
		return fmt.Errorf("too late to bet on match %v", bet.Match.Id())
	}
	g.bets[bet.Match.Id()][*player] = bet
	return nil
}

func (g *GameMock) GetPlayersPoints() map[models.Player]int {
	return g.points
}

// Test cases
func TestGameService_Play_SingleMatch(t *testing.T) {
	// Setup test data
	testTime := time.Now()
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	// Create a game
	game := NewGameMock(players, matches, &ScorerMock{})

	// Setup mock updates
	updates := []map[string]models.Match{
		{
			match.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0),
		},
	}

	// Create service with mocks
	repo := &GameRepositoryMock{}
	service, err := NewGameService(game, repo)
	if err != nil {
		t.Fatalf("Failed to create game service: %v", err)
	}
	service.watcher = NewMatchWatcherServiceMock(updates)

	// Add some bets
	bet1 := models.NewBet(match, 2, 1) // Correct bet
	bet2 := models.NewBet(match, 1, 1) // Wrong bet
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)

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
	testTime := time.Now()
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", testTime.Add(time.Hour), 2)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match1, match2}

	// Create a game
	game := NewGameMock(players, matches, &ScorerMock{})

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
	service, err := NewGameService(game, repo)
	if err != nil {
		t.Fatalf("Failed to create game service: %v", err)
	}
	service.watcher = NewMatchWatcherServiceMock(updates)

	// Add some bets
	bet1 := models.NewBet(match1, 2, 1) // Correct bet for match 1
	bet2 := models.NewBet(match1, 1, 1) // Wrong bet for match 1
	bet3 := models.NewBet(match2, 2, 1) // Correct bet for match 2
	bet4 := models.NewBet(match2, 1, 1) // Wrong bet for match 2
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)
	game.AddPlayerBet(&players[0], bet3, testTime)
	game.AddPlayerBet(&players[1], bet4, testTime)

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
	testTime := time.Now()
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	// Create a game
	game := NewGameMock(players, matches, &ScorerMock{})

	// Setup mock updates
	updates := []map[string]models.Match{
		{
			match.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", testTime, 1, 1.0, 2.0, 3.0),
		},
	}

	// Create service with mocks
	repo := &GameRepositoryMock{}
	service, err := NewGameService(game, repo)
	if err != nil {
		t.Fatalf("Failed to create game service: %v", err)
	}
	service.watcher = NewMatchWatcherServiceMock(updates)

	// Add wrong bets for both players
	bet1 := models.NewBet(match, 1, 1) // Wrong bet
	bet2 := models.NewBet(match, 0, 2) // Wrong bet
	game.AddPlayerBet(&players[0], bet1, testTime)
	game.AddPlayerBet(&players[1], bet2, testTime)

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
