package api

import (
	"os"
	"testing"
)

// TestMain is used to setup integration tests and check if they should run
func TestMain(m *testing.M) {
	// Check if integration tests should be run
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		println("Skipping integration tests. Set INTEGRATION_TESTS=true to run them.")
		os.Exit(0)
	}

	// Get API token from environment
	apiToken := os.Getenv("SPORTSMONK_API_TOKEN")
	if apiToken == "" {
		println("SPORTSMONK_API_TOKEN environment variable is required for integration tests")
		os.Exit(1)
	}

	// Run tests
	os.Exit(m.Run())
}

func TestGetSeasonIds_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	api := NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN"))
	seasonCodes := []string{"2023/24"}

	seasonIds, err := api.GetSeasonIds(seasonCodes, 301)
	if err != nil {
		t.Fatalf("GetSeasonIds failed: %v", err)
	}

	if len(seasonIds) == 0 {
		t.Error("Expected at least one season ID, got none")
	}

	// Check if we got the expected season
	if _, ok := seasonIds["2023/24"]; !ok {
		t.Error("Expected to find season 2023/24")
	}
}

func TestGetFixturesInfos_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	api := NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN"))
	// Using known fixture IDs from Ligue 1
	fixtureIds := []int{1234567} // Replace with actual fixture IDs

	fixtures, err := api.GetFixturesInfos(fixtureIds)
	if err != nil {
		t.Fatalf("GetFixturesInfos failed: %v", err)
	}

	if len(fixtures) == 0 {
		t.Error("Expected at least one fixture, got none")
	}

	// Test the content of the fixtures
	for id, match := range fixtures {
		if match == nil {
			t.Errorf("Match with ID %d is nil", id)
		}
		// Add more specific checks based on your Match interface
	}
}

func TestGetSeasonFixtures_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	api := NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN"))
	// Use a known season ID from Ligue 1
	seasonId := 21787 // Replace with actual season ID

	fixtures, err := api.GetSeasonFixtures(seasonId)
	if err != nil {
		t.Fatalf("GetSeasonFixtures failed: %v", err)
	}

	if len(fixtures) == 0 {
		t.Error("Expected at least one fixture, got none")
	}

	// Test the content of the fixtures
	for id, match := range fixtures {
		if match == nil {
			t.Errorf("Match with ID %d is nil", id)
		}
		// Add more specific checks based on your Match interface
	}
}
