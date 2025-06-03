package postgres

import (
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
)

type PostgresMatchRepository struct {
	*PostgresRepository
	cache repositories.MatchRepository
}

func NewPostgresMatchRepository(db *sql.DB) repositories.MatchRepository {
	baseRepo := NewPostgresRepository(db)
	cache := repositories.NewInMemoryMatchRepository()
	return &PostgresMatchRepository{PostgresRepository: baseRepo, cache: cache}
}

func (r *PostgresMatchRepository) SaveMatch(match models.Match) (string, error) {
	/* We know that the match is a SeasonMatch */
	seasonMatch := match.(*models.SeasonMatch)
	query := `
		INSERT INTO match (game_id, home_team_id, away_team_id, home_team_score, away_team_score,
						 match_date, match_status, season_code, competition_code, matchday)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			home_team_score = EXCLUDED.home_team_score,
			away_team_score = EXCLUDED.away_team_score,
			match_status = EXCLUDED.match_status
		RETURNING id`

	var id string
	err := r.db.QueryRow(
		query,
		match.Id(),
		match.GetHomeTeam(),
		match.GetAwayTeam(),
		match.GetHomeGoals(),
		match.GetAwayGoals(),
		match.GetDate(),
		match.IsFinished(),
		match.GetSeasonCode(),
		match.GetCompetitionCode(),
		seasonMatch.Matchday,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("error saving match: %v", err)
	}

	// Update cache
	if _, err := r.cache.SaveMatch(match); err != nil {
		fmt.Printf("Warning: failed to update cache: %v\n", err)
	}

	return id, nil
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

func (r *PostgresMatchRepository) GetMatch(matchId string) (models.Match, error) {
	// Try cache first
	match, err := r.cache.GetMatch(matchId)
	if err == nil && match != nil {
		return match, nil
	}

	query := `
		SELECT id, home_team_id, away_team_id, home_team_score, away_team_score,
			   match_date, match_status, season_code, competition_code, matchday
		FROM match
		WHERE id = $1`

	rows, err := r.db.Query(query, matchId)
	if err != nil {
		return nil, fmt.Errorf("error getting match: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("match %s not found", matchId)
	}

	match, err = scanMatchRow(rows)
	if err != nil {
		return nil, err
	}

	// Populate cache
	if _, err := r.cache.SaveMatch(match); err != nil {
		fmt.Printf("Warning: failed to update cache: %v\n", err)
	}

	return match, nil
}

func (r *PostgresMatchRepository) GetMatchesByGame(gameId string) ([]models.Match, error) {
	query := `
		SELECT id, home_team_id, away_team_id, home_team_score, away_team_score,
			   match_date, match_status, season_code, competition_code, matchday
		FROM match
		WHERE game_id = $1
		ORDER BY match_date ASC`

	rows, err := r.db.Query(query, gameId)
	if err != nil {
		return nil, fmt.Errorf("error getting matches: %v", err)
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

	return matches, nil
}
