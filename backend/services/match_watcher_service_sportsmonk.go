package services

import (
	"context"
	"liguain/backend/api"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

type MatchWatcherServiceSportsmonk struct {
	watchedMatches map[string]models.Match
	repo           repositories.SportsmonkRepository
	subscribers    map[string]GameService
	stopChan       chan struct{}
	isRunning      bool
	pollInterval   time.Duration
}

var (
	watcher *MatchWatcherServiceSportsmonk
	once    sync.Once
)

func NewMatchWatcherServiceSportsmonk(env string) (*MatchWatcherServiceSportsmonk, error) {
	once.Do(func() {
		if env == "local" {
			err := localInit()
			if err != nil {
				log.Errorf("Error initializing local environment: %v", err)
			}
		}
		watcher = &MatchWatcherServiceSportsmonk{
			watchedMatches: make(map[string]models.Match),
			repo:           repositories.NewSportsmonkRepository(api.NewSportsmonkAPI(getAPIToken())),
			subscribers:    make(map[string]GameService),
			stopChan:       make(chan struct{}),
			pollInterval:   30 * time.Second,
		}
	})
	return watcher, nil
}

func (m *MatchWatcherServiceSportsmonk) Subscribe(handler GameService) error {
	gameID := handler.GetGameID()
	m.subscribers[gameID] = handler

	log.Infof("Game %s subscribed to match watcher", gameID)
	return nil
}

func (m *MatchWatcherServiceSportsmonk) Unsubscribe(gameID string) error {
	if _, exists := m.subscribers[gameID]; !exists {
		return nil
	}

	delete(m.subscribers, gameID)
	log.Infof("Game %s unsubscribed from match watcher", gameID)
	return nil
}

func (m *MatchWatcherServiceSportsmonk) Start(ctx context.Context) error {
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
	updates, err := m.getMatchesUpdates()
	if err != nil {
		log.Errorf("Error getting match updates: 	%v", err)
		return
	}

	if len(updates) == 0 {
		return
	}

	log.Infof("Found %d match updates, notifying subscribers", len(updates))

	for gameID, handler := range m.subscribers {
		go func(gameID string, handler GameService, updates map[string]models.Match) {
			if err := handler.HandleMatchUpdates(updates); err != nil {
				log.Errorf("Error handling updates for game %s: %v", gameID, err)
			}
		}(gameID, handler, updates)
	}
}

func (m *MatchWatcherServiceSportsmonk) getMatchesUpdates() (map[string]models.Match, error) {
	updates := make(map[string]models.Match)
	lastMatchInfos, err := m.repo.GetLastMatchInfos(m.watchedMatches)
	if err != nil {
		return nil, err
	}
	for matchId, match := range m.watchedMatches {
		lastMatch := lastMatchInfos[matchId]
		if matchWasUpdated(match, lastMatch) {
			updates[matchId] = lastMatch
			m.watchedMatches[matchId] = lastMatch
		}
	}
	return updates, nil
}

func matchWasUpdated(match models.Match, lastMatch models.Match) bool {
	if match.IsFinished() != lastMatch.IsFinished() {
		return true
	}
	if match.GetHomeTeamOdds() != lastMatch.GetHomeTeamOdds() {
		return true
	}
	if match.GetAwayTeamOdds() != lastMatch.GetAwayTeamOdds() {
		return true
	}
	if match.GetDrawOdds() != lastMatch.GetDrawOdds() {
		return true
	}
	if match.GetDate() != lastMatch.GetDate() {
		log.Warnf("Match %s date was updated from %s to %s", match.Id(), lastMatch.GetDate(), match.GetDate())
		return true
	}
	return false
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

// MockMatchWatcherService is a mock implementation for testing
type MockMatchWatcherService struct {
	subscribers map[string]GameService
}

func NewMockMatchWatcherService() *MockMatchWatcherService {
	return &MockMatchWatcherService{
		subscribers: make(map[string]GameService),
	}
}

func (m *MockMatchWatcherService) Subscribe(handler GameService) error {
	gameID := handler.GetGameID()
	m.subscribers[gameID] = handler
	log.Infof("Mock: Game %s subscribed", gameID)
	return nil
}

func (m *MockMatchWatcherService) Unsubscribe(gameID string) error {
	delete(m.subscribers, gameID)
	log.Infof("Mock: Game %s unsubscribed", gameID)
	return nil
}

func (m *MockMatchWatcherService) Start(ctx context.Context) error {
	log.Info("Mock: Match watcher service started")
	return nil
}

func (m *MockMatchWatcherService) Stop() error {
	log.Info("Mock: Match watcher service stopped")
	return nil
}

// SimulateUpdates is a test helper to simulate match updates
func (m *MockMatchWatcherService) SimulateUpdates(updates map[string]models.Match) {
	subscribers := make(map[string]GameService, len(m.subscribers))
	for gameID, handler := range m.subscribers {
		subscribers[gameID] = handler
	}

	for gameID, handler := range subscribers {
		if err := handler.HandleMatchUpdates(updates); err != nil {
			log.Errorf("Error handling updates for game %s: %v", gameID, err)
		}
	}
}
