package postgres

import (
	"testing"
	"time"

	"ligain/backend/models"

	"github.com/stretchr/testify/require"
)

func TestMatchRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		matchRepo := NewPostgresMatchRepository(testDB.db)

		t.Run("Save and Get Match", func(t *testing.T) {
			// Setup test data
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ('123e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')
			`)
			require.NoError(t, err)

			// Create match
			match := models.NewSeasonMatch("Team A", "Team B", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
			err = matchRepo.SaveMatch(match)
			require.NoError(t, err)

			// Get match
			retrievedMatch, err := matchRepo.GetMatch(match.Id())
			require.NoError(t, err)
			require.Equal(t, match.GetHomeTeam(), retrievedMatch.GetHomeTeam())
			require.Equal(t, match.GetAwayTeam(), retrievedMatch.GetAwayTeam())
			require.Equal(t, match.GetDate().Unix(), retrievedMatch.GetDate().Unix())
		})

		t.Run("Save and Get Match with Odds", func(t *testing.T) {
			// Setup test data
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ('223e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')
			`)
			require.NoError(t, err)

			// Create match with odds
			match := models.NewSeasonMatch("Team C", "Team D", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
			match.SetHomeTeamOdds(1.5)
			match.SetAwayTeamOdds(2.5)
			match.SetDrawOdds(3.0)
			err = matchRepo.SaveMatch(match)
			require.NoError(t, err)

			// Get match
			retrievedMatch, err := matchRepo.GetMatch(match.Id())
			require.NoError(t, err)
			require.Equal(t, match.GetHomeTeam(), retrievedMatch.GetHomeTeam())
			require.Equal(t, match.GetAwayTeam(), retrievedMatch.GetAwayTeam())
			require.Equal(t, match.GetDate().Unix(), retrievedMatch.GetDate().Unix())

			// Verify odds are correctly stored and retrieved
			require.Equal(t, 1.5, retrievedMatch.GetHomeTeamOdds())
			require.Equal(t, 2.5, retrievedMatch.GetAwayTeamOdds())
			require.Equal(t, 3.0, retrievedMatch.GetDrawOdds())
		})

		t.Run("Get All Matches", func(t *testing.T) {
			// Setup test data
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ('323e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')
			`)
			require.NoError(t, err)

			// Create multiple matches
			match1 := models.NewSeasonMatch("Team E", "Team F", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
			match2 := models.NewSeasonMatch("Team G", "Team H", "2024", "Premier League", testTime.Add(48*time.Hour), 2)

			err = matchRepo.SaveMatch(match1)
			require.NoError(t, err)
			err = matchRepo.SaveMatch(match2)
			require.NoError(t, err)

			// Get all matches
			matches, err := matchRepo.GetMatches()
			require.NoError(t, err)
			// We expect 4 matches total: 1 from first test + 1 from second test + 2 from this test
			require.Equal(t, 4, len(matches))
		})

		t.Run("Get Match for Non-existent Game", func(t *testing.T) {
			// Try to get match for non-existent match with proper UUID format
			_, err := matchRepo.GetMatch("423e4567-e89b-12d3-a456-426614174999")
			require.Error(t, err)
		})

		t.Run("SQL Scanning Issues Prevention - GetMatchesByCompetitionAndSeason with Odds", func(t *testing.T) {
			// This test specifically verifies that the SQL scanning issue we fixed doesn't occur
			// when getting matches by competition and season with odds data

			// Create multiple matches with odds for the same competition/season
			match1 := models.NewSeasonMatch("Team I", "Team J", "2024", "Ligue 1", testTime.Add(24*time.Hour), 1)
			match1.SetHomeTeamOdds(2.5)
			match1.SetAwayTeamOdds(3.2)
			match1.SetDrawOdds(3.0)

			match2 := models.NewSeasonMatch("Team K", "Team L", "2024", "Ligue 1", testTime.Add(48*time.Hour), 1)
			match2.SetHomeTeamOdds(1.8)
			match2.SetAwayTeamOdds(4.1)
			match2.SetDrawOdds(3.5)

			// Save matches
			err := matchRepo.SaveMatch(match1)
			require.NoError(t, err, "Saving match1 should succeed")
			err = matchRepo.SaveMatch(match2)
			require.NoError(t, err, "Saving match2 should succeed")

			// Test the method that was causing SQL scanning issues
			matches, err := matchRepo.GetMatchesByCompetitionAndSeason("Ligue 1", "2024")
			require.NoError(t, err, "GetMatchesByCompetitionAndSeason should not have SQL scanning issues")
			require.Len(t, matches, 2, "Should return exactly 2 matches")

			// Verify odds are properly loaded for all matches
			for i, match := range matches {
				require.Greater(t, match.GetHomeTeamOdds(), 0.0, "Match %d should have home odds", i+1)
				require.Greater(t, match.GetAwayTeamOdds(), 0.0, "Match %d should have away odds", i+1)
				require.Greater(t, match.GetDrawOdds(), 0.0, "Match %d should have draw odds", i+1)
			}
		})

		t.Run("SQL Scanning Issues Prevention - GetMatches with Odds", func(t *testing.T) {
			// This test verifies that GetMatches doesn't have SQL scanning issues with odds

			// Create a match with odds
			match := models.NewSeasonMatch("Team M", "Team N", "2024", "Serie A", testTime.Add(24*time.Hour), 1)
			match.SetHomeTeamOdds(2.0)
			match.SetAwayTeamOdds(3.5)
			match.SetDrawOdds(3.2)

			// Save match
			err := matchRepo.SaveMatch(match)
			require.NoError(t, err, "Saving match should succeed")

			// Test GetMatches - this should NOT fail with SQL scanning errors
			allMatches, err := matchRepo.GetMatches()
			require.NoError(t, err, "GetMatches should not have SQL scanning issues")
			require.GreaterOrEqual(t, len(allMatches), 1, "Should return at least 1 match")

			// Find our specific match and verify odds
			found := false
			for _, retrievedMatch := range allMatches {
				if retrievedMatch.GetHomeTeam() == "Team M" && retrievedMatch.GetAwayTeam() == "Team N" {
					require.Equal(t, 2.0, retrievedMatch.GetHomeTeamOdds(), "Home odds should match")
					require.Equal(t, 3.5, retrievedMatch.GetAwayTeamOdds(), "Away odds should match")
					require.Equal(t, 3.2, retrievedMatch.GetDrawOdds(), "Draw odds should match")
					found = true
					break
				}
			}
			require.True(t, found, "Should find the match we just created")
		})
	}, 10*time.Second)
}
