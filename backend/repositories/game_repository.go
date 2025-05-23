package repositories

import (
	"liguain/backend/models"
	"strconv"
)

type GameRepository interface {
	// GetGame returns a game if it exists
	//GetGame(gameId string) (rules.Game, error)
	// CreateGame creates a game and returns the game id, and an error if saving failed
	SaveGame(game models.Game) (string, models.Game, error)
}

type InMemoryGameRepository struct {
	games  map[string]models.Game
	lastId int
}

var instance *InMemoryGameRepository

func NewInMemoryGameRepository() GameRepository {
	if instance == nil {
		instance = &InMemoryGameRepository{
			games:  make(map[string]models.Game),
			lastId: 1,
		}
	}
	return instance
}

func (r *InMemoryGameRepository) SaveGame(game models.Game) (string, models.Game, error) {
	gameId := strconv.Itoa(r.lastId)
	r.games[gameId] = game
	r.lastId++
	return gameId, r.games[gameId], nil
}
