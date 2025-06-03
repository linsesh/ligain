package repositories

import (
	"liguain/backend/models"
)

const matchCacheSize = 1000 // Maximum number of matches to keep in cache

type MatchRepository interface {
	// SaveMatch saves or updates a match and returns the match id
	SaveMatch(match models.Match) (string, error)
	// GetMatch returns a match by its id
	GetMatch(matchId string) (models.Match, error)
	// GetMatchesByGame returns all matches for a given game
	GetMatchesByGame(gameId string) ([]models.Match, error)
}

// InMemoryMatchRepository is a simple in-memory implementation of MatchRepository
type InMemoryMatchRepository struct {
	cache *Cache[string, models.Match]
}

// NewInMemoryMatchRepository creates a new instance of InMemoryMatchRepository
func NewInMemoryMatchRepository() *InMemoryMatchRepository {
	return &InMemoryMatchRepository{
		cache: NewCache[string, models.Match](matchCacheSize),
	}
}

// SaveMatch saves or updates a match and returns the match id
func (r *InMemoryMatchRepository) SaveMatch(match models.Match) (string, error) {
	id := match.Id()
	r.cache.Set(id, match)
	return id, nil
}

// GetMatch returns a match by its id
func (r *InMemoryMatchRepository) GetMatch(matchId string) (models.Match, error) {
	match, err := r.cache.Get(matchId)
	if err == nil {
		return match, nil
	}
	return nil, nil
}

// GetMatchesByGame returns all matches for a given game
func (r *InMemoryMatchRepository) GetMatchesByGame(gameId string) ([]models.Match, error) {
	var matches []models.Match
	for _, entry := range r.cache.GetAll() {
		if entry.Value.GetCompetitionCode() == gameId {
			matches = append(matches, entry.Value)
		}
	}
	return matches, nil
}
