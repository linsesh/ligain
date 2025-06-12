package main

import (
	"context"
	"database/sql"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/repositories/postgres"
	"liguain/backend/routes"
	"liguain/backend/services"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type MatchWatcherServiceMock struct{}

func (m *MatchWatcherServiceMock) WatchMatches(matches []models.Match) {}

func (m *MatchWatcherServiceMock) GetUpdates(ctx context.Context, done chan services.MatchWatcherServiceResult) {
}

func NewMatchWatcherServiceMock() services.MatchWatcherService {
	return &MatchWatcherServiceMock{}
}

func main() {
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/ligain_test?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	gameRepo, err := postgres.NewPostgresGameRepository(db)
	if err != nil {
		log.Fatal("Failed to create game repository:", err)
	}

	// Use the hardcoded game ID from init_db.go
	gameId := "123e4567-e89b-12d3-a456-426614174000"

	// Get the game from the database
	game, err := gameRepo.GetGame(gameId)
	if err != nil {
		log.Fatal("Failed to get game:", err)
	}

	betRepo := postgres.NewPostgresBetRepository(db, repositories.NewInMemoryBetRepository())
	//watcher, err := services.NewMatchWatcherServiceSportsmonk("local")
	//if err != nil {
	//	log.Fatal("Failed to create match watcher service:", err)
	//}
	watcher := NewMatchWatcherServiceMock()
	gameService := services.NewGameService(gameId, game, gameRepo, betRepo, watcher, 10*time.Second)
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
	matchHandler := routes.NewMatchHandler(map[string]services.GameService{
		"123e4567-e89b-12d3-a456-426614174000": gameService,
	})
	matchHandler.SetupRoutes(router)

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
