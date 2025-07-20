package repositories

import (
	"ligain/backend/api"
	"ligain/backend/models"
	"ligain/backend/utils"
)

type SportsmonkRepository interface {
	// GetLastMatchInfos returns the last match infos for a given list of matches. The matches should be from the same season/competition
	GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error)
}

type SportsmonkRepositoryImpl struct {
	api api.SportsmonkAPI
	// Fixture ID is the Sportsmonk ID for a match, we use it as a cache to avoid reconverting the match ID to a fixture ID
	matchIdToFixtureId             map[string]int
	seasonCodeToSeasonId           map[string]int
	competitionCodeToCompetitionId map[string]int
}

func NewSportsmonkRepository(api api.SportsmonkAPI) SportsmonkRepository {
	return &SportsmonkRepositoryImpl{
		api:                            api,
		seasonCodeToSeasonId:           make(map[string]int),
		competitionCodeToCompetitionId: make(map[string]int),
		matchIdToFixtureId:             make(map[string]int),
	}
}

func (r *SportsmonkRepositoryImpl) GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error) {
	if len(matches) == 0 {
		return make(map[string]models.Match), nil
	}
	fixtureInfos, err := r.getFixtureInfos(matches)
	if err != nil {
		return nil, err
	}
	return fixtureInfos, nil
}

func (r *SportsmonkRepositoryImpl) getFixtureInfos(matches map[string]models.Match) (map[string]models.Match, error) {
	seasonId, err := r.askAndCacheSeasonId(matches)
	if err != nil {
		return nil, err
	}

	fixtureInfos, err := r.askAndCacheFixtureInfo(seasonId, matches)
	if err != nil {
		return nil, err
	}
	return fixtureInfos, nil
}

func (r *SportsmonkRepositoryImpl) askAndCacheFixtureInfo(seasonId int, matches map[string]models.Match) (map[string]models.Match, error) {
	fixtureIdToMatch, err := r.api.GetSeasonFixtures(seasonId)
	if err != nil {
		return nil, err
	}

	matchIdToMatch := make(map[string]models.Match)
	for fixtureId, match := range fixtureIdToMatch {
		if _, ok := matches[match.Id()]; ok {
			matchIdToMatch[match.Id()] = match
		}
		r.matchIdToFixtureId[match.Id()] = fixtureId
	}

	return matchIdToMatch, nil
}

func (r *SportsmonkRepositoryImpl) askAndCacheSeasonId(matches map[string]models.Match) (int, error) {
	var seasonId int
	matchesSlice := utils.MapValues(matches)

	seasonCode := matchesSlice[0].GetSeasonCode()
	competitionCode := matchesSlice[0].GetCompetitionCode()
	if _, ok := r.competitionCodeToCompetitionId[competitionCode]; !ok {
		competitionId, err := r.api.GetCompetitionId(competitionCode)
		if err != nil {
			return -1, err
		}
		r.competitionCodeToCompetitionId[competitionCode] = competitionId
	}

	if _, ok := r.seasonCodeToSeasonId[seasonCode]; !ok {
		seasonIds, err := r.api.GetSeasonIds([]string{seasonCode}, r.competitionCodeToCompetitionId[competitionCode])
		if err != nil {
			return -1, err
		}
		seasonId = seasonIds[seasonCode]
		r.seasonCodeToSeasonId[seasonCode] = seasonId
	}
	return seasonId, nil
}
