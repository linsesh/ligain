package services

import (
	"context"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"

	log "github.com/sirupsen/logrus"
)

// GameQueryServiceInterface defines the interface for game query operations
type GameQueryServiceInterface interface {
	// GetPlayerGames returns all games that a player is part of with full aggregation
	GetPlayerGames(player models.Player) ([]PlayerGame, error)
	// GetGame returns a single game by ID
	GetGame(gameID string) (models.Game, error)
}

// GameQueryService handles game query operations
type GameQueryService struct {
	gameRepo       repositories.GameRepository
	gamePlayerRepo repositories.GamePlayerRepository
	gameCodeRepo   repositories.GameCodeRepository
	betRepo        repositories.BetRepository
}

// NewGameQueryService creates a new GameQueryService instance
func NewGameQueryService(
	gameRepo repositories.GameRepository,
	gamePlayerRepo repositories.GamePlayerRepository,
	gameCodeRepo repositories.GameCodeRepository,
	betRepo repositories.BetRepository,
) *GameQueryService {
	return &GameQueryService{
		gameRepo:       gameRepo,
		gamePlayerRepo: gamePlayerRepo,
		gameCodeRepo:   gameCodeRepo,
		betRepo:        betRepo,
	}
}

// GetPlayerGames returns all games that a player is part of
func (s *GameQueryService) GetPlayerGames(player models.Player) ([]PlayerGame, error) {
	ctx := context.Background()

	gameIDs, err := s.gamePlayerRepo.GetPlayerGames(ctx, player.GetID())
	if err != nil {
		return nil, fmt.Errorf("error getting player games: %v", err)
	}

	var playerGames []PlayerGame
	for _, gameID := range gameIDs {
		game, err := s.gameRepo.GetGame(gameID)
		if err != nil {
			log.Errorf("error getting game %s: %v", gameID, err)
			continue // Graceful degradation - skip games that fail to load
		}

		// Fetch all players in the game
		players, err := s.gamePlayerRepo.GetPlayersInGame(ctx, gameID)
		if err != nil {
			log.Errorf("error getting players for game %s: %v", gameID, err)
			continue
		}

		// Fetch all scores for the game (by match and player)
		playerScoresByMatch, err := s.betRepo.GetScoresByMatchAndPlayer(gameID)
		if err != nil {
			log.Errorf("error getting scores for game %s: %v", gameID, err)
			continue
		}

		// Build player info
		var playerInfos []PlayerGameInfo
		for _, p := range players {
			total := 0
			scoresByMatch := make(map[string]int)
			for matchID, playerScores := range playerScoresByMatch {
				if score, ok := playerScores[p.GetID()]; ok {
					total += score
					scoresByMatch[matchID] = score
				}
			}
			playerInfos = append(playerInfos, PlayerGameInfo{
				ID:            p.GetID(),
				Name:          p.GetName(),
				TotalScore:    total,
				ScoresByMatch: scoresByMatch,
			})
		}

		// Get the game code
		gameCode, err := s.gameCodeRepo.GetGameCodeByGameID(gameID)
		code := ""
		if err == nil && gameCode != nil {
			code = gameCode.Code
		}

		gameStatus := string(game.GetGameStatus())
		playerGame := PlayerGame{
			GameID:          gameID,
			SeasonYear:      game.GetSeasonYear(),
			CompetitionName: game.GetCompetitionName(),
			Name:            game.GetName(),
			Status:          gameStatus,
			Players:         playerInfos,
			Code:            code,
		}

		playerGames = append(playerGames, playerGame)
	}

	return playerGames, nil
}

// GetGame returns a single game by ID
func (s *GameQueryService) GetGame(gameID string) (models.Game, error) {
	return s.gameRepo.GetGame(gameID)
}
