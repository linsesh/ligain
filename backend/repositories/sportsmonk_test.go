package repositories

import (
	"ligain/backend/models"
	"testing"
	"time"
)

var matchTime = time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)

// MockSportsmonkAPI implements the SportsmonkAPI interface for testing
type MockSportsmonkAPI struct {
	futureUpdates      []map[int]models.Match
	matchIdToFixtureId map[string]int
	futureUpdatesIndex int
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
		result[code] = i
	}
	return result, nil
}

func (m *MockSportsmonkAPI) GetFixturesInfos(fixtureIds []int) (map[int]models.Match, error) {
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
