package repositories

import (
	"errors"
	"ligain/backend/models"
)

// Error constants
var ErrGameCodeNotFound = errors.New("game code not found")

// GameCodeRepository defines the interface for game code operations
type GameCodeRepository interface {
	// CreateGameCode creates a new game code and returns the created code
	CreateGameCode(gameCode *models.GameCode) error

	// GetGameCodeByCode retrieves a game code by its code value
	GetGameCodeByCode(code string) (*models.GameCode, error)

	// GetGameCodeByGameID retrieves the active game code for a specific game
	GetGameCodeByGameID(gameID string) (*models.GameCode, error)

	// DeleteExpiredCodes removes all expired game codes
	DeleteExpiredCodes() error

	// CodeExists checks if a code exists and is not expired
	CodeExists(code string) (bool, error)

	// DeleteGameCode deletes a specific game code
	DeleteGameCode(code string) error
	DeleteGameCodeByGameID(gameID string) error
}
