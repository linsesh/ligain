package repositories

import (
	"context"
	"fmt"
	"ligain/backend/models"
	"sync"
)

// InMemoryGamePlayerRepository implements GamePlayerRepository using in-memory storage
type InMemoryGamePlayerRepository struct {
	mu            sync.RWMutex
	gameToPlayers map[string]map[string]bool // gameID -> set of playerIDs
	playerToGames map[string]map[string]bool // playerID -> set of gameIDs
	playerRepo    PlayerRepository           // Reference to player repository to fetch player data
}

// NewInMemoryGamePlayerRepository creates a new in-memory game-player repository
func NewInMemoryGamePlayerRepository(playerRepo PlayerRepository) *InMemoryGamePlayerRepository {
	return &InMemoryGamePlayerRepository{
		gameToPlayers: make(map[string]map[string]bool),
		playerToGames: make(map[string]map[string]bool),
		playerRepo:    playerRepo,
	}
}

func (r *InMemoryGamePlayerRepository) AddPlayerToGame(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.gameToPlayers[gameID] == nil {
		r.gameToPlayers[gameID] = make(map[string]bool)
	}
	r.gameToPlayers[gameID][playerID] = true

	if r.playerToGames[playerID] == nil {
		r.playerToGames[playerID] = make(map[string]bool)
	}
	r.playerToGames[playerID][gameID] = true

	return nil
}

func (r *InMemoryGamePlayerRepository) RemovePlayerFromGame(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.gameToPlayers[gameID] != nil {
		delete(r.gameToPlayers[gameID], playerID)
	}

	if r.playerToGames[playerID] != nil {
		delete(r.playerToGames[playerID], gameID)
	}

	return nil
}

func (r *InMemoryGamePlayerRepository) GetPlayersInGame(ctx context.Context, gameID string) ([]models.Player, error) {
	r.mu.RLock()
	playerIDs := r.gameToPlayers[gameID]
	r.mu.RUnlock()

	if playerIDs == nil {
		return []models.Player{}, nil
	}

	players := make([]models.Player, 0, len(playerIDs))
	for playerID := range playerIDs {
		// Fetch player from player repository to get current data (including updated names)
		player, err := r.playerRepo.GetPlayerByID(ctx, playerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get player %s: %v", playerID, err)
		}
		if player != nil {
			players = append(players, player)
		}
	}

	return players, nil
}

func (r *InMemoryGamePlayerRepository) GetPlayerGames(ctx context.Context, playerID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	gameIDs := r.playerToGames[playerID]
	if gameIDs == nil {
		return []string{}, nil
	}

	games := make([]string, 0, len(gameIDs))
	for gameID := range gameIDs {
		games = append(games, gameID)
	}

	return games, nil
}

func (r *InMemoryGamePlayerRepository) IsPlayerInGame(ctx context.Context, gameID string, playerID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.gameToPlayers[gameID] == nil {
		return false, nil
	}

	return r.gameToPlayers[gameID][playerID], nil
}
