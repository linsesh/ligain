package main

import (
	"context"
	"database/sql"
	"liguain/backend/middleware"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/repositories/postgres"
	"liguain/backend/routes"
	"liguain/backend/services"
	"log"
	"os"
	"strings"
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
	// Set Gin to release mode in production
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Successfully connected to database")

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

	// Setup CORS with specific origins
	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		allowedOriginsStr = "http://localhost:3000" // Default fallback
	}
	allowedOrigins := strings.Split(allowedOriginsStr, ",")

	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Apply API key authentication middleware
	router.Use(middleware.APIKeyAuth())

	// Setup routes
	matchHandler := routes.NewMatchHandler(map[string]services.GameService{
		"123e4567-e89b-12d3-a456-426614174000": gameService,
	})
	matchHandler.SetupRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
