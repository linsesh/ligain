package postgres

import (
	"database/sql"
	"testing"
	"time"

	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/rules"

	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

func TestGameRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		gameRepo, err := NewPostgresGameRepository(testDB.db)
		require.NoError(t, err)

		t.Run("Create and Get Game", func(t *testing.T) {
			// Create test data using raw SQL with proper UUID and schema columns
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status)
				VALUES ('123e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started');
			`)
			require.NoError(t, err)

			// Don't create any matches or bets - this should result in a "not started" game

			// Get game
			game, err := gameRepo.GetGame("123e4567-e89b-12d3-a456-426614174000")
			require.NoError(t, err)
			require.Equal(t, "2024", game.GetSeasonYear())
			require.Equal(t, "Premier League", game.GetCompetitionName())
			require.Equal(t, models.GameStatusNotStarted, game.GetGameStatus())
		})

		t.Run("Get Non-existent Game", func(t *testing.T) {
			// Try to get non-existent game
			_, err := gameRepo.GetGame("123e4567-e89b-12d3-a456-426614174999")
			require.Error(t, err)
		})

		t.Run("Save Game with ID", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				game := rules.NewFreshGame("2024", "Premier League", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
				gameID := "223e4567-e89b-12d3-a456-426614174000"
				err := gameRepo.SaveWithId(gameID, game)
				require.NoError(t, err)
				retrieved, err := gameRepo.GetGame(gameID)
				require.NoError(t, err)
				require.Equal(t, game.GetSeasonYear(), retrieved.GetSeasonYear())
				require.Equal(t, game.GetCompetitionName(), retrieved.GetCompetitionName())
				require.Equal(t, game.GetGameStatus(), retrieved.GetGameStatus())
			})
		})

		t.Run("Save Game with Invalid ID", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				game := rules.NewFreshGame("2024", "Premier League", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
				err := gameRepo.SaveWithId("invalid-uuid-format", game)
				require.Error(t, err)
			})
		})
	}, 10*time.Second)
}

func TestGameRepository_RestoreGameState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		// Create repositories
		gameRepo, err := NewPostgresGameRepository(testDB.db)
		require.NoError(t, err)

		// Create test data with proper UUID
		gameId := "323e4567-e89b-12d3-a456-426614174000"
		game := rules.NewFreshGame("2024", "Premier League", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
		err = gameRepo.SaveWithId(gameId, game)
		require.NoError(t, err)

		// Create players
		player1 := models.Player{Name: "Player1"}
		player2 := models.Player{Name: "Player2"}
		playerRepo := NewPostgresPlayerRepository(testDB.db)
		_, err = playerRepo.SavePlayer(player1)
		require.NoError(t, err)
		_, err = playerRepo.SavePlayer(player2)
		require.NoError(t, err)

		// Create matches
		matchRepo := NewPostgresMatchRepository(testDB.db)

		// Past match (finished)
		pastMatch := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", testTime.Add(-24*time.Hour), 1)
		pastMatch.Finish(2, 1)
		err = matchRepo.SaveMatch(pastMatch)
		require.NoError(t, err)

		// Current match (in progress)
		currentMatch := models.NewSeasonMatch("Liverpool", "Man City", "2024", "Premier League", testTime, 2)
		err = matchRepo.SaveMatch(currentMatch)
		require.NoError(t, err)

		// Future match (not started)
		futureMatch := models.NewSeasonMatch("Man United", "Tottenham", "2024", "Premier League", testTime.Add(24*time.Hour), 3)
		err = matchRepo.SaveMatch(futureMatch)
		require.NoError(t, err)

		// Create bets
		betCache := repositories.NewInMemoryBetRepository()
		betRepo := NewPostgresBetRepository(testDB.db, betCache)
		postgresBetRepo := betRepo.(*PostgresBetRepository)

		// Player 1 bets
		pastBet1 := models.NewBet(pastMatch, 2, 1) // Correct prediction
		_, _, err = postgresBetRepo.SaveBet(gameId, pastBet1, player1)
		require.NoError(t, err)

		currentBet1 := models.NewBet(currentMatch, 1, 1)
		_, _, err = postgresBetRepo.SaveBet(gameId, currentBet1, player1)
		require.NoError(t, err)

		futureBet1 := models.NewBet(futureMatch, 2, 0)
		_, _, err = postgresBetRepo.SaveBet(gameId, futureBet1, player1)
		require.NoError(t, err)

		// Player 2 bets
		pastBet2 := models.NewBet(pastMatch, 1, 2) // Wrong prediction
		_, _, err = postgresBetRepo.SaveBet(gameId, pastBet2, player2)
		require.NoError(t, err)

		currentBet2 := models.NewBet(currentMatch, 2, 1)
		_, _, err = postgresBetRepo.SaveBet(gameId, currentBet2, player2)
		require.NoError(t, err)

		futureBet2 := models.NewBet(futureMatch, 1, 1)
		_, _, err = postgresBetRepo.SaveBet(gameId, futureBet2, player2)
		require.NoError(t, err)

		// Save scores for past match
		_, _, err = postgresBetRepo.SaveBet(gameId, pastBet1, player1)
		require.NoError(t, err)
		_, _, err = postgresBetRepo.SaveBet(gameId, pastBet2, player2)
		require.NoError(t, err)
		err = postgresBetRepo.SaveScore(gameId, pastMatch, player1, 3) // Player 1 got 3 points for correct prediction
		require.NoError(t, err)
		err = postgresBetRepo.SaveScore(gameId, pastMatch, player2, 0) // Player 2 got 0 points for wrong prediction
		require.NoError(t, err)

		t.Run("Restore Game State", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Get the game
				restoredGame, err := gameRepo.GetGame(gameId)
				require.NoError(t, err)
				require.NotNil(t, restoredGame)

				// Verify basic game properties
				require.Equal(t, game.GetSeasonYear(), restoredGame.GetSeasonYear())
				require.Equal(t, game.GetCompetitionName(), restoredGame.GetCompetitionName())

				// Get past and incoming matches
				pastResults := restoredGame.GetPastResults()
				incomingMatches := restoredGame.GetIncomingMatches()

				// Verify past match
				pastMatchResult, exists := pastResults[pastMatch.Id()]
				require.True(t, exists, "Past match should exist in results")
				require.NotNil(t, pastMatchResult)
				require.Equal(t, pastMatch.Id(), pastMatchResult.Match.Id())
				require.True(t, pastMatchResult.Match.IsFinished())
				require.Equal(t, 2, pastMatchResult.Match.GetHomeGoals())
				require.Equal(t, 1, pastMatchResult.Match.GetAwayGoals())

				// Verify current match
				currentMatchResult, exists := incomingMatches[currentMatch.Id()]
				require.True(t, exists, "Current match should exist in incoming matches")
				require.NotNil(t, currentMatchResult)
				require.Equal(t, currentMatch.Id(), currentMatchResult.Match.Id())
				require.False(t, currentMatchResult.Match.IsFinished())

				// Verify future match
				futureMatchResult, exists := incomingMatches[futureMatch.Id()]
				require.True(t, exists, "Future match should exist in incoming matches")
				require.NotNil(t, futureMatchResult)
				require.Equal(t, futureMatch.Id(), futureMatchResult.Match.Id())
				require.False(t, futureMatchResult.Match.IsFinished())

				// Verify bets and scores for past match
				require.NotNil(t, pastMatchResult.Bets)
				require.NotNil(t, pastMatchResult.Scores)
				require.Equal(t, 2, len(pastMatchResult.Bets))   // Two players made bets
				require.Equal(t, 2, len(pastMatchResult.Scores)) // Two players have scores

				// Verify bets for current match
				require.NotNil(t, currentMatchResult.Bets)
				require.Equal(t, 2, len(currentMatchResult.Bets)) // Two players made bets

				// Verify bets for future match
				require.NotNil(t, futureMatchResult.Bets)
				require.Equal(t, 2, len(futureMatchResult.Bets)) // Two players made bets

				// Verify player points
				playerPoints := restoredGame.GetPlayersPoints()
				require.Equal(t, 2, len(playerPoints)) // Two players have points
				for player, points := range playerPoints {
					if player.Name == "Player1" {
						require.Equal(t, 3, points) // Player 1 got 3 points for correct prediction
					} else {
						require.Equal(t, 0, points) // Player 2 got 0 points for wrong prediction
					}
				}
			})
		})
	}, 10*time.Second)
}
