package postgres

import (
	"testing"
	"time"

	"liguain/backend/models"
	"liguain/backend/repositories"

	"github.com/stretchr/testify/require"
)

func TestGameCodeRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		gameCodeRepo := NewPostgresGameCodeRepository(testDB.db)

		t.Run("Create and Get Game Code", func(t *testing.T) {
			gameID := "123e4567-e89b-12d3-a456-426614174001"
			_, err := testDB.db.Exec(`INSERT INTO game (id, season_year, competition_name, status) VALUES ($1, $2, $3, $4)`, gameID, "2024", "Test League", "not started")
			require.NoError(t, err)

			expiresAt := time.Now().Add(24 * time.Hour)
			gameCode := models.NewGameCode(gameID, "TST1", expiresAt)
			err = gameCodeRepo.CreateGameCode(gameCode)
			require.NoError(t, err)
			require.NotEmpty(t, gameCode.ID)
			require.NotZero(t, gameCode.CreatedAt)

			retrievedCode, err := gameCodeRepo.GetGameCodeByCode("TST1")
			require.NoError(t, err)
			require.Equal(t, gameID, retrievedCode.GameID)
			require.Equal(t, "TST1", retrievedCode.Code)
			require.Equal(t, gameCode.ID, retrievedCode.ID)
		})

		t.Run("Get Game Code by Game ID", func(t *testing.T) {
			gameID := "123e4567-e89b-12d3-a456-426614174002"
			_, err := testDB.db.Exec(`INSERT INTO game (id, season_year, competition_name, status) VALUES ($1, $2, $3, $4)`, gameID, "2024", "Test League 2", "not started")
			require.NoError(t, err)

			expiresAt := time.Now().Add(24 * time.Hour)
			gameCode := models.NewGameCode(gameID, "CD02", expiresAt)
			err = gameCodeRepo.CreateGameCode(gameCode)
			require.NoError(t, err)

			retrievedCode, err := gameCodeRepo.GetGameCodeByGameID(gameID)
			require.NoError(t, err)
			require.Equal(t, gameID, retrievedCode.GameID)
			require.Equal(t, "CD02", retrievedCode.Code)
		})

		t.Run("Code Exists Check", func(t *testing.T) {
			gameID := "123e4567-e89b-12d3-a456-426614174003"
			_, err := testDB.db.Exec(`INSERT INTO game (id, season_year, competition_name, status) VALUES ($1, $2, $3, $4)`, gameID, "2024", "Test League 3", "not started")
			require.NoError(t, err)

			expiresAt := time.Now().Add(24 * time.Hour)
			gameCode := models.NewGameCode(gameID, "EXST", expiresAt)
			err = gameCodeRepo.CreateGameCode(gameCode)
			require.NoError(t, err)

			exists, err := gameCodeRepo.CodeExists("EXST")
			require.NoError(t, err)
			require.True(t, exists)

			exists, err = gameCodeRepo.CodeExists("NOPE")
			require.NoError(t, err)
			require.False(t, exists)
		})

		t.Run("Delete Game Code", func(t *testing.T) {
			gameID := "123e4567-e89b-12d3-a456-426614174004"
			_, err := testDB.db.Exec(`INSERT INTO game (id, season_year, competition_name, status) VALUES ($1, $2, $3, $4)`, gameID, "2024", "Test League 4", "not started")
			require.NoError(t, err)

			expiresAt := time.Now().Add(24 * time.Hour)
			gameCode := models.NewGameCode(gameID, "DELT", expiresAt)
			err = gameCodeRepo.CreateGameCode(gameCode)
			require.NoError(t, err)

			err = gameCodeRepo.DeleteGameCode("DELT")
			require.NoError(t, err)

			_, err = gameCodeRepo.GetGameCodeByCode("DELT")
			require.Error(t, err)
			require.Equal(t, repositories.ErrGameCodeNotFound, err)
		})

		t.Run("Expired Code Handling", func(t *testing.T) {
			gameID := "123e4567-e89b-12d3-a456-426614174005"
			_, err := testDB.db.Exec(`INSERT INTO game (id, season_year, competition_name, status) VALUES ($1, $2, $3, $4)`, gameID, "2024", "Test League 5", "not started")
			require.NoError(t, err)

			expiresAt := time.Now().Add(-24 * time.Hour)
			gameCode := models.NewGameCode(gameID, "EXPD", expiresAt)
			err = gameCodeRepo.CreateGameCode(gameCode)
			require.NoError(t, err)

			_, err = gameCodeRepo.GetGameCodeByCode("EXPD")
			require.Error(t, err)
			require.Equal(t, repositories.ErrGameCodeNotFound, err)

			exists, err := gameCodeRepo.CodeExists("EXPD")
			require.NoError(t, err)
			require.False(t, exists)
		})

		t.Run("Delete Expired Codes", func(t *testing.T) {
			gameID1 := "123e4567-e89b-12d3-a456-426614174006"
			gameID2 := "123e4567-e89b-12d3-a456-426614174007"
			_, err := testDB.db.Exec(`INSERT INTO game (id, season_year, competition_name, status) VALUES ($1, $2, $3, $4)`, gameID1, "2024", "Test League 6", "not started")
			require.NoError(t, err)
			_, err = testDB.db.Exec(`INSERT INTO game (id, season_year, competition_name, status) VALUES ($1, $2, $3, $4)`, gameID2, "2024", "Test League 6", "not started")
			require.NoError(t, err)

			expiresAt1 := time.Now().Add(-24 * time.Hour)
			gameCode1 := models.NewGameCode(gameID1, "EX01", expiresAt1)
			err = gameCodeRepo.CreateGameCode(gameCode1)
			require.NoError(t, err)

			expiresAt2 := time.Now().Add(24 * time.Hour)
			gameCode2 := models.NewGameCode(gameID2, "VL01", expiresAt2)
			err = gameCodeRepo.CreateGameCode(gameCode2)
			require.NoError(t, err)

			err = gameCodeRepo.DeleteExpiredCodes()
			require.NoError(t, err)

			_, err = gameCodeRepo.GetGameCodeByCode("EX01")
			require.Error(t, err)
			require.Equal(t, repositories.ErrGameCodeNotFound, err)

			_, err = gameCodeRepo.GetGameCodeByCode("VL01")
			require.NoError(t, err)
		})

		t.Run("Get Non-existent Game Code", func(t *testing.T) {
			_, err := gameCodeRepo.GetGameCodeByCode("NONE")
			require.Error(t, err)
			require.Equal(t, repositories.ErrGameCodeNotFound, err)
		})

		t.Run("Get Non-existent Game Code by Game ID", func(t *testing.T) {
			_, err := gameCodeRepo.GetGameCodeByGameID("123e4567-e89b-12d3-a456-426614174999")
			require.Error(t, err)
			require.Equal(t, repositories.ErrGameCodeNotFound, err)
		})

		t.Run("Delete Non-existent Game Code", func(t *testing.T) {
			err := gameCodeRepo.DeleteGameCode("NONE")
			require.Error(t, err)
			require.Equal(t, repositories.ErrGameCodeNotFound, err)
		})
	}, 30*time.Second)
}
