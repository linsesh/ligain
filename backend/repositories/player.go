package repositories

import (
	"context"
	"fmt"
	"liguain/backend/models"

	"github.com/google/uuid"
)

const playerCacheSize = 1000 // Maximum number of players to keep in cache

type PlayerRepository interface {
	GetPlayer(playerId string) (models.Player, error)
	// GetPlayers returns all players who have made bets in a game
	GetPlayers(gameId string) ([]models.Player, error)

	// Authentication methods
	CreatePlayer(ctx context.Context, player *models.PlayerData) error
	GetPlayerByID(ctx context.Context, id string) (*models.PlayerData, error)
	GetPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error)
	GetPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error)
	GetPlayerByName(ctx context.Context, name string) (*models.PlayerData, error)
	UpdatePlayer(ctx context.Context, player *models.PlayerData) error
	CreateAuthToken(ctx context.Context, token *models.AuthToken) error
	GetAuthToken(ctx context.Context, token string) (*models.AuthToken, error)
	DeleteAuthToken(ctx context.Context, token string) error
	DeleteExpiredTokens(ctx context.Context) error
}

// InMemoryPlayerRepository is a simple in-memory implementation of PlayerRepository
type InMemoryPlayerRepository struct {
	players map[string]models.Player
}

func NewInMemoryPlayerRepository() *InMemoryPlayerRepository {
	return &InMemoryPlayerRepository{
		players: make(map[string]models.Player),
	}
}

func (r *InMemoryPlayerRepository) GetPlayer(playerId string) (models.Player, error) {
	player, exists := r.players[playerId]
	if !exists {
		return &models.PlayerData{}, nil
	}
	return player, nil
}

func (r *InMemoryPlayerRepository) GetPlayers(gameId string) ([]models.Player, error) {
	// For in-memory, return all players
	var players []models.Player
	for _, player := range r.players {
		players = append(players, player)
	}
	return players, nil
}

// Authentication methods for InMemoryPlayerRepository
func (r *InMemoryPlayerRepository) CreatePlayer(ctx context.Context, player *models.PlayerData) error {
	player.ID = uuid.New().String()
	r.players[player.ID] = player
	return nil
}

func (r *InMemoryPlayerRepository) GetPlayerByID(ctx context.Context, id string) (*models.PlayerData, error) {
	player, exists := r.players[id]
	if !exists {
		return nil, fmt.Errorf("player not found")
	}
	if playerData, ok := player.(*models.PlayerData); ok {
		return playerData, nil
	}
	return nil, fmt.Errorf("player is not of type PlayerData")
}

func (r *InMemoryPlayerRepository) GetPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error) {
	for _, player := range r.players {
		if playerData, ok := player.(*models.PlayerData); ok && playerData.Email != nil && *playerData.Email == email {
			return playerData, nil
		}
	}
	return nil, fmt.Errorf("player not found")
}

func (r *InMemoryPlayerRepository) GetPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error) {
	for _, player := range r.players {
		if playerData, ok := player.(*models.PlayerData); ok && playerData.Provider != nil && *playerData.Provider == provider {
			return playerData, nil
		}
	}
	return nil, fmt.Errorf("player not found")
}

func (r *InMemoryPlayerRepository) GetPlayerByName(ctx context.Context, name string) (*models.PlayerData, error) {
	for _, player := range r.players {
		if player.GetName() == name {
			if playerData, ok := player.(*models.PlayerData); ok {
				return playerData, nil
			}
		}
	}
	return nil, fmt.Errorf("player not found")
}

func (r *InMemoryPlayerRepository) UpdatePlayer(ctx context.Context, player *models.PlayerData) error {
	r.players[player.ID] = player
	return nil
}

func (r *InMemoryPlayerRepository) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	// In-memory implementation - store in a simple map
	// This is a simplified version for testing
	return nil
}

func (r *InMemoryPlayerRepository) GetAuthToken(ctx context.Context, token string) (*models.AuthToken, error) {
	// In-memory implementation - return nil for testing
	return nil, nil
}

func (r *InMemoryPlayerRepository) DeleteAuthToken(ctx context.Context, token string) error {
	// In-memory implementation
	return nil
}

func (r *InMemoryPlayerRepository) DeleteExpiredTokens(ctx context.Context) error {
	// In-memory implementation
	return nil
}
