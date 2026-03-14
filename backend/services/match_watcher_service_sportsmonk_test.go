package services

import (
	"context"
	"errors"
	"ligain/backend/models"
	"testing"
	"time"

	"ligain/backend/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SportsmonkRepositoryMock struct {
	lastMatchInfos  []map[string]models.Match
	receivedMatches []map[string]models.Match
	err             error
	callCount       int
}

func (r *SportsmonkRepositoryMock) GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error) {
	r.receivedMatches = append(r.receivedMatches, matches)
	if r.callCount >= len(r.lastMatchInfos) {
		return make(map[string]models.Match), r.err
	}
	result := r.lastMatchInfos[r.callCount]
	r.callCount++
	return result, r.err
}

// MockGameService for testing
type MockGameService struct {
	gameID  string
	updates []map[string]models.Match
}

func (m *MockGameService) HandleMatchUpdates(updates map[string]models.Match) error {
	m.updates = append(m.updates, updates)
	return nil
}

func (m *MockGameService) GetGameID() string {
	return m.gameID
}

// Implement other GameService methods for testing
func (m *MockGameService) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}

func (m *MockGameService) GetMatchResults() map[string]*models.MatchResult {
	return make(map[string]*models.MatchResult)
}

func (m *MockGameService) UpdatePlayerBet(player models.Player, bet *models.Bet, now time.Time) error {
	return nil
}

func (m *MockGameService) GetPlayerBets(player models.Player) ([]*models.Bet, error) {
	return nil, nil
}

func (m *MockGameService) GetPlayers() []models.Player {
	return nil
}

func (m *MockGameService) AddPlayer(player models.Player) error {
	return nil
}

func (m *MockGameService) RemovePlayer(player models.Player) error {
	return nil
}

func NewMockGameService(gameID string) *MockGameService {
	return &MockGameService{
		gameID:  gameID,
		updates: make([]map[string]models.Match, 0),
	}
}

func TestMatchWatcherServiceSportsmonk_Subscribe(t *testing.T) {
	// Create mock repository
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{},
		err:            nil,
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
		subscribers:    make(map[string]GameService),
		stopChan:       make(chan struct{}),
		pollInterval:   30 * time.Second,
		matchRepo:      repositories.NewInMemoryMatchRepository(),
		now:            func() time.Time { return time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) },
	}

	// Create mock handler
	handler := NewMockGameService("game1")

	// Test subscription (should not error)
	err := service.Subscribe(handler)
	require.NoError(t, err)

	// Subscribing again with the same handler should not error
	err = service.Subscribe(handler)
	require.NoError(t, err)
}

func TestMatchWatcherServiceSportsmonk_Unsubscribe(t *testing.T) {
	// Create mock repository
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{},
		err:            nil,
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
		subscribers:    make(map[string]GameService),
		stopChan:       make(chan struct{}),
		pollInterval:   30 * time.Second,
		matchRepo:      repositories.NewInMemoryMatchRepository(),
		now:            func() time.Time { return time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) },
	}

	// Create mock handler and subscribe
	handler := NewMockGameService("game1")
	err := service.Subscribe(handler)
	require.NoError(t, err)

	// Test unsubscription (should not error)
	err = service.Unsubscribe("game1")
	require.NoError(t, err)

	// Unsubscribing again should not error
	err = service.Unsubscribe("game1")
	require.NoError(t, err)
}

func TestMatchWatcherServiceSportsmonk_StartStop(t *testing.T) {
	// Create mock repository
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{},
		err:            nil,
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
		subscribers:    make(map[string]GameService),
		stopChan:       make(chan struct{}),
		pollInterval:   30 * time.Second,
		matchRepo:      repositories.NewInMemoryMatchRepository(),
		now:            func() time.Time { return time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) },
	}

	// Test start (should not error)
	ctx := context.Background()
	err := service.Start(ctx)
	require.NoError(t, err)

	// Starting again should not error
	err = service.Start(ctx)
	require.NoError(t, err)

	// Test stop (should not error)
	err = service.Stop()
	require.NoError(t, err)

	// Stopping again should not error
	err = service.Stop()
	require.NoError(t, err)
}

func TestMatchWatcherServiceSportsmonk_CheckForUpdates(t *testing.T) {
	// Create test matches
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	initialMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	updatedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)

	// Create mock repository with updated match info
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{
			{
				initialMatch.Id(): updatedMatch,
			},
		},
		err: nil,
	}

	// Create service instance with initial match in watched matches
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: map[string]models.Match{initialMatch.Id(): initialMatch},
		repo:           mockRepo,
		subscribers:    make(map[string]GameService),
		stopChan:       make(chan struct{}),
		pollInterval:   30 * time.Second,
		matchRepo:      repositories.NewInMemoryMatchRepository(),
		now:            func() time.Time { return time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) },
	}

	// Create mock handler and subscribe
	handler := NewMockGameService("game1")
	err := service.Subscribe(handler)
	require.NoError(t, err)

	// Check for updates
	service.checkForUpdates()

	// Wait a bit for the async handler to be called
	time.Sleep(10 * time.Millisecond)

	// Verify handler was called with updates
	require.Len(t, handler.updates, 1)
	updates := handler.updates[0]
	assert.Len(t, updates, 1)
	assert.Equal(t, updatedMatch, updates[initialMatch.Id()])
}

