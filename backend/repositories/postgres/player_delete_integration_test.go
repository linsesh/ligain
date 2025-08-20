package postgres

import (
	"context"
	"database/sql"
	"ligain/backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresPlayerRepository_DeletePlayer_Integration(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewPostgresPlayerRepository(testDB.db)
	ctx := context.Background()

	t.Run("Successfully delete player and cascade related data", func(t *testing.T) {
		// Create test player
		player := &models.PlayerData{
			Name:       "Test Player for Delete",
			Email:      stringPtr("delete-test@example.com"),
			Provider:   stringPtr("google"),
			ProviderID: stringPtr("google-delete-123"),
			CreatedAt:  &time.Time{},
			UpdatedAt:  &time.Time{},
		}
		*player.CreatedAt = time.Now()
		*player.UpdatedAt = time.Now()

		err := repo.CreatePlayer(ctx, player)
		require.NoError(t, err)
		require.NotEmpty(t, player.ID)

		// Create auth token for the player
		authToken := &models.AuthToken{
			PlayerID:  player.ID,
			Token:     "test-token-for-delete",
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		err = repo.CreateAuthToken(ctx, authToken)
		require.NoError(t, err)

		// Create test game to associate with player
		gameID := createTestGame(t, testDB.db)

		// Create game_player association
		_, err = testDB.db.ExecContext(ctx, `
			INSERT INTO game_player (game_id, player_id)
			VALUES ($1, $2)
		`, gameID, player.ID)
		require.NoError(t, err)

		// Create test match
		matchID := createTestMatch(t, testDB.db)

		// Create bet for the player
		_, err = testDB.db.ExecContext(ctx, `
			INSERT INTO bet (game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
			VALUES ($1, $2, $3, $4, $5)
		`, gameID, matchID, player.ID, 2, 1)
		require.NoError(t, err)

		// Create score for the player
		_, err = testDB.db.ExecContext(ctx, `
			INSERT INTO score (game_id, match_id, player_id, points)
			VALUES ($1, $2, $3, $4)
		`, gameID, matchID, player.ID, 500)
		require.NoError(t, err)

		// Verify data exists before deletion
		verifyPlayerDataExists(t, testDB.db, player.ID, gameID, matchID)

		// Execute delete
		err = repo.DeletePlayer(ctx, player.ID)
		require.NoError(t, err)

		// Verify all related data is deleted due to CASCADE
		verifyPlayerDataDeleted(t, testDB.db, player.ID, gameID, matchID)
	})

	t.Run("Delete non-existent player returns ErrNoRows", func(t *testing.T) {
		// Use a valid UUID format but non-existent ID
		nonExistentID := "00000000-0000-0000-0000-000000000000"
		err := repo.DeletePlayer(ctx, nonExistentID)
		assert.Equal(t, sql.ErrNoRows, err)
	})
}

func verifyPlayerDataExists(t *testing.T, db *sql.DB, playerID, gameID, matchID string) {
	ctx := context.Background()

	// Check player exists
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM player WHERE id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Player should exist")

	// Check auth_tokens exist
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth_tokens WHERE player_id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Auth token should exist")

	// Check game_player exists
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM game_player WHERE player_id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Game player association should exist")

	// Check bet exists
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM bet WHERE player_id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Bet should exist")

	// Check score exists
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM score WHERE player_id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Score should exist")
}

func verifyPlayerDataDeleted(t *testing.T, db *sql.DB, playerID, gameID, matchID string) {
	ctx := context.Background()

	// Check player is deleted
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM player WHERE id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Player should be deleted")

	// Check auth_tokens are deleted (CASCADE)
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth_tokens WHERE player_id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Auth tokens should be deleted")

	// Check game_player is deleted (CASCADE)
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM game_player WHERE player_id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Game player association should be deleted")

	// Check bet is deleted (CASCADE)
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM bet WHERE player_id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Bet should be deleted")

	// Check score is deleted (CASCADE)
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM score WHERE player_id = $1", playerID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Score should be deleted")

	// Verify game and match still exist (should not be affected)
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM game WHERE id = $1", gameID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Game should still exist")

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM match WHERE id = $1", matchID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Match should still exist")
}

func createTestGame(t *testing.T, db *sql.DB) string {
	var gameID string
	err := db.QueryRow(`
		INSERT INTO game (season_year, competition_name, status, game_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "2023", "Test Competition", "ongoing", "Test Game for Delete").Scan(&gameID)
	require.NoError(t, err)
	return gameID
}

func createTestMatch(t *testing.T, db *sql.DB) string {
	var matchID string
	err := db.QueryRow(`
		INSERT INTO match (local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, "test-match-delete", "home-team", "away-team", time.Now(), "scheduled", "2023", "TEST", 1).Scan(&matchID)
	require.NoError(t, err)
	return matchID
}
