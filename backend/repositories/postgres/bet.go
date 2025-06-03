package postgres

import (
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
)

type PostgresBetRepository struct {
	*PostgresRepository
	cache     repositories.BetRepository
	matchRepo repositories.MatchRepository
	// betIds maps bet key (gameId:player:matchId) to bet ID
	betIds map[string]string
}

func NewPostgresBetRepository(db *sql.DB, cache repositories.BetRepository) repositories.BetRepository {
	matchRepo := NewPostgresMatchRepository(db)
	return &PostgresBetRepository{
		PostgresRepository: &PostgresRepository{db: db},
		cache:              repositories.NewInMemoryBetRepository(),
		matchRepo:          matchRepo,
		betIds:             make(map[string]string),
	}
}

func (r *PostgresBetRepository) GetBetId(gameId string, player models.Player, matchId string) string {
	key := fmt.Sprintf("%s:%s:%s", gameId, player.Name, matchId)
	return r.betIds[key]
}

func (r *PostgresBetRepository) GetBets(gameId string, player models.Player) ([]*models.Bet, error) {
	// Try to get from cache first
	if bets, err := r.cache.GetBets(gameId, player); err == nil && len(bets) > 0 {
		return bets, nil
	}

	// If not in cache, get from database
	rows, err := r.db.Query(`
		SELECT b.id, b.match_id, b.predicted_home_goals, b.predicted_away_goals,
			   m.home_team_score, m.away_team_score, m.match_date, m.match_status,
			   m.season_code, m.competition_code, m.matchday, p.name
		FROM bet b
		JOIN match m ON b.match_id = m.id
		JOIN player p ON b.player_id = p.id
		WHERE m.game_id = $1
		ORDER BY m.match_date DESC
	`, gameId)
	if err != nil {
		return nil, fmt.Errorf("error querying bets: %w", err)
	}
	defer rows.Close()

	var bets []*models.Bet
	for rows.Next() {
		var betId, matchId string
		var predictedHomeGoals, predictedAwayGoals int
		var homeTeamScore, awayTeamScore sql.NullInt32
		var matchDate sql.NullTime
		var matchStatus string
		var seasonCode, competitionCode string
		var matchday int
		var playerName string

		err := rows.Scan(
			&betId, &matchId, &predictedHomeGoals, &predictedAwayGoals,
			&homeTeamScore, &awayTeamScore, &matchDate, &matchStatus,
			&seasonCode, &competitionCode, &matchday, &playerName,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning bet: %w", err)
		}

		// Get match from repository
		match, err := r.matchRepo.GetMatch(matchId)
		if err != nil {
			return nil, fmt.Errorf("error getting match: %w", err)
		}

		// Create bet
		bet := models.NewBet(match, predictedHomeGoals, predictedAwayGoals)
		betPlayer := models.Player{Name: playerName}

		// Cache the bet with its PostgreSQL ID
		if err := r.cache.SaveWithId(gameId, betId, bet, betPlayer); err != nil {
			return nil, fmt.Errorf("error caching bet: %w", err)
		}

		// Only return bets for the requested player
		if betPlayer.Name == player.Name {
			bets = append(bets, bet)
		}
	}

	return bets, nil
}

func (r *PostgresBetRepository) SaveBet(gameId string, bet *models.Bet, player models.Player) (string, error) {
	// Save match first
	matchId, err := r.matchRepo.SaveMatch(bet.Match)
	if err != nil {
		return "", fmt.Errorf("error saving match: %w", err)
	}

	// Save bet and get player ID in a single query
	var betId string
	err = r.db.QueryRow(`
		WITH player_id AS (
			SELECT id FROM player WHERE name = $1
		)
		INSERT INTO bet (match_id, player_id, predicted_home_goals, predicted_away_goals)
		SELECT $2, id, $3, $4
		FROM player_id
		ON CONFLICT (match_id, player_id) DO UPDATE SET
			predicted_home_goals = EXCLUDED.predicted_home_goals,
			predicted_away_goals = EXCLUDED.predicted_away_goals
		RETURNING id
	`, player.Name, matchId, bet.PredictedHomeGoals, bet.PredictedAwayGoals).Scan(&betId)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("player not found: %s", player.Name)
		}
		return "", fmt.Errorf("error saving bet: %w", err)
	}

	// Update cache with the PostgreSQL-generated ID
	if err := r.cache.SaveWithId(gameId, betId, bet, player); err != nil {
		return betId, fmt.Errorf("error updating cache: %w", err)
	}

	return betId, nil
}

func (r *PostgresBetRepository) GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error) {
	// Try to get from cache first
	if bets, players, err := r.cache.GetBetsForMatch(match, gameId); err == nil && len(bets) > 0 {
		return bets, players, nil
	}

	// If not in cache, get from database
	rows, err := r.db.Query(`
		SELECT b.id, b.predicted_home_goals, b.predicted_away_goals, p.name
		FROM bet b
		JOIN player p ON b.player_id = p.id
		WHERE b.match_id = $1
	`, match.Id())
	if err != nil {
		return nil, nil, fmt.Errorf("error querying bets for match: %w", err)
	}
	defer rows.Close()

	bets := make([]*models.Bet, 0)
	players := make([]models.Player, 0)
	for rows.Next() {
		var betId string
		var predictedHomeGoals, predictedAwayGoals int
		var playerName string

		err := rows.Scan(
			&betId,
			&predictedHomeGoals,
			&predictedAwayGoals,
			&playerName,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("error scanning bet: %w", err)
		}

		bet := models.NewBet(match, predictedHomeGoals, predictedAwayGoals)
		player := models.Player{Name: playerName}

		// Cache the bet with its PostgreSQL ID
		if err := r.cache.SaveWithId(gameId, betId, bet, player); err != nil {
			return nil, nil, fmt.Errorf("error caching bet: %w", err)
		}

		bets = append(bets, bet)
		players = append(players, player)
	}

	return bets, players, nil
}

func (r *PostgresBetRepository) SaveWithId(gameId string, betId string, bet *models.Bet, player models.Player) error {
	// Save match first
	matchId, err := r.matchRepo.SaveMatch(bet.Match)
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
