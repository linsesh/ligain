package services

import (
	"context"
	"database/sql"
	"ligain/backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPlayerRepository for testing DeleteAccount
type MockPlayerRepositoryForDelete struct {
	mock.Mock
}

func (m *MockPlayerRepositoryForDelete) GetPlayer(playerId string) (models.Player, error) {
	args := m.Called(playerId)
	return args.Get(0).(models.Player), args.Error(1)
}

func (m *MockPlayerRepositoryForDelete) GetPlayers(gameId string) ([]models.Player, error) {
	args := m.Called(gameId)
	return args.Get(0).([]models.Player), args.Error(1)
}

func (m *MockPlayerRepositoryForDelete) CreatePlayer(ctx context.Context, player *models.PlayerData) error {
	args := m.Called(ctx, player)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) GetPlayerByID(ctx context.Context, id string) (*models.PlayerData, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockPlayerRepositoryForDelete) GetPlayerByEmail(ctx context.Context, email string) (*models.PlayerData, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockPlayerRepositoryForDelete) GetPlayerByProvider(ctx context.Context, provider, providerID string) (*models.PlayerData, error) {
	args := m.Called(ctx, provider, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockPlayerRepositoryForDelete) GetPlayerByName(ctx context.Context, name string) (*models.PlayerData, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PlayerData), args.Error(1)
}

func (m *MockPlayerRepositoryForDelete) UpdatePlayer(ctx context.Context, player *models.PlayerData) error {
	args := m.Called(ctx, player)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) DeletePlayer(ctx context.Context, playerID string) error {
	args := m.Called(ctx, playerID)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) CreateAuthToken(ctx context.Context, token *models.AuthToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) GetAuthToken(ctx context.Context, token string) (*models.AuthToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthToken), args.Error(1)
}

func (m *MockPlayerRepositoryForDelete) UpdateAuthToken(ctx context.Context, token *models.AuthToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) DeleteAuthToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) UpdateAvatar(ctx context.Context, playerID string, objectKey string, signedURL string, expiresAt time.Time) error {
	args := m.Called(ctx, playerID, objectKey, signedURL, expiresAt)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) UpdateAvatarSignedURL(ctx context.Context, playerID string, signedURL string, expiresAt time.Time) error {
	args := m.Called(ctx, playerID, signedURL, expiresAt)
	return args.Error(0)
}

func (m *MockPlayerRepositoryForDelete) ClearAvatar(ctx context.Context, playerID string) error {
	args := m.Called(ctx, playerID)
	return args.Error(0)
}

func TestAuthService_DeleteAccount_Success(t *testing.T) {
	// Setup
	mockRepo := new(MockPlayerRepositoryForDelete)
	mockOAuth := new(MockOAuthVerifier)
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFunc := func() time.Time { return fixedTime }

	authService := NewAuthServiceWithTimeFunc(mockRepo, mockOAuth, timeFunc)

	ctx := context.Background()
	playerID := "test-player-id"

	// Create test player
	testPlayer := &models.PlayerData{
		ID:   playerID,
		Name: "Test Player",
	}

	// Setup expectations
	mockRepo.On("GetPlayerByID", ctx, playerID).Return(testPlayer, nil)
	mockRepo.On("DeletePlayer", ctx, playerID).Return(nil)

	// Execute
	err := authService.DeleteAccount(ctx, playerID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_DeleteAccount_PlayerNotFound(t *testing.T) {
	// Setup
	mockRepo := new(MockPlayerRepositoryForDelete)
	mockOAuth := new(MockOAuthVerifier)
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFunc := func() time.Time { return fixedTime }

	authService := NewAuthServiceWithTimeFunc(mockRepo, mockOAuth, timeFunc)

	ctx := context.Background()
	playerID := "non-existent-player"

	// Setup expectations
	mockRepo.On("GetPlayerByID", ctx, playerID).Return(nil, sql.ErrNoRows)

	// Execute
	err := authService.DeleteAccount(ctx, playerID)

	// Assert
	assert.Error(t, err)
	var playerNotFoundErr *models.PlayerNotFoundError
	assert.ErrorAs(t, err, &playerNotFoundErr)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_DeleteAccount_PlayerNil(t *testing.T) {
	// Setup
	mockRepo := new(MockPlayerRepositoryForDelete)
	mockOAuth := new(MockOAuthVerifier)
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFunc := func() time.Time { return fixedTime }

	authService := NewAuthServiceWithTimeFunc(mockRepo, mockOAuth, timeFunc)

	ctx := context.Background()
	playerID := "test-player-id"

	// Setup expectations - return nil player without error
	mockRepo.On("GetPlayerByID", ctx, playerID).Return((*models.PlayerData)(nil), nil)

	// Execute
	err := authService.DeleteAccount(ctx, playerID)

	// Assert
	assert.Error(t, err)
	var playerNotFoundErr *models.PlayerNotFoundError
	assert.ErrorAs(t, err, &playerNotFoundErr)
	mockRepo.AssertExpectations(t)
}

func TestAuthService_DeleteAccount_DeleteFails(t *testing.T) {
	// Setup
	mockRepo := new(MockPlayerRepositoryForDelete)
	mockOAuth := new(MockOAuthVerifier)
	fixedTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFunc := func() time.Time { return fixedTime }

	authService := NewAuthServiceWithTimeFunc(mockRepo, mockOAuth, timeFunc)

	ctx := context.Background()
	playerID := "test-player-id"

	// Create test player
	testPlayer := &models.PlayerData{
		ID:   playerID,
		Name: "Test Player",
	}

	// Setup expectations
	mockRepo.On("GetPlayerByID", ctx, playerID).Return(testPlayer, nil)
	mockRepo.On("DeletePlayer", ctx, playerID).Return(assert.AnError)

	// Execute
	err := authService.DeleteAccount(ctx, playerID)

	// Assert
	assert.Error(t, err)
	var generalAuthErr *models.GeneralAuthError
	assert.ErrorAs(t, err, &generalAuthErr)
	mockRepo.AssertExpectations(t)
}
