package routes

import (
	"bytes"
	"encoding/json"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"liguain/backend/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var testTime = time.Date(2024, 3, 15, 15, 0, 0, 0, time.UTC)

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

func setupTestRouter() (*gin.Engine, *MockGame) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	gameRepo := repositories.NewInMemoryGameRepository()
	betRepo := repositories.NewInMemoryBetRepository()
	game := &MockGame{
		incomingMatches: make(map[string]*models.MatchResult),
		pastMatches:     make(map[string]*models.MatchResult),
		bets:            make(map[string]map[models.Player]*models.Bet),
	}
	gameRepo.SaveWithId("123e4567-e89b-12d3-a456-426614174000", game)
	gameService := services.NewGameService("123e4567-e89b-12d3-a456-426614174000", game, gameRepo, betRepo, nil, 10*time.Second)

	handler := NewMatchHandler(map[string]services.GameService{
		"123e4567-e89b-12d3-a456-426614174000": gameService,
	})
	handler.SetupRoutes(router)

	return router, game
}

func TestGetMatches(t *testing.T) {
	router, mockGame := setupTestRouter()

	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matchResult := models.NewMatchWithBets(match, nil)

	mockGame.GetIncomingMatches()[match.Id()] = matchResult

	// Create request
	req := httptest.NewRequest("GET", "/api/game/123e4567-e89b-12d3-a456-426614174000/matches", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	var err error
	err = json.Unmarshal(w.Body.Bytes(), &response)
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
	router, game := setupTestRouter()

	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matchResult := models.NewMatchWithBets(match, nil)

	game.GetIncomingMatches()[match.Id()] = matchResult

	// Create request body
	betRequest := SaveBetRequest{
		MatchID:            match.Id(),
		PredictedHomeGoals: 2,
		PredictedAwayGoals: 1,
	}
	jsonBody, err := json.Marshal(betRequest)
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest("POST", "/api/game/123e4567-e89b-12d3-a456-426614174000/bet", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the bet was saved in the mock game
	playerBets := game.bets[match.Id()]
	assert.NotNil(t, playerBets)
	playerBet := playerBets[models.Player{Name: "Player1"}]
	assert.NotNil(t, playerBet)
	assert.Equal(t, 2, playerBet.PredictedHomeGoals)
	assert.Equal(t, 1, playerBet.PredictedAwayGoals)
}

func TestSaveBet_InvalidRequest(t *testing.T) {
	router, _ := setupTestRouter()

	// Create invalid request body
	invalidBody := []byte(`{"invalid": "json"`)

	// Create request
	req := httptest.NewRequest("POST", "/api/game/123e4567-e89b-12d3-a456-426614174000/bet", bytes.NewBuffer(invalidBody))
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
	req := httptest.NewRequest("POST", "/api/game/123e4567-e89b-12d3-a456-426614174000/bet", bytes.NewBuffer(jsonBody))
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
	router, _ := setupTestRouter()

	// Create request
	req := httptest.NewRequest("GET", "/api/game/non-existent-game/matches", nil)
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

func TestSaveBet_UpdateExistingBet(t *testing.T) {
	router, game := setupTestRouter()

	// Setup test data
	match := models.NewSeasonMatch("Team1", "Team2", "2024", "Premier League", testTime, 1)
	matchResult := models.NewMatchWithBets(match, nil)

	game.GetIncomingMatches()[match.Id()] = matchResult

	// First bet
	initialBet := SaveBetRequest{
		MatchID:            match.Id(),
		PredictedHomeGoals: 2,
		PredictedAwayGoals: 1,
	}
	jsonBody, err := json.Marshal(initialBet)
	assert.NoError(t, err)

	// Create and perform initial bet request
	req := httptest.NewRequest("POST", "/api/game/123e4567-e89b-12d3-a456-426614174000/bet", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify initial bet in mock game
	playerBets := game.bets[match.Id()]
	assert.NotNil(t, playerBets)
	playerBet := playerBets[models.Player{Name: "Player1"}]
	assert.NotNil(t, playerBet)
	assert.Equal(t, 2, playerBet.PredictedHomeGoals)
	assert.Equal(t, 1, playerBet.PredictedAwayGoals)

	// Updated bet
	updatedBet := SaveBetRequest{
		MatchID:            match.Id(),
		PredictedHomeGoals: 3,
		PredictedAwayGoals: 2,
	}
	jsonBody, err = json.Marshal(updatedBet)
	assert.NoError(t, err)

	// Create and perform update bet request
	req = httptest.NewRequest("POST", "/api/game/123e4567-e89b-12d3-a456-426614174000/bet", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify updated bet in mock game
	playerBets = game.bets[match.Id()]
	assert.NotNil(t, playerBets)
	playerBet = playerBets[models.Player{Name: "Player1"}]
	assert.NotNil(t, playerBet)
	assert.Equal(t, 3, playerBet.PredictedHomeGoals)
	assert.Equal(t, 2, playerBet.PredictedAwayGoals)
}

// Add new test for missing gameId
func TestGetMatches_MissingGameId(t *testing.T) {
	router, _ := setupTestRouter()

	// Create request without gameId
	req := httptest.NewRequest("GET", "/api/game//matches", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "game-id is required", response["error"])
}

// Add new test for missing gameId in bet
func TestSaveBet_MissingGameId(t *testing.T) {
	router, _ := setupTestRouter()

	// Create request body
	betRequest := SaveBetRequest{
		MatchID:            "some-match",
		PredictedHomeGoals: 2,
		PredictedAwayGoals: 1,
	}
	jsonBody, _ := json.Marshal(betRequest)

	// Create request without gameId
	req := httptest.NewRequest("POST", "/api/game//bet", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "game-id is required", response["error"])
}
