package repositories

import "liguain/backend/models"

type SportsmonkRepository interface {
	GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error)
}

type SportsmonkRepositoryImpl struct {
	apiToken string
	// Fixture ID is the Sportsmonk ID for a match, we use it as a cache to avoid reconverting the match ID to a fixture ID
	matchIdToFixtureId map[string]string
}

func NewSportsmonkRepository(apiToken string) SportsmonkRepository {
	return &SportsmonkRepositoryImpl{
		apiToken:           apiToken,
		matchIdToFixtureId: make(map[string]string),
	}
}

func (r *SportsmonkRepositoryImpl) GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error) {
	return nil, nil
}
