package services

import (
	"context"
	"errors"
	"ligain/backend/models"
	"sync"
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
	gameID   string
	mu       sync.Mutex
	received [][]string // the match IDs of each notification
}

// notifications returns the match IDs of every notification received so far
func (m *MockGameService) notifications() [][]string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([][]string(nil), m.received...)
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

func (m *MockGameService) OnMatchesUpdated(matchIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.received = append(m.received, matchIDs)
	return nil
}

func NewMockGameService(gameID string) *MockGameService {
	return &MockGameService{gameID: gameID}
}

// newTestWatcher builds a watcher over a real MatchService and an in-memory match repository holding the given matches
func newTestWatcher(t *testing.T, sportsmonk *SportsmonkRepositoryMock, now func() time.Time, matches ...models.Match) (*MatchWatcherServiceSportsmonk, repositories.MatchRepository, MatchService) {
	matchRepo := repositories.NewInMemoryMatchRepository()
	for _, m := range matches {
		require.NoError(t, matchRepo.SaveMatch(m))
	}
	matchService := NewMatchServiceWithTime(matchRepo, now)
	return NewMatchWatcherServiceSportsmonk(sportsmonk, matchService, 30*time.Second), matchRepo, matchService
}

func fixedTime(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestMatchWatcherServiceSportsmonk_StartStop(t *testing.T) {
	service, _, _ := newTestWatcher(t, &SportsmonkRepositoryMock{}, fixedTime(time.Now()))

	ctx := context.Background()
	require.NoError(t, service.Start(ctx))
	require.NoError(t, service.Start(ctx)) // starting again should not error
	require.NoError(t, service.Stop())
	require.NoError(t, service.Stop()) // stopping again should not error
}

// pollOnce runs one poll of the watcher and waits for the asynchronous notification of the games
func pollOnce(service *MatchWatcherServiceSportsmonk) {
	service.checkForUpdates()
	time.Sleep(10 * time.Millisecond)
}

func TestMatchWatcherServiceSportsmonk_FinishedMatchIsSavedAndSentToGames(t *testing.T) {
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	initial := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
	finished := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{initial.Id(): finished}}}
	service, matchRepo, matchService := newTestWatcher(t, sportsmonk, fixedTime(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)), initial)
	handler := NewMockGameService("game1")
	require.NoError(t, matchService.Subscribe(handler))

	pollOnce(service)

	saved, err := matchRepo.GetMatch(initial.Id())
	require.NoError(t, err)
	assert.True(t, saved.IsFinished())
	assert.Equal(t, 2, saved.GetHomeGoals())
	assert.Equal(t, [][]string{{initial.Id()}}, handler.notifications())
}

func TestMatchWatcherServiceSportsmonk_NothingChangedNothingSent(t *testing.T) {
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{match.Id(): match.Clone()}}}
	service, _, matchService := newTestWatcher(t, sportsmonk, fixedTime(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)), match)
	handler := NewMockGameService("game1")
	require.NoError(t, matchService.Subscribe(handler))

	pollOnce(service)

	assert.Empty(t, handler.notifications())
}

func TestMatchWatcherServiceSportsmonk_ProviderErrorChangesNothing(t *testing.T) {
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	sportsmonk := &SportsmonkRepositoryMock{err: errors.New("API error")}
	service, _, matchService := newTestWatcher(t, sportsmonk, fixedTime(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)), match)
	handler := NewMockGameService("game1")
	require.NoError(t, matchService.Subscribe(handler))

	pollOnce(service)

	assert.Empty(t, handler.notifications())
}

