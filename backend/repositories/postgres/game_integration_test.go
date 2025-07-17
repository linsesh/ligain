package postgres

import (
	"context"
	"database/sql"
	"fmt"
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

		// Create players first
		player1 := &models.PlayerData{
			Name:       "Player1",
			Email:      stringPtr("player1@example.com"),
			Provider:   stringPtr("google"),
			ProviderID: stringPtr("google_user_1"),
		}
		player2 := &models.PlayerData{
			Name:       "Player2",
			Email:      stringPtr("player2@example.com"),
			Provider:   stringPtr("google"),
			ProviderID: stringPtr("google_user_2"),
		}
		playerRepo := NewPostgresPlayerRepository(testDB.db)
		err = playerRepo.CreatePlayer(context.Background(), player1)
		require.NoError(t, err)
		err = playerRepo.CreatePlayer(context.Background(), player2)
		require.NoError(t, err)

		// Now that player1 and player2 have IDs, create the game
		gameId := "323e4567-e89b-12d3-a456-426614174000"
		game := rules.NewFreshGame("2024", "Premier League", []models.Player{player1, player2}, []models.Match{}, &rules.ScorerOriginal{})
		err = gameRepo.SaveWithId(gameId, game)
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
		id1, bet1, err := postgresBetRepo.SaveBet(gameId, currentBet1, player1)
		fmt.Println(id1)
		fmt.Println(bet1)
		require.NoError(t, err)

		futureBet1 := models.NewBet(futureMatch, 2, 0)
		_, _, err = postgresBetRepo.SaveBet(gameId, futureBet1, player1)
		require.NoError(t, err)

		// Player 2 bets
		pastBet2 := models.NewBet(pastMatch, 1, 2) // Wrong prediction
		_, _, err = postgresBetRepo.SaveBet(gameId, pastBet2, player2)
		require.NoError(t, err)

		currentBet2 := models.NewBet(currentMatch, 2, 1)
		id, bet, err := postgresBetRepo.SaveBet(gameId, currentBet2, player2)
		fmt.Println(id)
		fmt.Println(bet)
		require.NoError(t, err)

		futureBet2 := models.NewBet(futureMatch, 1, 1)
		_, _, err = postgresBetRepo.SaveBet(gameId, futureBet2, player2)
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
				// Use the actual player1 from restoredGame.GetPlayers()
				var restoredPlayer1 models.Player
				for _, p := range restoredGame.GetPlayers() {
					if p.GetName() == "Player1" {
						restoredPlayer1 = p
						break
					}
				}
				require.NotNil(t, restoredPlayer1, "restored player1 should not be nil")
				// Use the testing method to get all bets for verification
				incomingMatches := restoredGame.GetIncomingMatchesForTesting()

				// Debug: print bets and scores for past match
				fmt.Printf("Past match bets: %v\n", pastResults[pastMatch.Id()].Bets)
				fmt.Printf("Past match scores: %v\n", pastResults[pastMatch.Id()].Scores)

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
				fmt.Printf("Here are all the bets for the match %s: %v\n", currentMatchResult.Match.Id(), currentMatchResult.Bets)
				require.Equal(t, 2, len(currentMatchResult.Bets)) // Two players made bets

				// Verify bets for future match
				require.NotNil(t, futureMatchResult.Bets)
				require.Equal(t, 2, len(futureMatchResult.Bets)) // Two players made bets

				// Verify player points
				playerPoints := restoredGame.GetPlayersPoints()
				require.Equal(t, 2, len(playerPoints)) // Two players have points
				for playerID, points := range playerPoints {
					// Find the player by ID to check their name
					var playerName string
					for _, player := range restoredGame.GetPlayers() {
						if player.GetID() == playerID {
							playerName = player.GetName()
							break
						}
					}
					if playerName == "Player1" {
						require.Equal(t, 3, points) // Player 1 got 3 points for correct prediction
					} else {
						require.Equal(t, 0, points) // Player 2 got 0 points for wrong prediction
					}
				}
			})
		})
	}, 10*time.Second)
}

