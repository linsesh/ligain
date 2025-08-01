package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/rules"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
)

// GameCreationServiceInterface defines the interface for game creation services
type GameCreationServiceInterface interface {
	CreateGame(req *CreateGameRequest, player models.Player) (*CreateGameResponse, error)
	JoinGame(code string, player models.Player) (*JoinGameResponse, error)
	GetPlayerGames(player models.Player) ([]PlayerGame, error)
	GetGameService(gameID string, player models.Player) (GameService, error)
	CleanupExpiredCodes() error
	LeaveGame(gameID string, player models.Player) error
}

// GameCreationService handles game creation with unique codes
type GameCreationService struct {
	gameRepo       repositories.GameRepository
	gameCodeRepo   repositories.GameCodeRepository
	gamePlayerRepo repositories.GamePlayerRepository
	betRepo        repositories.BetRepository
	matchRepo      repositories.MatchRepository
	watcher        MatchWatcherService
	gameServices   map[string]GameService
}

// NewGameCreationService creates a new GameCreationService instance
func NewGameCreationService(gameRepo repositories.GameRepository, gameCodeRepo repositories.GameCodeRepository, gamePlayerRepo repositories.GamePlayerRepository, betRepo repositories.BetRepository, matchRepo repositories.MatchRepository, watcher MatchWatcherService) GameCreationServiceInterface {
	return &GameCreationService{
		gameRepo:       gameRepo,
		gameCodeRepo:   gameCodeRepo,
		gamePlayerRepo: gamePlayerRepo,
		betRepo:        betRepo,
		matchRepo:      matchRepo,
		watcher:        watcher,
		gameServices:   make(map[string]GameService),
	}
}

// NewGameCreationServiceWithLoadedGames creates a new GameCreationService instance and loads all existing games from the repository
func NewGameCreationServiceWithLoadedGames(gameRepo repositories.GameRepository, gameCodeRepo repositories.GameCodeRepository, gamePlayerRepo repositories.GamePlayerRepository, betRepo repositories.BetRepository, matchRepo repositories.MatchRepository, watcher MatchWatcherService) (GameCreationServiceInterface, error) {
	service := &GameCreationService{
		gameRepo:       gameRepo,
		gameCodeRepo:   gameCodeRepo,
		gamePlayerRepo: gamePlayerRepo,
		betRepo:        betRepo,
		matchRepo:      matchRepo,
		watcher:        watcher,
		gameServices:   make(map[string]GameService),
	}

	// Load all games from the repository
	games, err := gameRepo.GetAllGames()
	if err != nil {
		return nil, fmt.Errorf("failed to load games from repository: %v", err)
	}

	// Create GameService instances for each game
	for gameID := range games {
		gameService := NewGameService(gameID, gameRepo, betRepo)
		service.gameServices[gameID] = gameService

		// Subscribe to the watcher if available
		if watcher != nil {
			if err := watcher.Subscribe(gameService); err != nil {
				log.WithError(err).Warnf("Failed to subscribe game %s to watcher", gameID)
			}
		}
	}

	log.WithField("gameCount", len(games)).Info("Loaded games from repository")
	return service, nil
}

// GetGameService returns a GameService by ID, but only if the player has access to it
func (s *GameCreationService) GetGameService(gameID string, player models.Player) (GameService, error) {
	// Check if player is in the game first
	isInGame, err := s.gamePlayerRepo.IsPlayerInGame(context.Background(), gameID, player.GetID())
	if err != nil {
		return nil, fmt.Errorf("error checking game access: %v", err)
	}
	if !isInGame {
		return nil, ErrPlayerNotInGame
	}

	gameService, exists := s.gameServices[gameID]
	if !exists {
		// Create a new game service
		gameService = NewGameService(gameID, s.gameRepo, s.betRepo)
		s.gameServices[gameID] = gameService

		// Subscribe to the watcher if available
		if s.watcher != nil {
			if err := s.watcher.Subscribe(gameService); err != nil {
				return nil, fmt.Errorf("failed to subscribe game to watcher: %v", err)
			}
		}
	}

	return gameService, nil
}

// CreateGameRequest represents the request to create a new game
type CreateGameRequest struct {
	SeasonYear      string `json:"seasonYear" binding:"required"`
	CompetitionName string `json:"competitionName" binding:"required"`
	Name            string `json:"name" binding:"required"`
}

