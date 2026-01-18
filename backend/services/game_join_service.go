package services

import (
	"context"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"time"
)

// GameJoinServiceInterface defines the interface for game join operations
type GameJoinServiceInterface interface {
	// JoinGame joins a player to a game using a code
	JoinGame(code string, player models.Player) (*JoinGameResponse, error)
}

// GameJoinService handles joining games by code
type GameJoinService struct {
	gameCodeRepo      repositories.GameCodeRepository
	gameRepo          repositories.GameRepository
	gamePlayerRepo    repositories.GamePlayerRepository
	membershipService GameMembershipServiceInterface
	registry          GameServiceRegistryInterface
	timeFunc          func() time.Time
}

// NewGameJoinService creates a new GameJoinService instance
func NewGameJoinService(
	gameCodeRepo repositories.GameCodeRepository,
	gameRepo repositories.GameRepository,
	membershipService GameMembershipServiceInterface,
	registry GameServiceRegistryInterface,
	timeFunc func() time.Time,
) *GameJoinService {
	// Extract gamePlayerRepo from membershipService for game limit checks
	var gamePlayerRepo repositories.GamePlayerRepository
	if ms, ok := membershipService.(*GameMembershipService); ok {
		gamePlayerRepo = ms.gamePlayerRepo
	}

	return &GameJoinService{
		gameCodeRepo:      gameCodeRepo,
		gameRepo:          gameRepo,
		gamePlayerRepo:    gamePlayerRepo,
		membershipService: membershipService,
		registry:          registry,
		timeFunc:          timeFunc,
	}
}

// JoinGame joins a player to a game using a 4-character code
func (s *GameJoinService) JoinGame(code string, player models.Player) (*JoinGameResponse, error) {
	ctx := context.Background()

	// Get the game code from the database
	gameCode, err := s.gameCodeRepo.GetGameCodeByCode(code)
	if err != nil {
		return nil, fmt.Errorf("invalid game code: %v", err)
	}

	// Check if the code is expired
	if gameCode.ExpiresAt.Before(s.timeFunc()) {
		return nil, fmt.Errorf("game code has expired")
	}

	// Get the game from the database
	game, err := s.gameRepo.GetGame(gameCode.GameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get game: %v", err)
	}

	// Prevent joining if finished
	if game.GetGameStatus() == models.GameStatusFinished {
		return nil, fmt.Errorf("cannot join a finished game")
	}

	// Check if player has reached the game limit (5 games)
	playerGames, err := s.gamePlayerRepo.GetPlayerGames(ctx, player.GetID())
	if err != nil {
		return nil, fmt.Errorf("failed to check player games: %v", err)
	}
	if len(playerGames) >= 5 {
		return nil, ErrPlayerGameLimit
	}

	// Add the player to the game (also updates cached game service)
	err = s.membershipService.AddPlayerToGame(gameCode.GameID, player)
	if err != nil {
		return nil, fmt.Errorf("failed to add player to game: %v", err)
	}

	return &JoinGameResponse{
		GameID:          gameCode.GameID,
		SeasonYear:      game.GetSeasonYear(),
		CompetitionName: game.GetCompetitionName(),
		Message:         "Successfully joined the game",
	}, nil
}