func TestMatchWatcherServiceSportsmonk_WhatChangesIsSentToGames(t *testing.T) {
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	now := fixedTime(time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC))
	odds := func(h, a, d float64) models.Match {
		return models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "Premier League", matchTime, 1, h, a, d)
	}
	started := func() models.Match { m := odds(1.5, 2.0, 3.0); m.Start(); return m }
	moved := func() *models.SeasonMatch {
		m := odds(1.5, 2.0, 3.0).(*models.SeasonMatch)
		m.Date = matchTime.Add(time.Hour)
		return m
	}

	tests := []struct {
		name     string
		reported models.Match
		sent     bool
	}{
		{"status change", started(), true},
		{"home odds change", odds(2.5, 2.0, 3.0), true},
		{"away odds change", odds(1.5, 3.0, 3.0), true},
		{"draw odds change", odds(1.5, 2.0, 2.8), true},
		{"date change", moved(), true},
		{"nothing changed", odds(1.5, 2.0, 3.0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := odds(1.5, 2.0, 3.0)
			sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{initial.Id(): tt.reported}}}
			service, _, matchService := newTestWatcher(t, sportsmonk, now, initial)
			handler := NewMockGameService("game1")
			require.NoError(t, matchService.Subscribe(handler))

			pollOnce(service)

			assert.Equal(t, tt.sent, len(handler.notifications()) == 1)
		})
	}
}

// Once a match is about to start, the provider keeps returning odds for it (bookmakers settle them at
// extreme prices once the match is over, such as a draw at 1.02). They must never overwrite the
// pre-match odds, neither in the database nor in what the games receive, because scores are computed from the odds.
func TestMatchWatcherServiceSportsmonk_PreMatchOddsAreKeptOnceMatchIsAboutToStart(t *testing.T) {
	kickoff := time.Date(2026, 11, 21, 19, 45, 0, 0, time.UTC)
	preMatch := func() models.Match {
		return models.NewSeasonMatchWithKnownOdds("Lens", "Toulouse", "2026/2027", "Ligue 1", kickoff, 11, 2.12, 3.25, 3.3)
	}
	started := func() models.Match {
		m := models.NewSeasonMatchWithKnownOdds("Lens", "Toulouse", "2026/2027", "Ligue 1", kickoff, 11, 4.35, 9, 1.35)
		m.Start()
		return m
	}
	finished := models.NewFinishedSeasonMatch("Lens", "Toulouse", 1, 1, "2026/2027", "Ligue 1", kickoff, 11, 21, 56, 1.02)

	tests := []struct {
		name     string
		now      time.Time
		reported models.Match
		odds     [3]float64 // expected home, away, draw
	}{
		{"odds reported when the match starts", kickoff.Add(time.Minute), started(), [3]float64{2.12, 3.25, 3.3}},
		{"odds reported when the match finishes", kickoff.Add(2 * time.Hour), finished, [3]float64{2.12, 3.25, 3.3}},
		{"odds still move before the freeze", kickoff.Add(-10 * time.Minute), models.NewSeasonMatchWithKnownOdds("Lens", "Toulouse", "2026/2027", "Ligue 1", kickoff, 11, 2.2, 3.1, 3.4), [3]float64{2.2, 3.1, 3.4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initial := preMatch()
			sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{initial.Id(): tt.reported}}}
			service, matchRepo, matchService := newTestWatcher(t, sportsmonk, fixedTime(tt.now), initial)
			handler := NewMockGameService("game1")
			require.NoError(t, matchService.Subscribe(handler))

			pollOnce(service)

			saved, err := matchRepo.GetMatch(initial.Id())
			require.NoError(t, err)
			assert.Equal(t, tt.odds, [3]float64{saved.GetHomeTeamOdds(), saved.GetAwayTeamOdds(), saved.GetDrawOdds()}, "odds saved")
			assert.Equal(t, [][]string{{initial.Id()}}, handler.notifications(), "games are told the match changed and read it from the MatchService")
		})
	}
}

func TestMatchWatcherServiceSportsmonk_FinishedMatchWithoutUsableOddsIsRejected(t *testing.T) {
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	initial := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1) // never had odds
	finishedWithoutOdds := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 0, 0, 0)
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{initial.Id(): finishedWithoutOdds}}}
	service, matchRepo, matchService := newTestWatcher(t, sportsmonk, fixedTime(matchTime.Add(3*time.Hour)), initial)
	handler := NewMockGameService("game1")
	require.NoError(t, matchService.Subscribe(handler))

	pollOnce(service)

	saved, err := matchRepo.GetMatch(initial.Id())
	require.NoError(t, err)
	assert.False(t, saved.IsFinished(), "a finished match nobody can score must not be saved")
	assert.Empty(t, handler.notifications())
}

