package main

import (
	"context"
	"database/sql"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/repositories/postgres"
	"ligain/backend/routes"
	"ligain/backend/services"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
)

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

	betRepo := postgres.NewPostgresBetRepository(db, repositories.NewInMemoryBetRepository())
	playerRepo := postgres.NewPostgresPlayerRepository(db)
	matchRepo := postgres.NewPostgresMatchRepository(db)

	// Create match watcher service
	var watcher services.MatchWatcherService
	env := os.Getenv("ENV")
	matches, err := matchRepo.GetMatchesByCompetitionAndSeason("Ligue 1", "2025-2026")
	if err != nil {
		log.Fatal("Failed to get matches:", err)
	}
	matchesMap := make(map[string]models.Match)
	for _, match := range matches {
		matchesMap[match.Id()] = match
	}
	watcher, err = services.NewMatchWatcherServiceSportsmonk(env, matchesMap, matchRepo)
	if err != nil {
		log.Fatal("Failed to create match watcher service:", err)
	}

	// Start the match watcher service
	ctx := context.Background()
	if err := watcher.Start(ctx); err != nil {
		log.Fatal("Failed to start match watcher service:", err)
	}

	// Initialize authentication service
	authService := services.NewAuthService(playerRepo)

	// Initialize game creation service
	gameCodeRepo := postgres.NewPostgresGameCodeRepository(db)
	gamePlayerRepo := postgres.NewPostgresGamePlayerRepository(db)
	gameCreationService := services.NewGameCreationService(gameRepo, gameCodeRepo, gamePlayerRepo, betRepo, matchRepo, watcher)

	// Load the hardcoded game explicitly
	hardcodedGameID := "123e4567-e89b-12d3-a456-426614174000"
	if _, err := gameCreationService.GetGameService(hardcodedGameID); err != nil {
		log.Printf("Warning: Failed to load hardcoded game: %v", err)
	}

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
	matchHandler := routes.NewMatchHandler(gameCreationService, authService)
	matchHandler.SetupRoutes(router)

	// Setup authentication routes
	authHandler := routes.NewAuthHandler(authService)
	authHandler.SetupRoutes(router)

	// Setup game creation routes
	gameHandler := routes.NewGameHandler(gameCreationService, authService)
	gameHandler.SetupRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
