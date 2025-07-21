package postgres

import (
	"database/sql"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"

	"time"

	log "github.com/sirupsen/logrus"
)

type PostgresBetRepository struct {
	db    DBExecutor
	cache repositories.BetRepository
}

func NewPostgresBetRepository(db DBExecutor, cache repositories.BetRepository) repositories.BetRepository {
	if cache == nil {
		cache = repositories.NewInMemoryBetRepository()
	}
	return &PostgresBetRepository{
		db:    db,
		cache: cache,
	}
}

func (r *PostgresBetRepository) SaveBet(gameId string, bet *models.Bet, player models.Player) (string, *models.Bet, error) {
	query := `
		WITH match_id AS (
			SELECT id FROM match 
			WHERE home_team_id = $1 
			AND away_team_id = $2 
			AND season_code = $3 
			AND competition_code = $4 
			AND matchday = $5
		)
		INSERT INTO bet (game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
		SELECT $6, m.id, p.id, $7, $8
		FROM match_id m
		JOIN player p ON p.name = $9
		ON CONFLICT (match_id, player_id) DO UPDATE 
		SET predicted_home_goals = $7, predicted_away_goals = $8
		RETURNING id`

	var id string
	err := r.db.QueryRow(
		query,
		bet.Match.GetHomeTeam(),
		bet.Match.GetAwayTeam(),
		bet.Match.GetSeasonCode(),
		bet.Match.GetCompetitionCode(),
		bet.Match.(*models.SeasonMatch).Matchday,
		gameId,
		bet.PredictedHomeGoals,
		bet.PredictedAwayGoals,
		player.GetName(),
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, fmt.Errorf("player not found: %s", player.GetName())
		}
		return "", nil, fmt.Errorf("error saving bet: %v", err)
	}

	// Update cache
	_, _, err = r.cache.SaveBet(gameId, bet, player)
	if err != nil {
		return "", nil, fmt.Errorf("error saving bet to cache: %v", err)
	}

	return id, bet, nil
}

// updateBetsCache updates the cache with the given bets and players
func (r *PostgresBetRepository) updateBetsCache(gameId string, bets []*models.Bet, players []models.Player) error {
	for i, bet := range bets {
		_, _, err := r.cache.SaveBet(gameId, bet, players[i])
		if err != nil {
			return fmt.Errorf("error saving bet to cache: %v", err)
		}
	}
	return nil
}

func (r *PostgresBetRepository) GetBets(gameId string, player models.Player) ([]*models.Bet, error) {
	// Try cache first
	if bets, err := r.cache.GetBets(gameId, player); err == nil {
		return bets, nil
	}

	// Query database with JOIN to player table
	query := `
		SELECT b.id, b.match_id, b.predicted_home_goals, b.predicted_away_goals,
			   m.home_team_id, m.away_team_id, m.match_date, m.match_status,
			   m.season_code, m.competition_code, m.matchday
		FROM bet b
		JOIN match m ON b.match_id = m.id
		JOIN player p ON b.player_id = p.id
		WHERE b.game_id = $1 AND p.name = $2`

	rows, err := r.db.Query(query, gameId, player.GetName())
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
			id                                     string
			matchId                                string
			predictedHomeGoals, predictedAwayGoals int
			homeTeam, awayTeam                     string
			matchDateStr                           string
			matchStatus                            string
			seasonCode, competitionCode            string
			matchday                               int
		)

		err := rows.Scan(
			&id, &matchId, &predictedHomeGoals, &predictedAwayGoals,
			&homeTeam, &awayTeam, &matchDateStr, &matchStatus,
			&seasonCode, &competitionCode, &matchday,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning bet row: %v", err)
		}

		// Parse match date
		matchDate, err := time.Parse(time.RFC3339, matchDateStr)
		if err != nil {
			return nil, fmt.Errorf("error parsing match date: %v", err)
		}

		// Create match
		match := models.NewSeasonMatch(homeTeam, awayTeam, seasonCode, competitionCode, matchDate, matchday)
		if matchStatus == "finished" {
			match.Finish(predictedHomeGoals, predictedAwayGoals)
		}

		// Create bet
		bet := models.NewBet(match, predictedHomeGoals, predictedAwayGoals)
		bets = append(bets, bet)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bet rows: %v", err)
	}

	// Update cache
	if err := r.updateBetsCache(gameId, bets, []models.Player{player}); err != nil {
		return nil, err
	}

	return bets, nil
}

func (r *PostgresBetRepository) GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error) {
	// Try cache first
	if bets, players, err := r.cache.GetBetsForMatch(match, gameId); err == nil {
		return bets, players, nil
	}

	// Query database with JOIN to get player names and IDs
	query := `
		SELECT b.id, p.id, p.name, b.predicted_home_goals, b.predicted_away_goals
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
			id                                     string
			playerId                               string
			playerName                             string
			predictedHomeGoals, predictedAwayGoals int
		)

		err := rows.Scan(&id, &playerId, &playerName, &predictedHomeGoals, &predictedAwayGoals)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning bet row: %v", err)
		}

		// Create bet
		bet := models.NewBet(match, predictedHomeGoals, predictedAwayGoals)
		bets = append(bets, bet)

		// Create player with proper ID
		player := models.NewSimplePlayer(playerId, playerName)
		players = append(players, player)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating bet rows: %v", err)
	}

	// Update cache
	if err := r.updateBetsCache(gameId, bets, players); err != nil {
		return nil, nil, err
	}

	return bets, players, nil
}

func (r *PostgresBetRepository) SaveWithId(gameId string, betId string, bet *models.Bet, player models.Player) error {
	// Save bet with specific ID
	_, err := r.db.Exec(`
		WITH player_id AS (
			SELECT id FROM player WHERE name = $1
		),
		match_id AS (
			SELECT id FROM match 
			WHERE home_team_id = $2 
			AND away_team_id = $3 
			AND season_code = $4 
			AND competition_code = $5 
			AND matchday = $6
		)
		INSERT INTO bet (id, match_id, player_id, predicted_home_goals, predicted_away_goals, game_id)
		SELECT $7, m.id, p.id, $8, $9, $10
		FROM player_id p, match_id m
		ON CONFLICT (id) DO UPDATE SET
			match_id = EXCLUDED.match_id,
			player_id = EXCLUDED.player_id,
			predicted_home_goals = EXCLUDED.predicted_home_goals,
			predicted_away_goals = EXCLUDED.predicted_away_goals,
			game_id = EXCLUDED.game_id
	`, player.GetName(), bet.Match.GetHomeTeam(), bet.Match.GetAwayTeam(), bet.Match.GetSeasonCode(), bet.Match.GetCompetitionCode(), bet.Match.(*models.SeasonMatch).Matchday, betId, bet.PredictedHomeGoals, bet.PredictedAwayGoals, gameId)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("player not found: %s", player.GetName())
		}
		return fmt.Errorf("error saving bet: %w", err)
	}

	// Update cache
	return r.cache.SaveWithId(gameId, betId, bet, player)
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
		SET points = EXCLUDED.points`,
		match.Id(),
		player.GetID(),
		gameId,
		points)
	if err != nil {
		return fmt.Errorf("error saving score: %v", err)
	}

	// Update cache
	return r.cache.SaveScore(gameId, match, player, points)
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
