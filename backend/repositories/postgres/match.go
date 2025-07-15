package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"time"
)

// PostgresMatchRepository is a PostgreSQL implementation of MatchRepository
type PostgresMatchRepository struct {
	*PostgresRepository
	cache repositories.MatchRepository
	db    *sql.DB
}

// NewPostgresMatchRepository creates a new instance of PostgresMatchRepository
func NewPostgresMatchRepository(db *sql.DB) *PostgresMatchRepository {
	baseRepo := NewPostgresRepository(db)
	cache := repositories.NewInMemoryMatchRepository()
	return &PostgresMatchRepository{PostgresRepository: baseRepo, cache: cache, db: db}
}

// SaveMatch saves or updates a match
func (r *PostgresMatchRepository) SaveMatch(match models.Match) error {
	localId := match.Id()

	_, err := r.db.Exec(`
		INSERT INTO match (
			local_id, home_team_id, away_team_id, 
			home_team_score, away_team_score, match_date,
			match_status, season_code, competition_code, matchday
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (local_id) DO UPDATE SET
			home_team_score = EXCLUDED.home_team_score,
			away_team_score = EXCLUDED.away_team_score,
			match_date = EXCLUDED.match_date,
			match_status = EXCLUDED.match_status,
			updated_at = CURRENT_TIMESTAMP
	`, localId, match.GetHomeTeam(), match.GetAwayTeam(),
		match.GetHomeGoals(), match.GetAwayGoals(), match.GetDate(),
		match.GetStatus(), match.GetSeasonCode(), match.GetCompetitionCode(),
		match.(*models.SeasonMatch).Matchday)

	if err != nil {
		return err
	}

	// Update cache
	if err := r.cache.SaveMatch(match); err != nil {
		fmt.Printf("Warning: failed to update cache: %v\n", err)
	}

	return nil
}

// scanMatchRow scans a single row from a match query into a Match object
func scanMatchRow(rows *sql.Rows) (models.Match, error) {
	var id, homeTeamId, awayTeamId string
	var homeTeamScore, awayTeamScore sql.NullInt32
	var matchDate sql.NullTime
	var matchStatus string
	var seasonCode, competitionCode string
	var matchday int

	if err := rows.Scan(
		&id,
		&homeTeamId,
		&awayTeamId,
		&homeTeamScore,
		&awayTeamScore,
		&matchDate,
		&matchStatus,
		&seasonCode,
		&competitionCode,
		&matchday,
	); err != nil {
		return nil, fmt.Errorf("error scanning match: %v", err)
	}

	return CreateMatchFromDB(homeTeamId, awayTeamId, seasonCode, competitionCode, matchDate.Time, matchday, matchStatus, homeTeamScore, awayTeamScore), nil
}

// GetMatch returns a match by its id
func (r *PostgresMatchRepository) GetMatch(matchId string) (models.Match, error) {
	// Try cache first
	match, err := r.cache.GetMatch(matchId)
	if err == nil && match != nil {
		return match, nil
	}

	var seasonMatch models.SeasonMatch
	var date time.Time

	err = r.db.QueryRow(`
		SELECT local_id, home_team_id, away_team_id,
			   home_team_score, away_team_score, match_date,
			   match_status, season_code, competition_code, matchday
		FROM match
		WHERE local_id = $1
	`, matchId).Scan(
		&seasonMatch.HomeTeam, &seasonMatch.AwayTeam,
		&seasonMatch.HomeGoals, &seasonMatch.AwayGoals, &date,
		&seasonMatch.Status, &seasonMatch.SeasonCode, &seasonMatch.CompetitionCode,
		&seasonMatch.Matchday)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("match not found: %s", matchId)
	}
	if err != nil {
		return nil, err
	}

	seasonMatch.Date = date
	return &seasonMatch, nil
}

// GetMatches returns all matches
func (r *PostgresMatchRepository) GetMatches() (map[string]models.Match, error) {
	query := `
		SELECT local_id, home_team_id, away_team_id,
			   home_team_score, away_team_score, match_date,
			   match_status, season_code, competition_code, matchday
		FROM match
	`

	rows, err := r.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("error getting matches: %v", err)
	}
	defer rows.Close()

	matches := make(map[string]models.Match)
	for rows.Next() {
		var seasonMatch models.SeasonMatch
		var date time.Time
		var localId string

		err := rows.Scan(
			&localId, &seasonMatch.HomeTeam, &seasonMatch.AwayTeam,
			&seasonMatch.HomeGoals, &seasonMatch.AwayGoals, &date,
			&seasonMatch.Status, &seasonMatch.SeasonCode, &seasonMatch.CompetitionCode,
			&seasonMatch.Matchday)

		if err != nil {
			return nil, err
		}

		seasonMatch.Date = date
		matches[localId] = &seasonMatch
	}

	return matches, nil
}

// GetMatchesByCompetitionAndSeason returns all matches for a specific competition and season
func (r *PostgresMatchRepository) GetMatchesByCompetitionAndSeason(competitionCode, seasonCode string) ([]models.Match, error) {
	query := `
		SELECT local_id, home_team_id, away_team_id,
			   home_team_score, away_team_score, match_date,
			   match_status, season_code, competition_code, matchday
		FROM match
		WHERE competition_code = $1 AND season_code = $2
		ORDER BY matchday, match_date
	`

	rows, err := r.db.QueryContext(context.Background(), query, competitionCode, seasonCode)
	if err != nil {
		return nil, fmt.Errorf("error getting matches by competition and season: %v", err)
	}
	defer rows.Close()

	var matches []models.Match
	for rows.Next() {
		match, err := scanMatchRow(rows)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating match rows: %v", err)
	}

	return matches, nil
}
