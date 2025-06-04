package postgres

import (
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
	rules "liguain/backend/rules"
)

type PostgresGameRepository struct {
	*PostgresRepository
	cache     repositories.GameRepository
	matchRepo repositories.MatchRepository
	scoreRepo repositories.ScoreRepository
}

func NewPostgresGameRepository(db *sql.DB) (repositories.GameRepository, error) {
	baseRepo := NewPostgresRepository(db)
	cache := repositories.NewInMemoryGameRepository()
	matchRepo := NewPostgresMatchRepository(db)
	scoreRepo := NewPostgresScoreRepository(db)
	return &PostgresGameRepository{
		PostgresRepository: baseRepo,
		cache:              cache,
		matchRepo:          matchRepo,
		scoreRepo:          scoreRepo,
	}, nil
}

func (r *PostgresGameRepository) CreateGame(game models.Game) (string, error) {
	query := `
		INSERT INTO game (season_year, competition_name, status)
		VALUES ($1, $2, $3)
		RETURNING id`

	var id string
	err := r.db.QueryRow(
		query,
		game.GetSeasonYear(),
		game.GetCompetitionName(),
		game.GetGameStatus(),
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("error saving game: %v", err)
	}

	// Update cache
	if err := r.cache.SaveWithId(id, game); err != nil {
		return "", fmt.Errorf("error saving game to cache: %v", err)
	}

	return id, nil
}

func (r *PostgresGameRepository) GetGame(gameId string) (models.Game, error) {
	if game, err := r.getGameFromCache(gameId); err == nil && game != nil {
		return game, nil
	}

	seasonYear, competitionName, err := r.getGameDetails(gameId)
	if err != nil {
		return nil, err
	}

	incomingMatches, pastMatches, bets, players, err := r.getMatchesAndBets(gameId)
	if err != nil {
		return nil, err
	}

	// Get scores organized by match and player
	playerScores, err := r.scoreRepo.GetScoresByMatchAndPlayer(gameId)
	if err != nil {
		return nil, fmt.Errorf("error getting scores: %v", err)
	}

	var gameImpl models.Game

	// Use NewFreshGame if no matches or bets exist, otherwise use NewStartedGame
	if len(incomingMatches) == 0 && len(pastMatches) == 0 && len(bets) == 0 {
		gameImpl = rules.NewFreshGame(
			seasonYear,
			competitionName,
			players,
			[]models.Match{}, // empty incoming matches
			&rules.ScorerOriginal{},
		)
	} else {
		gameImpl = rules.NewStartedGame(
			seasonYear,
			competitionName,
			players,
			pastMatches,
			incomingMatches,
			&rules.ScorerOriginal{},
			bets,
			playerScores,
		)
	}

	// Populate cache
	if err := r.saveGameToCache(gameId, gameImpl); err != nil {
		return nil, err
	}

	return gameImpl, nil
}

func (r *PostgresGameRepository) getGameFromCache(gameId string) (models.Game, error) {
	return r.cache.GetGame(gameId)
}

func (r *PostgresGameRepository) getGameDetails(gameId string) (string, string, error) {
	query := `
		SELECT season_year, competition_name
		FROM game
		WHERE id = $1`

	var seasonYear, competitionName string
	err := r.db.QueryRow(query, gameId).Scan(
		&seasonYear,
		&competitionName,
	)

	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("game %s not found", gameId)
	}
	if err != nil {
		return "", "", fmt.Errorf("error getting game: %v", err)
	}

	return seasonYear, competitionName, nil
}

func (r *PostgresGameRepository) getMatchesAndBets(gameId string) ([]models.Match, []models.Match, map[string]map[models.Player]*models.Bet, []models.Player, error) {
	query := `
		WITH match_data AS (
			SELECT DISTINCT
				m.id as match_id,
				m.home_team_id,
				m.away_team_id,
				m.home_team_score,
				m.away_team_score,
				m.match_date,
				m.match_status,
				m.season_code,
				m.competition_code,
				m.matchday,
				b.id as bet_id,
				b.predicted_home_goals,
				b.predicted_away_goals,
				p.name as player_name,
				s.points as score_points
			FROM bet b
			JOIN match m ON m.id = b.match_id
			LEFT JOIN player p ON b.player_id = p.id
			LEFT JOIN score s ON b.id = s.bet_id
			WHERE b.game_id = $1
		)
		SELECT * FROM match_data`

	rows, err := r.db.Query(query, gameId)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error getting matches and bets: %v", err)
	}
	defer rows.Close()

	return r.processMatchData(rows)
}

