package postgres

import (
	"database/sql"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"
	rules "ligain/backend/rules"

	log "github.com/sirupsen/logrus"
)

type PostgresGameRepository struct {
	*PostgresRepository
	matchRepo repositories.MatchRepository
	betRepo   repositories.BetRepository
}

func NewPostgresGameRepository(db *sql.DB) (repositories.GameRepository, error) {
	baseRepo := NewPostgresRepository(db)
	betCache := repositories.NewInMemoryBetRepository()
	matchRepo := NewPostgresMatchRepository(db)
	betRepo := NewPostgresBetRepository(db, betCache)
	return &PostgresGameRepository{
		PostgresRepository: baseRepo,
		matchRepo:          matchRepo,
		betRepo:            betRepo,
	}, nil
}

func (r *PostgresGameRepository) CreateGame(game models.Game) (string, error) {
	query := `
		INSERT INTO game (season_year, competition_name, status, game_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var id string
	err := r.db.QueryRow(
		query,
		game.GetSeasonYear(),
		game.GetCompetitionName(),
		game.GetGameStatus(),
		game.GetName(),
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("error saving game: %v", err)
	}

	return id, nil
}

func (r *PostgresGameRepository) GetGame(gameId string) (models.Game, error) {
	seasonYear, competitionName, name, err := r.getGameDetails(gameId)
	if err != nil {
		log.Errorf("error getting game details: %v", err)
		return nil, err
	}

	incomingMatches, pastMatches, bets, players, err := r.getMatchesAndBets(gameId)
	if err != nil {
		log.Errorf("error getting matches and bets: %v", err)
		return nil, err
	}

	// Get all players in the game (not just those who have made bets)
	gamePlayers, err := r.getGamePlayers(gameId)
	if err != nil {
		log.Errorf("error getting game players: %v", err)
		return nil, fmt.Errorf("error getting game players: %v", err)
	}

	// Merge players from bets and game_player table
	allPlayers := make(map[string]models.Player)
	for _, player := range players {
		allPlayers[player.GetID()] = player
	}
	for _, player := range gamePlayers {
		allPlayers[player.GetID()] = player
	}

	// Convert to slice
	playerSlice := make([]models.Player, 0, len(allPlayers))
	for _, player := range allPlayers {
		playerSlice = append(playerSlice, player)
	}

	// Get scores organized by match and player
	playerScores, err := r.betRepo.GetScoresByMatchAndPlayer(gameId)
	if err != nil {
		log.Errorf("error getting scores: %v", err)
		return nil, fmt.Errorf("error getting scores: %v", err)
	}

	var gameImpl models.Game

	// Use NewFreshGame if no matches or bets exist, otherwise use NewStartedGame
	if len(incomingMatches) == 0 && len(pastMatches) == 0 && len(bets) == 0 {
		gameImpl = rules.NewFreshGame(
			seasonYear,
			competitionName,
			name,
			playerSlice,
			[]models.Match{}, // empty incoming matches
			&rules.ScorerOriginal{},
		)
	} else {
		gameImpl = rules.NewStartedGame(
			seasonYear,
			competitionName,
			name,
			playerSlice,
			incomingMatches,
			pastMatches,
			&rules.ScorerOriginal{},
			bets,
			playerScores,
		)
	}

	return gameImpl, nil
}

func (r *PostgresGameRepository) GetAllGames() (map[string]models.Game, error) {
	query := `
		SELECT id, season_year, competition_name, status, game_name
		FROM game
		WHERE status != 'finished'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error getting all games: %v", err)
	}
	defer rows.Close()

	games := make(map[string]models.Game)
	for rows.Next() {
		var id, seasonYear, competitionName, status, name string
		err := rows.Scan(&id, &seasonYear, &competitionName, &status, &name)
		if err != nil {
			log.Errorf("error scanning game row: %v", err)
			continue
		}

		// Get the game using the existing GetGame method
		game, err := r.GetGame(id)
		if err != nil {
			log.Errorf("error getting game %s: %v", id, err)
			continue
		}

		games[id] = game
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating game rows: %v", err)
	}

	return games, nil
}

func (r *PostgresGameRepository) getGameDetails(gameId string) (string, string, string, error) {
	query := `
		SELECT g.season_year, g.competition_name, g.game_name
		FROM game g
		WHERE g.id = $1::uuid`

	var seasonYear, competitionName, name string
	err := r.db.QueryRow(query, gameId).Scan(
		&seasonYear,
		&competitionName,
		&name,
	)

	if err == sql.ErrNoRows {
		log.Errorf("The postgres query returned no rows for game %s", gameId)
		return "", "", "", fmt.Errorf("game %s not found", gameId)
	}
	if err != nil {
		return "", "", "", fmt.Errorf("error getting game: %v", err)
	}

	return seasonYear, competitionName, name, nil
}

