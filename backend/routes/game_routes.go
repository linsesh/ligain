package routes

import (
	"fmt"
	"liguain/backend/middleware"
	"liguain/backend/services"
	"net/http"

	"liguain/backend/models"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GameHandler handles all game-related routes
type GameHandler struct {
	gameCreationService services.GameCreationServiceInterface
	authService         services.AuthServiceInterface
}

// NewGameHandler creates a new GameHandler instance
func NewGameHandler(gameCreationService services.GameCreationServiceInterface, authService services.AuthServiceInterface) *GameHandler {
	return &GameHandler{
		gameCreationService: gameCreationService,
		authService:         authService,
	}
}

// SetupRoutes registers all game-related routes
func (h *GameHandler) SetupRoutes(router *gin.Engine) {
	router.POST("/api/games", middleware.PlayerAuth(h.authService), h.createGame)
	router.POST("/api/games/join", middleware.PlayerAuth(h.authService), h.joinGame)
}

// createGame handles the creation of a new game with a unique code
func (h *GameHandler) createGame(c *gin.Context) {
	var request services.CreateGameRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"request_body": c.Request.Body,
			"content_type": c.GetHeader("Content-Type"),
		}).Error("Failed to bind game creation request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           fmt.Sprintf("Invalid request format: %v", err),
			"expected_format": "{\"seasonYear\": \"string\", \"competitionName\": \"string\"}",
		})
		return
	}

	// Validate required fields
	if request.SeasonYear == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "seasonYear is required"})
		return
	}
	if request.CompetitionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "competitionName is required"})
		return
	}

	// Create the game
	response, err := h.gameCreationService.CreateGame(&request)
	if err != nil {
		if err == services.ErrInvalidCompetition || err == services.ErrInvalidSeasonYear {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.WithError(err).Error("Failed to create game")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create game"})
		return
	}

	log.WithFields(log.Fields{
		"gameId": response.GameID,
		"code":   response.Code,
	}).Info("Game created successfully")

	c.JSON(http.StatusCreated, response)
}

// joinGame handles joining a game by code
func (h *GameHandler) joinGame(c *gin.Context) {
	type joinRequest struct {
		Code string `json:"code" binding:"required"`
	}
	var req joinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	// Get authenticated player from context
	player, exists := c.Get("player")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player not found in context"})
		return
	}

	// Call service to join game by code
	response, err := h.gameCreationService.JoinGame(req.Code, player.(models.Player))
	if err != nil {
		if strings.Contains(err.Error(), "invalid game code") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.WithError(err).Error("Failed to join game")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join game"})
		return
	}

	log.WithFields(log.Fields{
		"gameId": response.GameID,
		"code":   req.Code,
		"player": player.(models.Player).GetName(),
	}).Info("Player joined game successfully")

	c.JSON(http.StatusOK, response)
}
