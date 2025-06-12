package services

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/rules"
	"liguain/backend/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var matchTime = time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

// Mock implementations
type gameRepositoryMock struct{}

func (r *gameRepositoryMock) CreateGame(game models.Game) (string, error) {
	return "test-game-id", nil
}

func (r *gameRepositoryMock) GetGame(gameId string) (models.Game, error) {
	return nil, nil // Not used in tests
}

func (r *gameRepositoryMock) SaveWithId(gameId string, game models.Game) error {
	return nil
}

type matchWatcherServiceMock struct {
	updates []map[string]models.Match
	index   int
}

func newMatchWatcherServiceMock(updates []map[string]models.Match) *matchWatcherServiceMock {
	return &matchWatcherServiceMock{
		updates: updates,
		index:   0,
	}
}

func (m *matchWatcherServiceMock) GetUpdates(ctx context.Context, done chan utils.TaskResult[map[string]models.Match]) {
	var result utils.TaskResult[map[string]models.Match]
	if m.index >= len(m.updates) {
		result = utils.TaskResult[map[string]models.Match]{Value: make(map[string]models.Match)}
	} else {
		update := m.updates[m.index]
		m.index++
		result = utils.TaskResult[map[string]models.Match]{Value: update}
	}
	select {
	case <-ctx.Done():
		return
	case done <- result:
	}
}

func (m *matchWatcherServiceMock) WatchMatches(matches []models.Match) {
	// Not used in tests
}

type scorerMock struct{}

func (s *scorerMock) Score(match models.Match, bets []*models.Bet) []int {
	scores := make([]int, len(bets))
	for i, bet := range bets {
		// Create a new bet with the same predictions but the finished match
		finishedBet := models.NewBet(match, bet.PredictedHomeGoals, bet.PredictedAwayGoals)
		if finishedBet.IsBetCorrect() {
			scores[i] = 500
		} else {
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

type MockMatchWatcher struct {
	updates map[string]models.Match
}

func (m *MockMatchWatcher) WatchMatches(matches []models.Match) {}

func (m *MockMatchWatcher) GetUpdates(ctx context.Context, done chan utils.TaskResult[map[string]models.Match]) {
	done <- utils.TaskResult[map[string]models.Match]{Value: m.updates}
}

type MockScorer struct{}

func (s *MockScorer) Score(match models.Match, bets []*models.Bet) []int {
	scores := make([]int, len(bets))
	for i, bet := range bets {
		if bet.PredictedHomeGoals == match.GetHomeGoals() && bet.PredictedAwayGoals == match.GetAwayGoals() {
			scores[i] = 3 // Perfect prediction
		} else {
			scores[i] = 1 // Wrong prediction
		}
	}
	return scores
}

type mockMatchWatcherService struct {
	updates map[string]models.Match
}

func newMockMatchWatcherService() *mockMatchWatcherService {
	return &mockMatchWatcherService{
		updates: make(map[string]models.Match),
	}
}

func (m *mockMatchWatcherService) WatchMatches(matches []models.Match) {
	// For testing, we don't need to do anything
}

func (m *mockMatchWatcherService) GetUpdates(ctx context.Context, done chan MatchWatcherServiceResult) {
	// For testing, we return an empty map
	done <- MatchWatcherServiceResult{
		Value: make(map[string]models.Match),
		Err:   nil,
	}
}

// Test cases
func TestGameService_Play_SingleMatch(t *testing.T) {
	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	// Create a game
	game := rules.NewFreshGame("2024", "Premier League", players, matches, &scorerMock{})

	// Setup mock updates
	updates := []map[string]models.Match{
		{
			match.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0),
		},
	}

	// Create service with mocks
	repo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, betRepo, newMatchWatcherServiceMock(updates), 10*time.Millisecond)

	// Add bets using the service
	goodBet := models.NewBet(match, 2, 1)
	wrongBet := models.NewBet(match, 1, 1)
	err := service.UpdatePlayerBet(players[0], goodBet, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)
	err = service.UpdatePlayerBet(players[1], wrongBet, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)

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
		require.NoError(t, playErr)
	}

	require.Len(t, winners, 1, "Expected exactly one winner")
	assert.Equal(t, "Player1", winners[0].Name, "Expected Player1 to win")
	assert.True(t, game.IsFinished(), "Expected game to be finished after all matches are played")
}

