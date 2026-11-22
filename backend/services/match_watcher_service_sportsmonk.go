package services

import (
	"context"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"time"

	log "github.com/sirupsen/logrus"
)

// watchHorizon is how far in the future the provider is asked about matches
const watchHorizon = 14 * 24 * time.Hour

// MatchWatcherServiceSportsmonk polls Sportsmonk and hands what it reports to the MatchService,
// which owns the match state. It keeps no match itself.
type MatchWatcherServiceSportsmonk struct {
	repo         repositories.SportsmonkRepository
	matchService MatchService
	stopChan     chan struct{}
	isRunning    bool
	pollInterval time.Duration
}

// NewMatchWatcherServiceSportsmonk creates a watcher polling the given repository every pollInterval
func NewMatchWatcherServiceSportsmonk(
	repo repositories.SportsmonkRepository,
	matchService MatchService,
	pollInterval time.Duration,
) *MatchWatcherServiceSportsmonk {
	return &MatchWatcherServiceSportsmonk{
		repo:         repo,
		matchService: matchService,
		stopChan:     make(chan struct{}),
		pollInterval: pollInterval,
	}
}

func (m *MatchWatcherServiceSportsmonk) Start(ctx context.Context) error {
	log.Infof("Starting match watcher service")
	if m.isRunning {
		return nil // Already running
	}

	m.isRunning = true
	m.stopChan = make(chan struct{})

	go m.pollLoop(ctx)
	log.Info("Match watcher service started")
	return nil
}

func (m *MatchWatcherServiceSportsmonk) Stop() error {
	if !m.isRunning {
		return nil // Already stopped
	}

	close(m.stopChan)
	m.isRunning = false
	log.Info("Match watcher service stopped")
	return nil
}

func (m *MatchWatcherServiceSportsmonk) pollLoop(ctx context.Context) {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Match watcher context cancelled, stopping poll loop")
			return
		case <-m.stopChan:
			log.Info("Match watcher stop signal received, stopping poll loop")
			return
		case <-ticker.C:
			m.checkForUpdates()
		}
	}
}

func (m *MatchWatcherServiceSportsmonk) checkForUpdates() {
	matchesToQuery, err := m.matchService.GetMatchesInHorizon(watchHorizon)
	if err != nil {
		log.Errorf("Error getting the matches to watch: %v", err)
		return
	}
	log.Infof("Getting last match infos for %d matches (scheduled within %v)", len(matchesToQuery), watchHorizon)

	lastMatchInfos, err := m.repo.GetLastMatchInfos(matchesToQuery)
	if err != nil {
		log.Errorf("Error getting match updates: \t%v", err)
		return
	}

	if err := m.matchService.TryUpdateMatches(models.MatchUpdatesOf(lastMatchInfos)); err != nil {
		log.Errorf("Error updating the matches: %v", err)
	}
}
