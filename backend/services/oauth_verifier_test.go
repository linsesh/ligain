package services

import (
	"context"
	"testing"
)

// Test the real OAuth verifier with mock tokens (these will now fail as expected)
func TestOAuthVerifier_VerifyToken_Google_Real(t *testing.T) {
	verifier := NewOAuthVerifier()

	ctx := context.Background()
	provider := "google"
	token := "mock_google_token_123"

	// This should now fail because we removed mock handling from production code
	_, err := verifier.VerifyToken(ctx, provider, token)
	if err == nil {
		t.Error("Expected error for mock token in real verifier")
	}

	// The token format is invalid, so we expect an invalid token format error
	// (validation happens before checking client ID configuration)
}

func TestOAuthVerifier_VerifyToken_Apple_Real(t *testing.T) {
	verifier := NewOAuthVerifier()

	ctx := context.Background()
	provider := "apple"
	token := "mock_apple_token_123"

	// This should now fail because we removed mock handling from production code
	_, err := verifier.VerifyToken(ctx, provider, token)
	if err == nil {
		t.Error("Expected error for mock token in real verifier")
	}

	// The token format is invalid, so we expect an invalid Apple token format error
	// (validation happens before checking client ID configuration)
}

// Test the mock OAuth verifier
func TestMockOAuthVerifier_VerifyToken_Google(t *testing.T) {
	verifier := NewMockOAuthVerifier()

	ctx := context.Background()
	provider := "google"
	token := "mock_google_token_123"

	result, err := verifier.VerifyToken(ctx, provider, token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify mock response structure
	if result["id"] != "google_user_123" {
		t.Errorf("Expected ID 'google_user_123', got %v", result["id"])
	}

	if result["email"] != "test@google.com" {
		t.Errorf("Expected email 'test@google.com', got %v", result["email"])
	}

	if result["name"] != "Test Google User" {
		t.Errorf("Expected name 'Test Google User', got %v", result["name"])
	}
}

func TestMockOAuthVerifier_VerifyToken_Apple(t *testing.T) {
	verifier := NewMockOAuthVerifier()

	ctx := context.Background()
	provider := "apple"
	token := "mock_apple_token_123"

	result, err := verifier.VerifyToken(ctx, provider, token)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify mock response structure
	if result["id"] != "apple_user_123" {
		t.Errorf("Expected ID 'apple_user_123', got %v", result["id"])
	}

	if result["email"] != "test@apple.com" {
		t.Errorf("Expected email 'test@apple.com', got %v", result["email"])
	}

	if result["name"] != "Test Apple User" {
		t.Errorf("Expected name 'Test Apple User', got %v", result["name"])
	}
}

func TestOAuthVerifier_VerifyToken_UnsupportedProvider(t *testing.T) {
	verifier := NewOAuthVerifier()

	ctx := context.Background()
	provider := "facebook" // Unsupported provider
	token := "real_token"  // Not a mock token

	_, err := verifier.VerifyToken(ctx, provider, token)
	if err == nil {
		t.Error("Expected error for unsupported provider")
	}

	expectedError := "unsupported provider: facebook"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestOAuthVerifier_VerifyToken_NonMockToken(t *testing.T) {
	verifier := NewOAuthVerifier()

	ctx := context.Background()
	provider := "google"
	token := "real_token_123" // Not a mock token

	_, err := verifier.VerifyToken(ctx, provider, token)
	if err == nil {
		t.Error("Expected error for non-mock token")
	}

	// The token format is invalid, so we expect an invalid token format error
	// (validation happens before checking client ID configuration)
}

func TestGoogleOAuthVerifier_VerifyToken_NoClientID(t *testing.T) {
	verifier := &GoogleOAuthVerifier{
		clientID: "", // No client ID
	}

	ctx := context.Background()
	token := "test_token"

	_, err := verifier.VerifyToken(ctx, token)
	if err == nil {
		t.Error("Expected error for missing client ID")
	}

	if err.Error() != "Google client ID not configured" {
		t.Errorf("Expected 'Google client ID not configured' error, got %s", err.Error())
	}
}

func TestAppleOAuthVerifier_VerifyToken_NoClientID(t *testing.T) {
	verifier := &AppleOAuthVerifier{
		clientID: "", // No client ID
	}

	ctx := context.Background()
	token := "test_token"

	_, err := verifier.VerifyToken(ctx, token)
	if err == nil {
		t.Error("Expected error for missing client ID")
	}

	if err.Error() != "Apple client ID not configured" {
		t.Errorf("Expected 'Apple client ID not configured' error, got %s", err.Error())
	}
}

func TestAppleOAuthVerifier_VerifyToken_InvalidFormat(t *testing.T) {
	verifier := &AppleOAuthVerifier{
		clientID: "test_client_id",
	}

	ctx := context.Background()
	token := "invalid_token_format" // Not a JWT format

	_, err := verifier.VerifyToken(ctx, token)
	if err == nil {
		t.Error("Expected error for invalid token format")
	}

	if err.Error() != "invalid Apple token format" {
		t.Errorf("Expected 'invalid Apple token format' error, got %s", err.Error())
	}
}

// Test the interface implementation
func TestOAuthVerifierInterface_Implementation(t *testing.T) {
	var _ OAuthVerifierInterface = NewOAuthVerifier()
}

// Test that the verifier can be used with dependency injection
func TestOAuthVerifier_DependencyInjection(t *testing.T) {
	verifier := NewOAuthVerifier()

	// This should not panic and should return a valid interface
	if verifier == nil {
		t.Error("Expected non-nil verifier")
	}

	// Test that it implements the interface
	var interfaceVerifier OAuthVerifierInterface = verifier
	if interfaceVerifier == nil {
		t.Error("Expected interface implementation")
	}
}

// Test mock verifier failure scenarios
func TestMockOAuthVerifier_VerifyToken_Failure(t *testing.T) {
	verifier := NewMockOAuthVerifierWithFailure()

	ctx := context.Background()
	provider := "google"
	token := "mock_google_token_123"

	_, err := verifier.VerifyToken(ctx, provider, token)
	if err == nil {
		t.Error("Expected error for mock failure")
	}

	if err.Error() != "mock verification failed" {
		t.Errorf("Expected 'mock verification failed' error, got %s", err.Error())
	}
}

func TestMockOAuthVerifier_VerifyToken_UnsupportedProvider(t *testing.T) {
	verifier := NewMockOAuthVerifier()

	ctx := context.Background()
	provider := "facebook" // Unsupported provider
	token := "any_token"

	_, err := verifier.VerifyToken(ctx, provider, token)
	if err == nil {
		t.Error("Expected error for unsupported provider")
	}

	expectedError := "unsupported provider: facebook"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Test mock verifier configuration
func TestMockOAuthVerifier_Configuration(t *testing.T) {
	verifier := NewMockOAuthVerifier()

	// Test setting custom user data
	customUser := map[string]interface{}{
		"id":    "custom_123",
		"email": "custom@example.com",
		"name":  "Custom User",
	}
	verifier.SetVerifiedUser("custom_provider", customUser)

	ctx := context.Background()
	result, err := verifier.VerifyToken(ctx, "custom_provider", "any_token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result["id"] != "custom_123" {
		t.Errorf("Expected ID 'custom_123', got %v", result["id"])
	}

	// Test setting failure mode
	verifier.SetShouldFail(true)
	_, err = verifier.VerifyToken(ctx, "google", "any_token")
	if err == nil {
		t.Error("Expected error when shouldFail is true")
	}
}
