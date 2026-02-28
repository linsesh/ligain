package api

import (
	"ligain/backend/models"
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// randomFixtureUpdate applies a random event to a fixture using the same distribution
// as FakeRandomSportsmonkAPI: 10% postpone, 50% goal, 40% finish.
func randomFixtureUpdate(fixture *models.SeasonMatch) {
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

// SimulatedSportsmonkAPI is a decorator that wraps a real SportsmonkAPI and injects
// random match updates into ~5% of fixtures returned by GetSeasonFixtures.
// This is used in fake mode with a real postgres DB to exercise real pgx code paths.
type SimulatedSportsmonkAPI struct {
	mu   sync.Mutex
	real SportsmonkAPI
}

func NewSimulatedSportsmonkAPI(real SportsmonkAPI) *SimulatedSportsmonkAPI {
	return &SimulatedSportsmonkAPI{real: real}
}

func (s *SimulatedSportsmonkAPI) GetSeasonIds(seasonCodes []string, competitionId int) (map[string]int, error) {
	return s.real.GetSeasonIds(seasonCodes, competitionId)
}

func (s *SimulatedSportsmonkAPI) GetFixturesInfos(fixtureIds []int) (map[int]models.Match, error) {
	return s.real.GetFixturesInfos(fixtureIds)
}

func (s *SimulatedSportsmonkAPI) GetCompetitionId(competitionCode string) (int, error) {
	return s.real.GetCompetitionId(competitionCode)
}

// GetSeasonFixtures calls the real API and then applies a random update to ~5% of fixtures.
func (s *SimulatedSportsmonkAPI) GetSeasonFixtures(seasonId int) (map[int]models.Match, error) {
	fixtures, err := s.real.GetSeasonFixtures(seasonId)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for id, m := range fixtures {
		if sm, ok := m.(*models.SeasonMatch); ok && rand.Float64() < 0.05 {
			randomFixtureUpdate(sm)
			fixtures[id] = sm
		}
	}
	return fixtures, nil
}