func TestGameService_Play_MultipleMatches(t *testing.T) {
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Corsica Championship", matchTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Corsica Championship", matchTime.Add(time.Hour), 1)
	match3 := models.NewSeasonMatch("Team4", "Team5", "2024", "Corsica Championship", matchTime.Add(2*time.Hour), 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match1, match2, match3}

	game := rules.NewFreshGame("2024", "Corsica Championship", players, matches, &scorerMock{})

	updates := []map[string]models.Match{
		{
			match1.Id(): models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Corsica Championship", matchTime, 1, 1.0, 2.0, 3.0),
		},
		{
			match2.Id(): models.NewFinishedSeasonMatch("Team3", "Team4", 2, 1, "2024", "Corsica Championship", matchTime.Add(time.Hour), 1, 1.0, 2.0, 3.0),
		},
		{
			match3.Id(): models.NewFinishedSeasonMatch("Team4", "Team5", 2, 1, "2024", "Corsica Championship", matchTime.Add(2*time.Hour), 1, 1.0, 2.0, 3.0),
		},
	}

	repo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, betRepo, newMatchWatcherServiceMock(updates), 10*time.Millisecond)

	// Add bets for all matches using the service
	// Match 1 bets
	goodBet1 := models.NewBet(match1, 2, 1)
	wrongBet1 := models.NewBet(match1, 1, 1)
	err := service.UpdatePlayerBet(players[0], goodBet1, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)
	err = service.UpdatePlayerBet(players[1], wrongBet1, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)

	// Match 2 bets
	goodBet2 := models.NewBet(match2, 2, 1)
	wrongBet2 := models.NewBet(match2, 1, 1)
	err = service.UpdatePlayerBet(players[0], goodBet2, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)
	err = service.UpdatePlayerBet(players[1], wrongBet2, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)

	// Match 3 bets
	goodBet3 := models.NewBet(match3, 2, 1)
	wrongBet3 := models.NewBet(match3, 1, 1)
	err = service.UpdatePlayerBet(players[0], goodBet3, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)
	err = service.UpdatePlayerBet(players[1], wrongBet3, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)

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
		require.NoError(t, playErr)
	}

	require.Len(t, winners, 1, "Expected exactly one winner")
	assert.Equal(t, "Player1", winners[0].Name, "Expected Player1 to win")
	assert.True(t, game.IsFinished(), "Expected game to be finished after all matches are played")

	// Get all bets for each player to find the bet IDs
	player1Bets, err := service.GetPlayerBets(players[0])
	require.NoError(t, err)
	require.Len(t, player1Bets, 3)
	player2Bets, err := service.GetPlayerBets(players[1])
	require.NoError(t, err)
	require.Len(t, player2Bets, 3)

	// Verify total points
	points := game.GetPlayersPoints()
	assert.Equal(t, 1500, points[players[0]], "Expected Player1 to have 1500 points total")
	assert.Equal(t, 0, points[players[1]], "Expected Player2 to have 0 points total")
}

func TestGameService_UpdatePlayerBet(t *testing.T) {
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{{Name: "Player1"}}
	matches := []models.Match{match}

	game := rules.NewFreshGame("2024", "Premier League", players, matches, &scorerMock{})
	repo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, betRepo, newMatchWatcherServiceMock(nil), 10*time.Millisecond)

	// Test valid bet
	bet := models.NewBet(match, 2, 1)
	err := service.UpdatePlayerBet(players[0], bet, matchTime.Add(-1*time.Second))
	require.NoError(t, err)

	// Test invalid bet (after match start)
	bet = models.NewBet(match, 1, 1)
	err = service.UpdatePlayerBet(players[0], bet, matchTime.Add(time.Second))
	require.Error(t, err)

	// Verify saved bet
	bets, err := service.GetPlayerBets(players[0])
	require.NoError(t, err)
	require.Len(t, bets, 1)
	assert.Equal(t, 2, bets[0].PredictedHomeGoals)
	assert.Equal(t, 1, bets[0].PredictedAwayGoals)
}

