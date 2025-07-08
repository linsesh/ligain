package models

import (
	"time"
)

// Player is the core interface for game logic
// It only exposes fields needed by the game system, not authentication details
type Player interface {
	GetID() string
	GetName() string
}

// SimplePlayer represents a player with only the essential fields for game logic
// This is used when we don't need authentication details
type SimplePlayer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PlayerData represents a human player with authentication
// This concrete type contains all the auth-related fields
type PlayerData struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	// Email is optional - some players might not provide an email during OAuth
	// Using *string allows NULL in database and omits from JSON when nil
	Email *string `json:"email,omitempty" db:"email"`
	// Provider indicates the OAuth provider ("google" or "apple")
	// Using *string because it's set during authentication, not at creation
	Provider *string `json:"provider,omitempty" db:"provider"`
	// ProviderID is the unique identifier from the OAuth provider
	// Using *string because it's only available after OAuth authentication
	ProviderID *string    `json:"provider_id,omitempty" db:"provider_id"`
	CreatedAt  *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

// Implement Player interface for SimplePlayer
func (p *SimplePlayer) GetID() string {
	return p.ID
}

func (p *SimplePlayer) GetName() string {
	return p.Name
}

// Implement Player interface for PlayerData
func (p *PlayerData) GetID() string {
	return p.ID
}

func (p *PlayerData) GetName() string {
	return p.Name
}

// AuthToken represents an authentication token
type AuthToken struct {
	ID        string    `json:"id" db:"id"`
	PlayerID  string    `json:"player_id" db:"player_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AuthRequest represents the authentication request from the frontend
type AuthRequest struct {
	Provider string `json:"provider"` // "google" or "apple"
	Token    string `json:"token"`    // ID token from the provider
	Email    string `json:"email"`
	Name     string `json:"name"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Player PlayerData `json:"player"`
	Token  string     `json:"token"`
}

// ToSimplePlayer converts a PlayerData to a SimplePlayer
func (p *PlayerData) ToSimplePlayer() *SimplePlayer {
	return &SimplePlayer{
		ID:   p.ID,
		Name: p.Name,
	}
}

// NewSimplePlayer creates a new SimplePlayer
func NewSimplePlayer(id, name string) *SimplePlayer {
	return &SimplePlayer{
		ID:   id,
		Name: name,
	}
}
