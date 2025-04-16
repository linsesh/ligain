package repositories

import (
	"liguain/backend/api"
	"liguain/backend/models"
)

type SportsmonkRepository interface {
	// GetLastMatchInfos returns the last match infos for a given list of matches
	GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error)
}

type SportsmonkRepositoryImpl struct {
	apiToken string
	api      api.SportsmonkAPI
	// Fixture ID is the Sportsmonk ID for a match, we use it as a cache to avoid reconverting the match ID to a fixture ID
	matchIdToFixtureId   map[string]string
	seasonCodeToSeasonId map[string]string
}

func NewSportsmonkRepository(apiToken string) SportsmonkRepository {
	return &SportsmonkRepositoryImpl{
		apiToken:             apiToken,
		api:                  api.NewSportsmonkAPI(apiToken),
		matchIdToFixtureId:   make(map[string]string),
		seasonCodeToSeasonId: make(map[string]string),
	}
}

func (r *SportsmonkRepositoryImpl) GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error) {
	fixtureIds, err := r.getFixtureIds(matches)
	if err != nil {
		return nil, err
	}
	fixtureInfos, err := r.api.GetFixturesInfos(fixtureIds)
	if err != nil {
		return nil, err
	}
	return fixtureInfos, nil
}

func (r *SportsmonkRepositoryImpl) getFixtureIds(matches map[string]models.Match) ([]string, error) {
	err := r.askAndCacheFixtureIdAndSeasonId(matches)
	if err != nil {
		return nil, err
	}

	fixtureIds := make([]string, 0)
	for _, match := range matches {
		fixtureIds = append(fixtureIds, r.matchIdToFixtureId[match.Id()])
	}

	return fixtureIds, nil
}

func (r *SportsmonkRepositoryImpl) askAndCacheFixtureIdAndSeasonId(matches map[string]models.Match) error {
	matchesToConvert := make([]models.Match, 0)
	matchesSeasonCodesToConvert := make([]string, 0)
	// Get the list of matches and seasons to convert at once to reduce the number of requests
	for _, match := range matches {
		if _, ok := r.matchIdToFixtureId[match.Id()]; !ok {
			matchesToConvert = append(matchesToConvert, match)
		}
		if _, ok := r.seasonCodeToSeasonId[match.GetSeasonCode()]; !ok {
			matchesSeasonCodesToConvert = append(matchesSeasonCodesToConvert, match.GetSeasonCode())
		}
	}
	fixtureIds, err := r.api.GetFixtureIds(matchesToConvert)
	if err != nil {
		return err
	}
	seasonsIds, err := r.api.GetSeasonIds(matchesSeasonCodesToConvert)
	if err != nil {
		return err
	}
	for _, match := range matchesToConvert {
		r.matchIdToFixtureId[match.Id()] = fixtureIds[match.Id()]
		r.seasonCodeToSeasonId[match.GetSeasonCode()] = seasonsIds[match.GetSeasonCode()]
	}
	return nil
}
