package services

import (
	"fmt"
	"ligain/backend/repositories"
	"sync"

	log "github.com/sirupsen/logrus"
)

// GameServiceRegistryInterface defines the interface for the game service registry
type GameServiceRegistryInterface interface {
	// Create creates a new GameService and registers it
	Create(gameID string) (GameService, error)
	// Get returns an existing GameService without creating one
	Get(gameID string) (GameService, bool)
	// Register explicitly adds a GameService to the registry
	Register(gameID string, gs GameService)
	// Unregister removes a GameService from the registry
	Unregister(gameID string)
}

// GameServiceRegistry manages GameService instances
type GameServiceRegistry struct {
	gameRepo       repositories.GameRepository
	betRepo        repositories.BetRepository
	gamePlayerRepo repositories.GamePlayerRepository
	watcher        MatchWatcherService
	gameServices   sync.Map
}

// NewGameServiceRegistry creates a new GameServiceRegistry instance and loads all existing games
func NewGameServiceRegistry(
	gameRepo repositories.GameRepository,
	betRepo repositories.BetRepository,
	gamePlayerRepo repositories.GamePlayerRepository,
	watcher MatchWatcherService,
) (*GameServiceRegistry, error) {
	r := &GameServiceRegistry{
		gameRepo:       gameRepo,
		betRepo:        betRepo,
		gamePlayerRepo: gamePlayerRepo,
		watcher:        watcher,
	}
	if err := r.loadAll(); err != nil {
		return nil, err
	}
	return r, nil
}

// Create creates a new GameService and registers it
func (r *GameServiceRegistry) Create(gameID string) (GameService, error) {
	gameService := NewGameService(gameID, r.gameRepo, r.betRepo, r.gamePlayerRepo)

	// Subscribe to watcher if available
	if r.watcher != nil {
		if err := r.watcher.Subscribe(gameService); err != nil {
			return nil, fmt.Errorf("failed to subscribe game to watcher: %v", err)
		}
	}

	r.gameServices.Store(gameID, gameService)
	return gameService, nil
}

// Get returns an existing GameService without creating one
func (r *GameServiceRegistry) Get(gameID string) (GameService, bool) {
	gs, exists := r.gameServices.Load(gameID)
	if !exists {
		return nil, false
	}
	return gs.(GameService), true
}

// Register explicitly adds a GameService to the registry
func (r *GameServiceRegistry) Register(gameID string, gs GameService) {
	r.gameServices.Store(gameID, gs)
}

// Unregister removes a GameService from the registry
func (r *GameServiceRegistry) Unregister(gameID string) {
	r.gameServices.Delete(gameID)
}

// loadAll loads all existing games from the repository
func (r *GameServiceRegistry) loadAll() error {
	games, err := r.gameRepo.GetAllGames()
	if err != nil {
		return fmt.Errorf("failed to load games from repository: %v", err)
	}

	for gameID := range games {
		gameService := NewGameService(gameID, r.gameRepo, r.betRepo, r.gamePlayerRepo)
		r.gameServices.Store(gameID, gameService)

		// Subscribe to watcher if available
		if r.watcher != nil {
			if err := r.watcher.Subscribe(gameService); err != nil {
				log.WithError(err).Warnf("Failed to subscribe game %s to watcher", gameID)
			}
		}
	}

	log.WithField("gameCount", len(games)).Info("Loaded games from repository")
	return nil
}
