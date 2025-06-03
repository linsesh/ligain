package postgres

import (
	"database/sql"
	"testing"

	"liguain/backend/models"

	"github.com/stretchr/testify/require"
)

func TestBetRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testDB := setupTestDB(t)
	defer testDB.Close()

	betRepo := NewPostgresBetRepository(testDB.db, nil)

	testDB.setupTestFixtures(t)

	t.Run("Save and Get Bet", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Setup test data using raw SQL
			_, err := tx.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('test-game', '2024', 'Premier League', 'started');

				INSERT INTO player (id, name)
				VALUES ('player1', 'TestPlayer');

				INSERT INTO match (id, game_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
				VALUES ('match1', 'test-game', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1);
			`, testTime)
			require.NoError(t, err)

			// Create bet
			match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", testTime, 1)
			bet := models.NewBet(match, 2, 1)
			player := models.Player{Name: "TestPlayer"}
			betId, err := betRepo.SaveBet("test-game", bet, player)
			require.NoError(t, err)
			require.NotEmpty(t, betId)

			// Get bet
			bets, err := betRepo.GetBets("test-game", player)
			require.NoError(t, err)
			require.Equal(t, 1, len(bets))
			require.Equal(t, bet.PredictedHomeGoals, bets[0].PredictedHomeGoals)
			require.Equal(t, bet.PredictedAwayGoals, bets[0].PredictedAwayGoals)
		})
	})

	t.Run("Get Bets for Match", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Setup test data using raw SQL
			_, err := tx.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('test-game', '2024', 'Premier League', 'started');

				INSERT INTO player (id, name)
				VALUES 
					('player1', 'Player1'),
					('player2', 'Player2');

				INSERT INTO match (id, game_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
				VALUES ('match1', 'test-game', 'Liverpool', 'Man City', $1, 'finished', '2024', 'Premier League', 1);
			`, testTime)
			require.NoError(t, err)

			// Create bets
			match := models.NewSeasonMatch("Liverpool", "Man City", "2024", "Premier League", testTime, 1)
			bet1 := models.NewBet(match, 2, 1)
			bet2 := models.NewBet(match, 1, 1)
			player1 := models.Player{Name: "Player1"}
			player2 := models.Player{Name: "Player2"}
			_, err = betRepo.SaveBet("test-game", bet1, player1)
			require.NoError(t, err)
			_, err = betRepo.SaveBet("test-game", bet2, player2)
			require.NoError(t, err)

			// Get bets for match
			bets, players, err := betRepo.GetBetsForMatch(match, "test-game")
			require.NoError(t, err)
			require.Equal(t, 2, len(bets))
			require.Equal(t, 2, len(players))
		})
	})

	t.Run("Save Bet with Invalid Match", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Setup test data using raw SQL
			_, err := tx.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('test-game', '2024', 'Premier League', 'started');

				INSERT INTO player (id, name)
				VALUES ('player1', 'TestPlayer');
			`)
			require.NoError(t, err)

			// Create invalid match (no ID)
			match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", testTime, 1)
			bet := models.NewBet(match, 2, 1)
			player := models.Player{Name: "TestPlayer"}

			// Try to save bet
			_, err = betRepo.SaveBet("test-game", bet, player)
			require.Error(t, err)
		})
	})

	t.Run("Get Bets for Non-existent Player", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Setup test data using raw SQL
			_, err := tx.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('test-game', '2024', 'Premier League', 'started');
			`)
			require.NoError(t, err)

			// Try to get bets for non-existent player
			nonExistentPlayer := models.Player{Name: "NonExistentPlayer"}
			bets, err := betRepo.GetBets("test-game", nonExistentPlayer)
			require.NoError(t, err)
			require.Empty(t, bets)
		})
	})
}