func TestGameRepository_LoadPlayersFromBothTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		// Create repositories
		gameRepo, err := NewPostgresGameRepository(testDB.db)
		require.NoError(t, err)

		// Create players first
		player1 := &models.PlayerData{
			Name:       "Player1",
			Email:      stringPtr("player1@example.com"),
			Provider:   stringPtr("google"),
			ProviderID: stringPtr("google_user_1"),
		}
		player2 := &models.PlayerData{
			Name:       "Player2",
			Email:      stringPtr("player2@example.com"),
			Provider:   stringPtr("google"),
			ProviderID: stringPtr("google_user_2"),
		}
		player3 := &models.PlayerData{
			Name:       "Player3",
			Email:      stringPtr("player3@example.com"),
			Provider:   stringPtr("google"),
			ProviderID: stringPtr("google_user_3"),
		}
		playerRepo := NewPostgresPlayerRepository(testDB.db)
		err = playerRepo.CreatePlayer(context.Background(), player1)
		require.NoError(t, err)
		err = playerRepo.CreatePlayer(context.Background(), player2)
		require.NoError(t, err)
		err = playerRepo.CreatePlayer(context.Background(), player3)
		require.NoError(t, err)

		// Create a game with all three players
		gameId := "423e4567-e89b-12d3-a456-426614174000"
		game := rules.NewFreshGame("2024", "Premier League", []models.Player{player1, player2, player3}, []models.Match{}, &rules.ScorerOriginal{})
		err = gameRepo.SaveWithId(gameId, game)
		require.NoError(t, err)

		// Add all players to the game_player table (this simulates the real scenario)
		gamePlayerRepo := NewPostgresGamePlayerRepository(testDB.db)
		err = gamePlayerRepo.AddPlayerToGame(context.Background(), gameId, player1.GetID())
		require.NoError(t, err)
		err = gamePlayerRepo.AddPlayerToGame(context.Background(), gameId, player2.GetID())
		require.NoError(t, err)
		err = gamePlayerRepo.AddPlayerToGame(context.Background(), gameId, player3.GetID())
		require.NoError(t, err)

		// Create a match
		matchRepo := NewPostgresMatchRepository(testDB.db)
		futureMatch := models.NewSeasonMatch("Arsenal", "Chelsea", "2024", "Premier League", testTime.Add(24*time.Hour), 1)
		err = matchRepo.SaveMatch(futureMatch)
		require.NoError(t, err)

		// Only Player1 and Player2 make bets, Player3 doesn't bet yet
		betCache := repositories.NewInMemoryBetRepository()
		betRepo := NewPostgresBetRepository(testDB.db, betCache)
		postgresBetRepo := betRepo.(*PostgresBetRepository)

		// Player 1 bets
		bet1 := models.NewBet(futureMatch, 2, 1)
		_, _, err = postgresBetRepo.SaveBet(gameId, bet1, player1)
		require.NoError(t, err)

		// Player 2 bets
		bet2 := models.NewBet(futureMatch, 1, 1)
		_, _, err = postgresBetRepo.SaveBet(gameId, bet2, player2)
		require.NoError(t, err)

		// Player 3 doesn't bet (this is the key test case)

		t.Run("Load All Players Including Non-Betting Players", func(t *testing.T) {
			testDB.withTransaction(t, func(tx *sql.Tx) {
				// Get the game
				restoredGame, err := gameRepo.GetGame(gameId)
				require.NoError(t, err)
				require.NotNil(t, restoredGame)

				// Verify that ALL three players are loaded, even though Player3 hasn't bet
				players := restoredGame.GetPlayers()
				require.Equal(t, 3, len(players), "All three players should be loaded")

				// Verify each player is present
				playerNames := make(map[string]bool)
				for _, player := range players {
					playerNames[player.GetName()] = true
				}
				require.True(t, playerNames["Player1"], "Player1 should be loaded")
				require.True(t, playerNames["Player2"], "Player2 should be loaded")
				require.True(t, playerNames["Player3"], "Player3 should be loaded even though they haven't bet")

				// Verify that Player3 can place a bet (this was the original bug)
				bet3 := models.NewBet(futureMatch, 0, 0)
				err = restoredGame.CheckPlayerBetValidity(players[2], bet3, testTime) // Player3
				require.NoError(t, err, "Player3 should be able to place a bet even though they haven't bet yet")

				// Verify incoming matches have bets from Player1 and Player2 only
				incomingMatches := restoredGame.GetIncomingMatchesForTesting()
				matchResult, exists := incomingMatches[futureMatch.Id()]
				require.True(t, exists, "Future match should exist in incoming matches")
				require.Equal(t, 2, len(matchResult.Bets), "Only Player1 and Player2 should have bets")
			})
		})
	}, 10*time.Second)
}
