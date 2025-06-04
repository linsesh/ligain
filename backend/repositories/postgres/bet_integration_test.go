package postgres

import (
	"database/sql"
	"testing"
	"time"

	"liguain/backend/models"

	"github.com/stretchr/testify/require"
)

var betTestTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

// matchWithUUID wraps a match to provide a custom UUID
type matchWithUUID struct {
	*models.SeasonMatch
	id string
}

func (m *matchWithUUID) Id() string {
	return m.id
}

func TestBetRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		t.Run("Save and Get Bet", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction dependency injection
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status)
					VALUES ('123e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started');
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES ('123e4567-e89b-12d3-a456-426614174001', 'TestPlayer');
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('123e4567-e89b-12d3-a456-426614174002', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1);
				`, betTestTime)
				require.NoError(t, err)

				// Create bet with actual match UUID
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet := models.NewBet(match, 2, 1)
				player := models.Player{Name: "TestPlayer"}

				// Override the match ID to use the UUID from database
				bet.Match = &matchWithUUID{
					SeasonMatch: match,
					id:          "123e4567-e89b-12d3-a456-426614174002",
				}

				// Use repository SaveBet method instead of raw SQL
				_, err = betRepo.SaveBet("123e4567-e89b-12d3-a456-426614174000", bet, player)
				require.NoError(t, err)

				// Get bet using repository
				bets, err := betRepo.GetBets("123e4567-e89b-12d3-a456-426614174000", player)
				require.NoError(t, err)
				require.Equal(t, 1, len(bets))
				require.Equal(t, bet.PredictedHomeGoals, bets[0].PredictedHomeGoals)
				require.Equal(t, bet.PredictedAwayGoals, bets[0].PredictedAwayGoals)
			})
		})

		t.Run("Get Bets for Match", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction dependency injection
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status)
					VALUES ('223e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started');
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES 
						('223e4567-e89b-12d3-a456-426614174001', 'Player1'),
						('223e4567-e89b-12d3-a456-426614174002', 'Player2');
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('223e4567-e89b-12d3-a456-426614174003', 'Liverpool', 'Man City', $1, 'finished', '2024', 'Premier League', 1);
				`, betTestTime)
				require.NoError(t, err)

				// Create bets with actual match UUIDs
				match := models.NewSeasonMatch("Liverpool", "Man City", "2024", "Premier League", betTestTime, 1)
				matchWithID := &matchWithUUID{
					SeasonMatch: match,
					id:          "223e4567-e89b-12d3-a456-426614174003",
				}
				bet1 := models.NewBet(matchWithID, 2, 1)
				bet2 := models.NewBet(matchWithID, 1, 1)
				player1 := models.Player{Name: "Player1"}
				player2 := models.Player{Name: "Player2"}
				_, err = betRepo.SaveBet("223e4567-e89b-12d3-a456-426614174000", bet1, player1)
				require.NoError(t, err)
				_, err = betRepo.SaveBet("223e4567-e89b-12d3-a456-426614174000", bet2, player2)
				require.NoError(t, err)

				// Get bets for match
				bets, players, err := betRepo.GetBetsForMatch(matchWithID, "223e4567-e89b-12d3-a456-426614174000")
				require.NoError(t, err)
				require.Equal(t, 2, len(bets))
				require.Equal(t, 2, len(players))
			})
		})

		t.Run("Save Bet with Invalid Match", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction dependency injection
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status)
					VALUES ('323e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started');

					INSERT INTO player (id, name)
					VALUES ('323e4567-e89b-12d3-a456-426614174001', 'TestPlayer');
				`)
				require.NoError(t, err)

				// Create invalid match (no ID)
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet := models.NewBet(match, 2, 1)
				player := models.Player{Name: "TestPlayer"}

				// Try to save bet
				_, err = betRepo.SaveBet("323e4567-e89b-12d3-a456-426614174000", bet, player)
				require.Error(t, err)
			})
		})

		t.Run("Get Bets for Non-existent Player", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction dependency injection
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status)
					VALUES ('423e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started');
				`)
				require.NoError(t, err)

				// Try to get bets for non-existent player
				nonExistentPlayer := models.Player{Name: "NonExistentPlayer"}
				bets, err := betRepo.GetBets("423e4567-e89b-12d3-a456-426614174000", nonExistentPlayer)
				require.NoError(t, err)
				require.Empty(t, bets)
			})
		})
	}, 10*time.Second)
}
