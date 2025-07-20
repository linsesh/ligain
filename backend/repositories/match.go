package repositories

import (
	"ligain/backend/models"
)

const matchCacheSize = 1000 // Maximum number of matches to keep in cache

type MatchRepository interface {
	// SaveMatch saves or updates a match
	SaveMatch(match models.Match) error
	// GetMatch returns a match by its id
	GetMatch(matchId string) (models.Match, error)

	GetMatches() (map[string]models.Match, error)
	// GetMatchesByCompetitionAndSeason returns all matches for a specific competition and season
	GetMatchesByCompetitionAndSeason(competitionCode, seasonCode string) ([]models.Match, error)
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
func (r *InMemoryMatchRepository) SaveMatch(match models.Match) error {
	r.cache.Set(match.Id(), match)
	return nil
}

// GetMatch returns a match by its id
func (r *InMemoryMatchRepository) GetMatch(matchId string) (models.Match, error) {
	match, err := r.cache.Get(matchId)
	if err == nil {
		return match, nil
	}
	return nil, nil
}

func (r *InMemoryMatchRepository) GetMatches() (map[string]models.Match, error) {
	matches := make(map[string]models.Match)
	for _, entry := range r.cache.GetAll() {
		matches[entry.Key] = entry.Value
	}
	return matches, nil
}

func (r *InMemoryMatchRepository) GetMatchesByCompetitionAndSeason(competitionCode, seasonCode string) ([]models.Match, error) {
	var matches []models.Match
	for _, entry := range r.cache.GetAll() {
		match := entry.Value
		if match.GetCompetitionCode() == competitionCode && match.GetSeasonCode() == seasonCode {
			matches = append(matches, match)
		}
	}
	return matches, nil
}
