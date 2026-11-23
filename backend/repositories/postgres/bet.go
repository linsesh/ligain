package postgres

import (
	"database/sql"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"

	log "github.com/sirupsen/logrus"
)

type PostgresBetRepository struct {
	db DBExecutor
}

func NewPostgresBetRepository(db DBExecutor) repositories.BetRepository {
	return &PostgresBetRepository{
		db: db,
	}
}

func (r *PostgresBetRepository) SaveBet(gameId string, bet *models.Bet, player models.Player) (string, *models.Bet, error) {
	query := `
		WITH match_id AS (
			SELECT id FROM match
			WHERE local_id = $1
		)
		INSERT INTO bet (game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
		SELECT $2, m.id, $3, $4, $5
		FROM match_id m
		ON CONFLICT (game_id, match_id, player_id) DO UPDATE
		SET predicted_home_goals = $4, predicted_away_goals = $5, updated_at = NOW()
		RETURNING id`

	var id string
	err := r.db.QueryRow(
		query,
		bet.MatchID,
		gameId,
		player.GetID(),
		bet.PredictedHomeGoals,
		bet.PredictedAwayGoals,
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("match or player not found")
		}
		return "", nil, fmt.Errorf("error saving bet: %v", err)
	}

	return id, bet, nil
}

func (r *PostgresBetRepository) GetBets(gameId string, player models.Player) ([]*models.Bet, error) {
	// Query database to get bets with their match IDs
	query := `
		SELECT b.predicted_home_goals, b.predicted_away_goals, m.local_id
		FROM bet b
		JOIN match m ON b.match_id = m.id
		JOIN player p ON b.player_id = p.id
		WHERE b.game_id = $1 AND p.id = $2`

	rows, err := r.db.Query(query, gameId, player.GetID())
	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.Bet{}, nil
		}
		return nil, fmt.Errorf("error getting bets: %v", err)
	}
	defer rows.Close()

	var bets []*models.Bet
	for rows.Next() {
		var (
			predictedHomeGoals, predictedAwayGoals int
			matchID                                string
		)

		err := rows.Scan(&predictedHomeGoals, &predictedAwayGoals, &matchID)
		if err != nil {
			return nil, fmt.Errorf("error scanning bet row: %v", err)
		}

		// Create bet with just the match ID (no embedded match)
		bet := &models.Bet{
			MatchID:            matchID,
			PredictedHomeGoals: predictedHomeGoals,
			PredictedAwayGoals: predictedAwayGoals,
		}
		bets = append(bets, bet)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bet rows: %v", err)
	}

	return bets, nil
}

func (r *PostgresBetRepository) GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error) {
	// Query database with JOIN to get player names and IDs
	query := `
		SELECT p.id, p.name, b.predicted_home_goals, b.predicted_away_goals
		FROM bet b
		JOIN player p ON b.player_id = p.id
		JOIN match m ON b.match_id = m.id
		WHERE b.game_id = $1 AND m.local_id = $2`

	rows, err := r.db.Query(query, gameId, match.Id())
	if err != nil {
		if err == sql.ErrNoRows {
			return []*models.Bet{}, []models.Player{}, nil
		}
		return nil, nil, fmt.Errorf("error getting bets for match: %v", err)
	}
	defer rows.Close()

	var (
		bets    []*models.Bet
		players []models.Player
	)

	for rows.Next() {
		var (
			playerId                               string
			playerName                             string
			predictedHomeGoals, predictedAwayGoals int
		)

		err := rows.Scan(&playerId, &playerName, &predictedHomeGoals, &predictedAwayGoals)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning bet row: %v", err)
		}

		// Create bet with just the match ID
		bet := &models.Bet{
			MatchID:            match.Id(),
			PredictedHomeGoals: predictedHomeGoals,
			PredictedAwayGoals: predictedAwayGoals,
		}
		bets = append(bets, bet)

		// Create player with proper ID
		player := models.NewSimplePlayer(playerId, playerName)
		players = append(players, player)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating bet rows: %v", err)
	}

	return bets, players, nil
}

