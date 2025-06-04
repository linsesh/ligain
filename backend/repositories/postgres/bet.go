package postgres

import (
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"time"
)

type PostgresBetRepository struct {
	db    DBExecutor
	cache repositories.BetRepository
	// betIds maps bet key (gameId:player:matchId) to bet ID
	betIds map[string]string
}

func NewPostgresBetRepository(db DBExecutor, cache repositories.BetRepository) repositories.BetRepository {
	if cache == nil {
		cache = repositories.NewInMemoryBetRepository()
	}
	return &PostgresBetRepository{
		db:     db,
		cache:  cache,
		betIds: make(map[string]string),
	}
}

func (r *PostgresBetRepository) GetBetId(gameId string, player models.Player, matchId string) string {
	key := fmt.Sprintf("%s:%s:%s", gameId, player.Name, matchId)
	return r.betIds[key]
}

func (r *PostgresBetRepository) SaveBet(gameId string, bet *models.Bet, player models.Player) (string, error) {
	query := `
		INSERT INTO bet (game_id, match_id, player_id, predicted_home_goals, predicted_away_goals)
		SELECT $1, $2, p.id, $3, $4
		FROM player p 
		WHERE p.name = $5
		ON CONFLICT (match_id, player_id) DO UPDATE 
		SET predicted_home_goals = $3, predicted_away_goals = $4, game_id = $1
		RETURNING id`

	var id string
	err := r.db.QueryRow(
		query,
		gameId,
		bet.Match.Id(),
		bet.PredictedHomeGoals,
		bet.PredictedAwayGoals,
		player.Name,
	).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("player not found: %s", player.Name)
		}
		return "", fmt.Errorf("error saving bet: %v", err)
	}

	// Update cache
	_, err = r.cache.SaveBet(gameId, bet, player)
	if err != nil {
		return "", fmt.Errorf("error saving bet to cache: %v", err)
	}

	// Update betIds map
	key := fmt.Sprintf("%s:%s:%s", gameId, player.Name, bet.Match.Id())
	r.betIds[key] = id

	return id, nil
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

	rows, err := r.db.Query(query, gameId, player.Name)
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

		// Update betIds map
		key := fmt.Sprintf("%s:%s:%s", gameId, player.Name, matchId)
		r.betIds[key] = id
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bet rows: %v", err)
	}

	// Update cache
	for _, bet := range bets {
		_, err = r.cache.SaveBet(gameId, bet, player)
		if err != nil {
			return nil, fmt.Errorf("error saving bet to cache: %v", err)
		}
	}

	return bets, nil
}

func (r *PostgresBetRepository) GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error) {
	// Try cache first
	if bets, players, err := r.cache.GetBetsForMatch(match, gameId); err == nil {
		return bets, players, nil
	}

	// Query database with JOIN to get player names
	query := `
		SELECT b.id, p.name, b.predicted_home_goals, b.predicted_away_goals
		FROM bet b
		JOIN player p ON b.player_id = p.id
		WHERE b.game_id = $1 AND b.match_id = $2`

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
			playerName                             string
			predictedHomeGoals, predictedAwayGoals int
		)

		err := rows.Scan(&id, &playerName, &predictedHomeGoals, &predictedAwayGoals)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning bet row: %v", err)
		}

		// Create bet
		bet := models.NewBet(match, predictedHomeGoals, predictedAwayGoals)
		bets = append(bets, bet)

		// Create player
		player := models.Player{Name: playerName}
		players = append(players, player)

		// Update betIds map
		key := fmt.Sprintf("%s:%s:%s", gameId, playerName, match.Id())
		r.betIds[key] = id
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating bet rows: %v", err)
	}

	// Update cache
	for i, bet := range bets {
		_, err = r.cache.SaveBet(gameId, bet, players[i])
		if err != nil {
			return nil, nil, fmt.Errorf("error saving bet to cache: %v", err)
		}
	}

	return bets, players, nil
}

func (r *PostgresBetRepository) SaveWithId(gameId string, betId string, bet *models.Bet, player models.Player) error {
	// Save match first using direct SQL
	seasonMatch := bet.Match.(*models.SeasonMatch)

	var matchId string
	err := r.db.QueryRow(`
		INSERT INTO match (home_team_id, away_team_id, home_team_score, away_team_score,
						 match_date, match_status, season_code, competition_code, matchday)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (home_team_id, away_team_id, season_code, competition_code, matchday) DO UPDATE SET
			home_team_score = EXCLUDED.home_team_score,
			away_team_score = EXCLUDED.away_team_score,
			match_status = EXCLUDED.match_status
		RETURNING id
	`, bet.Match.GetHomeTeam(), bet.Match.GetAwayTeam(), bet.Match.GetHomeGoals(),
		bet.Match.GetAwayGoals(), bet.Match.GetDate(), bet.Match.IsFinished(),
		bet.Match.GetSeasonCode(), bet.Match.GetCompetitionCode(), seasonMatch.Matchday).Scan(&matchId)

	if err != nil {
		return fmt.Errorf("error saving match: %w", err)
	}

	// Save bet with specific ID
	_, err = r.db.Exec(`
		WITH player_id AS (
			SELECT id FROM player WHERE name = $1
		)
		INSERT INTO bet (id, match_id, player_id, predicted_home_goals, predicted_away_goals)
		SELECT $2, $3, id, $4, $5
		FROM player_id
		ON CONFLICT (id) DO UPDATE SET
			match_id = EXCLUDED.match_id,
			player_id = EXCLUDED.player_id,
			predicted_home_goals = EXCLUDED.predicted_home_goals,
			predicted_away_goals = EXCLUDED.predicted_away_goals
	`, player.Name, betId, matchId, bet.PredictedHomeGoals, bet.PredictedAwayGoals)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("player not found: %s", player.Name)
		}
		return fmt.Errorf("error saving bet: %w", err)
	}

	// Update cache
	return r.cache.SaveWithId(gameId, betId, bet, player)
}
