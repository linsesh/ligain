package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/rules"
	"math/big"
	"time"
)

// GameCreationServiceInterface defines the interface for game creation services
type GameCreationServiceInterface interface {
	CreateGame(req *CreateGameRequest) (*CreateGameResponse, error)
	JoinGame(code string, player models.Player) (*JoinGameResponse, error)
	GetPlayerGames(player models.Player) ([]PlayerGame, error)
	CleanupExpiredCodes() error
}

// GameCreationService handles game creation with unique codes
type GameCreationService struct {
	gameRepo     repositories.GameRepository
	gameCodeRepo repositories.GameCodeRepository
	matchRepo    repositories.MatchRepository
}

// NewGameCreationService creates a new GameCreationService instance
func NewGameCreationService(gameRepo repositories.GameRepository, gameCodeRepo repositories.GameCodeRepository, matchRepo repositories.MatchRepository) GameCreationServiceInterface {
	return &GameCreationService{
		gameRepo:     gameRepo,
		gameCodeRepo: gameCodeRepo,
		matchRepo:    matchRepo,
	}
}

// CreateGameRequest represents the request to create a new game
type CreateGameRequest struct {
	SeasonYear      string `json:"seasonYear" binding:"required"`
	CompetitionName string `json:"competitionName" binding:"required"`
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

// PlayerGame represents a game that a player is part of
type PlayerGame struct {
	GameID          string `json:"gameId"`
	SeasonYear      string `json:"seasonYear"`
	CompetitionName string `json:"competitionName"`
	Status          string `json:"status"`
}

var (
	ErrInvalidCompetition = errors.New("only 'Ligue 1' is supported as competition name")
	ErrInvalidSeasonYear  = errors.New("only '2025/2026' is supported as season year")
)

// CreateGame creates a new game with a unique 4-character code
func (s *GameCreationService) CreateGame(req *CreateGameRequest) (*CreateGameResponse, error) {
	// Validate competition name - only Ligue 1 is supported
	if req.CompetitionName != "Ligue 1" {
		return nil, ErrInvalidCompetition
	}
	// Validate season year - only 2025/2026 is supported
	if req.SeasonYear != "2025/2026" {
		return nil, ErrInvalidSeasonYear
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
		[]models.Player{}, // No players initially
		matches,           // Load all matches for this competition/season
		&rules.ScorerOriginal{},
	)

	// Save the game to get its ID
	gameID, err := s.gameRepo.CreateGame(game)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %v", err)
	}

	// Generate a unique code for the game
	code, err := s.generateUniqueCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique code: %v", err)
	}

	// Create game code with 1 week expiration
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
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

	// Add the player to the game
	err = s.addPlayerToGame(gameCode.GameID, player)
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

// addPlayerToGame adds a player to a game if they're not already in it
func (s *GameCreationService) addPlayerToGame(gameID string, player models.Player) error {
	// For now, we'll assume the player is added successfully
	// In a real implementation, you'd check if the player is already in the game
	// and add them if they're not
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
	// For now, we'll return an empty list since we need to implement
	// the logic to find games where the player has made bets
	// This would typically involve querying the bet repository
	// to find all games where the player has bets

	// TODO: Implement proper logic to get games for player
	// This would involve:
	// 1. Querying the bet repository to find all game IDs where the player has bets
	// 2. For each game ID, getting the game details from the game repository
	// 3. Converting to PlayerGame structs

	return []PlayerGame{}, nil
}
