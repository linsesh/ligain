package repositories

import (
	"fmt"
	"ligain/backend/models"
	"strconv"
)

const gameCacheSize = 100 // Maximum number of games to keep in cache

type GameRepository interface {
	// GetGame returns a game if it exists
	GetGame(gameId string) (models.Game, error)
	// CreateGame creates a new game and returns the game id, and an error if saving failed
	CreateGame(game models.Game) (string, error)
	// SaveWithId saves a game with a provided ID and returns an error if saving failed
	SaveWithId(gameId string, game models.Game) error
}

type InMemoryGameRepository struct {
	cache  *Cache[string, models.Game]
	lastId int
}

func NewInMemoryGameRepository() GameRepository {
	return &InMemoryGameRepository{
		cache:  NewCache[string, models.Game](gameCacheSize),
		lastId: 1,
	}
}

func (r *InMemoryGameRepository) CreateGame(game models.Game) (string, error) {
	gameId := strconv.Itoa(r.lastId)
	r.cache.Set(gameId, game)
	r.lastId++
	return gameId, nil
}

func (r *InMemoryGameRepository) SaveWithId(gameId string, game models.Game) error {
	r.cache.Set(gameId, game)
	return nil
}

func (r *InMemoryGameRepository) GetGame(gameId string) (models.Game, error) {
	game, err := r.cache.Get(gameId)
	if err == nil {
		return game, nil
	}
	return nil, fmt.Errorf("game %s not found", gameId)
}
