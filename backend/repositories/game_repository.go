package repositories

import (
	"liguain/backend/models"
)

type GameRepository interface {
	// GetGame returns a game if it exists
	//GetGame(gameId string) (rules.Game, error)
	// SaveGame saves a game and returns the game id, and an error if saving failed
	//SaveGame(game rules.Game) (string, error)
	// UpdateScores updates the scores for the given players
	UpdateScores(match models.Match, scores map[models.Player]int) error
}
