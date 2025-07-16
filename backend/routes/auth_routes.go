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
		auth.POST("/signin/guest", h.SignInGuest)
		auth.POST("/signout", middleware.PlayerAuth(h.authService), h.SignOut)
		auth.GET("/me", middleware.PlayerAuth(h.authService), h.GetCurrentPlayer)
	}
}

// SignIn handles authentication with Google or Apple
func (h *AuthHandler) SignIn(c *gin.Context) {
	fmt.Printf("🔐 SignIn - Request received from %s\n", c.ClientIP())
	fmt.Printf("🔐 SignIn - Headers: %+v\n", c.Request.Header)

	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ SignIn - JSON binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	fmt.Printf("🔐 SignIn - Request body: provider=%s, email=%s, name=%s, token=%s\n",
		req.Provider, req.Email, req.Name,
		func() string {
			if len(req.Token) > 10 {
				return req.Token[:10] + "..."
			} else {
				return req.Token
			}
		}())

	if req.Provider == "" || req.Token == "" || req.Email == "" || req.Name == "" {
		fmt.Printf("❌ SignIn - Missing required fields: provider=%t, token=%t, email=%t, name=%t\n",
			req.Provider != "", req.Token != "", req.Email != "", req.Name != "")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
		return
	}

	if req.Provider != "google" && req.Provider != "apple" {
		fmt.Printf("❌ SignIn - Invalid provider: %s\n", req.Provider)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider"})
		return
	}

	fmt.Printf("🔐 SignIn - Calling authService.Authenticate\n")
	response, err := h.authService.Authenticate(c.Request.Context(), &req)
	if err != nil {
		fmt.Printf("❌ SignIn - Authentication error: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("✅ SignIn - Authentication successful for user: %s\n", response.Player.Name)
	c.JSON(http.StatusOK, response)
}

// SignInGuest handles guest authentication
func (h *AuthHandler) SignInGuest(c *gin.Context) {
	fmt.Printf("🔐 SignInGuest - Request received from %s\n", c.ClientIP())
	fmt.Printf("🔐 SignInGuest - Headers: %+v\n", c.Request.Header)

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("❌ SignInGuest - JSON binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	fmt.Printf("🔐 SignInGuest - Request body: name=%s\n", req.Name)

	if req.Name == "" {
		fmt.Printf("❌ SignInGuest - Missing name field\n")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	fmt.Printf("🔐 SignInGuest - Calling authService.AuthenticateGuest\n")
	response, err := h.authService.AuthenticateGuest(c.Request.Context(), req.Name)
	if err != nil {
		fmt.Printf("❌ SignInGuest - Authentication error: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("✅ SignInGuest - Guest authentication successful for user: %s\n", response.Player.Name)
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
	player, exists := c.Get("player")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"player": player})
}
