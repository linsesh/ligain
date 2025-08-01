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

type MatchWatcherServiceSportsmonk struct {
	watchedMatches map[string]models.Match
	repo           repositories.SportsmonkRepository
	matchRepo      repositories.MatchRepository
	subscribers    map[string]GameService
	stopChan       chan struct{}
	isRunning      bool
	pollInterval   time.Duration
}

var (
	watcher *MatchWatcherServiceSportsmonk
	once    sync.Once
)

func NewMatchWatcherServiceSportsmonk(env string, matches map[string]models.Match, matchRepo repositories.MatchRepository) (*MatchWatcherServiceSportsmonk, error) {
	once.Do(func() {
		if env == "local" {
			log.Infof("Initializing local environment")
			err := localInit()
			if err != nil {
				log.Errorf("Error initializing local environment: %v", err)
			}
		}
		seasonMatches := make([]models.SeasonMatch, 0, len(matches))
		for _, m := range matches {
			if sm, ok := m.(*models.SeasonMatch); ok {
				seasonMatches = append(seasonMatches, *sm)
			}
		}
		watcher = &MatchWatcherServiceSportsmonk{
			watchedMatches: matches,
			//repo:           repositories.NewSportsmonkRepository(api.NewSportsmonkAPI(getAPIToken())),
			repo:         repositories.NewSportsmonkRepository(api.NewFakeRandomSportsmonkAPI(seasonMatches)),
			matchRepo:    matchRepo,
			subscribers:  make(map[string]GameService),
			stopChan:     make(chan struct{}),
			pollInterval: time.Second * 30,
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
	updates, err := m.getMatchesUpdates()
	if err != nil {
		log.Errorf("Error getting match updates: \t%v", err)
		return
	}

	if len(updates) == 0 {
		return
	}

	// Save updates to the database before notifying subscribers
	for _, updatedMatch := range updates {
		if err := m.matchRepo.SaveMatch(updatedMatch); err != nil {
			log.Errorf("Failed to save updated match %s to database: %v", updatedMatch.Id(), err)
		}
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
	log.Infof("Getting last match infos for %d matches", len(m.watchedMatches))
	lastMatchInfos, err := m.repo.GetLastMatchInfos(m.watchedMatches)
	if err != nil {
		return nil, err
	}
	for matchId, match := range lastMatchInfos {
		if matchWasUpdated(match, m.watchedMatches[matchId]) {
			updates[matchId] = match
			m.watchedMatches[matchId] = match
			if match.IsFinished() {
				delete(m.watchedMatches, matchId)
			}
		}
	}
	log.Infof("Found %d updates", len(updates))
	return updates, nil
}

func matchWasUpdated(match models.Match, lastMatchState models.Match) bool {
	if match.IsFinished() != lastMatchState.IsFinished() {
		return true
	}
	if match.GetHomeTeamOdds() != lastMatchState.GetHomeTeamOdds() {
		return true
	}
	if match.GetAwayTeamOdds() != lastMatchState.GetAwayTeamOdds() {
		return true
	}
	if match.GetDrawOdds() != lastMatchState.GetDrawOdds() {
		return true
	}
	if match.GetDate() != lastMatchState.GetDate() {
		return true
	}
	if match.GetHomeGoals() != lastMatchState.GetHomeGoals() {
		return true
	}
	if match.GetAwayGoals() != lastMatchState.GetAwayGoals() {
		return true
	}
	if match.IsInProgress() != lastMatchState.IsInProgress() {
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
