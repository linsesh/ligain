package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
)

type PostgresGamePlayerRepository struct {
	db *sql.DB
}

func NewPostgresGamePlayerRepository(db *sql.DB) repositories.GamePlayerRepository {
	return &PostgresGamePlayerRepository{db: db}
}

func (r *PostgresGamePlayerRepository) AddPlayerToGame(ctx context.Context, gameID string, playerID string) error {
	query := `
		INSERT INTO game_player (game_id, player_id)
		VALUES ($1, $2)
		ON CONFLICT (game_id, player_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, gameID, playerID)
	if err != nil {
		return fmt.Errorf("error adding player to game: %v", err)
	}

	return nil
}

func (r *PostgresGamePlayerRepository) RemovePlayerFromGame(ctx context.Context, gameID string, playerID string) error {
	query := `
		DELETE FROM game_player
		WHERE game_id = $1 AND player_id = $2
	`

	_, err := r.db.ExecContext(ctx, query, gameID, playerID)
	if err != nil {
		return fmt.Errorf("error removing player from game: %v", err)
	}

	return nil
}

func (r *PostgresGamePlayerRepository) GetPlayersInGame(ctx context.Context, gameID string) ([]models.Player, error) {
	query := `
		SELECT p.id, p.name
		FROM game_player gp
		JOIN player p ON gp.player_id = p.id
		WHERE gp.game_id = $1
		ORDER BY p.name
	`

	rows, err := r.db.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, fmt.Errorf("error getting players in game: %v", err)
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.PlayerData
		err := rows.Scan(&player.ID, &player.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning player row: %v", err)
		}
		players = append(players, &player)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating player rows: %v", err)
	}

	return players, nil
}

func (r *PostgresGamePlayerRepository) GetPlayerGames(ctx context.Context, playerID string) ([]string, error) {
	query := `
		SELECT game_id
		FROM game_player
		WHERE player_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, fmt.Errorf("error getting player games: %v", err)
	}
	defer rows.Close()

	var gameIDs []string
	for rows.Next() {
		var gameID string
		err := rows.Scan(&gameID)
		if err != nil {
			return nil, fmt.Errorf("error scanning game ID row: %v", err)
		}
		gameIDs = append(gameIDs, gameID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating game ID rows: %v", err)
	}

	return gameIDs, nil
}

func (r *PostgresGamePlayerRepository) IsPlayerInGame(ctx context.Context, gameID string, playerID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM game_player
			WHERE game_id = $1 AND player_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, gameID, playerID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if player is in game: %v", err)
	}

	return exists, nil
}
