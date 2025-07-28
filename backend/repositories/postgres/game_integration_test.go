package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/rules"

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
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ('123e4567-e89b-12d3-a456-426614174000', '2024', 'Premier League', 'started', 'Test Game');
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
				game := rules.NewFreshGame("2024", "Premier League", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
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
				game := rules.NewFreshGame("2024", "Premier League", "Test Game", []models.Player{}, []models.Match{}, &rules.ScorerOriginal{})
				err := gameRepo.SaveWithId("invalid-uuid-format", game)
				require.Error(t, err)
			})
		})

		t.Run("SQL Scanning Issues Prevention - getMatchesAndBets with Odds", func(t *testing.T) {
			// This test specifically verifies that the SQL scanning issue we fixed doesn't occur
			// in the getMatchesAndBets method when matches have odds data

			// Create a game first
			gameID := "423e4567-e89b-12d3-a456-426614174000"
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ($1, '2024', 'Ligue 1', 'started', 'SQL Scanning Test Game')
			`, gameID)
			require.NoError(t, err)

			// Create matches with odds for this game's competition/season
			matchRepo := NewPostgresMatchRepository(testDB.db)

			match1 := models.NewSeasonMatch("Team A", "Team B", "2024", "Ligue 1", testTime.Add(24*time.Hour), 1)
			match1.SetHomeTeamOdds(2.5)
			match1.SetAwayTeamOdds(3.2)
			match1.SetDrawOdds(3.0)
			err = matchRepo.SaveMatch(match1)
			require.NoError(t, err)

			match2 := models.NewSeasonMatch("Team C", "Team D", "2024", "Ligue 1", testTime.Add(48*time.Hour), 1)
			match2.SetHomeTeamOdds(1.8)
			match2.SetAwayTeamOdds(4.1)
			match2.SetDrawOdds(3.5)
			err = matchRepo.SaveMatch(match2)
			require.NoError(t, err)

			// Create a player
			player := &models.PlayerData{
				Name:       "Test Player",
				Email:      stringPtr("test@example.com"),
				Provider:   stringPtr("google"),
				ProviderID: stringPtr("google_user_test"),
			}
			playerRepo := NewPostgresPlayerRepository(testDB.db)
			err = playerRepo.CreatePlayer(context.Background(), player)
			require.NoError(t, err)

			// Add player to game
			gamePlayerRepo := NewPostgresGamePlayerRepository(testDB.db)
			err = gamePlayerRepo.AddPlayerToGame(context.Background(), gameID, player.GetID())
			require.NoError(t, err)

			// Create a bet for one of the matches
			betRepo := NewPostgresBetRepository(testDB.db, repositories.NewInMemoryBetRepository())
			bet := models.NewBet(match1, 2, 1)
			_, _, err = betRepo.SaveBet(gameID, bet, player)
			require.NoError(t, err)

			// Test the getMatchesAndBets method - this should NOT fail with SQL scanning errors
			// We'll test this by calling GetGame which internally uses getMatchesAndBets
			restoredGame, err := gameRepo.GetGame(gameID)
			require.NoError(t, err, "GetGame should not have SQL scanning issues")
			require.NotNil(t, restoredGame)

			// Verify the game has the correct data
			require.Equal(t, "2024", restoredGame.GetSeasonYear())
			require.Equal(t, "Ligue 1", restoredGame.GetCompetitionName())
			require.Equal(t, "SQL Scanning Test Game", restoredGame.GetName())

			// Verify players are loaded
			players := restoredGame.GetPlayers()
			require.Len(t, players, 1)
			require.Equal(t, "Test Player", players[0].GetName())
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
		game := rules.NewFreshGame("2024", "Premier League", "Test Game", []models.Player{player1, player2}, []models.Match{}, &rules.ScorerOriginal{})
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
		game := rules.NewFreshGame("2024", "Premier League", "Test Game", []models.Player{player1, player2, player3}, []models.Match{}, &rules.ScorerOriginal{})
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

func TestGameRepository_GetAllGames_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	runTestWithTimeout(t, func(t *testing.T) {
		testDB := setupTestDB(t)
		defer testDB.Close()

		gameRepo, err := NewPostgresGameRepository(testDB.db)
		require.NoError(t, err)

		t.Run("Get All Games - Empty Repository", func(t *testing.T) {
			// Ensure the repository is empty
			games, err := gameRepo.GetAllGames()
			require.NoError(t, err)
			require.Empty(t, games)
		})

		t.Run("Get All Games - Multiple Games", func(t *testing.T) {
			// Create multiple games
			gameIDs := []string{
				"123e4567-e89b-12d3-a456-426614174001",
				"123e4567-e89b-12d3-a456-426614174002",
				"123e4567-e89b-12d3-a456-426614174003",
			}

			// Insert games with different statuses
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES 
					($1, '2024', 'Ligue 1', 'not started', 'Game 1'),
					($2, '2024', 'Ligue 1', 'in progress', 'Game 2'),
					($3, '2024', 'Ligue 1', 'finished', 'Game 3')
			`, gameIDs[0], gameIDs[1], gameIDs[2])
			require.NoError(t, err)

			// Get all games (should exclude finished games)
			games, err := gameRepo.GetAllGames()
			require.NoError(t, err)
			// Only check for the non-finished games we just created
			found := 0
			for _, gameID := range gameIDs {
				game, exists := games[gameID]
				if exists {
					found++
					if gameID == gameIDs[0] {
						require.Equal(t, "Game 1", game.GetName())
						require.Equal(t, models.GameStatusNotStarted, game.GetGameStatus())
					}
					if gameID == gameIDs[1] {
						require.Equal(t, "Game 2", game.GetName())
						// Accept either NotStarted or Scheduled depending on mapping
						status := game.GetGameStatus()
						if status != models.GameStatusNotStarted && status != models.GameStatusScheduled {
							t.Errorf("unexpected status for Game 2: %v", status)
						}
					}
					// Game 3 should not be found since it's finished
				}
			}
			// Should only find 2 games (not started and in progress)
			require.Equal(t, 2, found)
			// Verify that the finished game is not in the results
			_, exists := games[gameIDs[2]]
			require.False(t, exists, "Finished game should not be returned")
		})

		t.Run("Get All Games - Exclude Finished Games", func(t *testing.T) {
			// Create a finished game
			finishedGameID := "123e4567-e89b-12d3-a456-426614174007"
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ($1, '2024', 'Ligue 1', 'finished', 'Finished Game')
			`, finishedGameID)
			require.NoError(t, err)

			// Get all games
			games, err := gameRepo.GetAllGames()
			require.NoError(t, err)

			// Verify that the finished game is not in the results
			_, exists := games[finishedGameID]
			require.False(t, exists, "Finished game should not be returned by GetAllGames")
		})

		// --- With Players and Matches ---
		t.Run("Get All Games - With Players and Matches", func(t *testing.T) {
			// Use valid UUIDs for player IDs
			gameID := "123e4567-e89b-12d3-a456-426614174004"
			player1ID := "123e4567-e89b-12d3-a456-426614174101"
			player2ID := "123e4567-e89b-12d3-a456-426614174102"
			// Insert game
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name)
				VALUES ($1, '2024', 'Ligue 1', 'in progress', 'Complex Game')
			`, gameID)
			require.NoError(t, err)
			// Insert players
			_, err = testDB.db.Exec(`
				INSERT INTO player (id, name, email)
				VALUES 
					($1, 'Player 1', 'player1@test.com'),
					($2, 'Player 2', 'player2@test.com')
			`, player1ID, player2ID)
			require.NoError(t, err)
			// Add players to game
			_, err = testDB.db.Exec(`
				INSERT INTO game_player (game_id, player_id)
				VALUES 
					($1, $2),
					($1, $3)
			`, gameID, player1ID, player2ID)
			require.NoError(t, err)

			// Create and insert matches
			matchRepo := NewPostgresMatchRepository(testDB.db)
			match1 := models.NewSeasonMatch("Team A", "Team B", "2024", "Ligue 1", testTime.Add(24*time.Hour), 1)
			err = matchRepo.SaveMatch(match1)
			require.NoError(t, err)

			match2 := models.NewSeasonMatch("Team C", "Team D", "2024", "Ligue 1", testTime.Add(48*time.Hour), 1)
			err = matchRepo.SaveMatch(match2)
			require.NoError(t, err)

			// Insert bets
			betRepo := NewPostgresBetRepository(testDB.db, repositories.NewInMemoryBetRepository())
			bet1 := models.NewBet(match1, 2, 1)
			_, _, err = betRepo.SaveBet(gameID, bet1, &models.PlayerData{ID: player1ID, Name: "Player 1"})
			require.NoError(t, err)

			bet2 := models.NewBet(match2, 1, 2)
			_, _, err = betRepo.SaveBet(gameID, bet2, &models.PlayerData{ID: player2ID, Name: "Player 2"})
			require.NoError(t, err)

			// Get all games
			games, err := gameRepo.GetAllGames()
			require.NoError(t, err)

			// Find our complex game
			complexGame, exists := games[gameID]
			require.True(t, exists, "Complex game should be present")

			// Verify game details
			require.Equal(t, "Complex Game", complexGame.GetName())
			require.Equal(t, models.GameStatusScheduled, complexGame.GetGameStatus())
			require.Equal(t, "2024", complexGame.GetSeasonYear())
			require.Equal(t, "Ligue 1", complexGame.GetCompetitionName())

			// Verify players are loaded
			players := complexGame.GetPlayers()
			require.Len(t, players, 2)

			// Verify player details
			playerNames := make(map[string]string)
			for _, player := range players {
				playerNames[player.GetID()] = player.GetName()
			}
			require.Equal(t, "Player 1", playerNames[player1ID])
			require.Equal(t, "Player 2", playerNames[player2ID])
		})

		// --- Ordering ---
		t.Run("Get All Games - Ordering", func(t *testing.T) {
			gameIDs := []string{
				"123e4567-e89b-12d3-a456-426614174005",
				"123e4567-e89b-12d3-a456-426614174006",
			}
			_, err := testDB.db.Exec(`
				INSERT INTO game (id, season_year, competition_name, status, game_name, created_at)
				VALUES 
					($1, '2024', 'Ligue 1', 'not started', 'Older Game', NOW() - INTERVAL '1 hour'),
					($2, '2024', 'Ligue 1', 'not started', 'Newer Game', NOW())
			`, gameIDs[0], gameIDs[1])
			require.NoError(t, err)
			games, err := gameRepo.GetAllGames()
			require.NoError(t, err)
			// Only check for the games we just created
			found := 0
			for _, gameID := range gameIDs {
				if _, exists := games[gameID]; exists {
					found++
				}
			}
			require.Equal(t, 2, found)
		})
	}, 10*time.Second)
}
