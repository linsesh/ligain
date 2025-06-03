package postgres

import (
	"database/sql"
	"testing"
	"time"

	"liguain/backend/models"
	"liguain/backend/rules"

	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

func TestGameRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testDB := setupTestDB(t)
	defer testDB.Close()

	gameRepo, err := NewPostgresGameRepository(testDB.db)
	require.NoError(t, err)

	testDB.setupTestFixtures(t)

	t.Run("Create and Get Game", func(t *testing.T) {
		// Create test data using raw SQL
		_, err := testDB.db.Exec(`
			INSERT INTO game (id, name, description, created_at, updated_at)
			VALUES ('test-game', 'Test Game', 'Test Description', NOW(), NOW());
		`)
		require.NoError(t, err)

		// Create players
		_, err = testDB.db.Exec(`
			INSERT INTO player (name)
			VALUES ('Player1'), ('Player2');
		`)
		require.NoError(t, err)

		// Create matches
		_, err = testDB.db.Exec(`
			INSERT INTO match (id, game_id, home_team_id, away_team_id, match_date, match_status, season_code, competition_code, matchday)
			VALUES 
				('match1', 'test-game', 'Team A', 'Team B', NOW() + INTERVAL '1 day', 'scheduled', '2024', 'Premier League', 1),
				('match2', 'test-game', 'Team C', 'Team D', NOW() + INTERVAL '2 days', 'scheduled', '2024', 'Premier League', 2);
		`)
		require.NoError(t, err)

		// Get game
		game, err := gameRepo.GetGame("test-game")
		require.NoError(t, err)
		require.Equal(t, "2024", game.GetSeasonYear())
		require.Equal(t, "Premier League", game.GetCompetitionName())
		require.Equal(t, models.GameStatusNotStarted, game.GetGameStatus())
	})

	t.Run("Get Non-existent Game", func(t *testing.T) {
		// Try to get non-existent game
		_, err := gameRepo.GetGame("non-existent-game")
		require.Error(t, err)
	})

	t.Run("Save Game with ID", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			game := rules.NewFreshGame("2024", "Premier League", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
			gameID := "test-game-1"
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
			err := gameRepo.SaveWithId("non-existent-id", game)
			require.Error(t, err)
		})
	})
}

