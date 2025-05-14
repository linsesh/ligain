package services

import (
	"context"
	"liguain/backend/api"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/joho/godotenv"
)

type MatchWatcherServiceSportsmonk struct {
	watchedMatches map[string]models.Match
	repo           repositories.SportsmonkRepository
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
		}
	})
	return watcher, nil
}

func (m *MatchWatcherServiceSportsmonk) WatchMatches(matches []models.Match) {
	for _, match := range matches {
		m.watchedMatches[match.Id()] = match
	}
}

func (m *MatchWatcherServiceSportsmonk) GetUpdates(ctx context.Context, done chan MatchWatcherServiceResult) {
	var updates = make(map[string]models.Match)
	updates, err := m.getMatchesUpdates()
	done <- MatchWatcherServiceResult{
		Value: updates,
		Err:   err,
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
