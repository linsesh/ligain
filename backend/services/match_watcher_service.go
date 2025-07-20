package services

import (
	"context"
	"ligain/backend/models"
	"ligain/backend/utils"
)

type MatchWatcherServiceResult = utils.TaskResult[map[string]models.Match]

// MatchWatcherService is used to get notified when schedule, odds are changed or result are known
type MatchWatcherService interface {
	// Subscribe adds a game service to receive updates
	Subscribe(handler GameService) error
	// Unsubscribe removes a game service from receiving updates
	Unsubscribe(gameID string) error
	// Start begins the background polling and notification process
	Start(ctx context.Context) error
	// Stop stops the background polling process
	Stop() error
}
