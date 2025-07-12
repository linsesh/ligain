package models

import (
	"time"
)

// GameCode represents a short-lived code for joining a game
type GameCode struct {
	ID        string    `json:"id" db:"id"`
	GameID    string    `json:"gameId" db:"game_id"`
	Code      string    `json:"code" db:"code"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	ExpiresAt time.Time `json:"expiresAt" db:"expires_at"`
}

// NewGameCode creates a new GameCode instance
func NewGameCode(gameID, code string, expiresAt time.Time) *GameCode {
	return &GameCode{
		GameID:    gameID,
		Code:      code,
		ExpiresAt: expiresAt,
	}
}

// IsExpired checks if the game code has expired
func (gc *GameCode) IsExpired() bool {
	return time.Now().After(gc.ExpiresAt)
}

// IsValid checks if the game code is still valid (not expired)
func (gc *GameCode) IsValid() bool {
	return !gc.IsExpired()
}
