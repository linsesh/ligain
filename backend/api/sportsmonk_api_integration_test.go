package api

import (
	"context"
	"os"
	"testing"
	"time"
)

// runTestWithTimeout runs a test function with a global timeout
func runTestWithTimeout(t *testing.T, testFunc func(t *testing.T), timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		testFunc(t)
	}()

	select {
	case <-ctx.Done():
		t.Fatal("Test suite timed out after", timeout)
	case <-done:
		// Test completed successfully
	}
}

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

	runTestWithTimeout(t, func(t *testing.T) {
		api := NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN"))
		seasonCodes := []string{"2023/2024"}

		seasonIds, err := api.GetSeasonIds(seasonCodes, 301)
		if err != nil {
			t.Fatalf("GetSeasonIds failed: %v", err)
		}

		if len(seasonIds) == 0 {
			t.Error("Expected at least one season ID, got none")
		}

		// Check if we got the expected season
		if _, ok := seasonIds["2023/2024"]; !ok {
			t.Error("Expected to find season 2023/2024")
		}
	}, 10*time.Second)
}

func TestGetFixturesInfos_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		api := NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN"))
		// Marseille - Rennes 2025-05-17
		fixtureIds := []int{19139949}

		fixtures, err := api.GetFixturesInfos(fixtureIds)
		if err != nil {
			t.Fatalf("GetFixturesInfos failed: %v", err)
		}

		if len(fixtures) == 0 {
			t.Error("Expected at least one fixture, got none")
		}

		match := fixtures[19139949]
		if match.Id() != "Ligue 1-2024/2025-Olympique Marseille-Rennes-34" {
			t.Errorf("Expected match ID Ligue 1-2024/2025-Olympique Marseille-Rennes-34, got %s", match.Id())
		}
	}, 10*time.Second)
}

func TestGetSeasonFixtures_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		api := NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN"))
		seasonId := 23643 // Ligue 1 2024/2025
		fixtures, err := api.GetSeasonFixtures(seasonId)
		if err != nil {
			t.Fatalf("GetSeasonFixtures failed: %v", err)
		}

		if len(fixtures) == 0 {
			t.Error("Expected at least one fixture, got none")
		}

		// Check if we got the expected number of fixtures (34 * 9 matches per round)
		if len(fixtures) != 306 {
			t.Errorf("Expected 306 fixtures, got %d", len(fixtures))
		}
	}, 10*time.Second)
}
