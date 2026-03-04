package repositories

import (
	"fmt"
	"ligain/backend/models"
	"testing"
	"time"
)

var matchTime = time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

// MockSportsmonkAPI implements the SportsmonkAPI interface for testing
type MockSportsmonkAPI struct {
	futureUpdates        []map[int]models.Match
	matchIdToFixtureId   map[string]int
	futureUpdatesIndex   int
	seasonFixturesCalls  int
	fixtureInfosCalls    int
}

func NewMockSportsmonkAPI(futureUpdates []map[int]models.Match) *MockSportsmonkAPI {
	matchIdToFixtureId := make(map[string]int)
	for _, batchOfUpdates := range futureUpdates {
		for fixtureId, match := range batchOfUpdates {
			matchIdToFixtureId[match.Id()] = fixtureId
		}
	}
	return &MockSportsmonkAPI{
		futureUpdates:      futureUpdates,
		futureUpdatesIndex: 0,
		matchIdToFixtureId: matchIdToFixtureId,
	}
}

func (m *MockSportsmonkAPI) GetSeasonIds(seasonCodes []string, competitionId int) (map[string]int, error) {
	result := make(map[string]int)
	for i, code := range seasonCodes {
		result[code] = 100 + i // Use 100+ to avoid confusion with default 0 value
	}
	return result, nil
}

func (m *MockSportsmonkAPI) GetFixturesInfos(fixtureIds []int) (map[int]models.Match, error) {
	m.fixtureInfosCalls++
	if m.futureUpdatesIndex >= len(m.futureUpdates) {
		return make(map[int]models.Match), nil
	}

	result := make(map[int]models.Match)
	currentUpdates := m.futureUpdates[m.futureUpdatesIndex]
	m.futureUpdatesIndex++

	for _, fixtureId := range fixtureIds {
		if match, exists := currentUpdates[fixtureId]; exists {
			result[fixtureId] = match
		}
	}

	return result, nil
}

func (m *MockSportsmonkAPI) GetSeasonFixtures(seasonId int) (map[int]models.Match, error) {
	m.seasonFixturesCalls++
	if m.futureUpdatesIndex >= len(m.futureUpdates) {
		return make(map[int]models.Match), nil
	}

	currentUpdates := m.futureUpdates[m.futureUpdatesIndex]
	m.futureUpdatesIndex++

	return currentUpdates, nil
}

func (m *MockSportsmonkAPI) GetCompetitionId(competitionCode string) (int, error) {
	return 1, nil
}

func TestGetLastMatchInfosWithNoUpdatesOddsUpdatesResultUpdates(t *testing.T) {
	// Setup initial matches
	matches := make(map[string]models.Match)
	match1 := models.NewSeasonMatch(
		"team1",
		"team2",
		"season1",
		"comp1",
		matchTime,
		1,
	)
	match2 := models.NewSeasonMatch(
		"team3",
		"team4",
		"season2",
		"comp2",
		matchTime,
		1,
	)
	matches[match1.Id()] = match1
	matches[match2.Id()] = match2

	// Create future updates with different scenarios
	futureUpdates := []map[int]models.Match{
		// First update: No changes
		make(map[int]models.Match),
		// Second update: Only match1 is updated
		{
			1: models.NewSeasonMatchWithKnownOdds(
				"team1",
				"team2",
				"season1",
				"comp1",
				matchTime,
				1,
				2.0, 3.0, 4.0,
			),
		},
		// Third update: Only match2 is updated
		{
			2: models.NewFinishedSeasonMatch(
				"team3",
				"team4",
				2, 1,
				"season2",
				"comp2",
				matchTime,
				1,
				1.0, 2.0, 3.0,
			),
		},
	}

	mockAPI := NewMockSportsmonkAPI(futureUpdates)
	repo := NewSportsmonkRepository(mockAPI)

	// Test first call - should get initial state
	result, err := repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("GetLastMatchInfos returned an error: %v", err)
	}

	// Verify initial state - no updates expected
	if len(result) != 0 {
		t.Errorf("Expected no updates in initial state, got %d", len(result))
	}

	// Test second call - should get update for match1
	result, err = repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("GetLastMatchInfos returned an error: %v", err)
	}

	// Verify only match1 is updated
	if len(result) != 1 {
		t.Errorf("Expected 1 update, got %d", len(result))
	}
	if _, exists := result[match1.Id()]; !exists {
		t.Errorf("Expected match1 to be updated")
	}
	if _, exists := result[match2.Id()]; exists {
		t.Errorf("Did not expect match2 to be updated")
	}

	// Test third call - should get update for match2
	result, err = repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("GetLastMatchInfos returned an error: %v", err)
	}

	// Verify only match2 is updated
	if len(result) != 1 {
		t.Errorf("Expected 1 update, got %d", len(result))
	}
	if _, exists := result[match1.Id()]; exists {
		t.Errorf("Did not expect match1 to be updated")
	}
	if _, exists := result[match2.Id()]; !exists {
		t.Errorf("Expected match2 to be updated")
	}
}

