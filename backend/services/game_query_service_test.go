package services

import (
	"errors"
	"ligain/backend/models"
	"ligain/backend/rules"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGameQueryService_GetPlayerGames_Success(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameIDs := []string{"game1", "game2"}

	// Create real games
	realGame1 := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game 1", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})
	realGame2 := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game 2", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame1, nil)
	mockGameRepo.On("GetGame", "game2").Return(realGame2, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{player}, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game2").Return([]models.Player{player}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(map[string]map[string]int{}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game2").Return(map[string]map[string]int{}, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(&models.GameCode{Code: "ABC1"}, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game2").Return(&models.GameCode{Code: "XYZ9"}, nil)

	// Execute
	playerGames, err := service.GetPlayerGames(player)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, playerGames, 2)
	assert.Equal(t, "game1", playerGames[0].GameID)
	assert.Equal(t, "game2", playerGames[1].GameID)
	assert.Equal(t, "2025/2026", playerGames[0].SeasonYear)
	assert.Equal(t, "Ligue 1", playerGames[0].CompetitionName)
	assert.Equal(t, "ABC1", playerGames[0].Code)
	assert.Equal(t, "XYZ9", playerGames[1].Code)

	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
	mockGameCodeRepo.AssertExpectations(t)
}

func TestGameQueryService_GetPlayerGames_WithPlayersAndScores(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)

	player1 := &models.PlayerData{ID: "player1", Name: "Test Player"}
	player2 := &models.PlayerData{ID: "player2", Name: "Other Player"}
	gameIDs := []string{"game1"}

	// Create a real game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player1, player2}, []models.Match{}, &rules.ScorerOriginal{})

	mockPlayers := []models.Player{player1, player2}
	mockScores := map[string]map[string]int{
		"match1": {"player1": 10, "player2": 5},
		"match2": {"player1": 20, "player2": 15},
	}

	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return(mockPlayers, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(mockScores, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(nil, errors.New("code not found"))

	// Execute
	playerGames, err := service.GetPlayerGames(player1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, playerGames, 1)

	game := playerGames[0]
	assert.Equal(t, "game1", game.GameID)
	assert.Len(t, game.Players, 2)

	for _, p := range game.Players {
		if p.ID == "player1" {
			assert.Equal(t, 30, p.TotalScore)
			assert.Equal(t, 10, p.ScoresByMatch["match1"])
			assert.Equal(t, 20, p.ScoresByMatch["match2"])
		}
		if p.ID == "player2" {
			assert.Equal(t, 20, p.TotalScore)
			assert.Equal(t, 5, p.ScoresByMatch["match1"])
			assert.Equal(t, 15, p.ScoresByMatch["match2"])
		}
	}

	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockBetRepo.AssertExpectations(t)
}

func TestGameQueryService_GetPlayerGames_EmptyList(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return([]string{}, nil)

	// Execute
	playerGames, err := service.GetPlayerGames(player)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, playerGames, 0)

	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameQueryService_GetPlayerGames_RepositoryError(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(nil, errors.New("database error"))

	// Execute
	playerGames, err := service.GetPlayerGames(player)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, playerGames)
	assert.Contains(t, err.Error(), "error getting player games")

	mockGamePlayerRepo.AssertExpectations(t)
}

func TestGameQueryService_GetPlayerGames_GracefulDegradation(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameIDs := []string{"game1", "game2"}

	// Create one real game
	realGame1 := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game 1", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations - game2 fails but game1 succeeds
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame1, nil)
	mockGameRepo.On("GetGame", "game2").Return(nil, errors.New("game not found"))
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{player}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(map[string]map[string]int{}, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(&models.GameCode{Code: "ABC1"}, nil)

	// Execute
	playerGames, err := service.GetPlayerGames(player)

	// Assert - should still return game1 even though game2 failed
	assert.NoError(t, err)
	assert.Len(t, playerGames, 1)
	assert.Equal(t, "game1", playerGames[0].GameID)

	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}

func TestGameQueryService_GetGame_Success(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})

	// Mock expectations
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)

	// Execute
	game, err := service.GetGame("game1")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.Equal(t, "2025/2026", game.GetSeasonYear())
	assert.Equal(t, "Ligue 1", game.GetCompetitionName())

	mockGameRepo.AssertExpectations(t)
}

func TestGameQueryService_GetGame_NotFound(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)

	// Mock expectations
	mockGameRepo.On("GetGame", "nonexistent").Return(nil, errors.New("game not found"))

	// Execute
	game, err := service.GetGame("nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, game)

	mockGameRepo.AssertExpectations(t)
}

func TestGameQueryService_GetPlayerGames_FinishedStatus(t *testing.T) {
	// Setup
	mockGameRepo := new(MockGameRepository)
	mockGamePlayerRepo := new(MockGamePlayerRepository)
	mockGameCodeRepo := new(MockGameCodeRepository)
	mockBetRepo := new(MockBetRepository)

	service := NewGameQueryService(mockGameRepo, mockGamePlayerRepo, mockGameCodeRepo, mockBetRepo)

	player := &models.PlayerData{ID: "player1", Name: "Test Player"}
	gameIDs := []string{"game1"}

	// Create a finished game
	realGame := rules.NewFreshGame("2025/2026", "Ligue 1", "Test Game", []models.Player{player}, []models.Match{}, &rules.ScorerOriginal{})
	realGame.Finish() // Mark as finished

	// Mock expectations
	mockGamePlayerRepo.On("GetPlayerGames", mock.Anything, "player1").Return(gameIDs, nil)
	mockGameRepo.On("GetGame", "game1").Return(realGame, nil)
	mockGamePlayerRepo.On("GetPlayersInGame", mock.Anything, "game1").Return([]models.Player{player}, nil)
	mockBetRepo.On("GetScoresByMatchAndPlayer", "game1").Return(map[string]map[string]int{}, nil)
	mockGameCodeRepo.On("GetGameCodeByGameID", "game1").Return(&models.GameCode{Code: "ABC1"}, nil)

	// Execute
	playerGames, err := service.GetPlayerGames(player)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, playerGames, 1)
	assert.Equal(t, "finished", playerGames[0].Status)

	mockGamePlayerRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}
