package api

import (
	"liguain/backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupMatchRoutes(router *gin.Engine) {
	router.GET("/api/matches", getMatches)
	router.GET("/api/bets", getBets)
}

var date1, _ = time.Parse(time.RFC3339, "2024-03-20T15:00:00Z")
var date2, _ = time.Parse(time.RFC3339, "2024-03-21T15:00:00Z")

var match1 = models.NewFinishedSeasonMatch(
	"Bastia",
	"Liverpool",
	4,
	0,
	"2024",
	"Champions's League",
	date1,
	1,
	1.5,
	2.0,
	3.0,
)

var match2 = models.NewSeasonMatchWithKnownOdds(
	"Arsenal",
	"Chelsea",
	"2024",
	"Premier League",
	date2,
	1,
	1.8,
	2.2,
	3.5,
)

func getMatches(c *gin.Context) {
	matches := []*models.SeasonMatch{match1, match2}
	c.JSON(http.StatusOK, matches)
}

func getBets(c *gin.Context) {
	bet1 := models.NewBet(match1, 1, 2)

	bets := []*models.Bet{bet1}
	c.JSON(http.StatusOK, bets)
}
