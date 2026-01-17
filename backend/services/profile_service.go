package services

import (
	"context"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// signedURLRefreshThreshold is how long before expiration we refresh the URL
	signedURLRefreshThreshold = 24 * time.Hour
)

// ProfileError represents a profile service error with a code
type ProfileError struct {
	Code   string
	Reason string
}

func (e *ProfileError) Error() string {
	return e.Reason
}

// AvatarResult contains the result of an avatar upload operation
type AvatarResult struct {
	SignedURL string
	ExpiresAt time.Time
}

// ProfileService defines the interface for profile-related operations
type ProfileService interface {
	// UploadAvatar processes and uploads a new avatar image
	// It handles: image processing, old avatar deletion (if exists), upload, URL generation, and DB update
	UploadAvatar(ctx context.Context, playerID string, oldObjectKey *string, imageData []byte) (*AvatarResult, error)

	// DeleteAvatar removes the avatar for a player
	// It handles: storage deletion (fire-and-forget) and DB clearing
	DeleteAvatar(ctx context.Context, playerID string, objectKey *string) error

	// GetPlayerProfile retrieves a player's profile with refreshed signed URL if needed
	GetPlayerProfile(ctx context.Context, playerID string) (*models.PlayerData, error)
}

// ProfileServiceImpl implements ProfileService
type ProfileServiceImpl struct {
	storageService   StorageService
	imageProcessor   ImageProcessor
	playerRepository repositories.PlayerRepository
}

// NewProfileService creates a new ProfileService instance
func NewProfileService(
	storageService StorageService,
	imageProcessor ImageProcessor,
	playerRepository repositories.PlayerRepository,
) ProfileService {
	return &ProfileServiceImpl{
		storageService:   storageService,
		imageProcessor:   imageProcessor,
		playerRepository: playerRepository,
	}
}

// UploadAvatar implements ProfileService.UploadAvatar
func (s *ProfileServiceImpl) UploadAvatar(ctx context.Context, playerID string, oldObjectKey *string, imageData []byte) (*AvatarResult, error) {
	// 1. Process image (resize, crop, convert to WebP)
	processedImage, err := s.imageProcessor.ProcessAvatar(imageData)
	if err != nil {
		return nil, err
	}

	// 2. Delete old avatar if exists (fire-and-forget)
	if oldObjectKey != nil {
		go func(objectKey string) {
			if err := s.storageService.DeleteAvatar(context.Background(), objectKey); err != nil {
				log.Warnf("UploadAvatar - Failed to delete old avatar %s: %v", objectKey, err)
			}
		}(*oldObjectKey)
	}

	// 3. Upload new avatar
	objectKey, err := s.storageService.UploadAvatar(ctx, playerID, processedImage)
	if err != nil {
		log.Errorf("UploadAvatar - Failed to upload avatar for player %s: %v", playerID, err)
		return nil, err
	}

	// 4. Generate signed URL
	signedURL, err := s.storageService.GenerateSignedURL(ctx, objectKey, DefaultSignedURLTTL)
	if err != nil {
		log.Errorf("UploadAvatar - Failed to generate signed URL for player %s: %v", playerID, err)
		return nil, err
	}

	expiresAt := time.Now().Add(DefaultSignedURLTTL)

	// 5. Update database
	err = s.playerRepository.UpdateAvatar(ctx, playerID, objectKey, signedURL, expiresAt)
	if err != nil {
		log.Errorf("UploadAvatar - Failed to save avatar for player %s: %v", playerID, err)
		return nil, err
	}

	log.Infof("UploadAvatar - Avatar uploaded successfully for player %s", playerID)

	// 6. Return result
	return &AvatarResult{
		SignedURL: signedURL,
		ExpiresAt: expiresAt,
	}, nil
}

// DeleteAvatar implements ProfileService.DeleteAvatar
func (s *ProfileServiceImpl) DeleteAvatar(ctx context.Context, playerID string, objectKey *string) error {
	// If no avatar, just return success
	if objectKey == nil {
		return nil
	}

	// Delete from storage (fire-and-forget)
	go func(key string) {
		if err := s.storageService.DeleteAvatar(context.Background(), key); err != nil {
			log.Warnf("DeleteAvatar - Failed to delete avatar from storage %s: %v", key, err)
		}
	}(*objectKey)

	// Clear from database
	err := s.playerRepository.ClearAvatar(ctx, playerID)
	if err != nil {
		log.Errorf("DeleteAvatar - Failed to clear avatar for player %s: %v", playerID, err)
		return err
	}

	log.Infof("DeleteAvatar - Avatar deleted for player %s", playerID)
	return nil
}

// GetPlayerProfile implements ProfileService.GetPlayerProfile
func (s *ProfileServiceImpl) GetPlayerProfile(ctx context.Context, playerID string) (*models.PlayerData, error) {
	player, err := s.playerRepository.GetPlayerByID(ctx, playerID)
	if err != nil {
		log.Errorf("GetPlayerProfile - Error fetching player %s: %v", playerID, err)
		return nil, err
	}

	if player == nil {
		return nil, &ProfileError{Code: "PLAYER_NOT_FOUND", Reason: "player not found"}
	}

	// Refresh signed URL if needed
	if player.AvatarObjectKey != nil && s.needsSignedURLRefresh(player) {
		newURL, expiresAt, err := s.refreshSignedURL(ctx, player)
		if err != nil {
			log.Warnf("GetPlayerProfile - Failed to refresh signed URL for player %s: %v", playerID, err)
			// Continue with old URL or no URL
		} else {
			player.AvatarSignedURL = &newURL
			player.AvatarSignedURLExpiresAt = &expiresAt
		}
	}

	return player, nil
}

// needsSignedURLRefresh checks if the signed URL needs to be refreshed
func (s *ProfileServiceImpl) needsSignedURLRefresh(player *models.PlayerData) bool {
	if player.AvatarSignedURLExpiresAt == nil {
		return true
	}
	// Refresh if expiring within the threshold
	return time.Now().Add(signedURLRefreshThreshold).After(*player.AvatarSignedURLExpiresAt)
}

// refreshSignedURL generates a new signed URL and updates the database
func (s *ProfileServiceImpl) refreshSignedURL(ctx context.Context, player *models.PlayerData) (string, time.Time, error) {
	url, err := s.storageService.GenerateSignedURL(ctx, *player.AvatarObjectKey, DefaultSignedURLTTL)
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := time.Now().Add(DefaultSignedURLTTL)

	// Update in database (fire-and-forget for performance)
	go func() {
		if err := s.playerRepository.UpdateAvatarSignedURL(context.Background(), player.ID, url, expiresAt); err != nil {
			log.Warnf("refreshSignedURL - Failed to update signed URL in DB for player %s: %v", player.ID, err)
		}
	}()

	return url, expiresAt, nil
}
