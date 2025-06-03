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

	testDB := setupTestDB(t)
	defer testDB.Close()

	scoreRepo := NewPostgresScoreRepository(testDB.db)

	testDB.setupTestFixtures(t)

	t.Run("Save and Get Score", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Create test data using raw SQL
			_, err := tx.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('test-game', '2024', 'Premier League', 'started')
			`)
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO player (id, name)
				VALUES ('player1', 'TestPlayer')
			`)
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO match (id, game_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
				VALUES ('match1', 'test-game', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)
			`, testTime)
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO bet (id, game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
				VALUES ('bet1', 'test-game', 'match1', 'player1', 2, 1)
			`)
			require.NoError(t, err)

			// Save score
			err = scoreRepo.SaveScore("test-game", "bet1", 3)
			require.NoError(t, err)

			// Get score
			points, err := scoreRepo.GetScore("test-game", "bet1")
			require.NoError(t, err)
			require.Equal(t, 3, points)
		})
	})

	t.Run("Get All Scores for Game", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Create test data using raw SQL
			_, err := tx.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('test-game', '2024', 'Premier League', 'started')
			`)
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO player (id, name)
				VALUES 
					('player1', 'Player1'),
					('player2', 'Player2')
			`)
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO match (id, game_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
				VALUES ('match1', 'test-game', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1)
			`, testTime)
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO bet (id, game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
				VALUES 
					('bet1', 'test-game', 'match1', 'player1', 2, 1),
					('bet2', 'test-game', 'match1', 'player2', 1, 1)
			`)
			require.NoError(t, err)

			// Save scores
			err = scoreRepo.SaveScore("test-game", "bet1", 3)
			require.NoError(t, err)
			err = scoreRepo.SaveScore("test-game", "bet2", 1)
			require.NoError(t, err)

			// Get all scores
			scores, err := scoreRepo.GetScores("test-game")
			require.NoError(t, err)
			require.Equal(t, 2, len(scores))
			require.Equal(t, 3, scores["bet1"])
			require.Equal(t, 1, scores["bet2"])
		})
	})

	t.Run("Get Scores by Match and Player", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Create test data using raw SQL
			_, err := tx.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('test-game', '2024', 'Premier League', 'started')
			`)
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO player (id, name)
				VALUES 
					('player1', 'Player1'),
					('player2', 'Player2')
			`)
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO match (id, game_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
				VALUES 
					('match1', 'test-game', 'Arsenal', 'Chelsea', $1, 'finished', '2024', 'Premier League', 1),
					('match2', 'test-game', 'Liverpool', 'Man City', $2, 'finished', '2024', 'Premier League', 2)
			`, testTime, testTime.Add(24*time.Hour))
			require.NoError(t, err)

			_, err = tx.Exec(`
				INSERT INTO bet (id, game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
				VALUES 
					('bet1', 'test-game', 'match1', 'player1', 2, 1),
					('bet2', 'test-game', 'match1', 'player2', 1, 1),
					('bet3', 'test-game', 'match2', 'player1', 2, 0),
					('bet4', 'test-game', 'match2', 'player2', 1, 1)
			`)
			require.NoError(t, err)

			// Save scores
			err = scoreRepo.SaveScore("test-game", "bet1", 3)
			require.NoError(t, err)
			err = scoreRepo.SaveScore("test-game", "bet2", 1)
			require.NoError(t, err)
			err = scoreRepo.SaveScore("test-game", "bet3", 3)
			require.NoError(t, err)
			err = scoreRepo.SaveScore("test-game", "bet4", 1)
			require.NoError(t, err)

			// Get scores by match and player
			scores, err := scoreRepo.GetScoresByMatchAndPlayer("test-game")
			require.NoError(t, err)
			require.Equal(t, 2, len(scores))                                      // Two matches
			require.Equal(t, 2, len(scores["match1"]))                            // Two players for match1
			require.Equal(t, 2, len(scores["match2"]))                            // Two players for match2
			require.Equal(t, 3, scores["match1"][models.Player{Name: "Player1"}]) // Player1 got 3 points for match1
			require.Equal(t, 1, scores["match1"][models.Player{Name: "Player2"}]) // Player2 got 1 point for match1
			require.Equal(t, 3, scores["match2"][models.Player{Name: "Player1"}]) // Player1 got 3 points for match2
			require.Equal(t, 1, scores["match2"][models.Player{Name: "Player2"}]) // Player2 got 1 point for match2
		})
	})

	t.Run("Get Score for Non-existent Bet", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Setup test data using raw SQL
			_, err := tx.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('test-game', '2024', 'Premier League', 'started');
			`)
			require.NoError(t, err)

			// Try to get score for non-existent bet
			score, err := scoreRepo.GetScore("test-game", "non-existent-bet")
			require.NoError(t, err)
			require.Equal(t, 0, score) // Should return 0 for non-existent bet
		})
	})
}