func TestMatchWatcherServiceSportsmonk_GetMatchesUpdates(t *testing.T) {
	// Create test matches
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	initialMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	updatedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)

	// Create mock repository with updated match info
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{
			{
				initialMatch.Id(): updatedMatch,
			},
		},
		err: nil,
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: map[string]models.Match{initialMatch.Id(): initialMatch},
		repo:           mockRepo,
		subscribers:    make(map[string]GameService),
		stopChan:       make(chan struct{}),
		pollInterval:   30 * time.Second,
		matchRepo:      repositories.NewInMemoryMatchRepository(),
		now:            func() time.Time { return time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) },
	}

	// Get updates
	updates, err := service.getMatchesUpdates()
	require.NoError(t, err)

	// Verify results
	assert.Len(t, updates, 1)
	assert.Equal(t, updatedMatch, updates[initialMatch.Id()])
}

func TestMatchWatcherServiceSportsmonk_GetMatchesUpdates_NoChanges(t *testing.T) {
	// Create test match
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)

	// Create mock repository with same match info
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{
			{
				match.Id(): match,
			},
		},
		err: nil,
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: map[string]models.Match{match.Id(): match},
		repo:           mockRepo,
		subscribers:    make(map[string]GameService),
		stopChan:       make(chan struct{}),
		pollInterval:   30 * time.Second,
		matchRepo:      repositories.NewInMemoryMatchRepository(),
		now:            func() time.Time { return time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) },
	}

	// Get updates
	updates, err := service.getMatchesUpdates()
	require.NoError(t, err)

	// Verify results
	assert.Len(t, updates, 0)
}

func TestMatchWatcherServiceSportsmonk_GetMatchesUpdates_Error(t *testing.T) {
	// Create test match
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)

	// Create mock repository with error
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{},
		err:            errors.New("API error"),
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: map[string]models.Match{match.Id(): match},
		repo:           mockRepo,
		subscribers:    make(map[string]GameService),
		stopChan:       make(chan struct{}),
		pollInterval:   30 * time.Second,
		matchRepo:      repositories.NewInMemoryMatchRepository(),
		now:            func() time.Time { return time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC) },
	}

	// Get updates
	updates, err := service.getMatchesUpdates()
	require.Error(t, err)
	assert.Len(t, updates, 0)
}

func TestMatchWasUpdated(t *testing.T) {
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

	t.Run("detects finished status change", func(t *testing.T) {
		initialMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
		finishedMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)

		assert.True(t, matchWasUpdated(initialMatch, finishedMatch))
	})

	t.Run("detects odds changes", func(t *testing.T) {
		initialMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "Premier League", matchTime, 1, 1.5, 2.0, 3.0)
		updatedMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "Premier League", matchTime, 1, 2.5, 3.0, 2.8)

		assert.True(t, matchWasUpdated(initialMatch, updatedMatch))
	})

	t.Run("detects date changes", func(t *testing.T) {
		initialMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
		newTime := matchTime.Add(1 * time.Hour)
		updatedMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", newTime, 1)

		assert.True(t, matchWasUpdated(initialMatch, updatedMatch))
	})

	t.Run("no changes detected", func(t *testing.T) {
		match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
		match2 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)

		assert.False(t, matchWasUpdated(match1, match2))
	})
}

func TestMatchWatcherService_OnlyQueriesMatchesWithinTwoWeeks(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	matchWithin := models.NewSeasonMatch("Team1", "Team2", "2025", "Ligue 1", now.Add(7*24*time.Hour), 1)  // 7 days away
	matchEdge := models.NewSeasonMatch("Team3", "Team4", "2025", "Ligue 1", now.Add(13*24*time.Hour), 1)  // 13 days away
	matchBeyond := models.NewSeasonMatch("Team5", "Team6", "2025", "Ligue 1", now.Add(21*24*time.Hour), 1) // 21 days away

	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{{}},
	}

	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: map[string]models.Match{
			matchWithin.Id():  matchWithin,
			matchEdge.Id():   matchEdge,
			matchBeyond.Id(): matchBeyond,
		},
		repo:         mockRepo,
		subscribers:  make(map[string]GameService),
		stopChan:     make(chan struct{}),
		pollInterval: 30 * time.Second,
		matchRepo:    repositories.NewInMemoryMatchRepository(),
		now:          func() time.Time { return now },
	}

	_, err := service.getMatchesUpdates()
	require.NoError(t, err)

	require.Len(t, mockRepo.receivedMatches, 1)
	assert.Len(t, mockRepo.receivedMatches[0], 2)
	assert.Contains(t, mockRepo.receivedMatches[0], matchWithin.Id())
	assert.Contains(t, mockRepo.receivedMatches[0], matchEdge.Id())
	assert.NotContains(t, mockRepo.receivedMatches[0], matchBeyond.Id())
}

