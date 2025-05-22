package api

import (
	"liguain/backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetupMatchRoutes(router *gin.Engine) {
	router.GET("/api/matches", getMatches)
}

func getMatches(c *gin.Context) {
	// Create two sample matches
	date1, _ := time.Parse(time.RFC3339, "2024-03-20T15:00:00Z")
	date2, _ := time.Parse(time.RFC3339, "2024-03-21T15:00:00Z")

	match1 := models.NewFinishedSeasonMatch(
		"Manchester United",
		"Liverpool",
		2,
		1,
		"2024",
		"Premier League",
		date1,
		1,
		1.5,
		2.0,
		3.0,
	)

	match2 := models.NewSeasonMatchWithKnownOdds(
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

	matches := []*models.SeasonMatch{match1, match2}
	c.JSON(http.StatusOK, matches)
}
