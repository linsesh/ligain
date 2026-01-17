package routes

import (
	"errors"
	"io"
	"ligain/backend/middleware"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"ligain/backend/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	// signedURLRefreshThreshold is how long before expiration we refresh the URL
	signedURLRefreshThreshold = 24 * time.Hour
	// maxUploadSize is the maximum file size for avatar uploads (10MB)
	maxUploadSize = 10 * 1024 * 1024
)

// ProfileHandler handles profile-related HTTP requests
type ProfileHandler struct {
	storageService   services.StorageService
	imageProcessor   services.ImageProcessor
	playerRepository repositories.PlayerRepository
	authService      services.AuthServiceInterface
}

// NewProfileHandler creates a new ProfileHandler
func NewProfileHandler(
	storageService services.StorageService,
	imageProcessor services.ImageProcessor,
	playerRepository repositories.PlayerRepository,
	authService services.AuthServiceInterface,
) *ProfileHandler {
	return &ProfileHandler{
		storageService:   storageService,
		imageProcessor:   imageProcessor,
		playerRepository: playerRepository,
		authService:      authService,
	}
}

// SetupRoutes registers profile routes on the router
func (h *ProfileHandler) SetupRoutes(router *gin.Engine) {
	players := router.Group("/api/v1/players")
	{
		// Get player by ID (requires auth to access)
		players.GET("/:id", middleware.PlayerAuth(h.authService), h.GetPlayer)

		// Avatar operations for current user
		players.POST("/me/avatar", middleware.PlayerAuth(h.authService), h.UploadAvatar)
		players.DELETE("/me/avatar", middleware.PlayerAuth(h.authService), h.DeleteAvatar)
	}
}

// GetPlayer returns a player by ID
func (h *ProfileHandler) GetPlayer(c *gin.Context) {
	playerID := c.Param("id")

	player, err := h.playerRepository.GetPlayerByID(c.Request.Context(), playerID)
	if err != nil {
		log.Errorf("GetPlayer - Error fetching player %s: %v", playerID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch player"})
		return
	}

	if player == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	// Refresh signed URL if needed
	if player.AvatarObjectKey != nil && h.needsSignedURLRefresh(player) {
		newURL, expiresAt, err := h.refreshSignedURL(c.Request.Context(), player)
		if err != nil {
			log.Warnf("GetPlayer - Failed to refresh signed URL for player %s: %v", playerID, err)
			// Continue with old URL or no URL
		} else {
			player.AvatarSignedURL = &newURL
			player.AvatarSignedURLExpiresAt = &expiresAt
		}
	}

	c.JSON(http.StatusOK, gin.H{"player": h.toPlayerResponse(player)})
}

// UploadAvatar handles avatar upload for the current user
func (h *ProfileHandler) UploadAvatar(c *gin.Context) {
	// Get player from context (set by middleware)
	playerInterface, exists := c.Get("player")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player not found in context"})
		return
	}

	player, ok := playerInterface.(*models.PlayerData)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid player data"})
		return
	}

	// Get file from multipart form
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No avatar file provided",
			"code":  "INVALID_IMAGE",
		})
		return
	}

	// Validate file size
	if file.Size > maxUploadSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File too large (max 10MB)",
			"code":  "FILE_TOO_LARGE",
		})
		return
	}

	// Read file content
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read file",
			"code":  "UPLOAD_FAILED",
		})
		return
	}
	defer f.Close()

	imageData, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read file",
			"code":  "UPLOAD_FAILED",
		})
		return
	}

	// Process image (resize, crop, convert to WebP)
	processedImage, err := h.imageProcessor.ProcessAvatar(imageData)
	if err != nil {
		var imgErr *models.ImageProcessingError
		if errors.As(err, &imgErr) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": imgErr.Reason,
				"code":  imgErr.Code,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process image",
			"code":  "UPLOAD_FAILED",
		})
		return
	}

	// Delete old avatar if exists (fire and forget)
	if player.AvatarObjectKey != nil {
		go func(objectKey string) {
			if err := h.storageService.DeleteAvatar(c.Request.Context(), objectKey); err != nil {
				log.Warnf("UploadAvatar - Failed to delete old avatar %s: %v", objectKey, err)
			}
		}(*player.AvatarObjectKey)
	}

	// Upload new avatar
	objectKey, err := h.storageService.UploadAvatar(c.Request.Context(), player.ID, processedImage)
	if err != nil {
		log.Errorf("UploadAvatar - Failed to upload avatar for player %s: %v", player.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload avatar",
			"code":  "UPLOAD_FAILED",
		})
		return
	}

	// Generate signed URL
	signedURL, err := h.storageService.GenerateSignedURL(c.Request.Context(), objectKey, services.DefaultSignedURLTTL)
	if err != nil {
		log.Errorf("UploadAvatar - Failed to generate signed URL for player %s: %v", player.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate avatar URL",
			"code":  "UPLOAD_FAILED",
		})
		return
	}

	expiresAt := time.Now().Add(services.DefaultSignedURLTTL)

	// Update database
	err = h.playerRepository.UpdateAvatar(c.Request.Context(), player.ID, objectKey, signedURL, expiresAt)
	if err != nil {
		log.Errorf("UploadAvatar - Failed to save avatar for player %s: %v", player.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save avatar",
			"code":  "UPLOAD_FAILED",
		})
		return
	}

	log.Infof("UploadAvatar - Avatar uploaded successfully for player %s", player.ID)
	c.JSON(http.StatusOK, gin.H{
		"avatar_url": signedURL,
	})
}

