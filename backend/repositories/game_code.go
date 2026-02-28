package repositories

import (
	"errors"
	"ligain/backend/models"
	"time"

	"github.com/google/uuid"
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

type InMemoryGameCodeRepository struct {
	codes        map[string]*models.GameCode // code -> GameCode
	gameIDToCode map[string]*models.GameCode // gameID -> GameCode
}

func NewInMemoryGameCodeRepository() GameCodeRepository {
	return &InMemoryGameCodeRepository{
		codes:        make(map[string]*models.GameCode),
		gameIDToCode: make(map[string]*models.GameCode),
	}
}

func (r *InMemoryGameCodeRepository) CreateGameCode(gameCode *models.GameCode) error {
	gameCode.ID = uuid.New().String()
	gameCode.CreatedAt = time.Now()
	r.codes[gameCode.Code] = gameCode
	r.gameIDToCode[gameCode.GameID] = gameCode
	return nil
}

func (r *InMemoryGameCodeRepository) GetGameCodeByCode(code string) (*models.GameCode, error) {
	gc, exists := r.codes[code]
	if !exists || gc.IsExpired() {
		return nil, ErrGameCodeNotFound
	}
	return gc, nil
}

func (r *InMemoryGameCodeRepository) GetGameCodeByGameID(gameID string) (*models.GameCode, error) {
	gc, exists := r.gameIDToCode[gameID]
	if !exists || gc.IsExpired() {
		return nil, ErrGameCodeNotFound
	}
	return gc, nil
}

func (r *InMemoryGameCodeRepository) DeleteExpiredCodes() error {
	for code, gc := range r.codes {
		if gc.IsExpired() {
			delete(r.gameIDToCode, gc.GameID)
			delete(r.codes, code)
		}
	}
	return nil
}

func (r *InMemoryGameCodeRepository) CodeExists(code string) (bool, error) {
	gc, exists := r.codes[code]
	return exists && !gc.IsExpired(), nil
}

func (r *InMemoryGameCodeRepository) DeleteGameCode(code string) error {
	gc, exists := r.codes[code]
	if !exists {
		return ErrGameCodeNotFound
	}
	delete(r.gameIDToCode, gc.GameID)
	delete(r.codes, code)
	return nil
}

func (r *InMemoryGameCodeRepository) DeleteGameCodeByGameID(gameID string) error {
	gc, exists := r.gameIDToCode[gameID]
	if !exists {
		return nil
	}
	delete(r.codes, gc.Code)
	delete(r.gameIDToCode, gameID)
	return nil
}
