package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"liguain/backend/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var testTime = time.Date(2024, 3, 15, 15, 0, 0, 0, time.UTC)

// MockGameRepository implements repositories.GameRepository for testing
type MockGameRepository struct {
	game models.Game
	err  error
}

func (m *MockGameRepository) GetGame(gameId string) (models.Game, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.game, nil
}

func (m *MockGameRepository) CreateGame(game models.Game) (string, error) {
	return "1", nil
}

func (m *MockGameRepository) SaveWithId(gameId string, game models.Game) error {
	return nil
}

// MockGame implements models.Game for testing
type MockGame struct {
	incomingMatches map[string]*models.MatchResult
	pastMatches     map[string]*models.MatchResult
	bets            map[string]map[models.Player]*models.Bet
}

func (m *MockGame) GetIncomingMatches() map[string]*models.MatchResult {
	return m.incomingMatches
}

func (m *MockGame) GetPastResults() map[string]*models.MatchResult {
	return m.pastMatches
}

func (m *MockGame) GetSeasonYear() string {
	return "2024"
}

func (m *MockGame) GetCompetitionName() string {
	return "Premier League"
}

func (m *MockGame) GetGameStatus() models.GameStatus {
	return models.GameStatusScheduled
}

func (m *MockGame) CheckPlayerBetValidity(player models.Player, bet *models.Bet, datetime time.Time) error {
	return nil
}

func (m *MockGame) AddPlayerBet(player models.Player, bet *models.Bet) error {
	if m.bets == nil {
		m.bets = make(map[string]map[models.Player]*models.Bet)
	}
	if m.bets[bet.Match.Id()] == nil {
		m.bets[bet.Match.Id()] = make(map[models.Player]*models.Bet)
	}
	m.bets[bet.Match.Id()][player] = bet

	// Update the MatchResult in incomingMatches
	if _, exists := m.incomingMatches[bet.Match.Id()]; exists {
		m.incomingMatches[bet.Match.Id()] = models.NewMatchWithBets(bet.Match, m.bets[bet.Match.Id()])
	}
	return nil
}

func (m *MockGame) CalculateMatchScores(match models.Match) (map[models.Player]int, error) {
	return nil, nil
}

func (m *MockGame) ApplyMatchScores(match models.Match, scores map[models.Player]int) {
}

func (m *MockGame) UpdateMatch(match models.Match) error {
	return nil
}

func (m *MockGame) GetPlayersPoints() map[models.Player]int {
	return nil
}

func (m *MockGame) IsFinished() bool {
	return false
}

func (m *MockGame) GetWinner() []models.Player {
	return nil
}

func setupTestRouter() (*gin.Engine, *MockGameRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockRepo := &MockGameRepository{
		game: &MockGame{
			incomingMatches: make(map[string]*models.MatchResult),
			pastMatches:     make(map[string]*models.MatchResult),
			bets:            make(map[string]map[models.Player]*models.Bet),
		},
	}

	handler := NewMatchHandler(mockRepo)
	handler.SetupRoutes(router)

	return router, mockRepo
}

func TestGetMatches(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matchResult := models.NewMatchWithBets(match, nil)

	mockGame := mockRepo.game.(*MockGame)
	mockGame.incomingMatches[match.Id()] = matchResult

	// Create request
	req := httptest.NewRequest("GET", "/api/matches", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	incomingMatches, exists := response["incomingMatches"].(map[string]any)
	assert.True(t, exists)
	assert.NotNil(t, incomingMatches)

	pastMatches, exists := response["pastMatches"].(map[string]any)
	assert.True(t, exists)
	assert.NotNil(t, pastMatches)
}

func TestSaveBet_Success(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matchResult := models.NewMatchWithBets(match, nil)

	mockGame := mockRepo.game.(*MockGame)
	mockGame.incomingMatches[match.Id()] = matchResult

	// Create request body
	betRequest := SaveBetRequest{
		MatchID:            match.Id(),
		PredictedHomeGoals: 2,
		PredictedAwayGoals: 1,
	}
	jsonBody, _ := json.Marshal(betRequest)

	// Create request
	req := httptest.NewRequest("POST", "/api/bet", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the bet was saved by retrieving matches
	req = httptest.NewRequest("GET", "/api/matches", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	incomingMatches := response["incomingMatches"].(map[string]any)
	matchData := incomingMatches[match.Id()].(map[string]any)
	bets := matchData["bets"].(map[string]any)

	// Verify the bet exists and has correct values
	playerBet := bets["Player1"].(map[string]any)
	assert.Equal(t, float64(2), playerBet["predictedHomeGoals"])
	assert.Equal(t, float64(1), playerBet["predictedAwayGoals"])
}

func TestSaveBet_InvalidRequest(t *testing.T) {
	router, _ := setupTestRouter()

	// Create invalid request body
	invalidBody := []byte(`{"invalid": "json"`)

	// Create request
	req := httptest.NewRequest("POST", "/api/bet", bytes.NewBuffer(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Invalid request format")
}

func TestSaveBet_MatchNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	// Create request body with non-existent match ID
	betRequest := SaveBetRequest{
		MatchID:            "non-existent-match",
		PredictedHomeGoals: 2,
		PredictedAwayGoals: 1,
	}
	jsonBody, _ := json.Marshal(betRequest)

	// Create request
	req := httptest.NewRequest("POST", "/api/bet", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Match non-existent-match not found")
}

func TestGetMatches_GameNotFound(t *testing.T) {
	router, mockRepo := setupTestRouter()

	// Set the mock repository to return an error
	mockRepo.err = fmt.Errorf("game not found")

	// Create request
	req := httptest.NewRequest("GET", "/api/matches", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Your game was not found", response["error"])
}
