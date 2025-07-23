package models

import "fmt"

// InvalidDisplayNameError is returned when a display name is invalid
// (e.g., empty, too short, too long)
type InvalidDisplayNameError struct {
	Reason string
}

func (e *InvalidDisplayNameError) Error() string {
	return fmt.Sprintf("invalid display name: %s", e.Reason)
}

// PlayerNotFoundError is returned when a player cannot be found
// (e.g., by ID, token, etc.)
type PlayerNotFoundError struct {
	Reason string
}

func (e *PlayerNotFoundError) Error() string {
	return fmt.Sprintf("player not found: %s", e.Reason)
}

// TokenExpiredError is returned when an auth token is expired
type TokenExpiredError struct {
	Reason string
}

func (e *TokenExpiredError) Error() string {
	return fmt.Sprintf("token expired: %s", e.Reason)
}

// UnsupportedProviderError is returned when an unsupported auth provider is used
type UnsupportedProviderError struct {
	Provider string
}

func (e *UnsupportedProviderError) Error() string {
	return fmt.Sprintf("unsupported provider: %s", e.Provider)
}

// GeneralAuthError is a fallback for other auth errors
type GeneralAuthError struct {
	Reason string
}

func (e *GeneralAuthError) Error() string {
	return fmt.Sprintf("auth error: %s", e.Reason)
}

// NeedDisplayNameError is returned when a new user needs to provide a display name to complete registration
// Used for two-step authentication flow
// SuggestedName is optional and can be used to pre-fill the modal
// Reason is a human-readable message (e.g., validation error)
type NeedDisplayNameError struct {
	Reason        string
	SuggestedName string
}

func (e *NeedDisplayNameError) Error() string {
	return e.Reason
}
