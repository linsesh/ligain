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

type MatchWatcherServiceWithAPI struct {
	matches map[string]models.Match
}

func NewMatchWatcherServiceWithAPI() *MatchWatcherServiceWithAPI {
	return &MatchWatcherServiceWithAPI{
		matches: make(map[string]models.Match),
	}
}

func (m *MatchWatcherServiceWithAPI) WatchMatches(matches []models.Match) {
	for _, match := range matches {
		m.matches[match.Id()] = match
	}
}

func (m *MatchWatcherServiceWithAPI) GetUpdates(ctx context.Context, done chan MatchWatcherServiceResult) {
	done <- MatchWatcherServiceResult{
		Value: m.matches,
		Err:   nil,
	}
}
