package api

import (
	"ligain/backend/models"
	"math/rand"
)

// FakeRandomSportsmonkAPI is a fake implementation of the SportsmonkAPI interface that returns random fixtures updates, for testing purposes
type FakeRandomSportsmonkAPI struct {
	fixtures map[int][]models.SeasonMatch
}

func NewFakeRandomSportsmonkAPI(matches []models.SeasonMatch) *FakeRandomSportsmonkAPI {
	fixtures := make(map[int][]models.SeasonMatch)
	for _, match := range matches {
		fixtures[match.Matchday] = append(fixtures[match.Matchday], match)
	}
	return &FakeRandomSportsmonkAPI{
		fixtures: fixtures,
	}
}

func (a *FakeRandomSportsmonkAPI) GetSeasonIds(seasonCodes []string, competitionId int) (map[string]int, error) {
	return nil, nil
}

// GetFixturesInfos will sometimes return updated fixtures, sometimes not
func (a *FakeRandomSportsmonkAPI) GetFixturesInfos(fixtureIds []int) (map[int]models.Match, error) {
	eventHappened := rand.Intn(10)
	if eventHappened < 3 {
		return nil, nil
	}

	return nil, nil
}

func (a *FakeRandomSportsmonkAPI) GetSeasonFixtures(seasonId int) map[int]models.SeasonMatch {
	fixturesWithId := make(map[int]models.SeasonMatch)
	id := 0
	for _, matches := range a.fixtures {
		for _, match := range matches {
			fixturesWithId[id] = match
			id++
		}
	}
	return fixturesWithId
}

func (a *FakeRandomSportsmonkAPI) GetCompetitionId(competitionCode string) (int, error) {
	return 0, nil
}
