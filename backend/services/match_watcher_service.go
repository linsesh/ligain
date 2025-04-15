package services

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/utils"
)

type MatchWatcherServiceResult = utils.TaskResult[map[string]models.Match]

// MatchWatcherService is used to get notified when schedule, odds are changed or result are known
type MatchWatcherService interface {
	WatchMatches(matches []models.Match)
	// GetUpdates returns the updates of the matches in an async way
	GetUpdates(ctx context.Context, done chan MatchWatcherServiceResult)
}
