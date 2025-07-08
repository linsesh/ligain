package middleware

import (
	"liguain/backend/services"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// APIKeyAuth middleware checks for a valid API key in the request header
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		// Get allowed API keys from environment variables
		allowedKeys := []string{
			os.Getenv("API_KEY"),
		}

		// Check if the provided API key is valid
		isValid := false
		for _, key := range allowedKeys {
			if key != "" && key == apiKey {
				isValid = true
				break
			}
		}

		if !isValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// PlayerAuth middleware validates player authentication token
func PlayerAuth(authService services.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]
		player, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store player in Gin context
		c.Set("player", player)

		c.Next()
	}
}
