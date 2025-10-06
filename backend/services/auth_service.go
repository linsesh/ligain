package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"time"
)

// AuthServiceInterface defines the interface for authentication services
type AuthServiceInterface interface {
	Authenticate(ctx context.Context, req *models.AuthRequest) (*models.AuthResponse, error)
	AuthenticateGuest(ctx context.Context, displayName string) (*models.AuthResponse, error)
	ValidateToken(ctx context.Context, token string) (*models.PlayerData, error)
	RefreshToken(ctx context.Context, expiredToken string) (*models.AuthResponse, error)
	GetAuthTokenDirectly(ctx context.Context, token string) (*models.AuthToken, error)
	RefreshTokenByPlayerID(ctx context.Context, playerID string) (*models.AuthResponse, error)
	Logout(ctx context.Context, token string) error
	CleanupExpiredTokens(ctx context.Context) error
	GetOrCreatePlayer(ctx context.Context, verifiedUser map[string]interface{}, provider string, displayName string) (*models.PlayerData, error)
	UpdateDisplayName(ctx context.Context, playerID string, newDisplayName string) (*models.PlayerData, error)
	DeleteAccount(ctx context.Context, playerID string) error
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
		return nil, &models.UnsupportedProviderError{Provider: req.Provider}
	}

	// Verify the OAuth token
	verifiedUser, err := s.oauthVerifier.VerifyToken(ctx, req.Provider, req.Token)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to verify token: %v", err)}
	}

	// Get or create player (separate concern)
	player, err := s.GetOrCreatePlayer(ctx, verifiedUser, req.Provider, req.Name)
	if err != nil {
		// If it's a NeedDisplayNameError, propagate it for two-step flow
		var needNameErr *models.NeedDisplayNameError
		if errors.As(err, &needNameErr) {
			return nil, needNameErr
		}
		return nil, err
	}

	// Generate authentication token
	token, err := s.generateAuthToken(ctx, player.ID)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to generate auth token: %v", err)}
	}

	return &models.AuthResponse{
		Player: *player,
		Token:  token,
	}, nil
}

// AuthenticateGuest handles guest authentication
func (s *AuthService) AuthenticateGuest(ctx context.Context, displayName string) (*models.AuthResponse, error) {
	if displayName == "" {
		return nil, &models.InvalidDisplayNameError{Reason: "display name cannot be empty for guest authentication"}
	}

	// For guests, we still need some way to prevent abuse with identical names
	// Check if this exact display name is already taken by another guest user
	existingPlayerByName, err := s.playerRepo.GetPlayerByName(ctx, displayName)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to check existing player by name: %v", err)}
	}
	if existingPlayerByName != nil {
		// If the player exists and is a guest (no provider and no email), allow re-authentication
		if existingPlayerByName.Provider == nil && existingPlayerByName.Email == nil {
			// Generate authentication token for existing guest
			token, err := s.generateAuthToken(ctx, existingPlayerByName.ID)
			if err != nil {
				return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to generate auth token: %v", err)}
			}
			return &models.AuthResponse{
				Player: *existingPlayerByName,
				Token:  token,
			}, nil
		}
		// If it's an OAuth user with the same display name, still allow guest creation
		// since display names are no longer unique
	}

	// Create new guest player (no email, provider, or providerID)
	player := &models.PlayerData{
		Name:      displayName,
		CreatedAt: &time.Time{},
		UpdatedAt: &time.Time{},
	}

	err = s.playerRepo.CreatePlayer(ctx, player)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to create guest player: %v", err)}
	}

	// Generate authentication token
	token, err := s.generateAuthToken(ctx, player.ID)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to generate auth token: %v", err)}
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
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to get auth token: %v", err)}
	}

	if authToken == nil {
		return nil, &models.PlayerNotFoundError{Reason: "invalid token"}
	}

	if s.timeFunc().After(authToken.ExpiresAt) {
		// Token expired, but don't delete it yet - let the refresh mechanism handle it
		return nil, &models.TokenExpiredError{Reason: "token expired"}
	}

	player, err := s.playerRepo.GetPlayerByID(ctx, authToken.PlayerID)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to get player by ID: %v", err)}
	}

	if player == nil {
		return nil, &models.PlayerNotFoundError{Reason: "player not found for token"}
	}

	return player, nil
}