func TestGameService_GetPlayerBets(t *testing.T) {
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{{Name: "Player1"}}
	matches := []models.Match{match}

	game := rules.NewFreshGame("2024", "Premier League", players, matches, &scorerMock{})
	repo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, betRepo, newMatchWatcherServiceMock(nil), 10*time.Millisecond)

	// Test getting bets for player with no bets
	bets, err := service.GetPlayerBets(players[0])
	require.NoError(t, err)
	assert.Empty(t, bets)

	// Add a bet and verify
	bet := models.NewBet(match, 2, 1)
	betId, _, err := betRepo.SaveBet("1", bet, players[0])
	require.NoError(t, err)

	bets, err = service.GetPlayerBets(players[0])
	require.NoError(t, err)
	require.Len(t, bets, 1)
	assert.Equal(t, 2, bets[0].PredictedHomeGoals)
	assert.Equal(t, 1, bets[0].PredictedAwayGoals)

	// Verify bet ID was saved
	score := 500
	err = betRepo.SaveScore("1", match, players[0], score)
	require.NoError(t, err)
	savedScore, err := betRepo.GetScore("1", betId)
	require.NoError(t, err)
	assert.Equal(t, score, savedScore)
}

func TestGameService_HandleScoreUpdate(t *testing.T) {
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	game := rules.NewFreshGame("2024", "Premier League", players, matches, &scorerMock{})
	repo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()
	service := NewGameService("1", game, repo, betRepo, newMatchWatcherServiceMock(nil), 10*time.Millisecond)

	// Add bets using the service
	bet1 := models.NewBet(match, 2, 1) // Correct prediction
	bet2 := models.NewBet(match, 1, 1) // Wrong prediction
	err := service.UpdatePlayerBet(players[0], bet1, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)
	err = service.UpdatePlayerBet(players[1], bet2, matchTime.Add(-1*time.Hour))
	require.NoError(t, err)

	// Handle score update
	match.Finish(2, 1)
	err = service.HandleUpdates(map[string]models.Match{match.Id(): match})
	require.NoError(t, err)

	// Verify game state
	assert.True(t, game.IsFinished(), "Expected game to be finished after score update")

	points := game.GetPlayersPoints()
	assert.Equal(t, 500, points[players[0]], "Expected Player1 to have 500 points")
	assert.Equal(t, 0, points[players[1]], "Expected Player2 to have 0 points")

	// Get all bets for each player
	player1Bets, err := service.GetPlayerBets(players[0])
	require.NoError(t, err)
	require.Len(t, player1Bets, 1)
	player2Bets, err := service.GetPlayerBets(players[1])
	require.NoError(t, err)
	require.Len(t, player2Bets, 1)
}

func setupTestGameService() (*GameServiceImpl, *mockMatchWatcherService, models.Match) {
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
	matches := []models.Match{match}

	game := rules.NewFreshGame("2024", "Premier League", players, matches, &scorerMock{})
	gameRepo := &gameRepositoryMock{}
	betRepo := repositories.NewInMemoryBetRepository()
	watcher := newMockMatchWatcherService()
	waitTime := 10 * time.Millisecond

	service := NewGameService("test-game", game, gameRepo, betRepo, watcher, waitTime)
	return service, watcher, match
}

