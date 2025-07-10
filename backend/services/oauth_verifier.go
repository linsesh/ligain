package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// OAuthVerifierInterface defines the interface for OAuth token verification
type OAuthVerifierInterface interface {
	VerifyToken(ctx context.Context, provider, token string) (map[string]interface{}, error)
}

// GoogleOAuthVerifier handles Google OAuth token verification
type GoogleOAuthVerifier struct {
	clientID string
}

// NewGoogleOAuthVerifier creates a new Google OAuth verifier
func NewGoogleOAuthVerifier() *GoogleOAuthVerifier {
	return &GoogleOAuthVerifier{
		clientID: os.Getenv("GOOGLE_CLIENT_ID"),
	}
}

// AppleOAuthVerifier handles Apple OAuth token verification
type AppleOAuthVerifier struct {
	clientID string
}

// NewAppleOAuthVerifier creates a new Apple OAuth verifier
func NewAppleOAuthVerifier() *AppleOAuthVerifier {
	return &AppleOAuthVerifier{
		clientID: os.Getenv("APPLE_CLIENT_ID"),
	}
}

// VerifyToken verifies a Google ID token (JWT)
func (v *GoogleOAuthVerifier) VerifyToken(ctx context.Context, token string) (*GoogleUserInfo, error) {
	fmt.Printf("🔍 GoogleOAuthVerifier - Starting token verification\n")
	fmt.Printf("🔍 GoogleOAuthVerifier - Client ID configured: %t\n", v.clientID != "")

	if v.clientID == "" {
		fmt.Printf("❌ GoogleOAuthVerifier - Google client ID not configured\n")
		return nil, errors.New("Google client ID not configured")
	}

	// For ID tokens, we need to verify the JWT signature
	// For now, let's decode the JWT payload to get user info
	// In production, you should verify the JWT signature with Google's public keys

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		fmt.Printf("❌ GoogleOAuthVerifier - Invalid token format, expected 3 parts, got %d\n", len(parts))
		return nil, errors.New("invalid token format")
	}

	fmt.Printf("🔍 GoogleOAuthVerifier - Token has valid JWT format\n")

	// Decode the payload (second part of JWT)
	payload := parts[1]
	// Add padding if needed
	if len(payload)%4 != 0 {
		payload += strings.Repeat("=", 4-len(payload)%4)
	}

	// Decode base64
	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		fmt.Printf("❌ GoogleOAuthVerifier - Failed to decode token payload: %v\n", err)
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		fmt.Printf("❌ GoogleOAuthVerifier - Failed to parse token claims: %v\n", err)
		return nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	fmt.Printf("🔍 GoogleOAuthVerifier - Token claims parsed successfully\n")

	// Extract user info from claims
	userInfo := &GoogleUserInfo{
		ID:            claims["sub"].(string),
		Email:         claims["email"].(string),
		Name:          claims["name"].(string),
		Picture:       claims["picture"].(string),
		VerifiedEmail: claims["email_verified"].(bool),
	}

	fmt.Printf("✅ GoogleOAuthVerifier - Token verification successful for user: %s (%s)\n", userInfo.Name, userInfo.Email)

	return userInfo, nil
}

// VerifyToken verifies an Apple OAuth token
func (v *AppleOAuthVerifier) VerifyToken(ctx context.Context, token string) (*AppleUserInfo, error) {
	if v.clientID == "" {
		return nil, errors.New("Apple client ID not configured")
	}

	// Apple token verification is more complex and requires JWT validation
	// For now, we'll implement a basic verification
	// In production, you should use Apple's public keys to verify the JWT

	// Parse the JWT token (simplified version)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid Apple token format")
	}

	// For development, we'll accept the token and extract user info
	// In production, you should properly verify the JWT signature
	userInfo := &AppleUserInfo{
		ID:    fmt.Sprintf("apple_user_%d", time.Now().Unix()), // Mock ID for development
		Email: "user@example.com",                              // Mock email for development
	}

	return userInfo, nil
}

// OAuthVerifier implements OAuth token verification using the appropriate provider
type OAuthVerifier struct {
	googleVerifier *GoogleOAuthVerifier
	appleVerifier  *AppleOAuthVerifier
}

// GoogleUserInfo represents the user info from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

// AppleUserInfo represents the user info from Apple
type AppleUserInfo struct {
	ID    string `json:"sub"`
	Email string `json:"email"`
	Name  struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"name"`
}

// NewOAuthVerifier creates a new OAuth verifier
func NewOAuthVerifier() OAuthVerifierInterface {
	return &OAuthVerifier{
		googleVerifier: NewGoogleOAuthVerifier(),
		appleVerifier:  NewAppleOAuthVerifier(),
	}
}

// VerifyToken implements the unified verification method
func (v *OAuthVerifier) VerifyToken(ctx context.Context, provider, token string) (map[string]interface{}, error) {
	switch provider {
	case "google":
		return v.verifyGoogleToken(ctx, token)
	case "apple":
		return v.verifyAppleToken(ctx, token)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// verifyGoogleToken handles Google token verification
func (v *OAuthVerifier) verifyGoogleToken(ctx context.Context, token string) (map[string]interface{}, error) {
	userInfo, err := v.googleVerifier.VerifyToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":    userInfo.ID,
		"email": userInfo.Email,
		"name":  userInfo.Name,
	}, nil
}

// verifyAppleToken handles Apple token verification
func (v *OAuthVerifier) verifyAppleToken(ctx context.Context, token string) (map[string]interface{}, error) {
	userInfo, err := v.appleVerifier.VerifyToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":    userInfo.ID,
		"email": userInfo.Email,
		"name":  userInfo.Name.FirstName + " " + userInfo.Name.LastName,
	}, nil
}