// RefreshToken refreshes an authentication token (expired or not)
func (s *AuthService) RefreshToken(ctx context.Context, token string) (*models.AuthResponse, error) {
	authToken, err := s.playerRepo.GetAuthToken(ctx, token)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to get auth token for refresh: %v", err)}
	}

	if authToken == nil {
		return nil, &models.PlayerNotFoundError{Reason: "invalid token for refresh"}
	}

	// Get the player data first to ensure they exist
	player, err := s.playerRepo.GetPlayerByID(ctx, authToken.PlayerID)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to get player by ID for refresh: %v", err)}
	}

	if player == nil {
		return nil, &models.PlayerNotFoundError{Reason: "player not found for refresh"}
	}

	// Generate a new token
	newToken, err := s.generateAuthToken(ctx, authToken.PlayerID)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to generate new auth token: %v", err)}
	}

	// Delete the old token
	err = s.playerRepo.DeleteAuthToken(ctx, token)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to delete old auth token: %v", err)}
	}

	return &models.AuthResponse{
		Player: *player,
		Token:  newToken,
	}, nil
}

// GetAuthTokenDirectly gets an auth token without validation (for internal use)
func (s *AuthService) GetAuthTokenDirectly(ctx context.Context, token string) (*models.AuthToken, error) {
	return s.playerRepo.GetAuthToken(ctx, token)
}

// RefreshTokenByPlayerID refreshes a token using player ID directly
func (s *AuthService) RefreshTokenByPlayerID(ctx context.Context, playerID string) (*models.AuthResponse, error) {
	// Get the player data first to ensure they exist
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to get player by ID for refresh: %v", err)}
	}

	if player == nil {
		return nil, &models.PlayerNotFoundError{Reason: "player not found for refresh"}
	}

	// Generate a new token
	newToken, err := s.generateAuthToken(ctx, playerID)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to generate new auth token: %v", err)}
	}

	return &models.AuthResponse{
		Player: *player,
		Token:  newToken,
	}, nil
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
	// Extract and validate user information from OAuth provider
	userInfo, err := s.extractUserInfoFromOAuth(verifiedUser, provider)
	if err != nil {
		return nil, err
	}

	// Try to find existing player by provider ID
	existingPlayer, err := s.findExistingPlayerByProvider(ctx, provider, userInfo.providerID)
	if err != nil {
		return nil, err
	}

	if existingPlayer != nil {
		return s.handleExistingPlayer(ctx, existingPlayer, displayName)
	}

	// Try to find existing player by email (for account linking)
	existingPlayerByEmail, err := s.findExistingPlayerByEmail(ctx, userInfo.email)
	if err != nil {
		return nil, err
	}

	if existingPlayerByEmail != nil {
		return s.linkExistingAccount(ctx, existingPlayerByEmail, provider, userInfo.providerID, displayName)
	}

	// Create new player
	return s.createNewPlayer(ctx, provider, userInfo, displayName)
}

// userInfo holds extracted information from OAuth provider
type userInfo struct {
	providerID string
	email      string
	name       string
}

// extractUserInfoFromOAuth extracts and validates user information from OAuth provider
func (s *AuthService) extractUserInfoFromOAuth(verifiedUser map[string]interface{}, provider string) (*userInfo, error) {
	providerID, ok := verifiedUser["id"].(string)
	if !ok {
		return nil, &models.GeneralAuthError{Reason: "invalid user ID from OAuth provider"}
	}

	verifiedName := ""
	if name, ok := verifiedUser["name"].(string); ok {
		verifiedName = name
	}

	// Handle email - it's optional for Apple Sign-In
	var email string
	if emailVal, ok := verifiedUser["email"].(string); ok {
		email = emailVal
	}

	// For Google Sign-In, email is required
	if provider == "google" && email == "" {
		return nil, &models.GeneralAuthError{Reason: "email is required for Google Sign-In"}
	}

	return &userInfo{
		providerID: providerID,
		email:      email,
		name:       verifiedName,
	}, nil
}

// findExistingPlayerByProvider looks for an existing player with the same provider and provider ID
func (s *AuthService) findExistingPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error) {
	existingPlayer, err := s.playerRepo.GetPlayerByProvider(ctx, provider, providerID)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to get player by provider: %v", err)}
	}
	return existingPlayer, nil
}

// handleExistingPlayer handles the case where a player already exists with this provider
func (s *AuthService) handleExistingPlayer(ctx context.Context, existingPlayer *models.PlayerData, displayName string) (*models.PlayerData, error) {
	// Only update name if provided and different
	if displayName != "" && displayName != existingPlayer.Name {
		existingPlayer.Name = displayName
		existingPlayer.UpdatedAt = &time.Time{}
		*existingPlayer.UpdatedAt = s.timeFunc()

		err := s.playerRepo.UpdatePlayer(ctx, existingPlayer)
		if err != nil {
			return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to update player: %v", err)}
		}
	}

	return existingPlayer, nil
}

