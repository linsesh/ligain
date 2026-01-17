package postgres

import (
	"context"
	"ligain/backend/models"
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

		t.Run("UpdateAvatar sets all avatar fields", func(t *testing.T) {
			// Create player
			player := &models.PlayerData{
				Name:       "AvatarTestPlayer",
				Email:      stringPtr("avatartest@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_avatar_123"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			// Call UpdateAvatar with object key, signed URL, and expiration
			objectKey := "avatars/player123/avatar.jpg"
			signedURL := "https://storage.googleapis.com/bucket/avatars/player123/avatar.jpg?signature=abc"
			expiresAt := time.Now().Add(1 * time.Hour).Truncate(time.Microsecond)

			err = playerRepo.UpdateAvatar(context.Background(), player.ID, objectKey, signedURL, expiresAt)
			require.NoError(t, err)

			// Retrieve player and verify all 3 avatar fields are set correctly
			retrieved, err := playerRepo.GetPlayerByID(context.Background(), player.ID)
			require.NoError(t, err)
			require.NotNil(t, retrieved.AvatarObjectKey)
			require.NotNil(t, retrieved.AvatarSignedURL)
			require.NotNil(t, retrieved.AvatarSignedURLExpiresAt)
			require.Equal(t, objectKey, *retrieved.AvatarObjectKey)
			require.Equal(t, signedURL, *retrieved.AvatarSignedURL)
			require.WithinDuration(t, expiresAt, *retrieved.AvatarSignedURLExpiresAt, time.Second)
		})

		t.Run("UpdateAvatarSignedURL refreshes signed URL only", func(t *testing.T) {
			// Create player with avatar via UpdateAvatar
			player := &models.PlayerData{
				Name:       "AvatarRefreshTestPlayer",
				Email:      stringPtr("avatarrefresh@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_avatar_refresh_123"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			objectKey := "avatars/refresh/avatar.jpg"
			originalSignedURL := "https://storage.googleapis.com/bucket/avatars/refresh/avatar.jpg?signature=original"
			originalExpiresAt := time.Now().Add(1 * time.Hour).Truncate(time.Microsecond)

			err = playerRepo.UpdateAvatar(context.Background(), player.ID, objectKey, originalSignedURL, originalExpiresAt)
			require.NoError(t, err)

			// Call UpdateAvatarSignedURL with new URL and expiration
			newSignedURL := "https://storage.googleapis.com/bucket/avatars/refresh/avatar.jpg?signature=new"
			newExpiresAt := time.Now().Add(2 * time.Hour).Truncate(time.Microsecond)

			err = playerRepo.UpdateAvatarSignedURL(context.Background(), player.ID, newSignedURL, newExpiresAt)
			require.NoError(t, err)

			// Retrieve player and verify object_key unchanged, signed URL updated
			retrieved, err := playerRepo.GetPlayerByID(context.Background(), player.ID)
			require.NoError(t, err)
			require.NotNil(t, retrieved.AvatarObjectKey)
			require.NotNil(t, retrieved.AvatarSignedURL)
			require.NotNil(t, retrieved.AvatarSignedURLExpiresAt)
			require.Equal(t, objectKey, *retrieved.AvatarObjectKey) // Object key unchanged
			require.Equal(t, newSignedURL, *retrieved.AvatarSignedURL)
			require.WithinDuration(t, newExpiresAt, *retrieved.AvatarSignedURLExpiresAt, time.Second)
		})

		t.Run("ClearAvatar removes all avatar fields", func(t *testing.T) {
			// Create player with all avatar fields set via UpdateAvatar
			player := &models.PlayerData{
				Name:       "AvatarClearTestPlayer",
				Email:      stringPtr("avatarclear@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_avatar_clear_123"),
			}

			err := playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			objectKey := "avatars/clear/avatar.jpg"
			signedURL := "https://storage.googleapis.com/bucket/avatars/clear/avatar.jpg?signature=abc"
			expiresAt := time.Now().Add(1 * time.Hour).Truncate(time.Microsecond)

			err = playerRepo.UpdateAvatar(context.Background(), player.ID, objectKey, signedURL, expiresAt)
			require.NoError(t, err)

			// Call ClearAvatar
			err = playerRepo.ClearAvatar(context.Background(), player.ID)
			require.NoError(t, err)

			// Retrieve player and verify all avatar fields are NULL
			retrieved, err := playerRepo.GetPlayerByID(context.Background(), player.ID)
			require.NoError(t, err)
			require.Nil(t, retrieved.AvatarObjectKey)
			require.Nil(t, retrieved.AvatarSignedURL)
			require.Nil(t, retrieved.AvatarSignedURLExpiresAt)
		})

		t.Run("UpdateAvatar for non-existent player returns error", func(t *testing.T) {
			// Call UpdateAvatar with non-existent player ID
			nonExistentID := "00000000-0000-0000-0000-000000000000"
			objectKey := "avatars/nonexistent/avatar.jpg"
			signedURL := "https://storage.googleapis.com/bucket/avatars/nonexistent/avatar.jpg?signature=abc"
			expiresAt := time.Now().Add(1 * time.Hour)

			err := playerRepo.UpdateAvatar(context.Background(), nonExistentID, objectKey, signedURL, expiresAt)
			require.Error(t, err)
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