func TestGetLastMatchInfosWithEmptyInput(t *testing.T) {
	// Setup
	mockAPI := NewMockSportsmonkAPI([]map[int]models.Match{})
	repo := NewSportsmonkRepository(mockAPI)

	// Test with empty input
	emptyMatches := make(map[string]models.Match)
	result, err := repo.GetLastMatchInfos(emptyMatches)
	if err != nil {
		t.Fatalf("GetLastMatchInfos returned an error: %v", err)
	}

	// Verify empty result
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d matches", len(result))
	}
}

func TestGetLastMatchInfosWithCaching(t *testing.T) {
	// Setup matches with same season and competition codes
	matches := make(map[string]models.Match)
	match1 := models.NewSeasonMatch(
		"team1",
		"team2",
		"season1",
		"comp1",
		matchTime,
		1,
	)
	match2 := models.NewSeasonMatch(
		"team3",
		"team4",
		"season1", // Same season code as match1
		"comp1",   // Same competition code as match1
		matchTime,
		2,
	)
	matches[match1.Id()] = match1
	matches[match2.Id()] = match2

	// Create future updates
	futureUpdates := []map[int]models.Match{
		// First update: Both matches updated
		{
			1: models.NewSeasonMatchWithKnownOdds(
				"team1",
				"team2",
				"season1",
				"comp1",
				matchTime,
				1,
				2.0, 3.0, 4.0,
			),
			2: models.NewSeasonMatchWithKnownOdds(
				"team3",
				"team4",
				"season1",
				"comp1",
				matchTime,
				2,
				1.5, 2.5, 3.5,
			),
		},
		// Second update: Only match1 updated (to test caching)
		{
			1: models.NewFinishedSeasonMatch(
				"team1",
				"team2",
				2, 1,
				"season1",
				"comp1",
				matchTime,
				1,
				2.0, 3.0, 4.0,
			),
		},
	}

	mockAPI := NewMockSportsmonkAPI(futureUpdates)
	repo := NewSportsmonkRepository(mockAPI)

	// First call - should populate cache
	result, err := repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("First GetLastMatchInfos returned an error: %v", err)
	}

	// Verify both matches are returned
	if len(result) != 2 {
		t.Errorf("Expected 2 matches in first call, got %d", len(result))
	}

	// Second call - should use cached season ID
	result, err = repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("Second GetLastMatchInfos returned an error: %v", err)
	}

	// Verify only match1 is returned (as it's the only one updated in second batch)
	if len(result) != 1 {
		t.Errorf("Expected 1 match in second call, got %d", len(result))
	}
	if _, exists := result[match1.Id()]; !exists {
		t.Errorf("Expected match1 to be returned in second call")
	}
	if _, exists := result[match2.Id()]; exists {
		t.Errorf("Did not expect match2 to be returned in second call")
	}
}

