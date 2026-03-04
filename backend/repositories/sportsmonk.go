package repositories

import (
	"ligain/backend/api"
	"ligain/backend/models"
	"ligain/backend/utils"
	"time"
)

const fixtureCacheTTL = 12 * time.Hour

type SportsmonkRepository interface {
	// GetLastMatchInfos returns the last match infos for a given list of matches. The matches should be from the same season/competition
	GetLastMatchInfos(matches map[string]models.Match) (map[string]models.Match, error)
}

type SportsmonkRepositoryImpl struct {
	api                            api.SportsmonkAPI
	seasonCodeToSeasonId           map[string]int
	competitionCodeToCompetitionId map[string]int
	matchIdToFixtureId             map[string]int
	lastFullFetch                  time.Time
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

// needsFullFetch returns true if we should re-fetch the full season fixture list.
// This happens when: the cache TTL has expired (including after a failed fetch,
// since lastFullFetch is only updated on success), or a watched match has no
// cached fixture ID yet.
func (r *SportsmonkRepositoryImpl) needsFullFetch(matches map[string]models.Match) bool {
	if time.Since(r.lastFullFetch) > fixtureCacheTTL {
		return true
	}
	for matchId := range matches {
		if _, ok := r.matchIdToFixtureId[matchId]; !ok {
			return true
		}
	}
	return false
}

func (r *SportsmonkRepositoryImpl) askAndCacheFixtureInfo(seasonId int, matches map[string]models.Match) (map[string]models.Match, error) {
	if r.needsFullFetch(matches) {
		return r.fetchAndCacheAllFixtures(seasonId, matches)
	}
	return r.fetchWatchedFixturesFromCache(matches)
}

// fetchAndCacheAllFixtures performs a full season fetch, caches the matchId→fixtureId
// mapping, and returns only the watched matches.
func (r *SportsmonkRepositoryImpl) fetchAndCacheAllFixtures(seasonId int, matches map[string]models.Match) (map[string]models.Match, error) {
	fixtureIdToMatch, err := r.api.GetSeasonFixtures(seasonId)
	if err != nil {
		return nil, err
	}

	r.lastFullFetch = time.Now()
	for fixtureId, match := range fixtureIdToMatch {
		r.matchIdToFixtureId[match.Id()] = fixtureId
	}

	matchIdToMatch := make(map[string]models.Match)
	for _, match := range fixtureIdToMatch {
		if _, ok := matches[match.Id()]; ok {
			matchIdToMatch[match.Id()] = match
		}
	}
	return matchIdToMatch, nil
}

// fetchWatchedFixturesFromCache uses cached fixture IDs to fetch only the watched
// matches via the targeted GetFixturesInfos endpoint.
func (r *SportsmonkRepositoryImpl) fetchWatchedFixturesFromCache(matches map[string]models.Match) (map[string]models.Match, error) {
	fixtureIds := make([]int, 0, len(matches))
	for matchId := range matches {
		fixtureIds = append(fixtureIds, r.matchIdToFixtureId[matchId])
	}

	fixtureIdToMatch, err := r.api.GetFixturesInfos(fixtureIds)
	if err != nil {
		return nil, err
	}

	matchIdToMatch := make(map[string]models.Match)
	for _, match := range fixtureIdToMatch {
		matchIdToMatch[match.Id()] = match
	}
	return matchIdToMatch, nil
}

func (r *SportsmonkRepositoryImpl) askAndCacheSeasonId(matches map[string]models.Match) (int, error) {
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
		seasonId := seasonIds[seasonCode]
		r.seasonCodeToSeasonId[seasonCode] = seasonId
		return seasonId, nil
	}

	return r.seasonCodeToSeasonId[seasonCode], nil
}
