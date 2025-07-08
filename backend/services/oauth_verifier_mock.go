package services

import (
	"context"
	"errors"
	"fmt"
)

// MockOAuthVerifier implements OAuthVerifierInterface for testing
type MockOAuthVerifier struct {
	verifiedUsers map[string]map[string]interface{}
	shouldFail    bool
}

// NewMockOAuthVerifier creates a new mock OAuth verifier for testing
func NewMockOAuthVerifier() *MockOAuthVerifier {
	return &MockOAuthVerifier{
		verifiedUsers: map[string]map[string]interface{}{
			"google": {
				"id":    "google_123",
				"email": "test@example.com",
				"name":  "Test Google User",
			},
			"apple": {
				"id":    "apple_123",
				"email": "test@example.com",
				"name":  "Test Apple User",
			},
		},
		shouldFail: false,
	}
}

// NewMockOAuthVerifierWithFailure creates a mock verifier that will fail verification
func NewMockOAuthVerifierWithFailure() *MockOAuthVerifier {
	return &MockOAuthVerifier{
		verifiedUsers: make(map[string]map[string]interface{}),
		shouldFail:    true,
	}
}

// SetVerifiedUser allows setting custom verified user data for testing
func (m *MockOAuthVerifier) SetVerifiedUser(provider string, userInfo map[string]interface{}) {
	if m.verifiedUsers == nil {
		m.verifiedUsers = make(map[string]map[string]interface{})
	}
	m.verifiedUsers[provider] = userInfo
}

// SetShouldFail allows controlling whether the mock should fail
func (m *MockOAuthVerifier) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// VerifyToken implements OAuthVerifierInterface for testing
func (m *MockOAuthVerifier) VerifyToken(ctx context.Context, provider, token string) (map[string]interface{}, error) {
	if m.shouldFail {
		return nil, errors.New("mock verification failed")
	}

	// Check for mock tokens first
	if token != "" && (token == "mock_google_token_123" || token == "mock_apple_token_123") {
		switch provider {
		case "google":
			return map[string]interface{}{
				"id":    "google_user_123",
				"email": "test@google.com",
				"name":  "Test Google User",
			}, nil
		case "apple":
			return map[string]interface{}{
				"id":    "apple_user_123",
				"email": "test@apple.com",
				"name":  "Test Apple User",
			}, nil
		}
	}

	userInfo, exists := m.verifiedUsers[provider]
	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	return userInfo, nil
}