// CreateGameResponse represents the response when creating a new game
type CreateGameResponse struct {
	GameID string `json:"gameId"`
	Code   string `json:"code"`
}

// JoinGameResponse represents the response when joining a game
type JoinGameResponse struct {
	GameID          string `json:"gameId"`
	SeasonYear      string `json:"seasonYear"`
	CompetitionName string `json:"competitionName"`
	Message         string `json:"message"`
}

type PlayerGameInfo struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	TotalScore    int            `json:"totalScore"`
	ScoresByMatch map[string]int `json:"scoresByMatch"`
}

type PlayerGame struct {
	GameID          string           `json:"gameId"`
	SeasonYear      string           `json:"seasonYear"`
	CompetitionName string           `json:"competitionName"`
	Name            string           `json:"name"`
	Status          string           `json:"status"`
	Players         []PlayerGameInfo `json:"players"`
	Code            string           `json:"code"`
}

var (
	ErrInvalidCompetition = errors.New("only 'Ligue 1' is supported as competition name")
	ErrInvalidSeasonYear  = errors.New("only '2025/2026' is supported as season year")
	ErrPlayerNotInGame    = errors.New("player is not in the game")
	ErrPlayerGameLimit    = errors.New("player has reached the maximum limit of 5 games")
)

// CreateGame creates a new game with a unique 4-character code
func (s *GameCreationService) CreateGame(req *CreateGameRequest, player models.Player) (*CreateGameResponse, error) {
	// Validate competition name - only Ligue 1 is supported
	if req.CompetitionName != "Ligue 1" {
		return nil, ErrInvalidCompetition
	}
	// Validate season year - only 2025/2026 is supported
	if req.SeasonYear != "2025/2026" {
		return nil, ErrInvalidSeasonYear
	}
	// Validate name - must not be empty
	if req.Name == "" {
		return nil, fmt.Errorf("game name is required")
	}

	// Check if player has reached the game limit (5 games)
	playerGames, err := s.gamePlayerRepo.GetPlayerGames(context.Background(), player.GetID())
	if err != nil {
		return nil, fmt.Errorf("failed to check player games: %v", err)
	}
	if len(playerGames) >= 5 {
		return nil, ErrPlayerGameLimit
	}

	// Load all matches for this competition and season
	matches, err := s.matchRepo.GetMatchesByCompetitionAndSeason(req.CompetitionName, req.SeasonYear)
	if err != nil {
		return nil, fmt.Errorf("failed to load matches: %v", err)
	}

	// Create a new game with the loaded matches
	game := rules.NewFreshGame(
		req.SeasonYear,
		req.CompetitionName,
		req.Name,
		[]models.Player{player},
		matches,
		&rules.ScorerOriginal{},
	)

	// Save the game to get its ID
	gameID, err := s.gameRepo.CreateGame(game)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %v", err)
	}

	// Add the creator to the game in the database
	err = s.gamePlayerRepo.AddPlayerToGame(context.Background(), gameID, player.GetID())
	if err != nil {
		return nil, fmt.Errorf("failed to add creator to game: %v", err)
	}

	// Create and store the game service
	gameService := NewGameService(gameID, s.gameRepo, s.betRepo)
	s.gameServices[gameID] = gameService

	// Subscribe the new game to the watcher
	if s.watcher != nil {
		err := s.watcher.Subscribe(gameService)
		if err != nil {
			return nil, fmt.Errorf("failed to subscribe game to watcher: %v", err)
		}
	}

	// Generate a unique code for the game
	code, err := s.generateUniqueCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique code: %v", err)
	}

	// Create game code with 6 months expiration
	expiresAt := time.Now().AddDate(0, 6, 0)
	gameCode := models.NewGameCode(gameID, code, expiresAt)

	err = s.gameCodeRepo.CreateGameCode(gameCode)
	if err != nil {
		return nil, fmt.Errorf("failed to create game code: %v", err)
	}

	return &CreateGameResponse{
		GameID: gameID,
		Code:   code,
	}, nil
}

