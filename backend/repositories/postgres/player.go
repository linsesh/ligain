package postgres

import (
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
)

type PostgresPlayerRepository struct {
	*PostgresRepository
	cache repositories.PlayerRepository
}

func NewPostgresPlayerRepository(db *sql.DB) repositories.PlayerRepository {
	baseRepo := NewPostgresRepository(db)
	cache := repositories.NewInMemoryPlayerRepository()
	return &PostgresPlayerRepository{PostgresRepository: baseRepo, cache: cache}
}

func (r *PostgresPlayerRepository) SavePlayer(player models.Player) (string, error) {
	query := `
		WITH player_data AS (
			SELECT id, name
			FROM player
			WHERE name = $1
		)
		INSERT INTO player (name)
		SELECT $1
		WHERE NOT EXISTS (SELECT 1 FROM player_data)
		RETURNING id`

	var id string
	err := r.db.QueryRow(query, player.Name).Scan(&id)
	if err == sql.ErrNoRows {
		// If no rows were inserted, get the existing player's ID
		err = r.db.QueryRow("SELECT id FROM player WHERE name = $1", player.Name).Scan(&id)
		if err != nil {
			return "", fmt.Errorf("error getting existing player: %v", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("error saving player: %v", err)
	}

	// Update cache
	if _, err := r.cache.SavePlayer(player); err != nil {
		fmt.Printf("Warning: failed to update cache: %v\n", err)
	}

	return id, nil
}

func (r *PostgresPlayerRepository) GetPlayer(playerId string) (models.Player, error) {
	// Try cache first
	player, err := r.cache.GetPlayer(playerId)
	if err == nil && player.Name != "" {
		return player, nil
	}

	query := `
		SELECT id, name
		FROM player
		WHERE id = $1`

	var id, name string
	err = r.db.QueryRow(query, playerId).Scan(&id, &name)

	if err == sql.ErrNoRows {
		return models.Player{}, fmt.Errorf("player %s not found", playerId)
	}
	if err != nil {
		return models.Player{}, fmt.Errorf("error getting player: %v", err)
	}

	player = models.Player{Name: name}

	// Populate cache
	if _, err := r.cache.SavePlayer(player); err != nil {
		fmt.Printf("Warning: failed to update cache: %v\n", err)
	}

	return player, nil
}

func (r *PostgresPlayerRepository) GetPlayers(gameId string) ([]models.Player, error) {
	query := `
		WITH game_players AS (
			SELECT DISTINCT p.id, p.name
			FROM player p
			JOIN bet b ON p.id = b.player_id
			JOIN match m ON b.match_id = m.id
			WHERE m.game_id = $1
		)
		SELECT id, name
		FROM game_players
		ORDER BY name ASC`

	rows, err := r.db.Query(query, gameId)
	if err != nil {
		return nil, fmt.Errorf("error getting players: %v", err)
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("error scanning player: %v", err)
		}
		players = append(players, models.Player{Name: name})
	}

	return players, nil
}
