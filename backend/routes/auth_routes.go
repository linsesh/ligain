package routes

import (
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/services"
	"net/http"
	"strings"

	"errors"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService services.AuthServiceInterface
}

func NewAuthHandler(authService services.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) SetupRoutes(router *gin.Engine) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/signin", h.SignIn)
		auth.POST("/signin/guest", h.SignInGuest)
		auth.POST("/signout", middleware.PlayerAuth(h.authService), h.SignOut)
		auth.GET("/me", middleware.PlayerAuthWithRefresh(h.authService), h.GetCurrentPlayer)
		auth.DELETE("/account", middleware.PlayerAuth(h.authService), h.DeleteAccount)
	}
}

// SignIn handles authentication with Google or Apple
func (h *AuthHandler) SignIn(c *gin.Context) {
	log.Infof("🔐 SignIn - Request received from %s", c.ClientIP())
	log.Infof("🔐 SignIn - Headers: %+v", c.Request.Header)

	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("❌ SignIn - JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Infof("🔐 SignIn - Request body: provider=%s, email=%s, name=%s, token=%s",
		req.Provider, req.Email, req.Name,
		func() string {
			if len(req.Token) > 10 {
				return req.Token[:10] + "..."
			} else {
				return req.Token
			}
		}())

	if req.Provider == "" || req.Token == "" {
		log.Errorf("❌ SignIn - Missing required fields: provider=%t, token=%t",
			req.Provider != "", req.Token != "")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	// Email is optional for Apple Sign-In (Apple doesn't always provide it)
	// For Google Sign-In, email should be provided
	if req.Provider == "google" && req.Email == "" {
		log.Errorf("❌ SignIn - Email is required for Google Sign-In")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required for Google Sign-In"})
		return
	}

	if req.Provider != "google" && req.Provider != "apple" {
		log.Errorf("❌ SignIn - Invalid provider: %s", req.Provider)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}

	log.Infof("🔐 SignIn - Calling authService.Authenticate")
	resp, err := h.authService.Authenticate(c.Request.Context(), &req)
	if err != nil {
		var needNameErr *models.NeedDisplayNameError
		if errors.As(err, &needNameErr) {
			c.JSON(http.StatusOK, gin.H{
				"status":        "need_display_name",
				"suggestedName": needNameErr.SuggestedName,
				"error":         needNameErr.Reason,
			})
			return
		}
		log.Errorf("❌ SignIn - Authentication error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Infof("✅ SignIn - Authentication successful for user: %s", resp.Player.Name)
	c.JSON(http.StatusOK, resp)
}

// SignInGuest handles guest authentication
func (h *AuthHandler) SignInGuest(c *gin.Context) {
	log.Infof("🔐 SignInGuest - Request received from %s", c.ClientIP())
	log.Infof("🔐 SignInGuest - Headers: %+v", c.Request.Header)

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("❌ SignInGuest - JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Infof("🔐 SignInGuest - Request body: name=%s", req.Name)

	if req.Name == "" {
		log.Errorf("❌ SignInGuest - Missing name field")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	log.Infof("🔐 SignInGuest - Calling authService.AuthenticateGuest")
	response, err := h.authService.AuthenticateGuest(c.Request.Context(), req.Name)
	if err != nil {
		log.Errorf("❌ SignInGuest - Authentication error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	log.Infof("✅ SignInGuest - Guest authentication successful for user: %s", response.Player.Name)
	c.JSON(http.StatusOK, response)
}

// SignOut handles user logout
func (h *AuthHandler) SignOut(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	// Extract token from "Bearer <token>" format
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	token := tokenParts[1]
	err := h.authService.Logout(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// GetCurrentPlayer returns the current authenticated player
func (h *AuthHandler) GetCurrentPlayer(c *gin.Context) {
	// Get player from context (set by middleware)
	player, exists := c.Get("player")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player not found in context"})
		return
	}

	playerData, ok := player.(*models.PlayerData)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid player data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"player": toPlayerResponse(playerData)})
}

// toPlayerResponse converts a PlayerData to the API response format
// This ensures consistent field naming (avatar_url instead of avatar_signed_url)
func toPlayerResponse(player *models.PlayerData) map[string]interface{} {
	response := map[string]interface{}{
		"id":   player.ID,
		"name": player.Name,
	}

	if player.Email != nil {
		response["email"] = *player.Email
	}
	if player.Provider != nil {
		response["provider"] = *player.Provider
	}
	if player.AvatarSignedURL != nil {
		response["avatar_url"] = *player.AvatarSignedURL
	}

	return response
}

// DeleteAccount handles permanent account deletion
func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	log.Infof("🗑️ DeleteAccount - Request received from %s", c.ClientIP())

	// Get authenticated player from middleware
	playerInterface, exists := c.Get("player")
	if !exists {
		log.Error("❌ DeleteAccount - Player not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	playerData, ok := playerInterface.(*models.PlayerData)
	if !ok {
		log.Error("❌ DeleteAccount - Invalid player type in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Infof("🗑️ DeleteAccount - Deleting account for player %s (%s)", playerData.Name, playerData.ID)

	err := h.authService.DeleteAccount(c.Request.Context(), playerData.ID)
	if err != nil {
		log.Errorf("❌ DeleteAccount - Error deleting account: %v", err)

		// Check for specific error types
		var playerNotFoundErr *models.PlayerNotFoundError
		if errors.As(err, &playerNotFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		return
	}

	log.Infof("✅ DeleteAccount - Account deleted successfully for player: %s", playerData.Name)
	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
