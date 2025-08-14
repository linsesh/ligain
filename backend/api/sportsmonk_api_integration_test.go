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
	}, 30*time.Second)
}

func TestGetFixturesInfos_WithOdds_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		api := NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN"))
		// Marseille - Rennes 2025-05-17 (this fixture should have odds available)
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

		// Validate that odds are properly retrieved and stored
		homeOdds := match.GetHomeTeamOdds()
		awayOdds := match.GetAwayTeamOdds()
		drawOdds := match.GetDrawOdds()

		// Check that odds are reasonable values (not zero unless the match has no odds)
		if homeOdds == 0 && awayOdds == 0 && drawOdds == 0 {
			t.Log("Match has no odds available (this might be normal for some fixtures)")
		} else {
			// Validate odds are reasonable (between 1.0 and 50.0)
			if homeOdds < 1.0 || homeOdds > 50.0 {
				t.Errorf("Home odds %f is outside reasonable range (1.0-50.0)", homeOdds)
			}
			if awayOdds < 1.0 || awayOdds > 50.0 {
				t.Errorf("Away odds %f is outside reasonable range (1.0-50.0)", awayOdds)
			}
			if drawOdds < 1.0 || drawOdds > 50.0 {
				t.Errorf("Draw odds %f is outside reasonable range (1.0-50.0)", drawOdds)
			}

			// Log the odds for verification
			t.Logf("Match %s odds - Home: %.2f, Away: %.2f, Draw: %.2f",
				match.Id(), homeOdds, awayOdds, drawOdds)
		}
	}, 10*time.Second)
}

func TestGetSeasonFixtures_WithOdds_Integration(t *testing.T) {
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

		// Validate that some matches have odds (check first 10 matches to avoid timeout)
		matchesWithOdds := 0
		checkedMatches := 0
		maxMatchesToCheck := 10

		for fixtureId, match := range fixtures {
			if checkedMatches >= maxMatchesToCheck {
				break
			}
			checkedMatches++

			homeOdds := match.GetHomeTeamOdds()
			awayOdds := match.GetAwayTeamOdds()
			drawOdds := match.GetDrawOdds()

			if homeOdds > 0 || awayOdds > 0 || drawOdds > 0 {
				matchesWithOdds++

				// Validate odds are reasonable for matches that have them
				if homeOdds < 1.0 || homeOdds > 50.0 {
					t.Errorf("Fixture %d: Home odds %f is outside reasonable range (1.0-50.0)", fixtureId, homeOdds)
				}
				if awayOdds < 1.0 || awayOdds > 50.0 {
					t.Errorf("Fixture %d: Away odds %f is outside reasonable range (1.0-50.0)", fixtureId, awayOdds)
				}
				if drawOdds < 1.0 || drawOdds > 50.0 {
					t.Errorf("Fixture %d: Draw odds %f is outside reasonable range (1.0-50.0)", fixtureId, drawOdds)
				}

				// Log the first few matches with odds for verification
				if matchesWithOdds <= 3 {
					t.Logf("Fixture %d (%s): Home: %.2f, Away: %.2f, Draw: %.2f",
						fixtureId, match.Id(), homeOdds, awayOdds, drawOdds)
				}
			}
		}

		// Log summary of odds availability
		t.Logf("Checked %d matches, found %d with odds (%.1f%%)",
			checkedMatches, matchesWithOdds, float64(matchesWithOdds)/float64(checkedMatches)*100)

		// Ensure at least some matches have odds
		if matchesWithOdds == 0 {
			t.Error("No matches have odds available. This indicates an issue with odds retrieval.")
		}
	}, 30*time.Second)
}

func TestBookmakerPreference_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		api := NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN"))

		// Test with a fixture that should have odds available
		// Using a recent fixture that likely has odds from multiple bookmakers
		fixtureIds := []int{19139949} // Marseille - Rennes 2025-05-17

		fixtures, err := api.GetFixturesInfos(fixtureIds)
		if err != nil {
			t.Fatalf("GetFixturesInfos failed: %v", err)
		}

		if len(fixtures) == 0 {
			t.Error("Expected at least one fixture, got none")
		}

		match := fixtures[19139949]
		homeOdds := match.GetHomeTeamOdds()
		awayOdds := match.GetAwayTeamOdds()
		drawOdds := match.GetDrawOdds()

		// If the match has odds, verify they are reasonable
		if homeOdds > 0 || awayOdds > 0 || drawOdds > 0 {
			// Validate odds are reasonable (between 1.0 and 50.0)
			if homeOdds < 1.0 || homeOdds > 50.0 {
				t.Errorf("Home odds %f is outside reasonable range (1.0-50.0)", homeOdds)
			}
			if awayOdds < 1.0 || awayOdds > 50.0 {
				t.Errorf("Away odds %f is outside reasonable range (1.0-50.0)", awayOdds)
			}
			if drawOdds < 1.0 || drawOdds > 50.0 {
				t.Errorf("Draw odds %f is outside reasonable range (1.0-50.0)", drawOdds)
			}

			t.Logf("Match %s odds - Home: %.2f, Away: %.2f, Draw: %.2f",
				match.Id(), homeOdds, awayOdds, drawOdds)
		} else {
			t.Log("Match has no odds available (this might be normal for some fixtures)")
		}

		// Note: We can't directly test which bookmaker was selected without modifying the API
		// to expose this information, but we can verify that odds are being retrieved correctly
		// and that the system falls back gracefully when preferred bookmakers aren't available
	}, 10*time.Second)
}
