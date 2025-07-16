package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"time"
)

// AuthServiceInterface defines the interface for authentication services
type AuthServiceInterface interface {
	Authenticate(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error)
	AuthenticateGuest(ctx context.Context, displayName string) (*models.AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (*models.PlayerData, error)
	Logout(ctx context.Context, token string) error
	CleanupExpiredTokens(ctx context.Context) error
	GetOrCreatePlayer(ctx context.Context, verifiedUser map[string]interface{}, provider string, displayName string) (*models.PlayerData, error)
}

// AuthService implements authentication with Google or Apple
type AuthService struct {
	playerRepo    repositories.PlayerRepository
	oauthVerifier OAuthVerifierInterface
	timeFunc      func() time.Time // For testing - allows injection of frozen time
}

// NewAuthService creates a new AuthService instance
func NewAuthService(playerRepo repositories.PlayerRepository) *AuthService {
	return &AuthService{
		playerRepo:    playerRepo,
		oauthVerifier: NewOAuthVerifier(),
		timeFunc:      time.Now,
	}
}

// NewAuthServiceWithTimeFunc creates an AuthService with a custom time function for testing
func NewAuthServiceWithTimeFunc(playerRepo repositories.PlayerRepository, oauthVerifier OAuthVerifierInterface, timeFunc func() time.Time) *AuthService {
	return &AuthService{
		playerRepo:    playerRepo,
		oauthVerifier: oauthVerifier,
		timeFunc:      timeFunc,
	}
}

// Authenticate handles authentication with Google or Apple
// This function only verifies the token and generates an auth token
func (s *AuthService) Authenticate(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error) {
	if req.Provider != "google" && req.Provider != "apple" {
		return nil, errors.New("unsupported provider")
	}

	// Verify the OAuth token
	verifiedUser, err := s.oauthVerifier.VerifyToken(ctx, req.Provider, req.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Get or create player (separate concern)
	player, err := s.GetOrCreatePlayer(ctx, verifiedUser, req.Provider, req.Name)
	if err != nil {
		return nil, err
	}

	// Generate authentication token
	token, err := s.generateAuthToken(ctx, player.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Player: *player,
		Token:  token,
	}, nil
}

// AuthenticateGuest handles guest authentication
func (s *AuthService) AuthenticateGuest(ctx context.Context, displayName string) (*models.AuthResponse, error) {
	if displayName == "" {
		return nil, errors.New("display name cannot be empty for guest authentication")
	}

	// Check if display name is already taken
	existingPlayerByName, err := s.playerRepo.GetPlayerByName(ctx, displayName)
	if err != nil {
		return nil, err
	}
	if existingPlayerByName != nil {
		// If the player exists and is a guest (no provider and no email), allow re-authentication
		if existingPlayerByName.Provider == nil && existingPlayerByName.Email == nil {
			// Generate authentication token for existing guest
			token, err := s.generateAuthToken(ctx, existingPlayerByName.ID)
			if err != nil {
				return nil, err
			}
			return &models.AuthResponse{
				Player: *existingPlayerByName,
				Token:  token,
			}, nil
		}
		// Otherwise, name is taken by a real account
		return nil, errors.New("display name is already taken")
	}

	// Create new guest player (no email, provider, or providerID)
	player := &models.PlayerData{
		Name:      displayName,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}

	err = s.playerRepo.CreatePlayer(ctx, player)
	if err != nil {
		return nil, err
	}

	// Generate authentication token
	token, err := s.generateAuthToken(ctx, player.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Player: *player,
		Token:  token,
	}, nil
}

// ValidateToken validates an authentication token
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*models.PlayerData, error) {
	authToken, err := s.playerRepo.GetAuthToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if authToken == nil {
		return nil, errors.New("invalid token")
	}

	if s.timeFunc().After(authToken.ExpiresAt) {
		// Token expired, delete it
		s.playerRepo.DeleteAuthToken(ctx, token)
		return nil, errors.New("token expired")
	}

	player, err := s.playerRepo.GetPlayerByID(ctx, authToken.PlayerID)
	if err != nil {
		return nil, err
	}

	if player == nil {
		return nil, errors.New("player not found")
	}

	return player, nil
}

// Logout invalidates a token
func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.playerRepo.DeleteAuthToken(ctx, token)
}

// CleanupExpiredTokens removes expired tokens from the database
func (s *AuthService) CleanupExpiredTokens(ctx context.Context) error {
	return s.playerRepo.DeleteExpiredTokens(ctx)
}

// GetOrCreatePlayer handles player creation, updates, and account linking
// This is a separate concern from authentication
func (s *AuthService) GetOrCreatePlayer(ctx context.Context, verifiedUser map[string]interface{}, provider string, displayName string) (*models.PlayerData, error) {
	verifiedEmail := verifiedUser["email"].(string)
	verifiedID := verifiedUser["id"].(string)

	if verifiedEmail == "" {
		return nil, errors.New("invalid user information from OAuth provider")
	}

	// First, check if player already exists by provider ID
	existingPlayer, err := s.playerRepo.GetPlayerByProvider(ctx, provider, verifiedID)
	if err != nil {
		return nil, err
	}

	var player *models.PlayerData
	if existingPlayer != nil {
		// Player exists with this provider, return existing player
		// Only update name if provided and different
		if displayName != "" && displayName != existingPlayer.Name {
			player = existingPlayer
			player.Name = displayName
			player.UpdatedAt = &time.Time{}
			*player.UpdatedAt = s.timeFunc()

			err = s.playerRepo.UpdatePlayer(ctx, player)
			if err != nil {
				return nil, err
			}
		} else {
			player = existingPlayer
		}
	} else {
		// Check if player exists by email (different provider)
		existingPlayerByEmail, err := s.playerRepo.GetPlayerByEmail(ctx, verifiedEmail)
		if err != nil {
			return nil, err
		}

		if existingPlayerByEmail != nil {
			// Player exists with different provider, link accounts
			player = existingPlayerByEmail
			player.Provider = &provider
			player.ProviderID = &verifiedID
			player.UpdatedAt = &time.Time{}
			*player.UpdatedAt = s.timeFunc()

			// Update name if provided and different
			if displayName != "" && displayName != player.Name {
				player.Name = displayName
			}

			err = s.playerRepo.UpdatePlayer(ctx, player)
			if err != nil {
				return nil, err
			}
		} else {
			// New player - check if display name is available
			if displayName == "" {
				return nil, errors.New("display name is required for new users")
			}

			// Check if display name is already taken
			existingPlayerByName, err := s.playerRepo.GetPlayerByName(ctx, displayName)
			if err != nil {
				return nil, err
			}
			if existingPlayerByName != nil {
				return nil, errors.New("display name is already taken")
			}

			// Create new player
			player = &models.PlayerData{
				Name:       displayName,
				Email:      &verifiedEmail,
				Provider:   &provider,
				ProviderID: &verifiedID,
			}

			err = s.playerRepo.CreatePlayer(ctx, player)
			if err != nil {
				return nil, err
			}
		}
	}

	return player, nil
}

// generateAuthToken creates a new authentication token
func (s *AuthService) generateAuthToken(ctx context.Context, playerID string) (string, error) {
	// Generate a random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	// Create auth token record
	authToken := &models.AuthToken{
		PlayerID:  playerID,
		Token:     token,
		ExpiresAt: s.timeFunc().Add(24 * time.Hour), // 24 hour expiry
	}

	// Store in database
	err := s.playerRepo.CreateAuthToken(ctx, authToken)
	if err != nil {
		return "", err
	}

	return token, nil
}
