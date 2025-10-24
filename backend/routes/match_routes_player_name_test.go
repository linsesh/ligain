package routes

import (
	"context"
	"encoding/json"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestPlayerNameUpdateReflectedInMatches tests the bug where updating a player's name
// doesn't reflect in match bets/scores because GetPlayers() returns current DB state
// but the game was loaded before the name update
func TestPlayerNameUpdateReflectedInMatches(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// STEP 1: Create player with ORIGINAL name
	originalName := "OriginalName"
	updatedName := "UpdatedName"
	playerID := "player-test-123"
	gameID := "123e4567-e89b-12d3-a456-426614174000"

	player := &models.PlayerData{
		ID:   playerID,
		Name: originalName,
	}

	// STEP 2: Create a game with a match and a bet from the player
	gameRepo := repositories.NewInMemoryGameRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	playerRepo := repositories.NewInMemoryPlayerRepository()
	gamePlayerRepo := repositories.NewInMemoryGamePlayerRepository(playerRepo)

	// Save player with original name to the repository
	ctx := context.Background()
	playerRepo.CreatePlayer(ctx, player)

	// Create a match
	matchDate := time.Now().Add(24 * time.Hour)
	match := models.NewSeasonMatch("Team A", "Team B", "2024", "Ligue 1", matchDate, 1)
	bet := models.NewBet(match, 2, 1)

	// Create a mock game that has the player's bet
	mockGame := &MockGame{
		incomingMatches: make(map[string]*models.MatchResult),
		pastMatches:     make(map[string]*models.MatchResult),
		bets:            make(map[string]map[string]*models.Bet),
	}

	matchResult := models.NewMatchWithBets(match, map[models.Player]*models.Bet{
		player: bet,
	})
	mockGame.incomingMatches[match.Id()] = matchResult
	gameRepo.SaveWithId(gameID, mockGame)

	// Create game service
	gameService := services.NewGameService(gameID, gameRepo, betRepo, gamePlayerRepo)
	gameService.AddPlayer(player)

	// Add player to the game-player mapping in the repository
	gamePlayerRepo.AddPlayerToGame(ctx, gameID, player.GetID())

	// STEP 3: Simulate player name update in database
	// This is what happens when UpdateDisplayName is called
	player.Name = updatedName
	playerRepo.UpdatePlayer(ctx, player)

	// STEP 4: Reload the game service (simulating a new request after name update)
	// The game service will call GetPlayers() which loads from the player repo
	gameServiceAfterUpdate := services.NewGameService(gameID, gameRepo, betRepo, gamePlayerRepo)
	gameServiceAfterUpdate.AddPlayer(player) // Add the updated player

	// Setup mocks
	mockGameCreationService := &MockGameCreationService{}
	mockGameCreationService.On("GetGameService", gameID, mock.AnythingOfType("*models.PlayerData")).
		Return(gameServiceAfterUpdate, nil)

	mockAuthService := &MockBetAuthService{}

	// Create handler and router
	handler := NewMatchHandler(mockGameCreationService, mockAuthService)
	router := gin.New()
	router.GET("/api/game/:game-id/matches", middleware.PlayerAuth(mockAuthService), handler.getMatches)

	// STEP 5: Make API request to get matches
	req := httptest.NewRequest("GET", "/api/game/"+gameID+"/matches", nil)
	req.Header.Set("Authorization", "Bearer testtoken")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// STEP 6: Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	t.Logf("Full response: %+v", response)

	// Extract bet data
	incomingMatches := response["incomingMatches"].(map[string]any)
	t.Logf("Incoming matches: %+v", incomingMatches)

	matchData := incomingMatches[match.Id()].(map[string]any)
	t.Logf("Match data: %+v", matchData)

	bets := matchData["bets"]
	t.Logf("Bets (raw): %+v (type: %T)", bets, bets)

	if bets == nil {
		t.Fatal("Bets is nil - this means the game doesn't have the player's bet loaded properly")
	}

	betsMap := bets.(map[string]any)
	t.Logf("Bets map: %+v", betsMap)

	// Find the player's bet - since we don't know the exact UUID, just grab the first bet
	var playerBet map[string]any
	var actualPlayerID string
	for pid, betData := range betsMap {
		playerBet = betData.(map[string]any)
		actualPlayerID = pid
		break
	}

	assert.NotNil(t, playerBet, "Player bet should exist in response")
	t.Logf("Found bet for player ID: %s", actualPlayerID)

	// THE KEY ASSERTION: Player name should be the UPDATED name
	actualPlayerName := playerBet["playerName"].(string)

	t.Logf("Expected player name: %s", updatedName)
	t.Logf("Actual player name: %s", actualPlayerName)

	assert.Equal(t, updatedName, actualPlayerName,
		"BUG REPRODUCED: Player name in match bet should be '%s' (updated name), but got '%s' (original name). "+
			"This happens because GetPlayers() returns the current DB state, but the bet data is cached in the game.",
		updatedName, actualPlayerName)
}
