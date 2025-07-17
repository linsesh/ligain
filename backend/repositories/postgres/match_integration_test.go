package postgres

import (
	"testing"
	"time"

	"liguain/backend/models"

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

		t.Run("Get All Matches", func(t *testing.T) {
			// Setup test data
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ('223e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')
			`)
			require.NoError(t, err)

			// Create multiple matches
			match1 := models.NewSeasonMatch("Team A", "Team B", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
			match2 := models.NewSeasonMatch("Team C", "Team D", "2024", "Premier League", testTime.Add(48*time.Hour), 2)

			err = matchRepo.SaveMatch(match1)
			require.NoError(t, err)
			err = matchRepo.SaveMatch(match2)
			require.NoError(t, err)

			// Get all matches
			matches, err := matchRepo.GetMatches()
			require.NoError(t, err)
			require.Equal(t, 2, len(matches))
		})

		t.Run("Get Match for Non-existent Game", func(t *testing.T) {
			// Try to get match for non-existent match with proper UUID format
			_, err := matchRepo.GetMatch("323e4567-e89b-12d3-a456-426614174999")
			require.Error(t, err)
		})
	}, 10*time.Second)
}
