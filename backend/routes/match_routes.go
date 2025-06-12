package routes

import (
	"fmt"
	"liguain/backend/models"
	"liguain/backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// MatchHandler handles all match-related routes
type MatchHandler struct {
	gameServices map[string]services.GameService
}

// NewMatchHandler creates a new MatchHandler instance
func NewMatchHandler(gameServices map[string]services.GameService) *MatchHandler {
	return &MatchHandler{
		gameServices: gameServices,
	}
}

// SimplifiedBet represents a bet with only the essential information needed for the API
type SimplifiedBet struct {
	PredictedHomeGoals int `json:"predictedHomeGoals"`
	PredictedAwayGoals int `json:"predictedAwayGoals"`
}

// convertMatchResultToJSON converts a MatchResult to a JSON-friendly structure
func (h *MatchHandler) convertMatchResultToJSON(matchResult *models.MatchResult) map[string]any {
	result := map[string]any{
		"match": matchResult.Match,
	}

	log.Info("matchResult.Bets", matchResult.Bets)
	if matchResult.Bets != nil {
		simplifiedBets := make(map[string]SimplifiedBet)
		for player, bet := range matchResult.Bets {
			simplifiedBets[player.Name] = SimplifiedBet{
				PredictedHomeGoals: bet.PredictedHomeGoals,
				PredictedAwayGoals: bet.PredictedAwayGoals,
			}
		}
		result["bets"] = simplifiedBets
	} else {
		result["bets"] = nil
	}

	if matchResult.Scores != nil {
		simplifiedScores := make(map[string]int)
		for player, score := range matchResult.Scores {
			simplifiedScores[player.Name] = score
		}
		result["scores"] = simplifiedScores
	} else {
		result["scores"] = nil
	}

	return result
}

// SetupRoutes registers all match-related routes
func (h *MatchHandler) SetupRoutes(router *gin.Engine) {
	router.GET("/api/game/:game-id/matches", h.getMatches)
	router.POST("/api/game/:game-id/bet", h.saveBet)
}

func (h *MatchHandler) getMatches(c *gin.Context) {
	gameId := c.Param("game-id")
	if gameId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game-id is required"})
		return
	}

	gameService, exists := h.gameServices[gameId]
	if !exists {
		log.Error("Failed to get game service")
		c.JSON(http.StatusNotFound, gin.H{"error": "Your game was not found"})
		return
	}

	incomingMatches := gameService.GetIncomingMatches()
	pastMatches := gameService.GetMatchResults()

	// Convert MatchResults to JSON-friendly format
	jsonIncomingMatches := make(map[string]any)
	for id, matchResult := range incomingMatches {
		jsonIncomingMatches[id] = h.convertMatchResultToJSON(matchResult)
	}

	jsonPastMatches := make(map[string]any)
	for id, matchResult := range pastMatches {
		jsonPastMatches[id] = h.convertMatchResultToJSON(matchResult)
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

	gameService, exists := h.gameServices[gameId]
	if !exists {
		log.Error("Failed to get game service")
		c.JSON(http.StatusNotFound, gin.H{"error": "Your game was not found"})
		return
	}

	incomingMatches := gameService.GetIncomingMatches()
	match, exists := incomingMatches[request.MatchID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Match %s not found", request.MatchID)})
		return
	}

	bet := models.NewBet(match.Match, request.PredictedHomeGoals, request.PredictedAwayGoals)
	err := gameService.UpdatePlayerBet(models.Player{Name: "Player1"}, bet, match.Match.GetDate())
	if err != nil {
		log.Error("Failed to update player bet", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bet saved successfully",
		"bet":     bet,
	})
}
