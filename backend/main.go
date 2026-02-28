package main

import (
	"context"
	"database/sql"
	"fmt"
	"ligain/backend/api"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/repositories/postgres"
	"ligain/backend/routes"
	"ligain/backend/rules"
	"ligain/backend/services"
	"ligain/backend/storage"
	"net/http"
	_ "net/http/pprof"
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

	env := os.Getenv("ENV")

	// Set Gin to release mode in production
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	var (
		playerRepo     repositories.PlayerRepository
		gameRepo       repositories.GameRepository
		betRepo        repositories.BetRepository
		matchRepo      repositories.MatchRepository
		gameCodeRepo   repositories.GameCodeRepository
		gamePlayerRepo repositories.GamePlayerRepository
		uow            repositories.UnitOfWork
		watcher        services.MatchWatcherService
	)

	if env == "fake" {
		fakeDBURL := os.Getenv("DATABASE_URL")
		if fakeDBURL != "" {
			log.Info("Running in fake mode — using real postgres")

			db, err := sql.Open("pgx", fakeDBURL)
			if err != nil {
				log.Fatal("Failed to connect to database:", err)
			}
			defer db.Close()

			if err := db.Ping(); err != nil {
				log.Fatal("Failed to ping database:", err)
			}

			pgGameRepo, err := postgres.NewPostgresGameRepository(db)
			if err != nil {
				log.Fatal("Failed to create game repository:", err)
			}
			gameRepo = pgGameRepo
			betRepo = postgres.NewPostgresBetRepository(db)
			playerRepo = postgres.NewPostgresPlayerRepository(db)
			matchRepo = postgres.NewPostgresMatchRepository(db)
			gameCodeRepo = postgres.NewPostgresGameCodeRepository(db)
			gamePlayerRepo = postgres.NewPostgresGamePlayerRepository(db)
			uow = postgres.NewUnitOfWork(db)

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
			simulatedAPI := api.NewSimulatedSportsmonkAPI(api.NewSportsmonkAPI(os.Getenv("SPORTSMONK_API_TOKEN")))
			sportsmonkRepo := repositories.NewSportsmonkRepository(simulatedAPI)
			watcher = services.NewMatchWatcherServiceSportsmonkWithOptions(sportsmonkRepo, matchesMap, matchRepo, 500*time.Millisecond)
			log.Info("Running in fake mode: SimulatedSportsmonkAPI + real postgres (500ms poll)")
		} else {
			log.Info("Running in fake mode — all repositories are in-memory, no DB required")

			inMemPlayerRepo := repositories.NewInMemoryPlayerRepository()
			playerRepo = inMemPlayerRepo
			gameRepo = repositories.NewInMemoryGameRepository()
			betRepo = repositories.NewInMemoryBetRepository()
			matchRepo = repositories.NewInMemoryMatchRepository()
			gameCodeRepo = repositories.NewInMemoryGameCodeRepository()
			gamePlayerRepo = repositories.NewInMemoryGamePlayerRepository(inMemPlayerRepo)
			uow = repositories.NewNoopUnitOfWork()

			fakeSeasonMatches := []models.SeasonMatch{
				*models.NewSeasonMatchWithKnownOdds("PSG", "Lyon", "2025/2026", "Ligue 1", time.Now().Add(1*time.Hour), 1, 1.5, 4.0, 3.5),
				*models.NewSeasonMatchWithKnownOdds("Marseille", "Nice", "2025/2026", "Ligue 1", time.Now().Add(2*time.Hour), 1, 2.0, 3.0, 3.0),
				*models.NewSeasonMatchWithKnownOdds("Monaco", "Rennes", "2025/2026", "Ligue 1", time.Now().Add(3*time.Hour), 1, 1.8, 3.5, 3.2),
				*models.NewSeasonMatchWithKnownOdds("Lens", "Lille", "2025/2026", "Ligue 1", time.Now().Add(4*time.Hour), 2, 2.5, 2.8, 3.1),
				*models.NewSeasonMatchWithKnownOdds("Strasbourg", "Nantes", "2025/2026", "Ligue 1", time.Now().Add(5*time.Hour), 2, 2.2, 2.9, 3.2),
				*models.NewSeasonMatchWithKnownOdds("Toulouse", "Brest", "2025/2026", "Ligue 1", time.Now().Add(6*time.Hour), 2, 2.4, 2.7, 3.3),
				*models.NewSeasonMatchWithKnownOdds("Montpellier", "Metz", "2025/2026", "Ligue 1", time.Now().Add(7*time.Hour), 3, 2.1, 3.1, 3.2),
				*models.NewSeasonMatchWithKnownOdds("Auxerre", "Reims", "2025/2026", "Ligue 1", time.Now().Add(8*time.Hour), 3, 2.3, 3.0, 3.0),
			}
			matchesMap := make(map[string]models.Match)
			for i := range fakeSeasonMatches {
				matchesMap[fakeSeasonMatches[i].Id()] = &fakeSeasonMatches[i]
			}
			fakeAPI := api.NewFakeRandomSportsmonkAPI(fakeSeasonMatches)
			fakeRepo := repositories.NewSportsmonkRepository(fakeAPI)
			watcher = services.NewMatchWatcherServiceSportsmonkWithOptions(fakeRepo, matchesMap, matchRepo, 500*time.Millisecond)

			// Seed fake games so the watcher has subscribers and HandleMatchUpdates is exercised
			allFakeMatches := make([]models.Match, 0, len(fakeSeasonMatches))
			for i := range fakeSeasonMatches {
				allFakeMatches = append(allFakeMatches, &fakeSeasonMatches[i])
			}
			for i := 1; i <= 3; i++ {
				game := rules.NewFreshGame("2025/2026", "Ligue 1", fmt.Sprintf("Fake Game %d", i), []models.Player{}, allFakeMatches, &rules.ScorerOriginal{})
				if _, err := gameRepo.CreateGame(game); err != nil {
					log.Fatalf("Failed to seed fake game %d: %v", i, err)
				}
			}
			log.Info("Running in fake mode: FakeRandomSportsmonkAPI (500ms poll), 3 games seeded")
		}
	} else {
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			log.Fatal("DATABASE_URL environment variable is not set")
		}

		db, err := sql.Open("pgx", databaseURL)
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			log.Fatal("Failed to ping database:", err)
		}

		pgGameRepo, err := postgres.NewPostgresGameRepository(db)
		if err != nil {
			log.Fatal("Failed to create game repository:", err)
		}
		gameRepo = pgGameRepo
		betRepo = postgres.NewPostgresBetRepository(db)
		playerRepo = postgres.NewPostgresPlayerRepository(db)
		matchRepo = postgres.NewPostgresMatchRepository(db)
		gameCodeRepo = postgres.NewPostgresGameCodeRepository(db)
		gamePlayerRepo = postgres.NewPostgresGamePlayerRepository(db)
		uow = postgres.NewUnitOfWork(db)

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
		w, err := services.NewMatchWatcherServiceSportsmonk(env, matchesMap, matchRepo)
		if err != nil {
			log.Fatal("Failed to create match watcher service:", err)
		}
		watcher = w
	}

	if err := watcher.Start(ctx); err != nil {
		log.Fatal("Failed to start match watcher service:", err)
	}

	authService := services.NewAuthService(playerRepo)

	registry, err := services.NewGameServiceRegistry(gameRepo, betRepo, gamePlayerRepo, watcher)
	if err != nil {
		log.Fatal("Failed to create game registry:", err)
	}

	membershipService := services.NewGameMembershipService(uow, gamePlayerRepo, gameRepo, gameCodeRepo, registry, watcher)
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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, X-App-Version")
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

	// Apply version check middleware
	router.Use(middleware.VersionCheck())

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

	// Start pprof server on :6060 for heap profiling
	go func() {
		log.Info("Starting pprof server on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Errorf("pprof server error: %v", err)
		}
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
