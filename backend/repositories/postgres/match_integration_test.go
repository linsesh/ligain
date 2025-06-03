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

	testDB := setupTestDB(t)
	defer testDB.Close()

	matchRepo := NewPostgresMatchRepository(testDB.db)

	testDB.setupTestFixtures(t)

	t.Run("Save and Get Match", func(t *testing.T) {
		// Create test data using raw SQL
		_, err := testDB.db.Exec(`
			INSERT INTO game (id, name, description, created_at, updated_at)
			VALUES ('test-game', 'Test Game', 'Test Description', $1, $1);
		`, testTime)
		require.NoError(t, err)

		// Create match
		match := models.NewSeasonMatch("Team A", "Team B", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
		matchId, err := matchRepo.SaveMatch(match)
		require.NoError(t, err)
		require.NotEmpty(t, matchId)

		// Get match
		retrievedMatch, err := matchRepo.GetMatch(matchId)
		require.NoError(t, err)
		require.Equal(t, match.GetHomeTeam(), retrievedMatch.GetHomeTeam())
		require.Equal(t, match.GetAwayTeam(), retrievedMatch.GetAwayTeam())
		require.Equal(t, match.GetDate().Unix(), retrievedMatch.GetDate().Unix())
	})

	t.Run("Get All Matches", func(t *testing.T) {
		// Create test data using raw SQL
		_, err := testDB.db.Exec(`
			INSERT INTO game (id, name, description, created_at, updated_at)
			VALUES ('test-game-2', 'Test Game 2', 'Test Description', $1, $1);
		`, testTime)
		require.NoError(t, err)

		// Create multiple matches
		match1 := models.NewSeasonMatch("Team A", "Team B", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
		match2 := models.NewSeasonMatch("Team C", "Team D", "2024", "Premier League", testTime.Add(48*time.Hour), 2)

		_, err = matchRepo.SaveMatch(match1)
		require.NoError(t, err)
		_, err = matchRepo.SaveMatch(match2)
		require.NoError(t, err)

		// Get all matches
		matches, err := matchRepo.GetMatchesByGame("test-game-2")
		require.NoError(t, err)
		require.Equal(t, 2, len(matches))
	})

	t.Run("Get Match for Non-existent Game", func(t *testing.T) {
		// Try to get match for non-existent game
		_, err := matchRepo.GetMatch("non-existent-match")
		require.Error(t, err)
	})
}
