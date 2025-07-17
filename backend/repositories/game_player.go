package repositories

import (
	"context"
	"liguain/backend/models"
)

// GamePlayerRepository defines the interface for managing game-player relationships
type GamePlayerRepository interface {
	// AddPlayerToGame adds a player to a game
	AddPlayerToGame(ctx context.Context, gameID string, playerID string) error
	// RemovePlayerFromGame removes a player from a game
	RemovePlayerFromGame(ctx context.Context, gameID string, playerID string) error
	// GetPlayersInGame returns all players in a specific game
	GetPlayersInGame(ctx context.Context, gameID string) ([]models.Player, error)
	// GetPlayerGames returns all games that a player is part of
	GetPlayerGames(ctx context.Context, playerID string) ([]string, error)
	// IsPlayerInGame checks if a player is in a specific game
	IsPlayerInGame(ctx context.Context, gameID string, playerID string) (bool, error)
}
