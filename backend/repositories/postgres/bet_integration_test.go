package postgres

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"ligain/backend/models"
	"ligain/backend/repositories"

	"github.com/stretchr/testify/require"
)

var betTestTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

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
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('123e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game');
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES ('123e4567-e89b-12d3-a456-426614174001', 'TestPlayer');
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('123e4567-e89b-12d3-a456-426614174002', 'Premier League-2024-Arsenal-Chelsea-1', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1);
				`, betTestTime)
				require.NoError(t, err)

				// Create bet with actual match UUID
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet := models.NewBet(match, 2, 1)
				player := newTestPlayer("TestPlayer")

				// Use repository SaveBet method instead of raw SQL
				_, _, err = betRepo.SaveBet("123e4567-e89b-12d3-a456-426614174000", bet, player)
				if err != nil {
					t.Errorf("Failed to save bet: %v", err)
				}

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
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('223e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game');
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
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('223e4567-e89b-12d3-a456-426614174003', 'Premier League-2024-Liverpool-Man City-1', 'Liverpool', 'Man City', $1, 'finished', '2024', 'Premier League', 1);
				`, betTestTime)
				require.NoError(t, err)

				// Create bets with actual match UUIDs
				match := models.NewSeasonMatch("Liverpool", "Man City", "2024", "Premier League", betTestTime, 1)
				bet1 := models.NewBet(match, 2, 1)
				bet2 := models.NewBet(match, 1, 1)
				player1 := newTestPlayer("Player1")
				player2 := newTestPlayer("Player2")
				_, _, err = betRepo.SaveBet("223e4567-e89b-12d3-a456-426614174000", bet1, player1)
				if err != nil {
					t.Errorf("Failed to save bet1: %v", err)
				}
				_, _, err = betRepo.SaveBet("223e4567-e89b-12d3-a456-426614174000", bet2, player2)
				if err != nil {
					t.Errorf("Failed to save bet2: %v", err)
				}

				// Get bets for match
				bets, players, err := betRepo.GetBetsForMatch(match, "223e4567-e89b-12d3-a456-426614174000")
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
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('323e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES ('323e4567-e89b-12d3-a456-426614174001', 'TestPlayer')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('323e4567-e89b-12d3-a456-426614174002', 'Premier League-2024-Arsenal-Chelsea-1', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)`,
					betTestTime)
				require.NoError(t, err)

				// Create match with proper UUID
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet := models.NewBet(match, 2, 1)
				player := newTestPlayer("TestPlayer")

				// Try to save bet
				_, _, err = betRepo.SaveBet("323e4567-e89b-12d3-a456-426614174000", bet, player)
				require.NoError(t, err)
			})
		})

		t.Run("Get Bets for Non-existent Player", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction dependency injection
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('423e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game');
				`)
				require.NoError(t, err)

				// Try to get bets for non-existent player
				nonExistentPlayer := newTestPlayer("NonExistentPlayer")
				bets, err := betRepo.GetBets("423e4567-e89b-12d3-a456-426614174000", nonExistentPlayer)
				require.NoError(t, err)
				require.Empty(t, bets)
			})
		})

		t.Run("Save and Get Score", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction dependency injection
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('423e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES ('423e4567-e89b-12d3-a456-426614174001', 'TestPlayer')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('423e4567-e89b-12d3-a456-426614174002', 'Premier League-2024-Arsenal-Chelsea-1', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)`,
					betTestTime)
				require.NoError(t, err)

				// Create match and bet
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet := models.NewBet(match, 2, 1)
				player := models.NewSimplePlayer("423e4567-e89b-12d3-a456-426614174001", "TestPlayer")

				// Save bet
				betId, _, err := betRepo.SaveBet("423e4567-e89b-12d3-a456-426614174000", bet, player)
				if err != nil {
					t.Errorf("Failed to save bet: %v", err)
				}

				// Save score
				err = betRepo.SaveScore("423e4567-e89b-12d3-a456-426614174000", match, player, 3)
				require.NoError(t, err)

				// Get score
				score, err := betRepo.GetScore("423e4567-e89b-12d3-a456-426614174000", betId)
				require.NoError(t, err)
				require.Equal(t, 3, score)
			})
		})

		t.Run("Get Score Not Found", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('523e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES ('523e4567-e89b-12d3-a456-426614174001', 'TestPlayer')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('523e4567-e89b-12d3-a456-426614174002', 'Premier League-2024-Arsenal-Chelsea-1', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)`,
					betTestTime)
				require.NoError(t, err)

				// Create match and bet without score
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet := models.NewBet(match, 2, 1)
				player := models.NewSimplePlayer("523e4567-e89b-12d3-a456-426614174001", "TestPlayer")

				// Save bet
				betId, _, err := betRepo.SaveBet("523e4567-e89b-12d3-a456-426614174000", bet, player)
				if err != nil {
					t.Errorf("Failed to save bet: %v", err)
				}

				// Try to get non-existent score
				_, err = betRepo.GetScore("523e4567-e89b-12d3-a456-426614174000", betId)
				require.Equal(t, repositories.ErrScoreNotFound, err)
			})
		})

		t.Run("Get All Scores", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				gameId := "623e4567-e89b-12d3-a456-426614174000"
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('623e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES 
						('623e4567-e89b-12d3-a456-426614174001', 'Player1'),
						('623e4567-e89b-12d3-a456-426614174002', 'Player2')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('623e4567-e89b-12d3-a456-426614174003', 'Premier League-2024-Arsenal-Chelsea-1', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)`,
					betTestTime)
				require.NoError(t, err)

				// Create match and bets
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet1 := models.NewBet(match, 2, 1)
				bet2 := models.NewBet(match, 1, 1)
				player1 := models.NewSimplePlayer("623e4567-e89b-12d3-a456-426614174001", "Player1")
				player2 := models.NewSimplePlayer("623e4567-e89b-12d3-a456-426614174002", "Player2")

				// Save bets and scores
				_, _, err = betRepo.SaveBet(gameId, bet1, player1)
				if err != nil {
					t.Errorf("Failed to save bet1: %v", err)
				}
				_, _, err = betRepo.SaveBet(gameId, bet2, player2)
				if err != nil {
					t.Errorf("Failed to save bet2: %v", err)
				}

				err = betRepo.SaveScore(gameId, match, player1, 3)
				require.NoError(t, err)
				err = betRepo.SaveScore(gameId, match, player2, 1)
				require.NoError(t, err)

				// Get all scores
				scores, err := betRepo.GetScores(gameId)
				fmt.Printf("DEBUG: GetScores returned: %+v\n", scores)
				require.NoError(t, err)
				require.Equal(t, 1, len(scores))                    // 1 match
				matchUUID := "623e4567-e89b-12d3-a456-426614174003" // The UUID used in the SQL insert
				matchScores := scores[matchUUID]
				require.Equal(t, 2, len(matchScores)) // 2 players
				require.Equal(t, 3, matchScores[player1.GetID()])
				require.Equal(t, 1, matchScores[player2.GetID()])
			})
		})

		t.Run("Get Scores By Match And Player", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('723e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES 
						('723e4567-e89b-12d3-a456-426614174001', 'Player1'),
						('723e4567-e89b-12d3-a456-426614174002', 'Player2')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('723e4567-e89b-12d3-a456-426614174003', 'Premier League-2024-Arsenal-Chelsea-1', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)`,
					betTestTime)
				require.NoError(t, err)

				// Create match and bets
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet1 := models.NewBet(match, 2, 1)
				bet2 := models.NewBet(match, 1, 1)
				player1 := models.NewSimplePlayer("723e4567-e89b-12d3-a456-426614174001", "Player1")
				player2 := models.NewSimplePlayer("723e4567-e89b-12d3-a456-426614174002", "Player2")

				// Save bets and scores
				_, _, err = betRepo.SaveBet("723e4567-e89b-12d3-a456-426614174000", bet1, player1)
				if err != nil {
					t.Errorf("Failed to save bet1: %v", err)
				}
				_, _, err = betRepo.SaveBet("723e4567-e89b-12d3-a456-426614174000", bet2, player2)
				if err != nil {
					t.Errorf("Failed to save bet2: %v", err)
				}

				// Save scores using repository
				err = betRepo.SaveScore("723e4567-e89b-12d3-a456-426614174000", match, player1, 3)
				require.NoError(t, err)
				err = betRepo.SaveScore("723e4567-e89b-12d3-a456-426614174000", match, player2, 1)
				require.NoError(t, err)

				// Use the local_id instead of the database UUID
				matchId := match.Id()

				// Get scores by match and player
				scores, err := betRepo.GetScoresByMatchAndPlayer("723e4567-e89b-12d3-a456-426614174000")
				require.NoError(t, err)
				require.Equal(t, 1, len(scores))
				matchScores := scores[matchId]
				require.NotNil(t, matchScores)
				require.Equal(t, 2, len(matchScores))

				// The method now returns player IDs as keys
				require.Equal(t, 3, matchScores["723e4567-e89b-12d3-a456-426614174001"])
				require.Equal(t, 1, matchScores["723e4567-e89b-12d3-a456-426614174002"])
			})
		})

		t.Run("Save With ID", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				betRepo := NewPostgresBetRepository(tx, nil)

				// Setup test data
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('823e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES ('823e4567-e89b-12d3-a456-426614174001', 'TestPlayer')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('823e4567-e89b-12d3-a456-426614174002', 'Premier League-2024-Arsenal-Chelsea-1', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)`,
					betTestTime)
				require.NoError(t, err)

				// Create match with proper UUID
				match := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", betTestTime, 1)
				bet := models.NewBet(match, 2, 1)
				player := models.NewSimplePlayer("823e4567-e89b-12d3-a456-426614174001", "TestPlayer")

				// Save bet with specific ID
				err = betRepo.SaveWithId("823e4567-e89b-12d3-a456-426614174000", "823e4567-e89b-12d3-a456-426614174003", bet, player)
				require.NoError(t, err)

				// Verify bet was saved with the custom ID
				_, err = betRepo.GetScore("823e4567-e89b-12d3-a456-426614174000", "823e4567-e89b-12d3-a456-426614174003")
				require.Equal(t, repositories.ErrScoreNotFound, err)

				// Save a score and verify it's associated with the custom ID
				err = betRepo.SaveScore("823e4567-e89b-12d3-a456-426614174000", match, player, 5)
				require.NoError(t, err)

				score, err := betRepo.GetScore("823e4567-e89b-12d3-a456-426614174000", "823e4567-e89b-12d3-a456-426614174003")
				require.NoError(t, err)
				require.Equal(t, 5, score)
			})
		})

		t.Run("Save and Get Score for Forgotten Bet", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Setup test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status, game_name)
					VALUES ('999e4567-e89b-12d3-a456-426614174000', '2024', 'Test League', 'started', 'Test Game')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES ('999e4567-e89b-12d3-a456-426614174001', 'ForgotPlayer')`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, local_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('999e4567-e89b-12d3-a456-426614174002', 'Test League-2024-Forgotten-FC-0', 'Forgotten', 'FC', $1, 'finished', '2024', 'Test League', 0)`, betTestTime)
				require.NoError(t, err)

				// Simulate saving a score for a player who did not bet (no bet row)
				// Use repository to save a score for a player who did not bet (no bet row)
				betRepo := NewPostgresBetRepository(tx, nil)
				match := models.NewSeasonMatch("Forgotten", "FC", "2024", "Test League", betTestTime, 0)
				player := models.NewSimplePlayer("999e4567-e89b-12d3-a456-426614174001", "ForgotPlayer")
				err = betRepo.SaveScore("999e4567-e89b-12d3-a456-426614174000", match, player, -100)
				require.NoError(t, err)

				// Query the score table directly to verify
				var points int
				row := tx.QueryRow(`SELECT points FROM score WHERE match_id = '999e4567-e89b-12d3-a456-426614174002' AND bet_id IS NULL`)
				err = row.Scan(&points)
				require.NoError(t, err)
				require.Equal(t, -100, points)
			})
		})
	}, 10*time.Second)
}