// JoinGame joins a player to a game using a 4-character code
func (s *GameCreationService) JoinGame(code string, player models.Player) (*JoinGameResponse, error) {
	// Get the game code from the database
	gameCode, err := s.gameCodeRepo.GetGameCodeByCode(code)
	if err != nil {
		return nil, fmt.Errorf("invalid game code: %v", err)
	}

	// Check if the code is expired
	if gameCode.ExpiresAt.Before(time.Now()) {
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
	playerGames, err := s.gamePlayerRepo.GetPlayerGames(context.Background(), player.GetID())
	if err != nil {
		return nil, fmt.Errorf("failed to check player games: %v", err)
	}
	if len(playerGames) >= 5 {
		return nil, ErrPlayerGameLimit
	}

	// Add the player to the game
	err = s.addPlayerToGame(gameCode.GameID, player)
	if err != nil {
		return nil, fmt.Errorf("failed to add player to game: %v", err)
	}

	// Update the cached game service if it exists
	if gs, ok := s.gameServices[gameCode.GameID]; ok {
		_ = gs.AddPlayer(player)
	}

	return &JoinGameResponse{
		GameID:          gameCode.GameID,
		SeasonYear:      game.GetSeasonYear(),
		CompetitionName: game.GetCompetitionName(),
		Message:         "Successfully joined the game",
	}, nil
}

// addPlayerToGame adds a player to a game if they're not already in it
func (s *GameCreationService) addPlayerToGame(gameID string, player models.Player) error {
	// Check if player is already in the game
	isInGame, err := s.gamePlayerRepo.IsPlayerInGame(context.Background(), gameID, player.GetID())
	if err != nil {
		return fmt.Errorf("error checking if player is in game: %v", err)
	}

	if isInGame {
		return nil // Player is already in the game
	}

	// Add player to the game
	err = s.gamePlayerRepo.AddPlayerToGame(context.Background(), gameID, player.GetID())
	if err != nil {
		return fmt.Errorf("error adding player to game: %v", err)
	}

	return nil
}

// generateUniqueCode generates a unique 4-character alphanumeric code
func (s *GameCreationService) generateUniqueCode() (string, error) {
	const maxAttempts = 10
	const codeLength = 4
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code := s.generateRandomCode(codeLength, charset)

		// Check if code already exists
		exists, err := s.gameCodeRepo.CodeExists(code)
		if err != nil {
			return "", fmt.Errorf("error checking code existence: %v", err)
		}

		if !exists {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

// generateRandomCode generates a random code of specified length using the given charset
func (s *GameCreationService) generateRandomCode(length int, charset string) string {
	code := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, _ := rand.Int(rand.Reader, charsetLen)
		code[i] = charset[randomIndex.Int64()]
	}

	return string(code)
}

// CleanupExpiredCodes removes all expired game codes
func (s *GameCreationService) CleanupExpiredCodes() error {
	return s.gameCodeRepo.DeleteExpiredCodes()
}

// GetPlayerGames returns all games that a player is part of
func (s *GameCreationService) GetPlayerGames(player models.Player) ([]PlayerGame, error) {
	gameIDs, err := s.gamePlayerRepo.GetPlayerGames(context.Background(), player.GetID())
	if err != nil {
		return nil, fmt.Errorf("error getting player games: %v", err)
	}

	var playerGames []PlayerGame
	for _, gameID := range gameIDs {
		// Always get fresh data from the database, not from cache
		game, err := s.gameRepo.GetGame(gameID)
		if err != nil {
			log.Errorf("error getting game %s: %v", gameID, err)
			continue
		}

		// Fetch all players in the game
		players, err := s.gamePlayerRepo.GetPlayersInGame(context.Background(), gameID)
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
			Name:            game.GetName(), // Add game name here
			Status:          gameStatus,
			Players:         playerInfos,
			Code:            code,
		}

		playerGames = append(playerGames, playerGame)
	}

	return playerGames, nil
}

// Add LeaveGame implementation
func (s *GameCreationService) LeaveGame(gameID string, player models.Player) error {
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
	// Check if any players are left
	players, err := s.gamePlayerRepo.GetPlayersInGame(ctx, gameID)
	if err != nil {
		return fmt.Errorf("error checking remaining players: %v", err)
	}
	if len(players) == 0 {
		if err := s.deleteGame(gameID); err != nil {
			return err
		}
	}
	// Remove from cache if present
	delete(s.gameServices, gameID)
	return nil
}

// deleteGame marks a game as finished, persists it, unsubscribes from the watcher, and deletes the join code
func (s *GameCreationService) deleteGame(gameID string) error {
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
