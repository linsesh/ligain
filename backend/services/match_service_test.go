package services

import (
	"errors"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// These tests exercise MatchService alone: its repository is a mock and its subscribers are fakes.

var (
	serviceNow     = time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	serviceKickoff = serviceNow.Add(7 * time.Hour)
)

func scheduledMatch(home, away string) *models.SeasonMatch {
	return models.NewSeasonMatchWithKnownOdds(home, away, "2026/2027", "Ligue 1", serviceKickoff, 11, 2.1, 3.2, 3.3)
}

func finishedUpdate(match models.Match) models.MatchUpdate {
	update := models.MatchUpdateOf(match)
	update.Status, update.HomeGoals, update.AwayGoals = models.MatchStatusFinished, 2, 1
	return update
}

// notifiedGame is a GameService fake that reports what MatchService tells it on a channel
type notifiedGame struct {
	*MockGameService
	received chan map[string]models.Match
}

func newNotifiedGame(gameID string) *notifiedGame {
	return &notifiedGame{MockGameService: NewMockGameService(gameID), received: make(chan map[string]models.Match, 10)}
}

func (g *notifiedGame) HandleMatchUpdates(updates map[string]models.Match) error {
	g.received <- updates
	return nil
}

func (g *notifiedGame) nothingReceived(t *testing.T) {
	select {
	case updates := <-g.received:
		t.Fatalf("game %s was not expected to be notified, got %v", g.GetGameID(), updates)
	case <-time.After(30 * time.Millisecond):
	}
}

func (g *notifiedGame) receivedWithin(t *testing.T) map[string]models.Match {
	select {
	case updates := <-g.received:
		return updates
	case <-time.After(time.Second):
		t.Fatalf("game %s was not notified", g.GetGameID())
		return nil
	}
}

func TestMatchService_TryUpdateMatches_SavesTheUpdatedMatchAndTellsEverySubscriber(t *testing.T) {
	current := scheduledMatch("Lens", "Toulouse")
	repo := new(MockMatchRepository)
	repo.On("GetMatch", current.Id()).Return(current, nil)
	repo.On("SaveMatch", mock.MatchedBy(func(m models.Match) bool { return m.Id() == current.Id() && m.IsFinished() && m.GetHomeGoals() == 2 })).Return(nil)
	service := NewMatchServiceWithTime(repo, fixedTime(serviceKickoff.Add(2*time.Hour)))
	game1, game2 := newNotifiedGame("game1"), newNotifiedGame("game2")
	require.NoError(t, service.Subscribe(game1))
	require.NoError(t, service.Subscribe(game2))

	err := service.TryUpdateMatches(map[string]models.MatchUpdate{current.Id(): finishedUpdate(current)})

	require.NoError(t, err)
	repo.AssertNumberOfCalls(t, "SaveMatch", 1)
	updates1, updates2 := game1.receivedWithin(t), game2.receivedWithin(t)
	for _, updates := range []map[string]models.Match{updates1, updates2} {
		require.Contains(t, updates, current.Id())
		assert.True(t, updates[current.Id()].IsFinished())
	}
	// each game gets its own copy
	updates1[current.Id()].(*models.SeasonMatch).HomeGoals = 99
	assert.Equal(t, 2, updates2[current.Id()].GetHomeGoals())
}

func TestMatchService_TryUpdateMatches_NothingChangedMeansNothingSavedNorSent(t *testing.T) {
	current := scheduledMatch("Lens", "Toulouse")
	repo := new(MockMatchRepository)
	repo.On("GetMatch", current.Id()).Return(current, nil)
	service := NewMatchServiceWithTime(repo, fixedTime(serviceNow))
	game := newNotifiedGame("game1")
	require.NoError(t, service.Subscribe(game))

	err := service.TryUpdateMatches(map[string]models.MatchUpdate{current.Id(): models.MatchUpdateOf(current)})

	require.NoError(t, err)
	repo.AssertNotCalled(t, "SaveMatch", mock.Anything)
	game.nothingReceived(t)
}

func TestMatchService_TryUpdateMatches_OddsRefusedNearKickoffMeansNothingSavedNorSent(t *testing.T) {
	current := scheduledMatch("Lens", "Toulouse")
	repo := new(MockMatchRepository)
	repo.On("GetMatch", current.Id()).Return(current, nil)
	service := NewMatchServiceWithTime(repo, fixedTime(serviceKickoff.Add(-time.Minute)))
	game := newNotifiedGame("game1")
	require.NoError(t, service.Subscribe(game))
	update := models.MatchUpdateOf(current)
	update.HomeTeamOdds, update.DrawOdds = 21, 1.02

	err := service.TryUpdateMatches(map[string]models.MatchUpdate{current.Id(): update})

	require.NoError(t, err)
	repo.AssertNotCalled(t, "SaveMatch", mock.Anything)
	game.nothingReceived(t)
}

func TestMatchService_TryUpdateMatches_FinishedMatchWithoutUsableOddsIsRejected(t *testing.T) {
	current := models.NewSeasonMatch("Lens", "Toulouse", "2026/2027", "Ligue 1", serviceKickoff, 11) // never had odds
	repo := new(MockMatchRepository)
	repo.On("GetMatch", current.Id()).Return(current, nil)
	service := NewMatchServiceWithTime(repo, fixedTime(serviceKickoff.Add(3*time.Hour)))
	game := newNotifiedGame("game1")
	require.NoError(t, service.Subscribe(game))

	err := service.TryUpdateMatches(map[string]models.MatchUpdate{current.Id(): finishedUpdate(current)})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing odds")
	repo.AssertNotCalled(t, "SaveMatch", mock.Anything)
	game.nothingReceived(t)
}

func TestMatchService_TryUpdateMatches_AFailingMatchDoesNotStopTheOthers(t *testing.T) {
	unknown := scheduledMatch("Nice", "Lille")
	unsavable := scheduledMatch("Monaco", "Auxerre")
	fine := scheduledMatch("Lens", "Toulouse")
	repo := new(MockMatchRepository)
	repo.On("GetMatch", unknown.Id()).Return(nil, repositories.ErrMatchNotFound)
	repo.On("GetMatch", unsavable.Id()).Return(unsavable, nil)
	repo.On("GetMatch", fine.Id()).Return(fine, nil)
	repo.On("SaveMatch", mock.MatchedBy(func(m models.Match) bool { return m.Id() == unsavable.Id() })).Return(errors.New("database down"))
	repo.On("SaveMatch", mock.MatchedBy(func(m models.Match) bool { return m.Id() == fine.Id() })).Return(nil)
	service := NewMatchServiceWithTime(repo, fixedTime(serviceKickoff.Add(2*time.Hour)))
	game := newNotifiedGame("game1")
	require.NoError(t, service.Subscribe(game))

	err := service.TryUpdateMatches(map[string]models.MatchUpdate{
		unknown.Id():   finishedUpdate(unknown),
		unsavable.Id(): finishedUpdate(unsavable),
		fine.Id():      finishedUpdate(fine),
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, repositories.ErrMatchNotFound)
	assert.Contains(t, err.Error(), "database down")
	updates := game.receivedWithin(t)
	assert.Len(t, updates, 1, "only the match that was saved is sent")
	assert.Contains(t, updates, fine.Id())
}

func TestMatchService_UnsubscribedGameIsNotTold(t *testing.T) {
	current := scheduledMatch("Lens", "Toulouse")
	repo := new(MockMatchRepository)
	repo.On("GetMatch", current.Id()).Return(current, nil)
	repo.On("SaveMatch", mock.Anything).Return(nil)
	service := NewMatchServiceWithTime(repo, fixedTime(serviceKickoff.Add(2*time.Hour)))
	stays, leaves := newNotifiedGame("stays"), newNotifiedGame("leaves")
	require.NoError(t, service.Subscribe(stays))
	require.NoError(t, service.Subscribe(leaves))
	require.NoError(t, service.Unsubscribe("leaves"))

	require.NoError(t, service.TryUpdateMatches(map[string]models.MatchUpdate{current.Id(): finishedUpdate(current)}))

	stays.receivedWithin(t)
	leaves.nothingReceived(t)
}

func TestMatchService_SubscribeAndUnsubscribeAreIdempotent(t *testing.T) {
	service := NewMatchService(new(MockMatchRepository))
	game := newNotifiedGame("game1")

	require.NoError(t, service.Subscribe(game))
	require.NoError(t, service.Subscribe(game))
	require.NoError(t, service.Unsubscribe("game1"))
	require.NoError(t, service.Unsubscribe("game1"))
}

func TestMatchService_GetMatch(t *testing.T) {
	match := scheduledMatch("Lens", "Toulouse")
	repo := new(MockMatchRepository)
	repo.On("GetMatch", match.Id()).Return(match, nil)
	repo.On("GetMatch", "unknown").Return(nil, repositories.ErrMatchNotFound)
	service := NewMatchService(repo)

	found, err := service.GetMatch(match.Id())
	require.NoError(t, err)
	assert.Equal(t, match.Id(), found.Id())

	_, err = service.GetMatch("unknown")
	assert.ErrorIs(t, err, repositories.ErrMatchNotFound)
}

func TestMatchService_GetMatches(t *testing.T) {
	match := scheduledMatch("Lens", "Toulouse")
	t.Run("unknown ids are skipped", func(t *testing.T) {
		repo := new(MockMatchRepository)
		repo.On("GetMatch", match.Id()).Return(match, nil)
		repo.On("GetMatch", "unknown").Return(nil, repositories.ErrMatchNotFound)

		matches, err := NewMatchService(repo).GetMatches([]string{match.Id(), "unknown"})

		require.NoError(t, err)
		assert.Len(t, matches, 1)
		assert.Contains(t, matches, match.Id())
	})
	t.Run("any other error is returned", func(t *testing.T) {
		repo := new(MockMatchRepository)
		repo.On("GetMatch", match.Id()).Return(nil, errors.New("database down"))

		_, err := NewMatchService(repo).GetMatches([]string{match.Id()})

		assert.ErrorContains(t, err, "database down")
	})
}

func TestMatchService_GetMatchesInHorizon(t *testing.T) {
	inThreeDays := models.NewSeasonMatch("A", "B", "2026/2027", "Ligue 1", serviceNow.Add(3*24*time.Hour), 12)
	inThreeWeeks := models.NewSeasonMatch("C", "D", "2026/2027", "Ligue 1", serviceNow.Add(21*24*time.Hour), 14)
	finished := models.NewFinishedSeasonMatch("E", "F", 1, 0, "2026/2027", "Ligue 1", serviceNow.Add(-24*time.Hour), 11, 1.5, 2.5, 3.0)
	pastNotFinished := models.NewSeasonMatch("G", "H", "2026/2027", "Ligue 1", serviceNow.Add(-2*time.Hour), 11)
	otherCompetition := models.NewSeasonMatch("I", "J", "2026/2027", "Serie A", serviceNow.Add(24*time.Hour), 11)
	oldSeason := models.NewSeasonMatch("K", "L", "2025/2026", "Ligue 1", serviceNow.Add(-190*24*time.Hour), 38)
	all := map[string]models.Match{}
	for _, m := range []models.Match{inThreeDays, inThreeWeeks, finished, pastNotFinished, otherCompetition, oldSeason} {
		all[m.Id()] = m
	}
	repo := new(MockMatchRepository)
	repo.On("GetMatches").Return(all, nil)

	t.Run("not finished matches within the horizon", func(t *testing.T) {
		toWatch, err := NewMatchServiceWithTime(repo, fixedTime(serviceNow)).GetMatchesInHorizon(14 * 24 * time.Hour)

		require.NoError(t, err)
		assert.Len(t, toWatch, 4) // the unfinished past match may have been rescheduled
		assert.Contains(t, toWatch, inThreeDays.Id())
		assert.Contains(t, toWatch, pastNotFinished.Id())
		assert.NotContains(t, toWatch, inThreeWeeks.Id())
		assert.NotContains(t, toWatch, finished.Id())
	})
	t.Run("restricted to one competition and season", func(t *testing.T) {
		service := NewMatchServiceWithTime(repo, fixedTime(serviceNow), WatchOnly("Ligue 1", "2026/2027"))

		toWatch, err := service.GetMatchesInHorizon(14 * 24 * time.Hour)

		require.NoError(t, err)
		assert.Len(t, toWatch, 2)
		assert.NotContains(t, toWatch, otherCompetition.Id())
		assert.NotContains(t, toWatch, oldSeason.Id())
	})
	t.Run("repository error is returned", func(t *testing.T) {
		failing := new(MockMatchRepository)
		failing.On("GetMatches").Return(nil, errors.New("database down"))

		_, err := NewMatchService(failing).GetMatchesInHorizon(time.Hour)

		assert.ErrorContains(t, err, "database down")
	})
}