func TestGameRepository_RestoreGameState(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testDB := setupTestDB(t)
	defer testDB.cleanup(t)

	// Create repositories
	gameRepo, err := NewPostgresGameRepository(testDB.db)
	require.NoError(t, err)

	// Create test data
	gameId := "test-game-2024"
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
	postgresMatchRepo := matchRepo.(*PostgresMatchRepository)

	// Past match (finished)
	pastMatch := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", testTime.Add(-24*time.Hour), 1)
	pastMatch.Finish(2, 1)
	pastMatchId, err := postgresMatchRepo.SaveMatch(pastMatch)
	require.NoError(t, err)

	// Current match (in progress)
	currentMatch := models.NewSeasonMatch("Liverpool", "Man City", "2024", "Premier League", testTime, 2)
	currentMatchId, err := postgresMatchRepo.SaveMatch(currentMatch)
	require.NoError(t, err)

	// Future match (not started)
	futureMatch := models.NewSeasonMatch("Man United", "Tottenham", "2024", "Premier League", testTime.Add(24*time.Hour), 3)
	futureMatchId, err := postgresMatchRepo.SaveMatch(futureMatch)
	require.NoError(t, err)

	// Create bets
	betRepo := NewPostgresBetRepository(testDB.db, nil)
	postgresBetRepo := betRepo.(*PostgresBetRepository)

	// Player 1 bets
	pastBet1 := models.NewBet(pastMatch, 2, 1) // Correct prediction
	_, err = postgresBetRepo.SaveBet(gameId, pastBet1, player1)
	require.NoError(t, err)

	currentBet1 := models.NewBet(currentMatch, 1, 1)
	_, err = postgresBetRepo.SaveBet(gameId, currentBet1, player1)
	require.NoError(t, err)

	futureBet1 := models.NewBet(futureMatch, 2, 0)
	_, err = postgresBetRepo.SaveBet(gameId, futureBet1, player1)
	require.NoError(t, err)

	// Player 2 bets
	pastBet2 := models.NewBet(pastMatch, 1, 2) // Wrong prediction
	_, err = postgresBetRepo.SaveBet(gameId, pastBet2, player2)
	require.NoError(t, err)

	currentBet2 := models.NewBet(currentMatch, 2, 1)
	_, err = postgresBetRepo.SaveBet(gameId, currentBet2, player2)
	require.NoError(t, err)

	futureBet2 := models.NewBet(futureMatch, 1, 1)
	_, err = postgresBetRepo.SaveBet(gameId, futureBet2, player2)
	require.NoError(t, err)

	// Save scores for past match
	scoreRepo := NewPostgresScoreRepository(testDB.db)
	pastBet1Id := postgresBetRepo.GetBetId(gameId, player1, pastMatchId)
	pastBet2Id := postgresBetRepo.GetBetId(gameId, player2, pastMatchId)
	err = scoreRepo.SaveScore(gameId, pastBet1Id, 3) // Player 1 got 3 points for correct prediction
	require.NoError(t, err)
	err = scoreRepo.SaveScore(gameId, pastBet2Id, 0) // Player 2 got 0 points for wrong prediction
	require.NoError(t, err)

	// Now test restoring the game state
	t.Run("Restore Game State", func(t *testing.T) {
		testDB.withTransaction(t, func(tx *sql.Tx) {
			// Get the game
			restoredGame, err := gameRepo.GetGame(gameId)
			require.NoError(t, err)
			require.NotNil(t, restoredGame)
			require.Equal(t, game.GetSeasonYear(), restoredGame.GetSeasonYear())
			require.Equal(t, game.GetCompetitionName(), restoredGame.GetCompetitionName())

			// Get all matches
			matches, err := postgresMatchRepo.GetMatchesByGame(gameId)
			require.NoError(t, err)
			require.Equal(t, 3, len(matches))

			// Verify past match
			pastMatch := matches[0]
			require.Equal(t, pastMatchId, pastMatch.Id())
			require.True(t, pastMatch.IsFinished())
			require.Equal(t, 2, pastMatch.GetHomeGoals())
			require.Equal(t, 1, pastMatch.GetAwayGoals())

			// Verify current match
			currentMatch := matches[1]
			require.Equal(t, currentMatchId, currentMatch.Id())
			require.False(t, currentMatch.IsFinished())

			// Verify future match
			futureMatch := matches[2]
			require.Equal(t, futureMatchId, futureMatch.Id())
			require.False(t, futureMatch.IsFinished())

			// Get bets for each player
			player1Bets, err := postgresBetRepo.GetBets(gameId, player1)
			require.NoError(t, err)
			require.Equal(t, 3, len(player1Bets))

			player2Bets, err := postgresBetRepo.GetBets(gameId, player2)
			require.NoError(t, err)
			require.Equal(t, 3, len(player2Bets))

			// Verify scores
			scores, err := scoreRepo.GetScores(gameId)
			require.NoError(t, err)
			require.Equal(t, 2, len(scores))        // Two scores for the past match
			require.Equal(t, 3, scores[pastBet1Id]) // Player 1 got 3 points
			require.Equal(t, 0, scores[pastBet2Id]) // Player 2 got 0 points

			// Verify bets for past match
			pastBets, players, err := postgresBetRepo.GetBetsForMatch(pastMatch, gameId)
			require.NoError(t, err)
			require.Equal(t, 2, len(pastBets))
			require.Equal(t, 2, len(players))

			// Verify bets for current match
			currentBets, players, err := postgresBetRepo.GetBetsForMatch(currentMatch, gameId)
			require.NoError(t, err)
			require.Equal(t, 2, len(currentBets))
			require.Equal(t, 2, len(players))

			// Verify bets for future match
			futureBets, players, err := postgresBetRepo.GetBetsForMatch(futureMatch, gameId)
			require.NoError(t, err)
			require.Equal(t, 2, len(futureBets))
			require.Equal(t, 2, len(players))
		})
	})
}
