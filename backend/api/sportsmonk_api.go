package api

import (
	"fmt"
	"liguain/backend/models"
)

// SportsmonkAPI is a wrapper around the Sportsmonk API
type SportsmonkAPI interface {
	// GetSeasonIds creates a mapping between the season code and the season ID
	GetSeasonIds(seasonCodes []string) (map[string]string, error)
	// GetFixturesInfos returns the fixtures infos for a given list of fixture IDs
	GetFixturesInfos(fixtureIds []string) (map[string]models.Match, error)
	// GetFixturesIds returns the fixture IDs for a given list of matches
	GetFixturesIds(matches []models.Match) (map[string]string, error)
}

type SportsmonkAPIImpl struct {
	apiToken string
}

const baseURL = "https://api.sportmonks.com/v3/football/"
const ligue1LigueId = 301

var getAllFixturesLinkLigue1 = fmt.Sprintf("%sfixtures?filters=fixtureLeagues:%d", baseURL, ligue1LigueId)

func NewSportsmonkAPI(apiToken string) *SportsmonkAPIImpl {
	return &SportsmonkAPIImpl{apiToken: apiToken}
}

func (s *SportsmonkAPIImpl) GetSeasonIds(seasonCodes []string) (map[string]string, error) {
	return nil, nil
}

func (s *SportsmonkAPIImpl) GetFixturesIds(matches []models.Match) (map[string]string, error) {
	return nil, nil
}

func (s *SportsmonkAPIImpl) GetFixturesInfos(fixtureIds []string) (map[string]models.Match, error) {
	return nil, nil
}