func TestMatchWatcherService_MatchBecomesRelevantAsTimeAdvances(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	nearMatch := models.NewSeasonMatch("Team1", "Team2", "2025", "Ligue 1", baseTime.Add(5*24*time.Hour), 1)  // 5 days away
	farMatch := models.NewSeasonMatch("Team3", "Team4", "2025", "Ligue 1", baseTime.Add(20*24*time.Hour), 1) // 20 days away

	finishedNearMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2025", "Ligue 1", baseTime.Add(5*24*time.Hour), 1, 1.0, 2.0, 3.0)

	currentTime := baseTime
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{
			{nearMatch.Id(): finishedNearMatch},
			{farMatch.Id(): farMatch},
		},
	}

	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: map[string]models.Match{
			nearMatch.Id(): nearMatch,
			farMatch.Id():  farMatch,
		},
		repo:         mockRepo,
		subscribers:  make(map[string]GameService),
		stopChan:     make(chan struct{}),
		pollInterval: 30 * time.Second,
		matchRepo:    repositories.NewInMemoryMatchRepository(),
		now:          func() time.Time { return currentTime },
	}

	// First call: only near match (5 days away) is within 2-week window
	_, err := service.getMatchesUpdates()
	require.NoError(t, err)
	require.Len(t, mockRepo.receivedMatches, 1)
	assert.Contains(t, mockRepo.receivedMatches[0], nearMatch.Id())
	assert.NotContains(t, mockRepo.receivedMatches[0], farMatch.Id())

	// near match finished and removed from watchedMatches
	assert.NotContains(t, service.watchedMatches, nearMatch.Id())

	// Advance time by 8 days: farMatch is now 12 days away, within the 2-week window
	currentTime = baseTime.Add(8 * 24 * time.Hour)

	// Second call: far match (now 12 days away) should now be queried
	_, err = service.getMatchesUpdates()
	require.NoError(t, err)
	require.Len(t, mockRepo.receivedMatches, 2)
	assert.Contains(t, mockRepo.receivedMatches[1], farMatch.Id())
}

func TestMatchWatcherService_QueriesPastMatchesThatMayBeRescheduled(t *testing.T) {
	now := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)

	pastMatch := models.NewSeasonMatch("Team1", "Team2", "2025", "Ligue 1", now.Add(-21*24*time.Hour), 1)  // 3 weeks ago
	futureMatch := models.NewSeasonMatch("Team3", "Team4", "2025", "Ligue 1", now.Add(7*24*time.Hour), 2)  // 7 days ahead (control: included)
	farFutureMatch := models.NewSeasonMatch("Team5", "Team6", "2025", "Ligue 1", now.Add(21*24*time.Hour), 3) // 21 days ahead (control: excluded)

	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{{}},
	}

	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: map[string]models.Match{
			pastMatch.Id():      pastMatch,
			futureMatch.Id():    futureMatch,
			farFutureMatch.Id(): farFutureMatch,
		},
		repo:         mockRepo,
		subscribers:  make(map[string]GameService),
		stopChan:     make(chan struct{}),
		pollInterval: 30 * time.Second,
		matchRepo:    repositories.NewInMemoryMatchRepository(),
		now:          func() time.Time { return now },
	}

	_, err := service.getMatchesUpdates()
	require.NoError(t, err)
	require.Len(t, mockRepo.receivedMatches, 1)
	queried := mockRepo.receivedMatches[0]

	assert.Contains(t, queried, pastMatch.Id(), "past-dated unfinished match must be queried so reschedules are detected")
	assert.Contains(t, queried, futureMatch.Id())
	assert.NotContains(t, queried, farFutureMatch.Id())
	assert.Len(t, queried, 2)
}

func TestMatchWatcherService_DetectsRescheduleOfPastMatch(t *testing.T) {
	now := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)
	originalDate := now.Add(-21 * 24 * time.Hour)
	rescheduledDate := now.Add(10 * 24 * time.Hour)

	pastMatch := models.NewSeasonMatch("Team1", "Team2", "2025", "Ligue 1", originalDate, 1)
	rescheduledMatch := models.NewSeasonMatch("Team1", "Team2", "2025", "Ligue 1", rescheduledDate, 1)

	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{
			{pastMatch.Id(): rescheduledMatch},
		},
	}

	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: map[string]models.Match{
			pastMatch.Id(): pastMatch,
		},
		repo:         mockRepo,
		subscribers:  make(map[string]GameService),
		stopChan:     make(chan struct{}),
		pollInterval: 30 * time.Second,
		matchRepo:    repositories.NewInMemoryMatchRepository(),
		now:          func() time.Time { return now },
	}

	updates, err := service.getMatchesUpdates()
	require.NoError(t, err)
	require.Len(t, updates, 1, "rescheduled past match must be reported as an update")
	updatedMatch, ok := updates[pastMatch.Id()]
	require.True(t, ok)
	assert.Equal(t, rescheduledDate, updatedMatch.GetDate())
	assert.Equal(t, rescheduledDate, service.watchedMatches[pastMatch.Id()].GetDate())
}
