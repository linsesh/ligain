package services

import (
	"context"
	"errors"
	"liguain/backend/models"
	"testing"
	"time"
)

type SportsmonkRepositoryMock struct {
	lastMatchInfos []map[string]models.Match
	err            error
	callCount      int
}

func (r *SportsmonkRepositoryMock) GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error) {
	if r.callCount >= len(r.lastMatchInfos) {
		return make(map[string]models.Match), r.err
	}
	result := r.lastMatchInfos[r.callCount]
	r.callCount++
	return result, r.err
}

func TestMatchWatcherServiceSportsmonk_GetUpdates(t *testing.T) {
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
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
	}

	// Add match to watch
	service.WatchMatches([]models.Match{initialMatch})

	// Test GetUpdates
	ctx := context.Background()
	done := make(chan MatchWatcherServiceResult)
	go service.GetUpdates(ctx, done)

	// Wait for result
	result := <-done

	// Verify results
	if result.Err != nil {
		t.Errorf("Expected no error, got %v", result.Err)
	}
	if len(result.Value) != 1 {
		t.Errorf("Expected 1 update, got %d", len(result.Value))
	}
	if result.Value[initialMatch.Id()] != updatedMatch {
		t.Errorf("Expected updated match in result")
	}
	if service.watchedMatches[initialMatch.Id()] != updatedMatch {
		t.Errorf("Expected updated match in watchedMatches")
	}
}

func TestMatchWatcherServiceSportsmonk_GetUpdates_NoChanges(t *testing.T) {
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
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
	}

	// Add match to watch
	service.WatchMatches([]models.Match{match})

	// Test GetUpdates
	ctx := context.Background()
	done := make(chan MatchWatcherServiceResult)
	go service.GetUpdates(ctx, done)

	// Wait for result
	result := <-done

	// Verify results
	if result.Err != nil {
		t.Errorf("Expected no error, got %v", result.Err)
	}
	if len(result.Value) != 0 {
		t.Errorf("Expected no updates, got %d", len(result.Value))
	}
	if service.watchedMatches[match.Id()] != match {
		t.Errorf("Expected match to remain unchanged in watchedMatches")
	}
}

func TestMatchWatcherServiceSportsmonk_GetUpdates_Error(t *testing.T) {
	// Create test match
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)

	// Create mock repository with error
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: nil,
		err:            errors.New("test error"),
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
	}

	// Add match to watch
	service.WatchMatches([]models.Match{match})

	// Test GetUpdates
	ctx := context.Background()
	done := make(chan MatchWatcherServiceResult)
	go service.GetUpdates(ctx, done)

	// Wait for result
	result := <-done

	// Verify results
	if result.Err == nil {
		t.Error("Expected error, got nil")
	}
	if result.Value != nil {
		t.Error("Expected nil value, got non-nil")
	}
	if service.watchedMatches[match.Id()] != match {
		t.Errorf("Expected match to remain unchanged in watchedMatches")
	}
}

func TestMatchWatcherServiceSportsmonk_GetUpdates_MultipleMatches(t *testing.T) {
	// Create test matches
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime, 2)
	match3 := models.NewSeasonMatch("Team5", "Team6", "2024", "Premier League", matchTime, 3)

	updatedMatch1 := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
	updatedMatch2 := models.NewFinishedSeasonMatch("Team3", "Team4", 0, 0, "2024", "Premier League", matchTime, 2, 1.0, 2.0, 3.0)
	// match3 remains unchanged

	// Create mock repository with updated match info
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{
			{
				match1.Id(): updatedMatch1,
				match2.Id(): updatedMatch2,
				match3.Id(): match3,
			},
		},
		err: nil,
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
	}

	// Add matches to watch
	service.WatchMatches([]models.Match{match1, match2, match3})

	// Test GetUpdates
	ctx := context.Background()
	done := make(chan MatchWatcherServiceResult)
	go service.GetUpdates(ctx, done)

	// Wait for result
	result := <-done

	// Verify results
	if result.Err != nil {
		t.Errorf("Expected no error, got %v", result.Err)
	}
	if len(result.Value) != 2 {
		t.Errorf("Expected 2 updates, got %d", len(result.Value))
	}
	if result.Value[match1.Id()] != updatedMatch1 {
		t.Errorf("Expected match1 to be updated")
	}
	if result.Value[match2.Id()] != updatedMatch2 {
		t.Errorf("Expected match2 to be updated")
	}
	if result.Value[match3.Id()] != nil {
		t.Errorf("Expected match3 to remain unchanged")
	}
}

func TestMatchWatcherServiceSportsmonk_GetUpdates_OddsChange(t *testing.T) {
	// Create test match
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	initialMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "Premier League", matchTime, 1, 1.5, 2.0, 3.0)

	// Create updated match with different odds
	updatedMatch := models.NewSeasonMatchWithKnownOdds("Team1", "Team2", "2024", "Premier League", matchTime, 1, 2.5, 3.0, 2.8)

	// Create mock repository
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
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
	}

	// Add match to watch
	service.WatchMatches([]models.Match{initialMatch})

	// Test GetUpdates
	ctx := context.Background()
	done := make(chan MatchWatcherServiceResult)
	go service.GetUpdates(ctx, done)

	// Wait for result
	result := <-done

	// Verify results
	if result.Err != nil {
		t.Errorf("Expected no error, got %v", result.Err)
	}
	if len(result.Value) != 1 {
		t.Errorf("Expected 1 update, got %d", len(result.Value))
	}
	updatedMatchResult := result.Value[initialMatch.Id()]
	if updatedMatchResult.GetHomeTeamOdds() != 2.5 ||
		updatedMatchResult.GetAwayTeamOdds() != 3.0 ||
		updatedMatchResult.GetDrawOdds() != 2.8 {
		t.Errorf("Expected odds to be updated to 2.5/3.0/2.8, got %v/%v/%v",
			updatedMatchResult.GetHomeTeamOdds(),
			updatedMatchResult.GetAwayTeamOdds(),
			updatedMatchResult.GetDrawOdds())
	}
}