func TestMatchWatcherService_OnlyQueriesMatchesWithinTwoWeeks(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	matchWithin := models.NewSeasonMatch("Team1", "Team2", "2025", "Ligue 1", now.Add(7*24*time.Hour), 1)  // 7 days away
	matchEdge := models.NewSeasonMatch("Team3", "Team4", "2025", "Ligue 1", now.Add(13*24*time.Hour), 1)   // 13 days away
	matchBeyond := models.NewSeasonMatch("Team5", "Team6", "2025", "Ligue 1", now.Add(21*24*time.Hour), 1) // 21 days away
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{}}}
	service, _, _ := newTestWatcher(t, sportsmonk, fixedTime(now), matchWithin, matchEdge, matchBeyond)

	pollOnce(service)

	require.Len(t, sportsmonk.receivedMatches, 1)
	assert.Len(t, sportsmonk.receivedMatches[0], 2)
	assert.Contains(t, sportsmonk.receivedMatches[0], matchWithin.Id())
	assert.Contains(t, sportsmonk.receivedMatches[0], matchEdge.Id())
	assert.NotContains(t, sportsmonk.receivedMatches[0], matchBeyond.Id())
}

func TestMatchWatcherService_MatchBecomesRelevantAsTimeAdvances(t *testing.T) {
	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	nearMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2025", "Ligue 1", baseTime.Add(5*24*time.Hour), 1, 1.0, 2.0, 3.0) // 5 days away
	farMatch := models.NewSeasonMatchWithKnownOdds("Team3", "Team4", "2025", "Ligue 1", baseTime.Add(20*24*time.Hour), 1, 1.0, 2.0, 3.0) // 20 days away
	finishedNearMatch := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2025", "Ligue 1", baseTime.Add(5*24*time.Hour), 1, 1.0, 2.0, 3.0)
	currentTime := baseTime
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{
		{nearMatch.Id(): finishedNearMatch},
		{farMatch.Id(): farMatch},
	}}
	service, _, _ := newTestWatcher(t, sportsmonk, func() time.Time { return currentTime }, nearMatch, farMatch)

	// First poll: only the near match (5 days away) is within the 2-week window; it finishes
	pollOnce(service)
	require.Len(t, sportsmonk.receivedMatches, 1)
	assert.Contains(t, sportsmonk.receivedMatches[0], nearMatch.Id())
	assert.NotContains(t, sportsmonk.receivedMatches[0], farMatch.Id())

	// Advance time by 8 days: the far match is now 12 days away, and the finished one is not asked about anymore
	currentTime = baseTime.Add(8 * 24 * time.Hour)
	pollOnce(service)
	require.Len(t, sportsmonk.receivedMatches, 2)
	assert.Contains(t, sportsmonk.receivedMatches[1], farMatch.Id())
	assert.NotContains(t, sportsmonk.receivedMatches[1], nearMatch.Id())
}

func TestMatchWatcherService_QueriesPastMatchesThatMayBeRescheduled(t *testing.T) {
	now := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)
	pastMatch := models.NewSeasonMatch("Team1", "Team2", "2025", "Ligue 1", now.Add(-21*24*time.Hour), 1)     // 3 weeks ago
	futureMatch := models.NewSeasonMatch("Team3", "Team4", "2025", "Ligue 1", now.Add(7*24*time.Hour), 2)     // 7 days ahead (control: included)
	farFutureMatch := models.NewSeasonMatch("Team5", "Team6", "2025", "Ligue 1", now.Add(21*24*time.Hour), 3) // 21 days ahead (control: excluded)
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{}}}
	service, _, _ := newTestWatcher(t, sportsmonk, fixedTime(now), pastMatch, futureMatch, farFutureMatch)

	pollOnce(service)

	require.Len(t, sportsmonk.receivedMatches, 1)
	queried := sportsmonk.receivedMatches[0]
	assert.Contains(t, queried, pastMatch.Id(), "an unfinished past match may have been rescheduled")
	assert.Contains(t, queried, futureMatch.Id())
	assert.NotContains(t, queried, farFutureMatch.Id())
}