// findExistingPlayerByEmail looks for an existing player with the same email (for account linking)
func (s *AuthService) findExistingPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error) {
	if email == "" {
		return nil, nil // No email to search by
	}

	existingPlayer, err := s.playerRepo.GetPlayerByEmail(ctx, email)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to get player by email: %v", err)}
	}
	return existingPlayer, nil
}

// linkExistingAccount links an existing account to a new OAuth provider
func (s *AuthService) linkExistingAccount(ctx context.Context, existingPlayer *models.PlayerData, provider, providerID, displayName string) (*models.PlayerData, error) {
	existingPlayer.Provider = &provider
	existingPlayer.ProviderID = &providerID
	existingPlayer.UpdatedAt = &time.Time{}
	*existingPlayer.UpdatedAt = s.timeFunc()

	// Update name if provided and different
	if displayName != "" && displayName != existingPlayer.Name {
		existingPlayer.Name = displayName
	}

	err := s.playerRepo.UpdatePlayer(ctx, existingPlayer)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to update player: %v", err)}
	}

	return existingPlayer, nil
}

// createNewPlayer creates a new player account
func (s *AuthService) createNewPlayer(ctx context.Context, provider string, userInfo *userInfo, displayName string) (*models.PlayerData, error) {
	// Validate display name requirements
	if err := s.validateDisplayNameForNewUser(displayName, userInfo.name); err != nil {
		return nil, err
	}

	// Create new player
	player := &models.PlayerData{
		Name:       displayName,
		Provider:   &provider,
		ProviderID: &userInfo.providerID,
	}

	// Only set email if provided (Apple might not provide email)
	if userInfo.email != "" {
		player.Email = &userInfo.email
	}

	err := s.playerRepo.CreatePlayer(ctx, player)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to create player: %v", err)}
	}

	return player, nil
}

// validateDisplayNameForNewUser validates display name requirements for new users
func (s *AuthService) validateDisplayNameForNewUser(displayName, suggestedName string) error {
	if displayName == "" {
		return &models.NeedDisplayNameError{
			Reason:        "display name is required for new users",
			SuggestedName: suggestedName,
		}
	}

	if len(displayName) < 2 {
		return &models.NeedDisplayNameError{
			Reason:        "display name must be at least 2 characters long",
			SuggestedName: suggestedName,
		}
	}

	if len(displayName) > 20 {
		return &models.NeedDisplayNameError{
			Reason:        "display name must be 20 characters or less",
			SuggestedName: suggestedName,
		}
	}

	return nil
}

// UpdateDisplayName updates a player's display name
func (s *AuthService) UpdateDisplayName(ctx context.Context, playerID string, newDisplayName string) (*models.PlayerData, error) {
	if newDisplayName == "" {
		return nil, &models.InvalidDisplayNameError{Reason: "display name cannot be empty"}
	}

	if len(newDisplayName) < 2 {
		return nil, &models.InvalidDisplayNameError{Reason: "display name must be at least 2 characters long"}
	}

	if len(newDisplayName) > 20 {
		return nil, &models.InvalidDisplayNameError{Reason: "display name must be 20 characters or less"}
	}

	// Get the current player
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return nil, &models.PlayerNotFoundError{Reason: fmt.Sprintf("player not found: %v", err)}
	}

	if player == nil {
		return nil, &models.PlayerNotFoundError{Reason: "player not found"}
	}

	// Update the display name
	player.Name = newDisplayName
	player.UpdatedAt = &time.Time{}
	*player.UpdatedAt = s.timeFunc()

	err = s.playerRepo.UpdatePlayer(ctx, player)
	if err != nil {
		return nil, &models.GeneralAuthError{Reason: fmt.Sprintf("failed to update display name: %v", err)}
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

// DeleteAccount permanently deletes a player account and all associated data
func (s *AuthService) DeleteAccount(ctx context.Context, playerID string) error {
	// First verify the player exists
	player, err := s.playerRepo.GetPlayerByID(ctx, playerID)
	if err != nil {
		return &models.PlayerNotFoundError{Reason: "player not found"}
	}

	if player == nil {
		return &models.PlayerNotFoundError{Reason: "player not found"}
	}

	// Delete the player (this will cascade delete all related data due to our foreign key constraints)
	// This includes: auth_tokens, bets, scores, game_player entries
	err = s.playerRepo.DeletePlayer(ctx, playerID)
	if err != nil {
		return &models.GeneralAuthError{Reason: fmt.Sprintf("failed to delete player account: %v", err)}
	}

	return nil
}
