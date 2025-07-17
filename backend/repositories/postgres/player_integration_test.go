package postgres

import (
	"context"
	"liguain/backend/models"
	"testing"
	"time"

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

		t.Run("Create and Get Player", func(t *testing.T) {
			// Create player via new authentication flow
			player := &models.PlayerData{
				Name:       "TestPlayer",
				Email:      stringPtr("test@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_user_123"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)
			require.NotEmpty(t, player.ID)

			// Test GetPlayerByID
			retrieved, err := playerRepo.GetPlayerByID(context.Background(), player.ID)
			require.NoError(t, err)
			require.Equal(t, player.Name, retrieved.Name)
			require.Equal(t, *player.Email, *retrieved.Email)
			require.Equal(t, *player.Provider, *retrieved.Provider)
			require.Equal(t, *player.ProviderID, *retrieved.ProviderID)
		})

		t.Run("Get Player by Email", func(t *testing.T) {
			player := &models.PlayerData{
				Name:       "EmailTestPlayer",
				Email:      stringPtr("emailtest@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_user_456"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			retrieved, err := playerRepo.GetPlayerByEmail(context.Background(), "emailtest@example.com")
			require.NoError(t, err)
			require.Equal(t, player.Name, retrieved.Name)
			require.Equal(t, *player.Email, *retrieved.Email)
		})

		t.Run("Get Player by Provider", func(t *testing.T) {
			player := &models.PlayerData{
				Name:       "ProviderTestPlayer",
				Email:      stringPtr("providertest@example.com"),
				Provider:   stringPtr("apple"),
				ProviderID: stringPtr("apple_user_789"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			retrieved, err := playerRepo.GetPlayerByProvider(context.Background(), "apple", "apple_user_789")
			require.NoError(t, err)
			require.Equal(t, player.Name, retrieved.Name)
			require.Equal(t, *player.Provider, *retrieved.Provider)
			require.Equal(t, *player.ProviderID, *retrieved.ProviderID)
		})

		t.Run("Get Player by Name", func(t *testing.T) {
			player := &models.PlayerData{
				Name:       "NameTestPlayer",
				Email:      stringPtr("nametest@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_user_999"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			retrieved, err := playerRepo.GetPlayerByName(context.Background(), "NameTestPlayer")
			require.NoError(t, err)
			require.Equal(t, player.Name, retrieved.Name)
		})

		t.Run("Update Player", func(t *testing.T) {
			player := &models.PlayerData{
				Name:       "UpdateTestPlayer",
				Email:      stringPtr("updatetest@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_user_111"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			// Update the player
			player.Name = "UpdatedPlayerName"
			err = playerRepo.UpdatePlayer(context.Background(), player)
			require.NoError(t, err)

			// Verify the update
			retrieved, err := playerRepo.GetPlayerByID(context.Background(), player.ID)
			require.NoError(t, err)
			require.Equal(t, "UpdatedPlayerName", retrieved.Name)
		})

		t.Run("Create Auth Token", func(t *testing.T) {
			player := &models.PlayerData{
				Name:       "TokenTestPlayer",
				Email:      stringPtr("tokentest@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_user_222"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			authToken := &models.AuthToken{
				PlayerID:  player.ID,
				Token:     "test_token_123",
				ExpiresAt: time.Now().Add(24 * time.Hour),
			}

			err = playerRepo.CreateAuthToken(context.Background(), authToken)
			require.NoError(t, err)

			// Verify token was created
			retrieved, err := playerRepo.GetAuthToken(context.Background(), "test_token_123")
			require.NoError(t, err)
			require.Equal(t, player.ID, retrieved.PlayerID)
			require.Equal(t, "test_token_123", retrieved.Token)
		})

		t.Run("Get Players by Game", func(t *testing.T) {
			// Setup test data
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ('123e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')
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
			require.Contains(t, []string{players[0].GetName(), players[1].GetName()}, "Player1")
			require.Contains(t, []string{players[0].GetName(), players[1].GetName()}, "Player2")
		})
	}, 10*time.Second)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
