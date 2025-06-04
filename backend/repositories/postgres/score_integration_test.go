package postgres

import (
	"database/sql"
	"testing"
	"time"

	"liguain/backend/models"

	"github.com/stretchr/testify/require"
)

func TestScoreRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		t.Run("Save and Get Score", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction
				scoreRepo := NewPostgresScoreRepository(tx)

				// Create test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status)
					VALUES ('123e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started')
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES ('123e4567-e89b-12d3-a456-426614174001', 'TestPlayer')
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('123e4567-e89b-12d3-a456-426614174002', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)
				`, testTime)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO bet (id, game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
					VALUES ('123e4567-e89b-12d3-a456-426614174003', '123e4567-e89b-12d3-a456-426614174000', '123e4567-e89b-12d3-a456-426614174002', '123e4567-e89b-12d3-a456-426614174001', 2, 1)
				`)
				require.NoError(t, err)

				// Save score
				err = scoreRepo.SaveScore("123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174003", 3)
				require.NoError(t, err)

				// Get score
				points, err := scoreRepo.GetScore("123e4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174003")
				require.NoError(t, err)
				require.Equal(t, 3, points)
			})
		})

		t.Run("Get All Scores for Game", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction
				scoreRepo := NewPostgresScoreRepository(tx)

				// Create test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status)
					VALUES ('223e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started')
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES 
						('223e4567-e89b-12d3-a456-426614174001', 'Player1'),
						('223e4567-e89b-12d3-a456-426614174002', 'Player2')
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES ('223e4567-e89b-12d3-a456-426614174003', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)
				`, testTime)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO bet (id, game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
					VALUES 
						('223e4567-e89b-12d3-a456-426614174004', '223e4567-e89b-12d3-a456-426614174000', '223e4567-e89b-12d3-a456-426614174003', '223e4567-e89b-12d3-a456-426614174001', 2, 1),
						('223e4567-e89b-12d3-a456-426614174005', '223e4567-e89b-12d3-a456-426614174000', '223e4567-e89b-12d3-a456-426614174003', '223e4567-e89b-12d3-a456-426614174002', 1, 1)
				`)
				require.NoError(t, err)

				// Save scores
				err = scoreRepo.SaveScore("223e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174004", 3)
				require.NoError(t, err)
				err = scoreRepo.SaveScore("223e4567-e89b-12d3-a456-426614174000", "223e4567-e89b-12d3-a456-426614174005", 1)
				require.NoError(t, err)

				// Get all scores
				scores, err := scoreRepo.GetScores("223e4567-e89b-12d3-a456-426614174000")
				require.NoError(t, err)
				require.Equal(t, 2, len(scores))
				require.Equal(t, 3, scores["223e4567-e89b-12d3-a456-426614174004"])
				require.Equal(t, 1, scores["223e4567-e89b-12d3-a456-426614174005"])
			})
		})

		t.Run("Get Scores by Match and Player", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction
				scoreRepo := NewPostgresScoreRepository(tx)

				// Create test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status)
					VALUES ('323e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started')
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO player (id, name)
					VALUES 
						('323e4567-e89b-12d3-a456-426614174001', 'Player1'),
						('323e4567-e89b-12d3-a456-426614174002', 'Player2')
				`)
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO match (id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
					VALUES 
						('323e4567-e89b-12d3-a456-426614174003', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1),
						('323e4567-e89b-12d3-a456-426614174004', 'Liverpool', 'Man City', $2, 'finished', '2024', 'Premier League', 2)
				`, testTime, testTime.Add(24*time.Hour))
				require.NoError(t, err)

				_, err = tx.Exec(`
					INSERT INTO bet (id, game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
					VALUES 
						('323e4567-e89b-12d3-a456-426614174005', '323e4567-e89b-12d3-a456-426614174000', '323e4567-e89b-12d3-a456-426614174003', '323e4567-e89b-12d3-a456-426614174001', 2, 1),
						('323e4567-e89b-12d3-a456-426614174006', '323e4567-e89b-12d3-a456-426614174000', '323e4567-e89b-12d3-a456-426614174003', '323e4567-e89b-12d3-a456-426614174002', 1, 1),
						('323e4567-e89b-12d3-a456-426614174007', '323e4567-e89b-12d3-a456-426614174000', '323e4567-e89b-12d3-a456-426614174004', '323e4567-e89b-12d3-a456-426614174001', 2, 0),
						('323e4567-e89b-12d3-a456-426614174008', '323e4567-e89b-12d3-a456-426614174000', '323e4567-e89b-12d3-a456-426614174004', '323e4567-e89b-12d3-a456-426614174002', 1, 1)
				`)
				require.NoError(t, err)

				// Save scores
				err = scoreRepo.SaveScore("323e4567-e89b-12d3-a456-426614174000", "323e4567-e89b-12d3-a456-426614174005", 3)
				require.NoError(t, err)
				err = scoreRepo.SaveScore("323e4567-e89b-12d3-a456-426614174000", "323e4567-e89b-12d3-a456-426614174006", 1)
				require.NoError(t, err)
				err = scoreRepo.SaveScore("323e4567-e89b-12d3-a456-426614174000", "323e4567-e89b-12d3-a456-426614174007", 3)
				require.NoError(t, err)
				err = scoreRepo.SaveScore("323e4567-e89b-12d3-a456-426614174000", "323e4567-e89b-12d3-a456-426614174008", 1)
				require.NoError(t, err)

				// Get scores by match and player
				scores, err := scoreRepo.GetScoresByMatchAndPlayer("323e4567-e89b-12d3-a456-426614174000")
				require.NoError(t, err)
				require.Equal(t, 2, len(scores))                                                                    // Two matches
				require.Equal(t, 2, len(scores["323e4567-e89b-12d3-a456-426614174003"]))                            // Two players for match1
				require.Equal(t, 2, len(scores["323e4567-e89b-12d3-a456-426614174004"]))                            // Two players for match2
				require.Equal(t, 3, scores["323e4567-e89b-12d3-a456-426614174003"][models.Player{Name: "Player1"}]) // Player1 got 3 points for match1
				require.Equal(t, 1, scores["323e4567-e89b-12d3-a456-426614174003"][models.Player{Name: "Player2"}]) // Player2 got 1 point for match1
				require.Equal(t, 3, scores["323e4567-e89b-12d3-a456-426614174004"][models.Player{Name: "Player1"}]) // Player1 got 3 points for match2
				require.Equal(t, 1, scores["323e4567-e89b-12d3-a456-426614174004"][models.Player{Name: "Player2"}]) // Player2 got 1 point for match2
			})
		})

		t.Run("Get Score for Non-existent Bet", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Create repository with transaction
				scoreRepo := NewPostgresScoreRepository(tx)

				// Setup test data using raw SQL with proper UUIDs
				_, err := tx.Exec(`
					INSERT INTO game (id, season_year, competition_name, status)
					VALUES ('423e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started');
				`)
				require.NoError(t, err)

				// Try to get score for non-existent bet
				score, err := scoreRepo.GetScore("423e4567-e89b-12d3-a456-426614174000", "non-existent-bet")
				require.NoError(t, err)
				require.Equal(t, 0, score) // Should return 0 for non-existent bet
			})
		})
	}, 10*time.Second)
}
