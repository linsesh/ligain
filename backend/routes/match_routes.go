package api

/*
import (
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
)

var gameRepo = repositories.NewInMemoryGameRepository()

func SetupMatchRoutes(router *gin.Engine) {
	router.GET("/api/matches", getMatches)
	router.GET("/api/bets", getBets)
	router.POST("/api/bet", saveBet)
}

func getMatches(c *gin.Context) {
	matches := gameRepo.GetMatches()
	c.JSON(http.StatusOK, matches)
}

func getBets(c *gin.Context) {
	bet1 := models.NewBet(match1, 1, 2)

	bets := []*models.Bet{bet1}
	c.JSON(http.StatusOK, bets)
}

type SaveBetRequest struct {
	MatchID            string `json:"matchId" binding:"required"`
	PredictedHomeGoals int    `json:"predictedHomeGoals" binding:"required"`
	PredictedAwayGoals int    `json:"predictedAwayGoals" binding:"required"`
}

func saveBet(c *gin.Context) {
	var request SaveBetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the matchId corresponds to one of our test matches
	var match models.Match
	if request.MatchID == match1.Id() {
		match = match1
	} else if request.MatchID == match2.Id() {
		match = match2
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Match %s not found", request.MatchID)})
		return
	}

	bet := models.NewBet(match, request.PredictedHomeGoals, request.PredictedAwayGoals)

	// TODO: Get the actual player from the session/authentication
	// For now, we'll use a test player
	_ = models.Player{Name: "TestPlayer"}

	// TODO: Get the game service from dependency injection
	// For now, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Bet saved successfully",
		"bet":     bet,
	})
}
*/