func TestGameService_GetMatchResult(t *testing.T) {
	service, _, match := setupTestGameService()

	t.Run("returns nil for non-existent match", func(t *testing.T) {
		matchResults := service.GetMatchResults()
		require.NotNil(t, matchResults)
		result := matchResults["non-existent"]
		assert.Nil(t, result)
	})

	t.Run("returns match result for incoming match with bets", func(t *testing.T) {
		// Get the first match from the game
		incomingMatches := service.GetIncomingMatches()
		require.NotEmpty(t, incomingMatches)
		matchId := match.Id()
		matchResult, exists := incomingMatches[matchId]
		require.True(t, exists)
		match := matchResult.Match
		require.NotNil(t, match)

		// Add a bet for the match
		player := models.Player{Name: "Player1"}
		bet := models.NewBet(match, 2, 1)
		err := service.UpdatePlayerBet(player, bet, match.GetDate())
		require.NoError(t, err)

		// Get match result and verify it has bets but no scores
		incomingMatches = service.GetIncomingMatches()
		require.NotNil(t, incomingMatches)
		result := incomingMatches[matchId]
		require.NotNil(t, result)
		assert.Equal(t, matchId, result.Match.Id())
		assert.NotNil(t, result.Bets)
		assert.Equal(t, 1, len(result.Bets))
		assert.Nil(t, result.Scores)

		// Finish the match
		match.Finish(2, 1)
		err = service.HandleUpdates(map[string]models.Match{match.Id(): match})
		require.NoError(t, err)

		// Get match result and verify it has both bets and scores
		pastMatches := service.GetMatchResults()
		require.NotNil(t, pastMatches)
		result = pastMatches[matchId]
		require.NotNil(t, result)
		assert.Equal(t, matchId, result.Match.Id())
		assert.NotNil(t, result.Bets)
		assert.Equal(t, 1, len(result.Bets))
		assert.NotNil(t, result.Scores)
		assert.Equal(t, 1, len(result.Scores))

		// Verify that all matches are included
		incomingMatches = service.GetIncomingMatches()
		pastMatches = service.GetMatchResults()
		assert.Equal(t, 0, len(incomingMatches), "No incoming matches after match is finished")
		assert.Equal(t, 1, len(pastMatches), "One past match after match is finished")
	})
}

func TestGameService_GetAllMatchResults(t *testing.T) {
	t.Run("returns all matches with proper bets and scores", func(t *testing.T) {
		// Create a service with multiple matches
		match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
		match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime.Add(time.Hour), 2)
		match3 := models.NewSeasonMatch("Team5", "Team6", "2024", "Premier League", matchTime.Add(2*time.Hour), 3)
		players := []models.Player{{Name: "Player1"}, {Name: "Player2"}}
		matches := []models.Match{match1, match2, match3}

		game := rules.NewFreshGame("2024", "Premier League", players, matches, &scorerMock{})
		gameRepo := &gameRepositoryMock{}
		betRepo := repositories.NewInMemoryBetRepository()
		watcher := newMockMatchWatcherService()
		service := NewGameService("test-game", game, gameRepo, betRepo, watcher, 10*time.Millisecond)

		// Add bets for all matches
		player := models.Player{Name: "Player1"}
		bet1 := models.NewBet(match1, 2, 1)
		bet2 := models.NewBet(match2, 1, 1)
		bet3 := models.NewBet(match3, 3, 0)
		err := service.UpdatePlayerBet(player, bet1, match1.GetDate())
		require.NoError(t, err)
		err = service.UpdatePlayerBet(player, bet2, match2.GetDate())
		require.NoError(t, err)
		err = service.UpdatePlayerBet(player, bet3, match3.GetDate())
		require.NoError(t, err)

		// Initially all matches should be incoming
		incomingMatches := service.GetIncomingMatches()
		pastMatches := service.GetMatchResults()
		assert.Equal(t, 3, len(incomingMatches), "Expected three incoming matches initially")
		assert.Equal(t, 0, len(pastMatches), "Expected no past matches initially")

		// Finish first match
		match1.Finish(2, 1)
		err = service.HandleUpdates(map[string]models.Match{match1.Id(): match1})
		require.NoError(t, err)

		// Should have 2 incoming and 1 past match
		incomingMatches = service.GetIncomingMatches()
		pastMatches = service.GetMatchResults()
		assert.Equal(t, 2, len(incomingMatches), "Expected two incoming matches after finishing one")
		assert.Equal(t, 1, len(pastMatches), "Expected one past match after finishing one")

		// Finish second match
		match2.Finish(1, 1)
		err = service.HandleUpdates(map[string]models.Match{match2.Id(): match2})
		require.NoError(t, err)

		// Should have 1 incoming and 2 past matches
		incomingMatches = service.GetIncomingMatches()
		pastMatches = service.GetMatchResults()
		assert.Equal(t, 1, len(incomingMatches), "Expected one incoming match after finishing two")
		assert.Equal(t, 2, len(pastMatches), "Expected two past matches after finishing two")

		// Finish last match
		match3.Finish(3, 0)
		err = service.HandleUpdates(map[string]models.Match{match3.Id(): match3})
		require.NoError(t, err)

		// Should have 0 incoming and 3 past matches
		incomingMatches = service.GetIncomingMatches()
		pastMatches = service.GetMatchResults()
		assert.Equal(t, 0, len(incomingMatches), "Expected no incoming matches after finishing all")
		assert.Equal(t, 3, len(pastMatches), "Expected three past matches after finishing all")

		// Verify the content of past matches
		for _, match := range []models.Match{match1, match2, match3} {
			result := pastMatches[match.Id()]
			require.NotNil(t, result, "Expected to find result for match %s", match.Id())
			assert.Equal(t, match.Id(), result.Match.Id())
			assert.NotNil(t, result.Bets)
			assert.Equal(t, 1, len(result.Bets))
			assert.NotNil(t, result.Scores)
			assert.Equal(t, 1, len(result.Scores))
		}
	})
}

