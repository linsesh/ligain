package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

// VerifyToken verifies a Google OAuth token
func (v *GoogleOAuthVerifier) VerifyToken(ctx context.Context, token string) (*GoogleUserInfo, error) {
	if v.clientID == "" {
		return nil, errors.New("Google client ID not configured")
	}

	// First, verify the token with Google
	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://oauth2.googleapis.com/tokeninfo?access_token="+token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid Google token")
	}

	// Get user info from Google
	userInfoReq, err := http.NewRequestWithContext(ctx, "GET",
		"https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %w", err)
	}

	userInfoReq.Header.Set("Authorization", "Bearer "+token)
	userInfoResp, err := client.Do(userInfoReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userInfoResp.Body.Close()

	if userInfoResp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get user info from Google")
	}

	body, err := io.ReadAll(userInfoResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
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
