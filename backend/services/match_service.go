package services

import (
	"errors"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// MatchService is the single owner of match state. It is the only place accepting provider (Sportsmonk)
// updates: it applies the freeze rules, persists the match and tells the subscribed games.
type MatchService interface {
	// GetMatch returns a copy of the match, or repositories.ErrMatchNotFound
	GetMatch(id string) (models.Match, error)
	// GetMatches returns copies of the matches with these ids (unknown ids are skipped)
	GetMatches(ids []string) (map[string]models.Match, error)
	// GetMatchesInHorizon returns the not finished matches scheduled within the horizon
	// (matches already past but not finished are included: they may have been rescheduled)
	GetMatchesInHorizon(horizon time.Duration) (map[string]models.Match, error)
	// TryUpdateMatches tries to apply the updates (a match may refuse parts of one, such as odds near kickoff),
	// persists the matches that really changed and notifies the subscribed listeners with their ids
	TryUpdateMatches(updates map[string]models.MatchUpdate) error
	// Subscribe makes a listener receive the ids of the matches that changed
	Subscribe(handler MatchUpdateListener) error
	// Unsubscribe stops a listener from receiving the matches that changed
	Unsubscribe(gameID string) error
}

// MatchUpdateListener receives notifications when matches change.
// It is implemented by services that need to react to match updates (e.g., GameService).
type MatchUpdateListener interface {
	GetGameID() string
	OnMatchesUpdated(ids []string) error
}

type MatchServiceImpl struct {
	repo        repositories.MatchRepository
	now         func() time.Time
	competition string // when set, only this competition and season are watched
	season      string
	mu          sync.RWMutex
	subscribers map[string]MatchUpdateListener
}

// MatchServiceOption customizes a MatchService
type MatchServiceOption func(*MatchServiceImpl)

// WatchOnly restricts the matches asked to the provider to one competition and season
func WatchOnly(competition, season string) MatchServiceOption {
	return func(s *MatchServiceImpl) { s.competition, s.season = competition, season }
}

func NewMatchService(repo repositories.MatchRepository, opts ...MatchServiceOption) *MatchServiceImpl {
	return NewMatchServiceWithTime(repo, time.Now, opts...)
}

// NewMatchServiceWithTime creates a MatchService with a custom clock for testing
func NewMatchServiceWithTime(repo repositories.MatchRepository, now func() time.Time, opts ...MatchServiceOption) *MatchServiceImpl {
	s := &MatchServiceImpl{repo: repo, now: now, subscribers: make(map[string]MatchUpdateListener)}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *MatchServiceImpl) GetMatch(id string) (models.Match, error) {
	return s.repo.GetMatch(id)
}

func (s *MatchServiceImpl) GetMatches(ids []string) (map[string]models.Match, error) {
	matches := make(map[string]models.Match, len(ids))
	for _, id := range ids {
		match, err := s.repo.GetMatch(id)
		if errors.Is(err, repositories.ErrMatchNotFound) {
			continue
		}
		if err != nil {
			return nil, err
		}
		matches[id] = match
	}
	return matches, nil
}

func (s *MatchServiceImpl) GetMatchesInHorizon(horizon time.Duration) (map[string]models.Match, error) {
	all, err := s.repo.GetMatches()
	if err != nil {
		return nil, err
	}
	limit := s.now().Add(horizon)
	toWatch := make(map[string]models.Match)
	for id, match := range all {
		if s.competition != "" && (match.GetCompetitionCode() != s.competition || match.GetSeasonCode() != s.season) {
			continue
		}
		if !match.IsFinished() && !match.GetDate().After(limit) {
			toWatch[id] = match
		}
	}
	return toWatch, nil
}

func (s *MatchServiceImpl) TryUpdateMatches(updates map[string]models.MatchUpdate) error {
	changed := make(map[string]models.Match)
	var errs []error
	for id, update := range updates {
		current, err := s.repo.GetMatch(id)
		if err != nil {
			errs = append(errs, fmt.Errorf("match %s: %w", id, err))
			continue
		}
		result, updated := current.TryUpdate(update, s.now())
		if !updated {
			continue
		}
		logChanges(current, result)
		if result.IsFinished() && !hasUsableOdds(result) {
			errs = append(errs, fmt.Errorf("match %s: missing odds for finished match", id))
			continue
		}
		if err := s.repo.SaveMatch(result); err != nil {
			errs = append(errs, fmt.Errorf("match %s: save: %w", id, err))
			continue
		}
		changed[id] = result
	}
	s.publish(changed)
	return errors.Join(errs...)
}

// publish sends the changed match IDs to every subscribed listener, each one in its own goroutine
func (s *MatchServiceImpl) publish(changed map[string]models.Match) {
	if len(changed) == 0 {
		return
	}
	log.Infof("Found %d match updates, notifying subscribers", len(changed))
	// Subscribers change while we notify (a game is created or emptied by an HTTP request), so they are read
	// under the lock. Nothing running under it can block: every listener is called in its own goroutine,
	// so a listener can subscribe or unsubscribe without deadlocking.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Extract IDs from changed matches
	ids := make([]string, 0, len(changed))
	for id := range changed {
		ids = append(ids, id)
	}

	for _, handler := range s.subscribers {
		go func(handler MatchUpdateListener, ids []string) {
			if err := handler.OnMatchesUpdated(ids); err != nil {
				log.Errorf("Error handling updates for game %s: %v", handler.GetGameID(), err)
			}
		}(handler, ids)
	}
}

func (s *MatchServiceImpl) Subscribe(handler MatchUpdateListener) error {
	gameID := handler.GetGameID()
	s.mu.Lock()
	s.subscribers[gameID] = handler
	s.mu.Unlock()
	log.Infof("Game %s subscribed to match updates", gameID)
	return nil
}

func (s *MatchServiceImpl) Unsubscribe(gameID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.subscribers[gameID]; !exists {
		return nil
	}
	delete(s.subscribers, gameID)
	log.Infof("Game %s unsubscribed from match updates", gameID)
	return nil
}

// logChanges logs what differs between two states of the same match
func logChanges(before, after models.Match) {
	id := after.Id()
	if before.GetStatus() != after.GetStatus() {
		log.Infof("Status updated from %v to %v in match %s", before.GetStatus(), after.GetStatus(), id)
	}
	if before.GetHomeGoals() != after.GetHomeGoals() {
		log.Infof("Home goals updated from %v to %v in match %s", before.GetHomeGoals(), after.GetHomeGoals(), id)
	}
	if before.GetAwayGoals() != after.GetAwayGoals() {
		log.Infof("Away goals updated from %v to %v in match %s", before.GetAwayGoals(), after.GetAwayGoals(), id)
	}
	if before.GetHomeTeamOdds() != after.GetHomeTeamOdds() {
		log.Infof("Home team odds updated from %v to %v in match %s", before.GetHomeTeamOdds(), after.GetHomeTeamOdds(), id)
	}
	if before.GetAwayTeamOdds() != after.GetAwayTeamOdds() {
		log.Infof("Away team odds updated from %v to %v in match %s", before.GetAwayTeamOdds(), after.GetAwayTeamOdds(), id)
	}
	if before.GetDrawOdds() != after.GetDrawOdds() {
		log.Infof("Draw odds updated from %v to %v in match %s", before.GetDrawOdds(), after.GetDrawOdds(), id)
	}
	if !before.GetDate().Equal(after.GetDate()) {
		log.Infof("Date updated from %v to %v in match %s", before.GetDate(), after.GetDate(), id)
	}
}

func hasUsableOdds(match models.Match) bool {
	return match.GetHomeTeamOdds() > 0 && match.GetAwayTeamOdds() > 0 && match.GetDrawOdds() > 0
}
