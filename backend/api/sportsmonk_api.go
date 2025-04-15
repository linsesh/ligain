package api

import "liguain/backend/models"

// SportsmonkAPI is a wrapper around the Sportsmonk API
type SportsmonkAPI interface {
	// GetFixturesIdsForSeason returns all fixtures ids, for a given season, and convert to our Match model
	GetFixturesIdsForSeason(seasonId string) (map[string]models.Match, error)
}

type SportsmonkAPIImpl struct {
	apiToken string
}

func NewSportsmonkAPI(apiToken string) *SportsmonkAPIImpl {
	return &SportsmonkAPIImpl{apiToken: apiToken}
}

func (s *SportsmonkAPIImpl) GetFixturesIdsForSeason(seasonId string) (map[string]models.Match, error) {
	return nil, nil
}
