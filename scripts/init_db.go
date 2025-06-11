package main

import (
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"log"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const (
	dbUser     = "postgres"
	dbPassword = "postgres"
	dbName     = "ligain_test"
	dbHost     = "localhost"
	dbPort     = 5432
)

func main() {
	// Connect to database
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	err = runMigrations(db)
	if err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Insert test data
	err = insertTestData(db)
	if err != nil {
		log.Fatal("Failed to insert test data:", err)
	}

	log.Println("Database initialized successfully!")
}

func runMigrations(db *sql.DB) error {
	// Initialize golang-migrate
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://backend/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	return nil
}

func insertTestData(db *sql.DB) error {
	// Insert game with hardcoded UUID
	gameId := "123e4567-e89b-12d3-a456-426614174000" // Hardcoded UUID that can be used in main.go
	_, err := db.Exec(`
		INSERT INTO game (id, season_year, competition_name, status)
		VALUES ($1, $2, $3, $4)`,
		gameId, "2024", "Premier League", "started")
	if err != nil {
		return fmt.Errorf("failed to insert game: %v", err)
	}

	// Insert players
	playerIds := make(map[string]string)
	for _, name := range []string{"Player1", "Player2"} {
		var playerId string
		err := db.QueryRow(`
			INSERT INTO player (name)
			VALUES ($1)
			RETURNING id`,
			name).Scan(&playerId)
		if err != nil {
			return fmt.Errorf("failed to insert player %s: %v", name, err)
		}
		playerIds[name] = playerId
	}

	// Insert matches
	matches := []struct {
		homeTeam    string
		awayTeam    string
		homeScore   *int
		awayScore   *int
		date        time.Time
		status      string
		season      string
		competition string
		matchday    int
	}{
		{
			homeTeam:    "Bastia",
			awayTeam:    "Liverpool",
			homeScore:   intPtr(4),
			awayScore:   intPtr(0),
			date:        parseTime("2024-03-20T15:00:00Z"),
			status:      string(models.MatchStatusFinished),
			season:      "2024",
			competition: "Champions's League",
			matchday:    1,
		},
		{
			homeTeam:    "Olympique de Marseille",
			awayTeam:    "Le Raincy",
			homeScore:   intPtr(0),
			awayScore:   intPtr(5),
			date:        parseTime("2024-03-20T17:00:00Z"),
			status:      string(models.MatchStatusFinished),
			season:      "2024",
			competition: "Ligue 1",
			matchday:    1,
		},
		{
			homeTeam:    "Paris Saint-Germain",
			awayTeam:    "Inter Milan",
			date:        parseTime("2024-03-20T20:00:00Z"),
			status:      string(models.MatchStatusStarted),
			season:      "2024",
			competition: "Champions's League",
			matchday:    1,
		},
		{
			homeTeam:    "Arsenal",
			awayTeam:    "Chelsea",
			date:        parseTime("2024-03-21T15:00:00Z"),
			status:      string(models.MatchStatusScheduled),
			season:      "2024",
			competition: "Premier League",
			matchday:    1,
		},
	}

	matchIds := make(map[string]string)
	for _, m := range matches {
		var matchId string
		err := db.QueryRow(`
			INSERT INTO match (
				home_team_id, away_team_id, home_team_score, away_team_score,
				match_date, match_status, season_code, competition_code, matchday
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`,
			m.homeTeam, m.awayTeam, m.homeScore, m.awayScore,
			m.date, m.status, m.season, m.competition, m.matchday,
		).Scan(&matchId)
		if err != nil {
			return fmt.Errorf("failed to insert match %s vs %s: %v", m.homeTeam, m.awayTeam, err)
		}
		matchIds[fmt.Sprintf("%s-%s", m.homeTeam, m.awayTeam)] = matchId
	}

	// Insert bets
	bets := []struct {
		matchKey           string
		playerName         string
		predictedHomeGoals int
		predictedAwayGoals int
	}{
		{"Bastia-Liverpool", "Player1", 0, 2},
		{"Bastia-Liverpool", "Player2", 0, 3},
		{"Olympique de Marseille-Le Raincy", "Player1", 0, 3},
		{"Olympique de Marseille-Le Raincy", "Player2", 5, 1},
		{"Paris Saint-Germain-Inter Milan", "Player1", 0, 3},
		{"Paris Saint-Germain-Inter Milan", "Player2", 5, 1},
	}

	for _, b := range bets {
		var betId string
		err := db.QueryRow(`
			INSERT INTO bet (
				game_id, match_id, player_id,
				predicted_home_goals, predicted_away_goals
			)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			gameId, matchIds[b.matchKey], playerIds[b.playerName],
			b.predictedHomeGoals, b.predictedAwayGoals,
		).Scan(&betId)
		if err != nil {
			return fmt.Errorf("failed to insert bet for %s by %s: %v", b.matchKey, b.playerName, err)
		}
	}

	// Insert scores
	scores := []struct {
		matchKey   string
		playerName string
		points     int
	}{
		{"Bastia-Liverpool", "Player1", 0},
		{"Bastia-Liverpool", "Player2", 0},
		{"Olympique de Marseille-Le Raincy", "Player1", 400},
		{"Olympique de Marseille-Le Raincy", "Player2", 0},
	}

	for _, s := range scores {
		_, err := db.Exec(`
			INSERT INTO score (bet_id, points)
			SELECT b.id, $1
			FROM bet b
			JOIN match m ON b.match_id = m.id
			JOIN player p ON b.player_id = p.id
			WHERE m.home_team_id = $2
			AND m.away_team_id = $3
			AND p.name = $4`,
			s.points,
			s.matchKey[:strings.Index(s.matchKey, "-")],
			s.matchKey[strings.Index(s.matchKey, "-")+1:],
			s.playerName,
		)
		if err != nil {
			return fmt.Errorf("failed to insert score for %s by %s: %v", s.matchKey, s.playerName, err)
		}
	}

	return nil
}

func intPtr(i int) *int {
	return &i
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