// TestAskAndCacheSeasonIdCaching tests the specific caching bug in askAndCacheSeasonId
func TestAskAndCacheSeasonIdCaching(t *testing.T) {
	mockAPI := NewMockSportsmonkAPI([]map[int]models.Match{})
	repo := &SportsmonkRepositoryImpl{
		api:                            mockAPI,
		seasonCodeToSeasonId:           make(map[string]int),
		competitionCodeToCompetitionId: make(map[string]int),
	}

	matches := make(map[string]models.Match)
	match := models.NewSeasonMatch(
		"team1",
		"team2",
		"season1",
		"comp1",
		matchTime,
		1,
	)
	matches[match.Id()] = match

	// First call - should populate cache with seasonId = 100
	seasonId1, err := repo.askAndCacheSeasonId(matches)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if seasonId1 != 100 {
		t.Errorf("Expected first call to return seasonId = 100, got %d", seasonId1)
	}

	// Second call - should return cached seasonId = 100, not 0
	seasonId2, err := repo.askAndCacheSeasonId(matches)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if seasonId2 != 100 {
		t.Errorf("Expected second call to return cached seasonId = 100, got %d", seasonId2)
	}
}

// --- Fixture ID caching tests ---

// newTwoMatchesSameSeason creates two matches in the same season/competition for use in caching tests.
func newTwoMatchesSameSeason() (models.Match, models.Match, map[string]models.Match) {
	m1 := models.NewSeasonMatch("team1", "team2", "season1", "comp1", matchTime, 1)
	m2 := models.NewSeasonMatch("team3", "team4", "season1", "comp1", matchTime, 2)
	return m1, m2, map[string]models.Match{m1.Id(): m1, m2.Id(): m2}
}

// FailingMockSportsmonkAPI wraps a MockSportsmonkAPI and makes GetSeasonFixtures fail a fixed number of times.
type FailingMockSportsmonkAPI struct {
	inner         *MockSportsmonkAPI
	failRemaining int
}

func (f *FailingMockSportsmonkAPI) GetSeasonIds(codes []string, competitionId int) (map[string]int, error) {
	return f.inner.GetSeasonIds(codes, competitionId)
}
func (f *FailingMockSportsmonkAPI) GetFixturesInfos(ids []int) (map[int]models.Match, error) {
	return f.inner.GetFixturesInfos(ids)
}
func (f *FailingMockSportsmonkAPI) GetCompetitionId(code string) (int, error) {
	return f.inner.GetCompetitionId(code)
}
func (f *FailingMockSportsmonkAPI) GetSeasonFixtures(seasonId int) (map[int]models.Match, error) {
	if f.failRemaining > 0 {
		f.failRemaining--
		f.inner.seasonFixturesCalls++ // count the failing call manually
		return nil, fmt.Errorf("simulated GetSeasonFixtures failure")
	}
	return f.inner.GetSeasonFixtures(seasonId) // inner counts successful calls itself
}

// TestFixtureIdCaching_UsesGetFixturesInfosAfterCachePopulated verifies that after
// a full season fetch the repository switches to targeted GetFixturesInfos calls.
func TestFixtureIdCaching_UsesGetFixturesInfosAfterCachePopulated(t *testing.T) {
	m1, m2, matches := newTwoMatchesSameSeason()

	futureUpdates := []map[int]models.Match{
		{1: m1, 2: m2},                           // full fetch (index 0)
		{1: models.NewFinishedSeasonMatch(        // GetFixturesInfos (index 1)
			"team1", "team2", 2, 1, "season1", "comp1", matchTime, 1, 2.0, 3.0, 4.0,
		)},
	}

	mockAPI := NewMockSportsmonkAPI(futureUpdates)
	repo := NewSportsmonkRepository(mockAPI).(*SportsmonkRepositoryImpl)

	// First call: must do a full season fetch to populate the cache.
	_, err := repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if mockAPI.seasonFixturesCalls != 1 {
		t.Errorf("expected 1 GetSeasonFixtures call after first poll, got %d", mockAPI.seasonFixturesCalls)
	}
	if mockAPI.fixtureInfosCalls != 0 {
		t.Errorf("expected 0 GetFixturesInfos calls after first poll, got %d", mockAPI.fixtureInfosCalls)
	}

	// Second call (within TTL, all IDs cached): must use GetFixturesInfos only.
	_, err = repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if mockAPI.seasonFixturesCalls != 1 {
		t.Errorf("expected still 1 GetSeasonFixtures call after second poll, got %d", mockAPI.seasonFixturesCalls)
	}
	if mockAPI.fixtureInfosCalls != 1 {
		t.Errorf("expected 1 GetFixturesInfos call after second poll, got %d", mockAPI.fixtureInfosCalls)
	}
}