func (r *PostgresGameRepository) getMatchesAndBets(gameId string) ([]models.Match, []models.Match, map[string]map[string]*models.Bet, []models.Player, error) {
	query := `
		WITH match_data AS (
			SELECT
				m.local_id as match_id,
				m.home_team_id,
				m.away_team_id,
				m.home_team_score,
				m.away_team_score,
				m.match_date,
				m.home_win_odds,
				m.away_win_odds,
				m.draw_odds,
				m.match_status,
				m.season_code,
				m.competition_code,
				m.matchday,
				b.id as bet_id,
				b.predicted_home_goals,
				b.predicted_away_goals,
				p.id as player_id,
				p.name as player_name
			FROM match m
			LEFT JOIN bet b ON m.id = b.match_id AND b.game_id = $1
			LEFT JOIN player p ON b.player_id = p.id
			LEFT JOIN score s ON b.id = s.bet_id
			WHERE m.season_code = (SELECT season_year FROM game WHERE id = $1::uuid)
			AND m.competition_code = (SELECT competition_name FROM game WHERE id = $1::uuid)
		)
		SELECT * FROM match_data`

	rows, err := r.db.Query(query, gameId)
	if err != nil {
		log.Errorf("error getting matches and bets: %v", err)
		return nil, nil, nil, nil, fmt.Errorf("error getting matches and bets: %v", err)
	}
	defer rows.Close()

	return r.processMatchData(rows)
}

func (r *PostgresGameRepository) processMatchData(rows *sql.Rows) ([]models.Match, []models.Match, map[string]map[string]*models.Bet, []models.Player, error) {
	matchesById := make(map[string]*models.SeasonMatch)
	match_id_to_player_id_to_bet := make(map[string]map[string]*models.Bet)
	players := make(map[string]models.Player)

	for rows.Next() {
		var matchId, homeTeamId, awayTeamId string
		var homeTeamScore, awayTeamScore sql.NullInt32
		var matchDate sql.NullTime
		var homeWinOdds, awayWinOdds, drawOdds float64
		var matchStatus string
		var seasonCode, competitionCode string
		var matchday int
		var betId sql.NullString
		var predictedHomeGoals, predictedAwayGoals sql.NullInt32
		var playerId sql.NullString
		var playerName sql.NullString

		err := rows.Scan(
			&matchId,
			&homeTeamId,
			&awayTeamId,
			&homeTeamScore,
			&awayTeamScore,
			&matchDate,
			&homeWinOdds,
			&awayWinOdds,
			&drawOdds,
			&matchStatus,
			&seasonCode,
			&competitionCode,
			&matchday,
			&betId,
			&predictedHomeGoals,
			&predictedAwayGoals,
			&playerId,
			&playerName,
		)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("error scanning match data: %v", err)
		}

		match, exists := matchesById[matchId]
		if !exists {
			match = CreateMatchFromDB(homeTeamId, awayTeamId, seasonCode, competitionCode, matchDate.Time, matchday, matchStatus, homeTeamScore, awayTeamScore, homeWinOdds, awayWinOdds, drawOdds)
			matchesById[matchId] = match
		}
		// Create bet if all required fields are present
		if betId.Valid && predictedHomeGoals.Valid && predictedAwayGoals.Valid && playerName.Valid {
			playerID := ""
			if playerId.Valid {
				playerID = playerId.String
			}
			player := models.NewSimplePlayer(playerID, playerName.String)
			players[playerName.String] = player

			if _, ok := match_id_to_player_id_to_bet[matchId]; !ok {
				match_id_to_player_id_to_bet[matchId] = make(map[string]*models.Bet)
			}
			bet := r.createBet(match, int(predictedHomeGoals.Int32), int(predictedAwayGoals.Int32))
			match_id_to_player_id_to_bet[matchId][playerID] = bet

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

	return incomingMatches, pastMatches, match_id_to_player_id_to_bet, playerSlice, nil
}

func (r *PostgresGameRepository) createBet(match models.Match, predictedHomeGoals, predictedAwayGoals int) *models.Bet {
	return models.NewBet(match, predictedHomeGoals, predictedAwayGoals)
}

func (r *PostgresGameRepository) SaveWithId(gameId string, game models.Game) error {
	query := `
		INSERT INTO game (id, season_year, competition_name, status, game_name)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			season_year = EXCLUDED.season_year,
			competition_name = EXCLUDED.competition_name,
			status = EXCLUDED.status,
			game_name = EXCLUDED.game_name,
			updated_at = NOW()`

	_, err := r.db.Exec(
		query,
		gameId,
		game.GetSeasonYear(),
		game.GetCompetitionName(),
		game.GetGameStatus(),
		game.GetName(),
	)

	if err != nil {
		return fmt.Errorf("error saving game: %v", err)
	}

	return nil
}

func (r *PostgresGameRepository) getGamePlayers(gameId string) ([]models.Player, error) {
	query := `
		SELECT p.id, p.name
		FROM game_player gp
		JOIN player p ON gp.player_id = p.id
		WHERE gp.game_id = $1::uuid
		ORDER BY p.name
	`

	rows, err := r.db.Query(query, gameId)
	if err != nil {
		return nil, fmt.Errorf("error getting game players: %v", err)
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.PlayerData
		err := rows.Scan(&player.ID, &player.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning game player row: %v", err)
		}
		players = append(players, &player)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating game player rows: %v", err)
	}

	return players, nil
}
