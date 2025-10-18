package middleware

import (
	"errors"
	"ligain/backend/models"
	"ligain/backend/services"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// APIKeyAuth middleware checks for a valid API key in the request header
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		log.Infof("🔑 APIKeyAuth - Request to %s from %s", c.Request.URL.Path, c.ClientIP())

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
			log.Error("PlayerAuth - Authorization header is required")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Error("PlayerAuth - Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]
		player, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			log.Error("PlayerAuth - Invalid or expired token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store player in Gin context
		c.Set("player", player)

		c.Next()
	}
}

// PlayerAuthWithRefresh middleware validates player authentication token and attempts refresh if expired
func PlayerAuthWithRefresh(authService services.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Error("PlayerAuthWithRefresh - Authorization header is required")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Error("PlayerAuthWithRefresh - Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]
		player, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			// Check if it's a token expired error
			var tokenExpiredErr *models.TokenExpiredError

			if errors.As(err, &tokenExpiredErr) {
				log.Info("PlayerAuthWithRefresh - Token expired, attempting refresh")

				// For expired tokens, we need to get the player ID first before the token is deleted
				// Let's try to get the auth token directly to extract the player ID
				authToken, getTokenErr := authService.GetAuthTokenDirectly(c.Request.Context(), token)
				if getTokenErr != nil || authToken == nil {
					log.Error("PlayerAuthWithRefresh - Could not get token info for refresh")
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired and refresh failed"})
					c.Abort()
					return
				}

				// Attempt to refresh the token (this will delete the old token)
				refreshResp, refreshErr := authService.RefreshToken(c.Request.Context(), token)
				if refreshErr != nil {
					log.Error("PlayerAuthWithRefresh - Token refresh failed")
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired and refresh failed"})
					c.Abort()
					return
				}

				// Set the new token in response headers
				c.Header("X-New-Token", refreshResp.Token)
				c.Header("X-Token-Refreshed", "true")

				// Use the player from the refresh response
				player = &refreshResp.Player
			} else {
				log.Error("PlayerAuthWithRefresh - Invalid token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}
		}

		// Store player in Gin context
		c.Set("player", player)

		c.Next()
	}
}
