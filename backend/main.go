package main

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/routes"
	"liguain/backend/rules"
	"liguain/backend/services"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

var date1, _ = time.Parse(time.RFC3339, "2024-03-20T15:00:00Z")
var date2, _ = time.Parse(time.RFC3339, "2024-03-20T17:00:00Z")
var date3, _ = time.Parse(time.RFC3339, "2024-03-20T20:00:00Z")
var date4, _ = time.Parse(time.RFC3339, "2024-03-21T15:00:00Z")

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

var match2 = models.NewFinishedSeasonMatch(
	"Olympique de Marseille",
	"Le Raincy",
	0,
	5,
	"2024",
	"Ligue 1",
	date2,
	1,
	1.01,
	14.0,
	32.0,
)

var match3 = models.NewSeasonMatchWithKnownOdds(
	"Paris Saint-Germain",
	"Inter Milan",
	"2024",
	"Champions's League",
	date3,
	1,
	1.8,
	2.2,
	3.5,
)

var match4 = models.NewSeasonMatchWithKnownOdds(
	"Arsenal",
	"Chelsea",
	"2024",
	"Premier League",
	date4,
	1,
	1.8,
	2.2,
	3.5,
)

var players = []models.Player{
	{Name: "Player1"},
	{Name: "Player2"},
}

var bets = map[string]map[models.Player]*models.Bet{
	match1.Id(): {
		players[0]: models.NewBet(match1, 0, 2),
		players[1]: models.NewBet(match1, 0, 3),
	},
	match2.Id(): {
		players[0]: models.NewBet(match2, 0, 3),
		players[1]: models.NewBet(match2, 5, 1),
	},
	match3.Id(): {
		players[0]: models.NewBet(match3, 0, 3),
		players[1]: models.NewBet(match3, 5, 1),
	},
}

var scores = map[string]map[models.Player]int{
	match1.Id(): {
		players[0]: 0,
		players[1]: 0,
	},
	match2.Id(): {
		players[0]: 400,
		players[1]: 0,
	},
}

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
	game := rules.NewStartedGame("2023/2024", "Premier League", players, []models.Match{match4}, []models.Match{match1, match2, match3}, &scorer, bets, scores)
	gameId, err := gameRepo.CreateGame(game)
	if err != nil {
		log.Fatal("Failed to save game:", err)
	}
	match3.Start()
	scoresRepo := repositories.NewInMemoryScoresRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	//watcher, err := services.NewMatchWatcherServiceSportsmonk("local")
	//if err != nil {
	//	log.Fatal("Failed to create match watcher service:", err)
	//}
	watcher := NewMatchWatcherServiceMock()
	gameService := services.NewGameService(gameId, game, gameRepo, scoresRepo, betRepo, watcher, 10*time.Second)
	go gameService.Play()
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
	matchHandler := routes.NewMatchHandler(gameRepo)
	matchHandler.SetupRoutes(router)

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
