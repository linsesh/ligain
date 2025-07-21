package api

import (
	"ligain/backend/models"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

// FakeRandomSportsmonkAPI is a fake implementation of the SportsmonkAPI interface that returns random fixtures updates, for testing purposes
type FakeRandomSportsmonkAPI struct {
	fixtures           map[int][]models.SeasonMatch
	matchIdToFixtureId map[string]int
	currentMatchday    int
}

func NewFakeRandomSportsmonkAPI(matches []models.SeasonMatch) *FakeRandomSportsmonkAPI {
	fixtures := make(map[int][]models.SeasonMatch)
	for _, match := range matches {
		fixtures[match.Matchday] = append(fixtures[match.Matchday], match)
	}
	matchIdToFixtureId := make(map[string]int)
	id := 0
	for _, matches := range fixtures {
		for _, match := range matches {
			matchIdToFixtureId[match.Id()] = id
			id++
		}
	}
	return &FakeRandomSportsmonkAPI{
		fixtures:           fixtures,
		matchIdToFixtureId: matchIdToFixtureId,
		currentMatchday:    1,
	}
}

func (a *FakeRandomSportsmonkAPI) GetSeasonIds(seasonCodes []string, competitionId int) (map[string]int, error) {
	return nil, nil
}

// GetFixturesInfos will sometimes return updated fixtures, sometimes not
func (a *FakeRandomSportsmonkAPI) GetFixturesInfos(fixtureIds []int) (map[int]models.Match, error) {
	if a.currentMatchday > len(a.fixtures) {
		return nil, nil
	}

	// 70% chance of an event happening
	eventHappened := rand.Intn(10)
	if eventHappened < 3 {
		return nil, nil
	}
	fixtures := a.fixtures[a.currentMatchday]
	// An event happened, we need to choose on how many fixtures we want to update
	numberOfFixturesToUpdate := rand.Intn(len(fixtures))

	// We update the fixtures
	for i := 0; i < numberOfFixturesToUpdate; i++ {
		fixture := fixtures[i]
		a.randomFixtureUpdate(fixture)
		if fixture.IsFinished() {
			delete(a.fixtures, i)
		}
	}

	if len(a.fixtures[a.currentMatchday]) == 0 {
		a.currentMatchday++
	}
	return nil, nil
}

func (a *FakeRandomSportsmonkAPI) randomFixtureUpdate(fixture models.SeasonMatch) {
	log.Infof("Updating fixture %s", fixture.Id())
	whichEvent := rand.Intn(10)
	// 10% of chance of being postponed (between 1 and 3 days)
	if whichEvent == 0 {
		log.Infof("Postponing fixture %s", fixture.Id())
		fixture.Date = fixture.Date.Add(time.Duration(rand.Intn(2)+1) * time.Hour * 24)
		return
	}
	// 50% of chance of having a random goal
	if whichEvent < 6 {
		log.Infof("Goal !! %s", fixture.Id())
		if !fixture.IsInProgress() {
			fixture.Start()
		}
		whichTeam := rand.Intn(2)
		if whichTeam == 0 {
			fixture.HomeGoals++
		} else {
			fixture.AwayGoals++
		}
		return
	}
	// 40% of chance of being finished now
	log.Infof("Finishing fixture %s", fixture.Id())
	fixture.Finish(fixture.HomeGoals, fixture.AwayGoals)
}

func (a *FakeRandomSportsmonkAPI) GetSeasonFixtures(seasonId int) map[int]models.SeasonMatch {
	fixtures := make(map[int]models.SeasonMatch)
	for i := 0; i < len(a.fixtures); i++ {
		for _, match := range a.fixtures[i] {
			fixtures[a.matchIdToFixtureId[match.Id()]] = match
		}
	}
	return fixtures
}

func (a *FakeRandomSportsmonkAPI) GetCompetitionId(competitionCode string) (int, error) {
	return 0, nil
}