func TestMatchWatcherService_DetectsRescheduleOfPastMatch(t *testing.T) {
	now := time.Date(2025, 1, 22, 0, 0, 0, 0, time.UTC)
	rescheduledDate := now.Add(10 * 24 * time.Hour)
	pastMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2025", "Ligue 1", now.Add(-21*24*time.Hour), 1, 1.5, 2.0, 3.0)
	rescheduledMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2025", "Ligue 1", rescheduledDate, 1, 1.5, 2.0, 3.0)
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{pastMatch.Id(): rescheduledMatch}}}
	service, matchRepo, matchService := newTestWatcher(t, sportsmonk, fixedTime(now), pastMatch)
	handler := NewMockGameService("game1")
	require.NoError(t, matchService.Subscribe(handler))

	pollOnce(service)

	saved, err := matchRepo.GetMatch(pastMatch.Id())
	require.NoError(t, err)
	assert.Equal(t, rescheduledDate, saved.GetDate())
	assert.Equal(t, [][]string{{pastMatch.Id()}}, handler.notifications(), "a rescheduled past match must be reported as an update")
}

func TestMatchWatcherService_OnlyQueriesTheConfiguredCompetitionAndSeason(t *testing.T) {
	now := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	current := models.NewSeasonMatch("Team1", "Team2", "2026/2027", "Ligue 1", now.Add(24*time.Hour), 1)
	oldSeasonStuck := models.NewSeasonMatch("Team3", "Team4", "2025/2026", "Ligue 1", now.Add(-190*24*time.Hour), 38) // never finished
	otherCompetition := models.NewSeasonMatch("Team5", "Team6", "2026/2027", "Serie A", now.Add(24*time.Hour), 1)
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{}}}
	matchRepo := repositories.NewInMemoryMatchRepository()
	for _, m := range []models.Match{current, oldSeasonStuck, otherCompetition} {
		require.NoError(t, matchRepo.SaveMatch(m))
	}
	matchService := NewMatchServiceWithTime(matchRepo, fixedTime(now), WatchOnly("Ligue 1", "2026/2027"))
	service := NewMatchWatcherServiceSportsmonk(sportsmonk, matchService, 30*time.Second)

	pollOnce(service)

	require.Len(t, sportsmonk.receivedMatches, 1)
	assert.Len(t, sportsmonk.receivedMatches[0], 1)
	assert.Contains(t, sportsmonk.receivedMatches[0], current.Id())
}

// After the freeze, the provider keeps changing the odds (bookmakers settle them at extreme prices once the match
// is over). With nothing else changed, there is nothing to save and nothing to tell the games.
func TestMatchWatcherServiceSportsmonk_OddsNoiseAfterKickoffIsNotAnUpdate(t *testing.T) {
	kickoff := time.Date(2026, 11, 21, 19, 45, 0, 0, time.UTC)
	started := models.NewSeasonMatchWithKnownOdds("Lens", "Toulouse", "2026/2027", "Ligue 1", kickoff, 11, 2.12, 3.25, 3.3)
	started.Start()
	settled := models.NewSeasonMatchWithKnownOdds("Lens", "Toulouse", "2026/2027", "Ligue 1", kickoff, 11, 4.35, 9, 1.35)
	settled.Start()
	sportsmonk := &SportsmonkRepositoryMock{lastMatchInfos: []map[string]models.Match{{started.Id(): settled}}}
	service, matchRepo, matchService := newTestWatcher(t, sportsmonk, fixedTime(kickoff.Add(time.Hour)), started)
	handler := NewMockGameService("game1")
	require.NoError(t, matchService.Subscribe(handler))

	pollOnce(service)

	saved, err := matchRepo.GetMatch(started.Id())
	require.NoError(t, err)
	assert.Equal(t, 2.12, saved.GetHomeTeamOdds())
	assert.Empty(t, handler.notifications())
}
