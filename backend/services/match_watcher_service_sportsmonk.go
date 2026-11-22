package services

import (
	"context"
	"ligain/backend/api"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
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

var (
	watcher *MatchWatcherServiceSportsmonk
	once    sync.Once
)

func NewMatchWatcherServiceSportsmonk(env string, matchService MatchService) (*MatchWatcherServiceSportsmonk, error) {
	once.Do(func() {
		if env == "local" {
			log.Infof("Initializing local environment")
			err := localInit()
			if err != nil {
				log.Errorf("Error initializing local environment: %v", err)
			}
		}
		watcher = &MatchWatcherServiceSportsmonk{
			repo:         repositories.NewSportsmonkRepository(api.NewSportsmonkAPI(getAPIToken())),
			matchService: matchService,
			stopChan:     make(chan struct{}),
			pollInterval: time.Minute * 1,
		}
	})
	return watcher, nil
}

// NewMatchWatcherServiceSportsmonkWithOptions creates a MatchWatcherServiceSportsmonk with a custom
// repository and poll interval. Intended for local development and testing (e.g. FakeRandomSportsmonkAPI).
func NewMatchWatcherServiceSportsmonkWithOptions(
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

// Subscribe makes a game receive the matches that changed (the MatchService publishes them)
func (m *MatchWatcherServiceSportsmonk) Subscribe(handler GameService) error {
	return m.matchService.Subscribe(handler)
}

func (m *MatchWatcherServiceSportsmonk) Unsubscribe(gameID string) error {
	return m.matchService.Unsubscribe(gameID)
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

// localInit is a function that initializes the environment variables for the local environment
func localInit() error {
	// Load reads the .env file and injects variables into os.Environ
	if err := godotenv.Load(); err != nil {
		log.Errorf("No .env file found or error loading it: %v", err)
		return err
	}
	return nil
}

func getAPIToken() string {
	return os.Getenv("SPORTSMONK_API_TOKEN")
}
