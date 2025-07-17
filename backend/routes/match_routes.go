package routes

import (
	"fmt"
	"liguain/backend/middleware"
	"liguain/backend/models"
	"liguain/backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// MatchHandler handles all match-related routes
type MatchHandler struct {
	gameCreationService services.GameCreationServiceInterface
	authService         services.AuthServiceInterface
}

// NewMatchHandler creates a new MatchHandler instance
func NewMatchHandler(gameCreationService services.GameCreationServiceInterface, authService services.AuthServiceInterface) *MatchHandler {
	return &MatchHandler{
		gameCreationService: gameCreationService,
		authService:         authService,
	}
}

// SimplifiedBet represents a bet with only the essential information needed for the API
type SimplifiedBet struct {
	PlayerID           string `json:"playerId"`
	PlayerName         string `json:"playerName"`
	PredictedHomeGoals int    `json:"predictedHomeGoals"`
	PredictedAwayGoals int    `json:"predictedAwayGoals"`
}

// SimplifiedScore represents a score with player information
type SimplifiedScore struct {
	PlayerID   string `json:"playerId"`
	PlayerName string `json:"playerName"`
	Points     int    `json:"points"`
}

// getAuthenticatedPlayer extracts the authenticated player from the context
func (h *MatchHandler) getAuthenticatedPlayer(c *gin.Context) (models.Player, error) {
	playerInterface, exists := c.Get("player")
	if !exists || playerInterface == nil {
		return nil, fmt.Errorf("player not found in context")
	}
	player, ok := playerInterface.(models.Player)
	if !ok {
		return nil, fmt.Errorf("invalid player type in context")
	}
	return player, nil
}

// convertMatchResultToJSON converts a MatchResult to a JSON-friendly structure
func (h *MatchHandler) convertMatchResultToJSON(matchResult *models.MatchResult, gameService services.GameService) map[string]any {
	result := map[string]any{
		"match": matchResult.Match,
	}
	playerIDToName := h.getPlayerIDToName(gameService)
	if matchResult.Bets != nil {
		simplifiedBets := h.simplifyBets(matchResult.Bets, playerIDToName)
		result["bets"] = simplifiedBets
	} else {
		result["bets"] = nil
	}

	if matchResult.Scores != nil {
		result["scores"] = h.simplifyScores(matchResult.Scores, playerIDToName)
	} else {
		result["scores"] = nil
	}

	return result
}

func (h *MatchHandler) getPlayerIDToName(gameService services.GameService) map[string]string {
	playerIDToName := make(map[string]string)
	for _, player := range gameService.GetPlayers() {
		playerIDToName[player.GetID()] = player.GetName()
	}
	return playerIDToName
}

func (h *MatchHandler) simplifyBets(bets map[string]*models.Bet, playerIDToName map[string]string) map[string]SimplifiedBet {
	simplifiedBets := make(map[string]SimplifiedBet)
	for playerID, bet := range bets {
		simplifiedBets[playerID] = SimplifiedBet{
			PlayerID:           playerID,
			PlayerName:         playerIDToName[playerID],
			PredictedHomeGoals: bet.PredictedHomeGoals,
			PredictedAwayGoals: bet.PredictedAwayGoals,
		}
	}
	return simplifiedBets
}

func (h *MatchHandler) simplifyScores(scores map[string]int, playerIDToName map[string]string) map[string]SimplifiedScore {
	simplifiedScores := make(map[string]SimplifiedScore)
	for playerID, score := range scores {
		simplifiedScores[playerID] = SimplifiedScore{
			PlayerID:   playerID,
			PlayerName: playerIDToName[playerID],
			Points:     score,
		}
	}
	return simplifiedScores
}

// SetupRoutes registers all match-related routes
func (h *MatchHandler) SetupRoutes(router *gin.Engine) {
	router.GET("/api/game/:game-id/matches", middleware.PlayerAuth(h.authService), h.getMatches)
	router.POST("/api/game/:game-id/bet", middleware.PlayerAuth(h.authService), h.saveBet)
}

func (h *MatchHandler) getMatches(c *gin.Context) {
	gameId := c.Param("game-id")
	if gameId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game-id is required"})
		return
	}

	gameService, err := h.gameCreationService.GetGameService(gameId)
	if err != nil {
		log.Errorf("Failed to get game service: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Your game was not found"})
		return
	}

	// Get authenticated player from context
	player, err := h.getAuthenticatedPlayer(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	incomingMatches := gameService.GetIncomingMatches(player) // Get matches filtered for this player
	pastMatches := gameService.GetMatchResults()

	// Convert MatchResults to JSON-friendly format
	jsonIncomingMatches := make(map[string]any)
	for id, matchResult := range incomingMatches {
		jsonIncomingMatches[id] = h.convertMatchResultToJSON(matchResult, gameService)
	}

	jsonPastMatches := make(map[string]any)
	for id, matchResult := range pastMatches {
		jsonPastMatches[id] = h.convertMatchResultToJSON(matchResult, gameService)
	}

	c.JSON(http.StatusOK, gin.H{
		"incomingMatches": jsonIncomingMatches,
		"pastMatches":     jsonPastMatches,
	})
}

type SaveBetRequest struct {
	MatchID            string `json:"matchId" binding:"required"`
	PredictedHomeGoals int    `json:"predictedHomeGoals" binding:"required"`
	PredictedAwayGoals int    `json:"predictedAwayGoals" binding:"required"`
}

func (h *MatchHandler) saveBet(c *gin.Context) {
	gameId := c.Param("game-id")
	if gameId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game-id is required"})
		return
	}

	var request SaveBetRequest
	log.Info("request", request)
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"request_body": c.Request.Body,
			"content_type": c.GetHeader("Content-Type"),
		}).Error("Failed to bind bet request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           fmt.Sprintf("Invalid request format: %v", err),
			"expected_format": "{\"matchId\": \"string\", \"predictedHomeGoals\": number, \"predictedAwayGoals\": number}",
		})
		return
	}

	gameService, err := h.gameCreationService.GetGameService(gameId)
	if err != nil {
		log.Errorf("Failed to get game service: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Your game was not found"})
		return
	}

	// Get authenticated player from context
	player, err := h.getAuthenticatedPlayer(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Get incoming matches filtered for this player only
	incomingMatches := gameService.GetIncomingMatches(player)
	matchResult, exists := incomingMatches[request.MatchID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Match %s not found", request.MatchID)})
		return
	}
	match := matchResult.Match

	bet := models.NewBet(match, request.PredictedHomeGoals, request.PredictedAwayGoals)
	updateErr := gameService.UpdatePlayerBet(player, bet, match.GetDate())
	if updateErr != nil {
		log.Error("Failed to update player bet", updateErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": updateErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bet saved successfully",
		"bet":     bet,
	})
}
