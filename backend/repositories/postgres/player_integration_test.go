package postgres

import (
	"testing"
	"time"

	"liguain/backend/models"

	"github.com/stretchr/testify/require"
)

func TestPlayerRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		playerRepo := NewPostgresPlayerRepository(testDB.db)

		t.Run("Save and Get Player", func(t *testing.T) {
			// Create test data using raw SQL
			_, err := testDB.db.Exec(`
				INSERT INTO player (name)
				VALUES ('TestPlayer');
			`)
			require.NoError(t, err)

			// Get player
			player := models.Player{Name: "TestPlayer"}
			id, err := playerRepo.SavePlayer(player)
			require.NoError(t, err)
			require.NotEmpty(t, id)

			retrieved, err := playerRepo.GetPlayer(id)
			require.NoError(t, err)
			require.Equal(t, player.Name, retrieved.Name)
		})

		t.Run("Save Duplicate Player", func(t *testing.T) {
			// Create test data using raw SQL
			_, err := testDB.db.Exec(`
				INSERT INTO player (name)
				VALUES ('DuplicatePlayer');
			`)
			require.NoError(t, err)

			player := models.Player{Name: "DuplicatePlayer"}
			id1, err := playerRepo.SavePlayer(player)
			require.NoError(t, err)
			require.NotEmpty(t, id1)

			id2, err := playerRepo.SavePlayer(player)
			require.NoError(t, err)
			require.Equal(t, id1, id2) // Should return the same ID
		})

		t.Run("Get Non-existent Player", func(t *testing.T) {
			// Try to get non-existent player
			_, err := playerRepo.GetPlayer("non-existent-id")
			require.Error(t, err)
		})

		t.Run("Get Players by Game", func(t *testing.T) {
			// Create test data using raw SQL with proper UUID and schema columns
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('123e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started');
			`)
			require.NoError(t, err)

			// Create players with proper UUIDs
			_, err = testDB.db.Exec(`
				INSERT INTO player (id, name)
				VALUES 
					('123e4567-e89b-12d3-a456-426614174001', 'Player1'), 
					('123e4567-e89b-12d3-a456-426614174002', 'Player2');
			`)
			require.NoError(t, err)

			// Create match with proper UUID
			_, err = testDB.db.Exec(`
				INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
				VALUES ('123e4567-e89b-12d3-a456-426614174003', 'Premier League-2024-Team A-Team B-1', 'Team A', 'Team B', $1, 'scheduled', '2024', 'Premier League', 1);
			`, testTime.Add(24*time.Hour))
			require.NoError(t, err)

			// Create bets with proper UUIDs
			_, err = testDB.db.Exec(`
				INSERT INTO bet (id, game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
				VALUES 
					('123e4567-e89b-12d3-a456-426614174004', '123e4567-e89b-12d3-a456-426614174000', '123e4567-e89b-12d3-a456-426614174003', '123e4567-e89b-12d3-a456-426614174001', 2, 1),
					('123e4567-e89b-12d3-a456-426614174005', '123e4567-e89b-12d3-a456-426614174000', '123e4567-e89b-12d3-a456-426614174003', '123e4567-e89b-12d3-a456-426614174002', 1, 1)
			`)
			require.NoError(t, err)

			// Get players for game
			players, err := playerRepo.GetPlayers("123e4567-e89b-12d3-a456-426614174000")
			require.NoError(t, err)
			require.Equal(t, 2, len(players))
			require.Contains(t, []string{players[0].Name, players[1].Name}, "Player1")
			require.Contains(t, []string{players[0].Name, players[1].Name}, "Player2")
		})
	}, 10*time.Second)
}
