package services

import (
	"context"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"

	log "github.com/sirupsen/logrus"
)

// GameMembershipServiceInterface defines the interface for game membership operations
type GameMembershipServiceInterface interface {
	// AddPlayerToGame adds a player to a game (idempotent - returns nil if already in game)
	AddPlayerToGame(gameID string, player models.Player) error
	// RemovePlayerFromGame removes a player from a game
	RemovePlayerFromGame(gameID string, player models.Player) error
	// IsPlayerInGame checks if a player is in a specific game
	IsPlayerInGame(gameID string, playerID string) (bool, error)
	// GetPlayersInGame returns all players in a specific game
	GetPlayersInGame(gameID string) ([]models.Player, error)
	// LeaveGame allows a player to leave a game (combines removal + cleanup)
	LeaveGame(gameID string, player models.Player) error
}

// GameMembershipService handles player membership in games
type GameMembershipService struct {
	gamePlayerRepo repositories.GamePlayerRepository
	gameRepo       repositories.GameRepository
	gameCodeRepo   repositories.GameCodeRepository
	registry       GameServiceRegistryInterface
	watcher        MatchWatcherService
}

// NewGameMembershipService creates a new GameMembershipService instance
func NewGameMembershipService(
	gamePlayerRepo repositories.GamePlayerRepository,
	gameRepo repositories.GameRepository,
	gameCodeRepo repositories.GameCodeRepository,
	registry GameServiceRegistryInterface,
	watcher MatchWatcherService,
) *GameMembershipService {
	return &GameMembershipService{
		gamePlayerRepo: gamePlayerRepo,
		gameRepo:       gameRepo,
		gameCodeRepo:   gameCodeRepo,
		registry:       registry,
		watcher:        watcher,
	}
}

// AddPlayerToGame adds a player to a game if they're not already in it
func (s *GameMembershipService) AddPlayerToGame(gameID string, player models.Player) error {
	ctx := context.Background()

	// Check if player is already in the game
	isInGame, err := s.gamePlayerRepo.IsPlayerInGame(ctx, gameID, player.GetID())
	if err != nil {
		return fmt.Errorf("error checking if player is in game: %v", err)
	}

	if isInGame {
		return nil // Idempotent - player already in the game
	}

	// Add player to the game
	err = s.gamePlayerRepo.AddPlayerToGame(ctx, gameID, player.GetID())
	if err != nil {
		return fmt.Errorf("error adding player to game: %v", err)
	}

	// Update cached game service if it exists
	s.addPlayerToGameService(gameID, player)

	return nil
}

// RemovePlayerFromGame removes a player from a game
func (s *GameMembershipService) RemovePlayerFromGame(gameID string, player models.Player) error {
	ctx := context.Background()

	// Check if player is in the game
	isInGame, err := s.gamePlayerRepo.IsPlayerInGame(ctx, gameID, player.GetID())
	if err != nil {
		return fmt.Errorf("error checking if player is in game: %v", err)
	}

	if !isInGame {
		return ErrPlayerNotInGame
	}

	// Remove player from game
	err = s.gamePlayerRepo.RemovePlayerFromGame(ctx, gameID, player.GetID())
	if err != nil {
		return fmt.Errorf("error removing player from game: %v", err)
	}

	// Update cached game service if it exists
	s.removePlayerFromGameService(gameID, player)

	// Check if any players are left
	players, err := s.gamePlayerRepo.GetPlayersInGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("error checking remaining players: %v", err)
	}

	if len(players) == 0 {
		if err := s.deleteGame(gameID); err != nil {
			return err
		}
		// Only remove from registry cache when game is deleted
		s.registry.Unregister(gameID)
	}

	return nil
}

// IsPlayerInGame checks if a player is in a specific game
func (s *GameMembershipService) IsPlayerInGame(gameID string, playerID string) (bool, error) {
	return s.gamePlayerRepo.IsPlayerInGame(context.Background(), gameID, playerID)
}

// GetPlayersInGame returns all players in a specific game
func (s *GameMembershipService) GetPlayersInGame(gameID string) ([]models.Player, error) {
	return s.gamePlayerRepo.GetPlayersInGame(context.Background(), gameID)
}

// LeaveGame allows a player to leave a game (alias for RemovePlayerFromGame)
func (s *GameMembershipService) LeaveGame(gameID string, player models.Player) error {
	return s.RemovePlayerFromGame(gameID, player)
}

// deleteGame marks a game as finished, persists it, unsubscribes from the watcher, and deletes the join code
func (s *GameMembershipService) deleteGame(gameID string) error {
	game, err := s.gameRepo.GetGame(gameID)
	if err != nil {
		return fmt.Errorf("error loading game to finish: %v", err)
	}

	game.Finish()

	err = s.gameRepo.SaveWithId(gameID, game)
	if err != nil {
		return fmt.Errorf("error saving finished game: %v", err)
	}

	// Always unsubscribe from watcher if present
	if s.watcher != nil {
		err := s.watcher.Unsubscribe(gameID)
		if err != nil {
			log.WithError(err).Warnf("Failed to unsubscribe game %s from watcher", gameID)
		}
	}

	if s.gameCodeRepo != nil {
		_ = s.gameCodeRepo.DeleteGameCodeByGameID(gameID)
	}

	return nil
}

// addPlayerToGameService updates the cached game service when a player is added
func (s *GameMembershipService) addPlayerToGameService(gameID string, player models.Player) {
	if gs, exists := s.registry.Get(gameID); exists {
		_ = gs.AddPlayer(player)
	}
}

// removePlayerFromGameService updates the cached game service when a player is removed
func (s *GameMembershipService) removePlayerFromGameService(gameID string, player models.Player) {
	if gs, exists := s.registry.Get(gameID); exists {
		_ = gs.RemovePlayer(player)
	}
}