// TestFixtureIdCaching_RefetchAfter12Hours verifies that the cache TTL triggers a new
// full season fetch after 12 hours.
func TestFixtureIdCaching_RefetchAfter12Hours(t *testing.T) {
	m1, m2, matches := newTwoMatchesSameSeason()

	futureUpdates := []map[int]models.Match{
		{1: m1, 2: m2}, // first full fetch
		{1: m1, 2: m2}, // second full fetch after TTL expires
	}

	mockAPI := NewMockSportsmonkAPI(futureUpdates)
	repo := NewSportsmonkRepository(mockAPI).(*SportsmonkRepositoryImpl)

	// First call: populates cache.
	_, err := repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Simulate 13 hours passing.
	repo.lastFullFetch = time.Now().Add(-13 * time.Hour)

	// Second call: cache is stale, must do a full fetch again.
	_, err = repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if mockAPI.seasonFixturesCalls != 2 {
		t.Errorf("expected 2 GetSeasonFixtures calls after TTL expired, got %d", mockAPI.seasonFixturesCalls)
	}
	if mockAPI.fixtureInfosCalls != 0 {
		t.Errorf("expected 0 GetFixturesInfos calls, got %d", mockAPI.fixtureInfosCalls)
	}
}

// TestFixtureIdCaching_RefetchAfterFailure verifies that a failed full fetch causes the
// next call to retry the full fetch. Since lastFullFetch is only updated on success, a
// failed fetch leaves the timestamp stale, which naturally triggers a retry.
func TestFixtureIdCaching_RefetchAfterFailure(t *testing.T) {
	m1, m2, matches := newTwoMatchesSameSeason()

	innerMock := NewMockSportsmonkAPI([]map[int]models.Match{
		{1: m1, 2: m2}, // succeeds on the retry
	})
	failingAPI := &FailingMockSportsmonkAPI{inner: innerMock, failRemaining: 1}

	repo := NewSportsmonkRepository(failingAPI).(*SportsmonkRepositoryImpl)

	// First call: GetSeasonFixtures fails; lastFullFetch is not updated.
	_, err := repo.GetLastMatchInfos(matches)
	if err == nil {
		t.Fatal("expected first call to fail")
	}

	// Second call: lastFullFetch is still zero (stale), so a retry is triggered.
	_, err = repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if innerMock.seasonFixturesCalls != 2 {
		t.Errorf("expected 2 GetSeasonFixtures calls (1 fail + 1 retry), got %d", innerMock.seasonFixturesCalls)
	}
}

// TestFixtureIdCaching_RefetchWhenNewMatchAdded verifies that adding a new watched match
// whose fixture ID is not yet cached triggers a full season fetch.
func TestFixtureIdCaching_RefetchWhenNewMatchAdded(t *testing.T) {
	m1 := models.NewSeasonMatch("team1", "team2", "season1", "comp1", matchTime, 1)
	matches := map[string]models.Match{m1.Id(): m1}

	m2 := models.NewSeasonMatch("team3", "team4", "season1", "comp1", matchTime, 2)

	futureUpdates := []map[int]models.Match{
		{1: m1},      // first full fetch — only m1 in season
		{1: m1, 2: m2}, // second full fetch after m2 is added
	}

	mockAPI := NewMockSportsmonkAPI(futureUpdates)
	repo := NewSportsmonkRepository(mockAPI).(*SportsmonkRepositoryImpl)

	// First call: caches only m1's fixture ID.
	_, err := repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if mockAPI.seasonFixturesCalls != 1 {
		t.Errorf("expected 1 GetSeasonFixtures call, got %d", mockAPI.seasonFixturesCalls)
	}

	// Add m2 to the watched list.
	matches[m2.Id()] = m2

	// Second call: m2 has no cached fixture ID, so a full fetch is required.
	_, err = repo.GetLastMatchInfos(matches)
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if mockAPI.seasonFixturesCalls != 2 {
		t.Errorf("expected 2 GetSeasonFixtures calls after adding new match, got %d", mockAPI.seasonFixturesCalls)
	}
	if mockAPI.fixtureInfosCalls != 0 {
		t.Errorf("expected 0 GetFixturesInfos calls, got %d", mockAPI.fixtureInfosCalls)
	}
}