func TestGameService_SaveBet(t *testing.T) {
	service, _, match := setupTestGameService()

	t.Run("saves valid bet", func(t *testing.T) {
		// Get the first match from the game

		player := models.Player{Name: "Player1"}
		bet := models.NewBet(match, 2, 1)

		err := service.UpdatePlayerBet(player, bet, match.GetDate())
		require.NoError(t, err)

		// Verify bet was saved
		bets, err := service.GetPlayerBets(player)
		require.NoError(t, err)
		assert.Len(t, bets, 1)
		assert.Equal(t, 2, bets[0].PredictedHomeGoals)
		assert.Equal(t, 1, bets[0].PredictedAwayGoals)
	})

	t.Run("fails to save bet for non-existent match", func(t *testing.T) {
		player := models.Player{Name: "Player1"}
		nonExistentMatch := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime, 1)
		bet := models.NewBet(nonExistentMatch, 2, 1)

		err := service.UpdatePlayerBet(player, bet, nonExistentMatch.GetDate())
		assert.Error(t, err)
	})
}

func TestGameService_GetBetsForMatch(t *testing.T) {
	service, _, match := setupTestGameService()

	t.Run("returns empty for non-existent player", func(t *testing.T) {
		bets, err := service.GetPlayerBets(models.Player{Name: "non-existent"})
		require.NoError(t, err)
		assert.Empty(t, bets)
	})

	t.Run("returns bets for match with bets", func(t *testing.T) {
		// Get the first match from the game

		// Add a bet
		player := models.Player{Name: "Player1"}
		bet := models.NewBet(match, 2, 1)
		err := service.UpdatePlayerBet(player, bet, match.GetDate())
		require.NoError(t, err)

		// Get bets for match
		bets, err := service.GetPlayerBets(player)
		require.NoError(t, err)
		assert.Len(t, bets, 1)
		assert.Equal(t, 2, bets[0].PredictedHomeGoals)
		assert.Equal(t, 1, bets[0].PredictedAwayGoals)
	})
}

func TestGameService_UpdateMatch(t *testing.T) {
	service, _, match := setupTestGameService()

	t.Run("updates existing match", func(t *testing.T) {
		// Create an updated match
		match.Finish(2, 1)
		err := service.HandleUpdates(map[string]models.Match{match.Id(): match})
		require.NoError(t, err)

		// Verify match was updated
		pastMatches := service.GetMatchResults()
		require.NotNil(t, pastMatches)
		result := pastMatches[match.Id()]
		assert.NotNil(t, result)
		assert.True(t, result.Match.IsFinished())
		assert.Equal(t, 2, result.Match.GetHomeGoals())
		assert.Equal(t, 1, result.Match.GetAwayGoals())

		// Verify no incoming matches
		incomingMatches := service.GetIncomingMatches()
		assert.Equal(t, 0, len(incomingMatches), "Expected no incoming matches after match is finished")
	})

	t.Run("fails to update non-existent match", func(t *testing.T) {
		nonExistentMatch := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime, 1)
		err := service.HandleUpdates(map[string]models.Match{nonExistentMatch.Id(): nonExistentMatch})
		assert.Error(t, err)
	})
}
