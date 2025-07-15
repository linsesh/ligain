package main

import (
	"database/sql"
	"fmt"
	"liguain/backend/api"
	"liguain/backend/models"
	postgresRepo "liguain/backend/repositories/postgres"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	// Ligue 1 competition ID from Sportsmonk API
	ligue1CompetitionId = 301
	// Season code for 2025/2026
	seasonCode = "2025/2026"
)

func main() {
	// Get database URL from environment or use default
	dbURL := getDatabaseURL()
	apiToken := getAPIToken()

	log.Printf("Connecting to database with URL: %s", dbURL)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Run migrations to ensure schema is up to date
	log.Println("Running migrations...")
	err = runMigrations(db)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Migrations completed successfully")

	// Initialize Sportsmonk API
	log.Println("Initializing Sportsmonk API...")
	sportsmonkAPI := api.NewSportsmonkAPI(apiToken)

	// Get season ID for 2025/2026
	log.Printf("Getting season ID for %s...", seasonCode)
	seasonIds, err := sportsmonkAPI.GetSeasonIds([]string{seasonCode}, ligue1CompetitionId)
	if err != nil {
		log.Fatalf("Failed to get season IDs: %v", err)
	}

	seasonId, exists := seasonIds[seasonCode]
	if !exists {
		log.Fatalf("Season %s not found in the API response", seasonCode)
	}
	log.Printf("Found season ID: %d", seasonId)

	// Fetch all fixtures for the season
	log.Printf("Fetching fixtures for season %s (ID: %d)...", seasonCode, seasonId)
	fixtures, err := sportsmonkAPI.GetSeasonFixtures(seasonId)
	if err != nil {
		log.Fatalf("Failed to get season fixtures: %v", err)
	}
	log.Printf("Retrieved %d fixtures", len(fixtures))

	// Initialize match repository
	matchRepo := postgresRepo.NewPostgresMatchRepository(db)

	// Save matches to database (deduplicated)
	log.Println("Saving matches to database...")
	savedCount := 0
	errorCount := 0
	seenMatches := make(map[string]bool) // Track seen match IDs to avoid duplicates

	for fixtureId, match := range fixtures {
		matchId := match.Id()
		if seenMatches[matchId] {
			log.Printf("Skipping duplicate match for database: %s (fixture ID: %d)", matchId, fixtureId)
			continue
		}
		seenMatches[matchId] = true

		err := matchRepo.SaveMatch(match)
		if err != nil {
			log.Printf("Error saving match %s (fixture ID: %d): %v", matchId, fixtureId, err)
			errorCount++
		} else {
			log.Printf("Saved match: %s", matchId)
			savedCount++
		}
	}

	log.Printf("Database operation completed!")
	log.Printf("Successfully saved: %d matches", savedCount)
	if errorCount > 0 {
		log.Printf("Failed to save: %d matches", errorCount)
	}

	// Print summary of matches by matchday (deduplicated)
	log.Println("\n=== MATCH SUMMARY BY MATCHDAY ===")
	matchesByMatchday := make(map[int][]models.Match)
	seenMatchesSummary := make(map[string]bool) // Track seen match IDs to avoid duplicates

	for _, match := range fixtures {
		matchId := match.Id()
		if seenMatchesSummary[matchId] {
			log.Printf("Skipping duplicate match: %s", matchId)
			continue
		}
		seenMatchesSummary[matchId] = true

		seasonMatch := match.(*models.SeasonMatch)
		matchesByMatchday[seasonMatch.Matchday] = append(matchesByMatchday[seasonMatch.Matchday], match)
	}

	for matchday := 1; matchday <= 34; matchday++ {
		if matches, exists := matchesByMatchday[matchday]; exists {
			log.Printf("Matchday %d: %d matches", matchday, len(matches))
			for _, match := range matches {
				log.Printf("  - %s vs %s (%s)", match.GetHomeTeam(), match.GetAwayTeam(), match.GetDate().Format("2006-01-02 15:04"))
			}
		} else {
			log.Printf("Matchday %d: No matches found", matchday)
		}
	}
}

func getDatabaseURL() string {
	// Check if DATABASE_URL environment variable is set
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	// Fall back to default local database configuration
	const (
		dbUser     = "postgres"
		dbPassword = "postgres"
		dbName     = "ligain_test"
		dbHost     = "localhost"
		dbPort     = 5432
	)

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)
}

func getAPIToken() string {
	apiToken := os.Getenv("SPORTSMONK_API_TOKEN")
	if apiToken == "" {
		log.Fatal("SPORTSMONK_API_TOKEN environment variable is required")
	}
	return apiToken
}

func runMigrations(db *sql.DB) error {
	// Initialize golang-migrate
	driver, err := migratePostgres.WithInstance(db, &migratePostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../backend/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}
