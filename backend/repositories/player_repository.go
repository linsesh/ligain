package repositories

import (
	"liguain/backend/models"
)

const playerCacheSize = 1000 // Maximum number of players to keep in cache

type PlayerRepository interface {
	// SavePlayer saves or updates a player and returns the player id
	SavePlayer(player models.Player) (string, error)
	GetPlayer(playerId string) (models.Player, error)
}

// InMemoryPlayerRepository is a simple in-memory implementation of PlayerRepository
type InMemoryPlayerRepository struct {
	cache *Cache[string, models.Player]
}

func NewInMemoryPlayerRepository() *InMemoryPlayerRepository {
	return &InMemoryPlayerRepository{
		cache: NewCache[string, models.Player](playerCacheSize),
	}
}

// SavePlayer saves or updates a player and returns the player id
func (r *InMemoryPlayerRepository) SavePlayer(player models.Player) (string, error) {
	id := player.Name // Using name as id for now
	r.cache.Set(id, player)
	return id, nil
}

func (r *InMemoryPlayerRepository) GetPlayer(playerId string) (models.Player, error) {
	player, err := r.cache.Get(playerId)
	if err == nil {
		return player, nil
	}
	return models.Player{}, nil
}