func (r *PostgresGameRepository) processMatchData(rows *sql.Rows) ([]models.Match, []models.Match, map[string]map[models.Player]*models.Bet, []models.Player, error) {
	matchesById := make(map[string]*models.SeasonMatch)
	bets := make(map[string]map[models.Player]*models.Bet)
	players := make(map[string]models.Player)
	scores := make(map[string]int)

	for rows.Next() {
		var matchId, homeTeamId, awayTeamId string
		var homeTeamScore, awayTeamScore sql.NullInt32
		var matchDate sql.NullTime
		var matchStatus string
		var seasonCode, competitionCode string
		var matchday int
		var betId sql.NullString
		var predictedHomeGoals, predictedAwayGoals sql.NullInt32
		var playerName sql.NullString
		var scorePoints sql.NullInt32

		err := rows.Scan(
			&matchId,
			&homeTeamId,
			&awayTeamId,
			&homeTeamScore,
			&awayTeamScore,
			&matchDate,
			&matchStatus,
			&seasonCode,
			&competitionCode,
			&matchday,
			&betId,
			&predictedHomeGoals,
			&predictedAwayGoals,
			&playerName,
			&scorePoints,
		)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error scanning match data: %v", err)
		}

		// Create or get match
		match, exists := matchesById[matchId]
		if !exists {
			match = CreateMatchFromDB(homeTeamId, awayTeamId, seasonCode, competitionCode, matchDate.Time, matchday, matchStatus, homeTeamScore, awayTeamScore)
			matchesById[matchId] = match
		}

		// Create bet if all required fields are present
		if betId.Valid && predictedHomeGoals.Valid && predictedAwayGoals.Valid && playerName.Valid {
			player := models.Player{Name: playerName.String}
			players[playerName.String] = player

			if _, ok := bets[matchId]; !ok {
				bets[matchId] = make(map[models.Player]*models.Bet)
			}
			bet := r.createBet(match, int(predictedHomeGoals.Int32), int(predictedAwayGoals.Int32))
			bets[matchId][player] = bet

			// Store score if present
			if scorePoints.Valid {
				scores[betId.String] = int(scorePoints.Int32)
			}
		}
	}

	// Separate matches into incoming and past
	var incomingMatches []models.Match
	var pastMatches []models.Match
	for _, match := range matchesById {
		if match.IsFinished() {
			pastMatches = append(pastMatches, match)
		} else {
			incomingMatches = append(incomingMatches, match)
		}
	}

	// Convert players map to slice
	playerSlice := make([]models.Player, 0, len(players))
	for _, player := range players {
		playerSlice = append(playerSlice, player)
	}

	return incomingMatches, pastMatches, bets, playerSlice, nil
}

func (r *PostgresGameRepository) createBet(match models.Match, predictedHomeGoals, predictedAwayGoals int) *models.Bet {
	return models.NewBet(match, predictedHomeGoals, predictedAwayGoals)
}

func (r *PostgresGameRepository) saveGameToCache(gameId string, game models.Game) error {
	return r.cache.SaveWithId(gameId, game)
}

func (r *PostgresGameRepository) SaveWithId(gameId string, game models.Game) error {
	query := `
		INSERT INTO game (id, season_year, competition_name, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			season_year = EXCLUDED.season_year,
			competition_name = EXCLUDED.competition_name,
			status = EXCLUDED.status`

	_, err := r.db.Exec(
		query,
		gameId,
		game.GetSeasonYear(),
		game.GetCompetitionName(),
		game.GetGameStatus(),
	)

	if err != nil {
		return fmt.Errorf("error saving game: %v", err)
	}

	// Update cache
	return r.cache.SaveWithId(gameId, game)
}