func (r *PostgresBetRepository) SaveWithId(gameId string, betId string, bet *models.Bet, player models.Player) error {
	// Save bet with specific ID
	_, err := r.db.Exec(`
		WITH match_id AS (
			SELECT id FROM match
			WHERE local_id = $1
		)
		INSERT INTO bet (id, match_id, player_id, predicted_home_goals, predicted_away_goals, game_id)
		SELECT $2, m.id, $3, $4, $5, $6
		FROM match_id m
		ON CONFLICT (id) DO UPDATE SET
			match_id = EXCLUDED.match_id,
			player_id = EXCLUDED.player_id,
			predicted_home_goals = EXCLUDED.predicted_home_goals,
			predicted_away_goals = EXCLUDED.predicted_away_goals,
			game_id = EXCLUDED.game_id,
			updated_at = NOW()
	`, bet.MatchID, betId, player.GetID(), bet.PredictedHomeGoals, bet.PredictedAwayGoals, gameId)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("match or player not found")
		}
		return fmt.Errorf("error saving bet: %w", err)
	}
	return nil
}

func (r *PostgresBetRepository) SaveScore(gameId string, match models.Match, player models.Player, points int) error {
	_, err := r.db.Exec(`
		WITH match_id AS (
			SELECT id FROM match 
			WHERE local_id = $1
		)
		INSERT INTO score (game_id, match_id, player_id, bet_id, points)
		SELECT $3, m.id, $2, b.id, $4
		FROM match_id m
		LEFT JOIN bet b ON b.match_id = m.id
		AND b.player_id = $2
		AND b.game_id = $3
		ON CONFLICT (game_id, match_id, player_id) DO UPDATE
		SET points = EXCLUDED.points, updated_at = NOW()`,
		match.Id(),
		player.GetID(),
		gameId,
		points)
	if err != nil {
		return fmt.Errorf("error saving score: %v", err)
	}

	return nil
}

func (r *PostgresBetRepository) GetScore(gameId string, betId string) (int, error) {
	var points sql.NullInt32
	err := r.db.QueryRow(`
		SELECT s.points
		FROM bet b
		LEFT JOIN score s ON s.bet_id = b.id
		WHERE b.game_id = $1 AND b.id = $2`,
		gameId, betId).Scan(&points)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Errorf("No score found for bet %s in game %s\n", betId, gameId)
			return 0, repositories.ErrScoreNotFound
		}
		return 0, fmt.Errorf("error getting score: %v", err)
	}
	if !points.Valid {
		return 0, repositories.ErrScoreNotFound
	}
	return int(points.Int32), nil
}

func (r *PostgresBetRepository) GetScores(gameId string) (map[string]map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT s.match_id, s.player_id, s.points
		FROM score s
		WHERE s.game_id = $1`,
		gameId)
	if err != nil {
		return nil, fmt.Errorf("error getting scores: %v", err)
	}
	defer rows.Close()

	scores := make(map[string]map[string]int)
	for rows.Next() {
		var matchId, playerId string
		var points int
		err := rows.Scan(&matchId, &playerId, &points)
		if err != nil {
			return nil, fmt.Errorf("error scanning score row: %v", err)
		}
		if _, ok := scores[matchId]; !ok {
			scores[matchId] = make(map[string]int)
		}
		scores[matchId][playerId] = points
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating score rows: %v", err)
	}
	return scores, nil
}

func (r *PostgresBetRepository) GetScoresByMatchAndPlayer(gameId string) (map[string]map[string]int, error) {
	rows, err := r.db.Query(`
		SELECT m.local_id, p.id, p.name, s.points
		FROM score s
		JOIN player p ON s.player_id = p.id
		JOIN match m ON s.match_id = m.id
		WHERE s.game_id = $1`,
		gameId)
	if err != nil {
		return nil, fmt.Errorf("error getting scores by match and player: %v", err)
	}
	defer rows.Close()

	scores := make(map[string]map[string]int)
	for rows.Next() {
		var matchId, playerId, playerName string
		var points int
		err := rows.Scan(&matchId, &playerId, &playerName, &points)
		if err != nil {
			return nil, fmt.Errorf("error scanning score row: %v", err)
		}
		if _, ok := scores[matchId]; !ok {
			scores[matchId] = make(map[string]int)
		}
		scores[matchId][playerId] = points
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating score rows: %v", err)
	}
	return scores, nil
}
