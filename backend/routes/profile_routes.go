package routes

import (
	"errors"
	"io"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	// maxUploadSize is the maximum file size for avatar uploads (10MB)
	maxUploadSize = 10 * 1024 * 1024
)

// ProfileHandler handles profile-related HTTP requests
type ProfileHandler struct {
	profileService services.ProfileService
	authService    services.AuthServiceInterface
}

// NewProfileHandler creates a new ProfileHandler
func NewProfileHandler(
	profileService services.ProfileService,
	authService services.AuthServiceInterface,
) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		authService:    authService,
	}
}

// SetupRoutes registers profile routes on the router
func (h *ProfileHandler) SetupRoutes(router *gin.Engine) {
	players := router.Group("/api/v1/players")
	{
		// Get player by ID (requires auth to access)
		players.GET("/:id", middleware.PlayerAuth(h.authService), h.GetPlayer)

		// Profile operations for current user
		players.PUT("/me/display-name", middleware.PlayerAuth(h.authService), h.UpdateDisplayName)

		// Avatar operations for current user
		players.POST("/me/avatar", middleware.PlayerAuth(h.authService), h.UploadAvatar)
		players.DELETE("/me/avatar", middleware.PlayerAuth(h.authService), h.DeleteAvatar)
	}
}

// GetPlayer returns a player by ID
func (h *ProfileHandler) GetPlayer(c *gin.Context) {
	playerID := c.Param("id")

	player, err := h.profileService.GetPlayerProfile(c.Request.Context(), playerID)
	if err != nil {
		var profileErr *services.ProfileError
		if errors.As(err, &profileErr) && profileErr.Code == "PLAYER_NOT_FOUND" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
			return
		}
		log.Errorf("GetPlayer - Error fetching player %s: %v", playerID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"player": h.toPlayerResponse(player)})
}

// UploadAvatar handles avatar upload for the current user
func (h *ProfileHandler) UploadAvatar(c *gin.Context) {
	// Get player from context (set by middleware)
	player, err := getPlayerFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Parse multipart form file
	imageData, err := parseAvatarFromRequest(c)
	if err != nil {
		var uploadErr *uploadError
		if errors.As(err, &uploadErr) {
			c.JSON(uploadErr.StatusCode, gin.H{
				"error": uploadErr.Message,
				"code":  uploadErr.Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read file",
			"code":  "UPLOAD_FAILED",
		})
		return
	}

	// Call service (service knows what to do)
	result, err := h.profileService.UploadAvatar(c.Request.Context(), player.ID, player.AvatarObjectKey, imageData)
	if err != nil {
		var imgErr *models.ImageProcessingError
		if errors.As(err, &imgErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": imgErr.Reason,
				"code":  imgErr.Code,
			})
			return
		}
		log.Errorf("UploadAvatar - Error uploading avatar for player %s: %v", player.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload avatar",
			"code":  "UPLOAD_FAILED",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"avatar_url": result.SignedURL,
	})
}

// DeleteAvatar removes the avatar for the current user
func (h *ProfileHandler) DeleteAvatar(c *gin.Context) {
	// Get player from context (set by middleware)
	player, err := getPlayerFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// If no avatar, just return success
	if player.AvatarObjectKey == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No avatar to delete"})
		return
	}

	// Call service (service handles storage deletion + DB update)
	err = h.profileService.DeleteAvatar(c.Request.Context(), player.ID, player.AvatarObjectKey)
	if err != nil {
		log.Errorf("DeleteAvatar - Error deleting avatar for player %s: %v", player.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Avatar deleted successfully"})
}

// UpdateDisplayName updates the current user's display name
func (h *ProfileHandler) UpdateDisplayName(c *gin.Context) {
	// Get player from context (set by middleware)
	player, err := getPlayerFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		DisplayName string `json:"displayName" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Errorf("UpdateDisplayName - JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Infof("UpdateDisplayName - Updating display name for player %s to %s", player.ID, req.DisplayName)

	updatedPlayer, err := h.authService.UpdateDisplayName(c.Request.Context(), player.ID, req.DisplayName)
	if err != nil {
		log.Errorf("UpdateDisplayName - Error updating display name: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Infof("UpdateDisplayName - Display name updated successfully for user: %s", updatedPlayer.Name)
	c.JSON(http.StatusOK, gin.H{"player": updatedPlayer})
}

// toPlayerResponse converts a PlayerData to the API response format
func (h *ProfileHandler) toPlayerResponse(player *models.PlayerData) map[string]interface{} {
	response := map[string]interface{}{
		"id":   player.ID,
		"name": player.Name,
	}

	if player.AvatarSignedURL != nil {
		response["avatar_url"] = *player.AvatarSignedURL
	}

	return response
}

// uploadError represents an upload-related error with HTTP status code
type uploadError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *uploadError) Error() string {
	return e.Message
}

// getPlayerFromContext extracts the player from the gin context
func getPlayerFromContext(c *gin.Context) (*models.PlayerData, error) {
	playerInterface, exists := c.Get("player")
	if !exists {
		return nil, errors.New("Player not found in context")
	}

	player, ok := playerInterface.(*models.PlayerData)
	if !ok {
		return nil, errors.New("Invalid player data")
	}

	return player, nil
}

// parseAvatarFromRequest extracts and validates avatar data from the request
func parseAvatarFromRequest(c *gin.Context) ([]byte, error) {
	// Get file from multipart form
	file, err := c.FormFile("avatar")
	if err != nil {
		return nil, &uploadError{
			StatusCode: http.StatusBadRequest,
			Code:       "INVALID_IMAGE",
			Message:    "No avatar file provided",
		}
	}

	// Validate file size
	if file.Size > maxUploadSize {
		return nil, &uploadError{
			StatusCode: http.StatusBadRequest,
			Code:       "FILE_TOO_LARGE",
			Message:    "File too large (max 10MB)",
		}
	}

	// Read file content
	f, err := file.Open()
	if err != nil {
		return nil, &uploadError{
			StatusCode: http.StatusInternalServerError,
			Code:       "UPLOAD_FAILED",
			Message:    "Failed to read file",
		}
	}
	defer f.Close()

	imageData, err := io.ReadAll(f)
	if err != nil {
		return nil, &uploadError{
			StatusCode: http.StatusInternalServerError,
			Code:       "UPLOAD_FAILED",
			Message:    "Failed to read file",
		}
	}

	return imageData, nil
}
