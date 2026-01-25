package main

import (
	"context"
	"database/sql"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/repositories/postgres"
	"ligain/backend/routes"
	"ligain/backend/services"
	"ligain/backend/storage"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Configure structured logging for Cloud Run
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	// Set Gin to release mode in production
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Get database URL from environment variable
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test the database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	gameRepo, err := postgres.NewPostgresGameRepository(db)
	if err != nil {
		log.Fatal("Failed to create game repository:", err)
	}

	betRepo := postgres.NewPostgresBetRepository(db)
	playerRepo := postgres.NewPostgresPlayerRepository(db)
	matchRepo := postgres.NewPostgresMatchRepository(db)

	env := os.Getenv("ENV")
	matches, err := matchRepo.GetMatchesByCompetitionAndSeason("Ligue 1", "2025/2026")
	log.Infof("Got %d matches", len(matches))
	if err != nil {
		log.Fatal("Failed to get matches:", err)
	}
	matchesMap := make(map[string]models.Match)
	for _, match := range matches {
		if !match.IsFinished() {
			matchesMap[match.Id()] = match
		}
	}
	watcher, err := services.NewMatchWatcherServiceSportsmonk(env, matchesMap, matchRepo)
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

	// Initialize game repositories
	gameCodeRepo := postgres.NewPostgresGameCodeRepository(db)
	gamePlayerRepo := postgres.NewPostgresGamePlayerRepository(db)

	// Initialize specialized game services
	registry, err := services.NewGameServiceRegistry(gameRepo, betRepo, gamePlayerRepo, watcher)
	if err != nil {
		log.Fatal("Failed to create game registry:", err)
	}

	membershipService := services.NewGameMembershipService(gamePlayerRepo, gameRepo, gameCodeRepo, registry, watcher)
	queryService := services.NewGameQueryService(gameRepo, gamePlayerRepo, gameCodeRepo, betRepo)
	joinService := services.NewGameJoinService(gameCodeRepo, gameRepo, gamePlayerRepo, membershipService, registry, time.Now)
	creationService := services.NewGameCreationServiceWithServices(
		gameRepo, gameCodeRepo, gamePlayerRepo, matchRepo,
		registry, membershipService, queryService, joinService,
		time.Now,
	)

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

	// Apply metrics middleware (should be before auth to capture all requests)
	router.Use(middleware.MetricsMiddleware())

	// Apply API key authentication middleware
	router.Use(middleware.APIKeyAuth())

	// Setup routes
	matchHandler := routes.NewMatchHandler(creationService, authService)
	matchHandler.SetupRoutes(router)

	// Setup authentication routes
	authHandler := routes.NewAuthHandler(authService)
	authHandler.SetupRoutes(router)

	// Setup game routes with all specialized services
	gameHandler := routes.NewGameHandler(creationService, joinService, queryService, membershipService, authService)
	gameHandler.SetupRoutes(router)

	// Setup profile routes (avatar upload requires GCS, display name works without it)
	var profileService services.ProfileService
	if bucketName := os.Getenv("GCS_BUCKET_NAME"); bucketName != "" {
		gcsStorage, err := storage.NewGCSBlobStorage(ctx, bucketName)
		if err != nil {
			log.Fatalf("Failed to create GCS storage: %v", err)
		}
		defer gcsStorage.Close()

		storageService := services.NewStorageService(gcsStorage)
		imageProcessor := services.NewImageProcessor()
		profileService = services.NewProfileService(storageService, imageProcessor, playerRepo)
		log.Infof("Profile routes enabled with GCS bucket: %s", bucketName)
	} else {
		log.Warn("GCS_BUCKET_NAME not set, avatar upload disabled")
	}
	profileHandler := routes.NewProfileHandler(profileService, authService)
	profileHandler.SetupRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
