package routes

import (
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var gameRepo = repositories.NewInMemoryGameRepository()

// SimplifiedBet represents a bet with only the essential information needed for the API
type SimplifiedBet struct {
	PredictedHomeGoals int `json:"predictedHomeGoals"`
	PredictedAwayGoals int `json:"predictedAwayGoals"`
}

// convertMatchResultToJSON converts a MatchResult to a JSON-friendly structure
func convertMatchResultToJSON(matchResult *models.MatchResult) map[string]any {
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

func SetupMatchRoutes(router *gin.Engine) {
	router.GET("/api/matches", getMatches)
	router.POST("/api/bet", saveBet)
}

func getMatches(c *gin.Context) {
	game, _ := gameRepo.GetGame("1")
	incomingMatches := game.GetIncomingMatches()
	pastMatches := game.GetPastResults()

	// Convert MatchResults to JSON-friendly format
	jsonIncomingMatches := make(map[string]any)
	for id, matchResult := range incomingMatches {
		jsonIncomingMatches[id] = convertMatchResultToJSON(matchResult)
	}

	jsonPastMatches := make(map[string]any)
	for id, matchResult := range pastMatches {
		jsonPastMatches[id] = convertMatchResultToJSON(matchResult)
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

func saveBet(c *gin.Context) {
	var request SaveBetRequest
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

	game, _ := gameRepo.GetGame("1")
	incomingMatches := game.GetIncomingMatches()
	match, exists := incomingMatches[request.MatchID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Match %s not found", request.MatchID)})
		return
	}

	bet := models.NewBet(match.Match, request.PredictedHomeGoals, request.PredictedAwayGoals)
	game.AddPlayerBet(models.Player{Name: "Player1"}, bet)

	c.JSON(http.StatusOK, gin.H{
		"message": "Bet saved successfully",
		"bet":     bet,
	})
}
