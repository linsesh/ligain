package postgres

import (
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
)

type PostgresScoreRepository struct {
	db    DBExecutor
	cache repositories.ScoreRepository
}

func NewPostgresScoreRepository(db DBExecutor) repositories.ScoreRepository {
	return &PostgresScoreRepository{
		db:    db,
		cache: repositories.NewInMemoryScoreRepository(),
	}
}

func (r *PostgresScoreRepository) SaveScore(gameId string, betId string, points int) error {
	query := `
		WITH bet_match AS (
			SELECT b.game_id
			FROM bet b
			WHERE b.id = $1
		)
		INSERT INTO score (bet_id, points)
		SELECT $1, $2
		FROM bet_match
		WHERE game_id = $3
		ON CONFLICT (bet_id) DO UPDATE
		SET points = $2, updated_at = CURRENT_TIMESTAMP`

	_, err := r.db.Exec(query, betId, points, gameId)
	if err != nil {
		return fmt.Errorf("error saving score: %v", err)
	}

	return r.cache.SaveScore(gameId, betId, points)
}

func (r *PostgresScoreRepository) GetScore(gameId string, betId string) (int, error) {
	// Try cache first
	if points, err := r.cache.GetScore(gameId, betId); err == nil {
		return points, nil
	}

	query := `
		WITH bet_match AS (
			SELECT b.game_id
			FROM bet b
			WHERE b.id = $1
		)
		SELECT s.points
		FROM score s
		JOIN bet_match bm ON bm.game_id = $2
		WHERE s.bet_id = $1`

	var points int
	err := r.db.QueryRow(query, betId, gameId).Scan(&points)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("error getting score: %v", err)
	}

	// Update cache
	if err := r.cache.SaveScore(gameId, betId, points); err != nil {
		return points, fmt.Errorf("error updating cache: %v", err)
	}

	return points, nil
}

func (r *PostgresScoreRepository) GetScores(gameId string) (map[string]int, error) {
	query := `
		WITH game_bets AS (
			SELECT b.id as bet_id
			FROM bet b
			WHERE b.game_id = $1
		)
		SELECT gb.bet_id, COALESCE(s.points, 0) as points
		FROM game_bets gb
		LEFT JOIN score s ON gb.bet_id = s.bet_id`

	rows, err := r.db.Query(query, gameId)
	if err != nil {
		return nil, fmt.Errorf("error getting scores: %v", err)
	}
	defer rows.Close()

	scores := make(map[string]int)
	for rows.Next() {
		var betId string
		var points int
		err := rows.Scan(&betId, &points)
		if err != nil {
			return nil, fmt.Errorf("error scanning score: %v", err)
		}

		scores[betId] = points
		if err := r.cache.SaveScore(gameId, betId, points); err != nil {
			return nil, fmt.Errorf("error updating cache: %v", err)
		}
	}

	return scores, nil
}

func (r *PostgresScoreRepository) GetScoresByMatchAndPlayer(gameId string) (map[string]map[models.Player]int, error) {
	query := `
		SELECT m.id as match_id, p.name as player_name, s.points
		FROM score s
		JOIN bet b ON s.bet_id = b.id
		JOIN match m ON b.match_id = m.id
		JOIN player p ON b.player_id = p.id
		WHERE b.game_id = $1`

	rows, err := r.db.Query(query, gameId)
	if err != nil {
		return nil, fmt.Errorf("error getting scores: %v", err)
	}
	defer rows.Close()

	scores := make(map[string]map[models.Player]int)
	for rows.Next() {
		var matchId string
		var playerName string
		var points int
		err := rows.Scan(&matchId, &playerName, &points)
		if err != nil {
			return nil, fmt.Errorf("error scanning score: %v", err)
		}

		player := models.Player{Name: playerName}
		if _, ok := scores[matchId]; !ok {
			scores[matchId] = make(map[models.Player]int)
		}
		scores[matchId][player] = points
	}

	return scores, nil
}
