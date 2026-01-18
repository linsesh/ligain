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
// This is now a facade that delegates to specialized services
type GameCreationService struct {
	gameRepo          repositories.GameRepository
	gameCodeRepo      repositories.GameCodeRepository
	gamePlayerRepo    repositories.GamePlayerRepository
	matchRepo         repositories.MatchRepository
	registry          GameServiceRegistryInterface
	membershipService GameMembershipServiceInterface
	queryService      GameQueryServiceInterface
	joinService       GameJoinServiceInterface
	timeFunc          func() time.Time
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

// NewGameCreationService creates a new GameCreationService instance (legacy constructor for compatibility)
func NewGameCreationService(gameRepo repositories.GameRepository, gameCodeRepo repositories.GameCodeRepository, gamePlayerRepo repositories.GamePlayerRepository, betRepo repositories.BetRepository, matchRepo repositories.MatchRepository, watcher MatchWatcherService) GameCreationServiceInterface {
	return NewGameCreationServiceWithTimeFunc(gameRepo, gameCodeRepo, gamePlayerRepo, betRepo, matchRepo, watcher, time.Now)
}

// NewGameCreationServiceWithTimeFunc creates a new GameCreationService instance with a custom time function (legacy constructor for compatibility)
func NewGameCreationServiceWithTimeFunc(gameRepo repositories.GameRepository, gameCodeRepo repositories.GameCodeRepository, gamePlayerRepo repositories.GamePlayerRepository, betRepo repositories.BetRepository, matchRepo repositories.MatchRepository, watcher MatchWatcherService, timeFunc func() time.Time) GameCreationServiceInterface {
	// Create the underlying services
	registry := NewGameServiceRegistry(gameRepo, betRepo, gamePlayerRepo, watcher)
	membershipService := NewGameMembershipService(gamePlayerRepo, gameRepo, gameCodeRepo, registry, watcher)
	queryService := NewGameQueryService(gameRepo, gamePlayerRepo, gameCodeRepo, betRepo)
	joinService := NewGameJoinService(gameCodeRepo, gameRepo, membershipService, registry, timeFunc)

	return &GameCreationService{
		gameRepo:          gameRepo,
		gameCodeRepo:      gameCodeRepo,
		gamePlayerRepo:    gamePlayerRepo,
		matchRepo:         matchRepo,
		registry:          registry,
		membershipService: membershipService,
		queryService:      queryService,
		joinService:       joinService,
		timeFunc:          timeFunc,
	}
}

// NewGameCreationServiceWithLoadedGames creates a new GameCreationService instance and loads all existing games from the repository
func NewGameCreationServiceWithLoadedGames(gameRepo repositories.GameRepository, gameCodeRepo repositories.GameCodeRepository, gamePlayerRepo repositories.GamePlayerRepository, betRepo repositories.BetRepository, matchRepo repositories.MatchRepository, watcher MatchWatcherService) (GameCreationServiceInterface, error) {
	return NewGameCreationServiceWithLoadedGamesAndTimeFunc(gameRepo, gameCodeRepo, gamePlayerRepo, betRepo, matchRepo, watcher, time.Now)
}

// NewGameCreationServiceWithLoadedGamesAndTimeFunc creates a new GameCreationService instance with a custom time function and loads all existing games from the repository
func NewGameCreationServiceWithLoadedGamesAndTimeFunc(gameRepo repositories.GameRepository, gameCodeRepo repositories.GameCodeRepository, gamePlayerRepo repositories.GamePlayerRepository, betRepo repositories.BetRepository, matchRepo repositories.MatchRepository, watcher MatchWatcherService, timeFunc func() time.Time) (GameCreationServiceInterface, error) {
	// Create the underlying services
	registry := NewGameServiceRegistry(gameRepo, betRepo, gamePlayerRepo, watcher)
	membershipService := NewGameMembershipService(gamePlayerRepo, gameRepo, gameCodeRepo, registry, watcher)
	queryService := NewGameQueryService(gameRepo, gamePlayerRepo, gameCodeRepo, betRepo)
	joinService := NewGameJoinService(gameCodeRepo, gameRepo, membershipService, registry, timeFunc)

	// Load all games via registry
	if err := registry.LoadAll(); err != nil {
		return nil, err
	}

	return &GameCreationService{
		gameRepo:          gameRepo,
		gameCodeRepo:      gameCodeRepo,
		gamePlayerRepo:    gamePlayerRepo,
		matchRepo:         matchRepo,
		registry:          registry,
		membershipService: membershipService,
		queryService:      queryService,
		joinService:       joinService,
		timeFunc:          timeFunc,
	}, nil
}

// NewGameCreationServiceWithServices creates a GameCreationService with explicit service dependencies (preferred for new code)
func NewGameCreationServiceWithServices(
	gameRepo repositories.GameRepository,
	gameCodeRepo repositories.GameCodeRepository,
	gamePlayerRepo repositories.GamePlayerRepository,
	matchRepo repositories.MatchRepository,
	registry GameServiceRegistryInterface,
	membershipService GameMembershipServiceInterface,
	queryService GameQueryServiceInterface,
	joinService GameJoinServiceInterface,
	timeFunc func() time.Time,
) *GameCreationService {
	return &GameCreationService{
		gameRepo:          gameRepo,
		gameCodeRepo:      gameCodeRepo,
		gamePlayerRepo:    gamePlayerRepo,
		matchRepo:         matchRepo,
		registry:          registry,
		membershipService: membershipService,
		queryService:      queryService,
		joinService:       joinService,
		timeFunc:          timeFunc,
	}
}

// GetGameService returns a GameService by ID, but only if the player has access to it
func (s *GameCreationService) GetGameService(gameID string, player models.Player) (GameService, error) {
	// Check if player is in the game first
	isInGame, err := s.membershipService.IsPlayerInGame(gameID, player.GetID())
	if err != nil {
		return nil, fmt.Errorf("error checking game access: %v", err)
	}
	if !isInGame {
		return nil, ErrPlayerNotInGame
	}

	return s.registry.GetOrCreate(gameID)
}

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

	// Register the game service in the registry
	_, err = s.registry.GetOrCreate(gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to create game service: %v", err)
	}

	// Generate a unique code for the game
	code, err := s.generateUniqueCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique code: %v", err)
	}

	// Create game code with 6 months expiration
	expiresAt := s.timeFunc().AddDate(0, 6, 0)
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

// JoinGame delegates to GameJoinService
func (s *GameCreationService) JoinGame(code string, player models.Player) (*JoinGameResponse, error) {
	return s.joinService.JoinGame(code, player)
}

// GetPlayerGames delegates to GameQueryService
func (s *GameCreationService) GetPlayerGames(player models.Player) ([]PlayerGame, error) {
	return s.queryService.GetPlayerGames(player)
}

// LeaveGame delegates to GameMembershipService
func (s *GameCreationService) LeaveGame(gameID string, player models.Player) error {
	return s.membershipService.LeaveGame(gameID, player)
}

// CleanupExpiredCodes removes all expired game codes
func (s *GameCreationService) CleanupExpiredCodes() error {
	return s.gameCodeRepo.DeleteExpiredCodes()
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