func TestMatchWatcherServiceSportsmonk_GetUpdates_Postponed(t *testing.T) {
	// Create test match
	initialTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	postponedTime := time.Date(2024, 1, 17, 15, 0, 0, 0, time.UTC)
	initialMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", initialTime, 1)

	// Create postponed match
	postponedMatch := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", postponedTime, 1)

	// Create mock repository
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{
			{
				initialMatch.Id(): postponedMatch,
			},
		},
		err: nil,
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
	}

	// Add match to watch
	service.WatchMatches([]models.Match{initialMatch})

	// Test GetUpdates
	ctx := context.Background()
	done := make(chan MatchWatcherServiceResult)
	go service.GetUpdates(ctx, done)

	// Wait for result
	result := <-done

	// Verify results
	if result.Err != nil {
		t.Errorf("Expected no error, got %v", result.Err)
	}
	if len(result.Value) != 1 {
		t.Errorf("Expected 1 update, got %d", len(result.Value))
	}
	updatedMatch := result.Value[initialMatch.Id()]
	if updatedMatch.GetDate() != postponedTime {
		t.Errorf("Expected match date to be updated to %v, got %v",
			postponedTime, updatedMatch.GetDate())
	}
}

func TestMatchWatcherServiceSportsmonk_GetUpdates_BatchedUpdates(t *testing.T) {
	// Create test matches
	matchTime := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
	match1 := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", matchTime, 1)
	match2 := models.NewSeasonMatch("Team3", "Team4", "2024", "Premier League", matchTime, 2)
	match3 := models.NewSeasonMatch("Team5", "Team6", "2024", "Premier League", matchTime, 3)

	// First batch of updates
	updatedMatch1 := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.0, 2.0, 3.0)
	updatedMatch2 := models.NewFinishedSeasonMatch("Team3", "Team4", 0, 0, "2024", "Premier League", matchTime, 2, 1.0, 2.0, 3.0)
	// match3 remains unchanged

	// Second batch of updates
	updatedMatch3 := models.NewFinishedSeasonMatch("Team5", "Team6", 1, 0, "2024", "Premier League", matchTime, 3, 1.0, 2.0, 3.0)
	updatedMatch1WithNewOdds := models.NewFinishedSeasonMatch("Team1", "Team2", 2, 1, "2024", "Premier League", matchTime, 1, 1.5, 2.5, 3.5)

	// Create mock repository with batched updates
	mockRepo := &SportsmonkRepositoryMock{
		lastMatchInfos: []map[string]models.Match{
			{
				match1.Id(): updatedMatch1,
				match2.Id(): updatedMatch2,
				match3.Id(): match3,
			},
			{
				match1.Id(): updatedMatch1WithNewOdds,
				match2.Id(): updatedMatch2,
				match3.Id(): updatedMatch3,
			},
		},
		err: nil,
	}

	// Create service instance
	service := &MatchWatcherServiceSportsmonk{
		watchedMatches: make(map[string]models.Match),
		repo:           mockRepo,
	}

	// Add matches to watch
	service.WatchMatches([]models.Match{match1, match2, match3})

	// First update
	ctx := context.Background()
	done1 := make(chan MatchWatcherServiceResult)
	go service.GetUpdates(ctx, done1)
	result1 := <-done1

	// Verify first batch results
	if result1.Err != nil {
		t.Errorf("Expected no error in first batch, got %v", result1.Err)
	}
	if len(result1.Value) != 2 {
		t.Errorf("Expected 2 updates in first batch, got %d", len(result1.Value))
	}
	if result1.Value[match1.Id()] != updatedMatch1 {
		t.Errorf("Expected match1 to be updated in first batch")
	}
	if result1.Value[match2.Id()] != updatedMatch2 {
		t.Errorf("Expected match2 to be updated in first batch")
	}
	if result1.Value[match3.Id()] != nil {
		t.Errorf("Expected match3 to remain unchanged in first batch")
	}

	// Second update
	done2 := make(chan MatchWatcherServiceResult)
	go service.GetUpdates(ctx, done2)
	result2 := <-done2

	// Verify second batch results
	if result2.Err != nil {
		t.Errorf("Expected no error in second batch, got %v", result2.Err)
	}
	if len(result2.Value) != 2 {
		t.Errorf("Expected 2 updates in second batch, got %d", len(result2.Value))
	}
	if result2.Value[match1.Id()] != updatedMatch1WithNewOdds {
		t.Errorf("Expected match1 to have new odds in second batch")
	}
	if result2.Value[match3.Id()] != updatedMatch3 {
		t.Errorf("Expected match3 to be updated in second batch")
	}
	if result2.Value[match2.Id()] != nil {
		t.Errorf("Expected match2 to remain unchanged in second batch")
	}

	// Verify final state
	if service.watchedMatches[match1.Id()] != updatedMatch1WithNewOdds {
		t.Errorf("Expected match1 to have final updated state")
	}
	if service.watchedMatches[match2.Id()] != updatedMatch2 {
		t.Errorf("Expected match2 to have final updated state")
	}
	if service.watchedMatches[match3.Id()] != updatedMatch3 {
		t.Errorf("Expected match3 to have final updated state")
	}
}