// DeleteAvatar removes the avatar for the current user
func (h *ProfileHandler) DeleteAvatar(c *gin.Context) {
	// Get player from context (set by middleware)
	playerInterface, exists := c.Get("player")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player not found in context"})
		return
	}

	player, ok := playerInterface.(*models.PlayerData)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid player data"})
		return
	}

	// If no avatar, just return success
	if player.AvatarObjectKey == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No avatar to delete"})
		return
	}

	// Delete from storage (fire and forget)
	go func(objectKey string) {
		if err := h.storageService.DeleteAvatar(c.Request.Context(), objectKey); err != nil {
			log.Warnf("DeleteAvatar - Failed to delete avatar from storage %s: %v", objectKey, err)
		}
	}(*player.AvatarObjectKey)

	// Clear from database
	err := h.playerRepository.ClearAvatar(c.Request.Context(), player.ID)
	if err != nil {
		log.Errorf("DeleteAvatar - Failed to clear avatar for player %s: %v", player.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete avatar"})
		return
	}

	log.Infof("DeleteAvatar - Avatar deleted for player %s", player.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Avatar deleted successfully"})
}

// needsSignedURLRefresh checks if the signed URL needs to be refreshed
func (h *ProfileHandler) needsSignedURLRefresh(player *models.PlayerData) bool {
	if player.AvatarSignedURLExpiresAt == nil {
		return true
	}
	// Refresh if expiring within the threshold
	return time.Now().Add(signedURLRefreshThreshold).After(*player.AvatarSignedURLExpiresAt)
}

// refreshSignedURL generates a new signed URL and updates the database
func (h *ProfileHandler) refreshSignedURL(ctx interface{}, player *models.PlayerData) (string, time.Time, error) {
	// Type assert ctx to context.Context if needed
	requestCtx, ok := ctx.(interface {
		Request() interface{}
	})
	_ = requestCtx
	_ = ok

	url, err := h.storageService.GenerateSignedURL(nil, *player.AvatarObjectKey, services.DefaultSignedURLTTL)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(services.DefaultSignedURLTTL)

	// Update in database (fire and forget for performance)
	go func() {
		if err := h.playerRepository.UpdateAvatarSignedURL(nil, player.ID, url, expiresAt); err != nil {
			log.Warnf("refreshSignedURL - Failed to update signed URL in DB for player %s: %v", player.ID, err)
		}
	}()

	return url, expiresAt, nil
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
