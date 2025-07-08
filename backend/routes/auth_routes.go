package routes

import (
	"fmt"
	"liguain/backend/middleware"
	"liguain/backend/models"
	"liguain/backend/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
		auth.POST("/signout", middleware.PlayerAuth(h.authService), h.SignOut)
		auth.GET("/me", middleware.PlayerAuth(h.authService), h.GetCurrentPlayer)
	}
}

// SignIn handles authentication with Google or Apple
func (h *AuthHandler) SignIn(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Provider == "" || req.Token == "" || req.Email == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	if req.Provider != "google" && req.Provider != "apple" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}

	response, err := h.authService.Authenticate(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

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
	fmt.Println("ta grosse mere")
	player, exists := c.Get("player")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"player": player})
}
