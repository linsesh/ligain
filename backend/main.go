package main

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/rules"
	"liguain/backend/services"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

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

var players = []models.Player{
	{Name: "Player1"},
	{Name: "Player2"},
}

var matches = []models.Match{match1, match2}

type MatchWatcherServiceMock struct{}

func (m *MatchWatcherServiceMock) WatchMatches(matches []models.Match) {}

func (m *MatchWatcherServiceMock) GetUpdates(ctx context.Context, done chan services.MatchWatcherServiceResult) {
}

func NewMatchWatcherServiceMock() services.MatchWatcherService {
	return &MatchWatcherServiceMock{}
}

func main() {

	gameRepo := repositories.NewInMemoryGameRepository()
	scorer := rules.ScorerOriginal{}
	game := rules.NewGame("2023/2024", "Premier League", players, matches, &scorer)
	gameId, game, err := gameRepo.SaveGame(game)
	if err != nil {
		log.Fatal("Failed to save game:", err)
	}
	scoresRepo := repositories.NewInMemoryScoresRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	//watcher, err := services.NewMatchWatcherServiceSportsmonk("local")
	//if err != nil {
	//	log.Fatal("Failed to create match watcher service:", err)
	//}
	watcher := NewMatchWatcherServiceMock()
	gameService := services.NewGameService(gameId, game, gameRepo, scoresRepo, betRepo, watcher, 10*time.Second)
	gameService.Play()
	router := gin.Default()

	// Setup CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Setup routes
	//api.SetupMatchRoutes(router)

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
