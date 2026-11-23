package repositories

import (
	"ligain/backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMatch() *models.SeasonMatch {
	return models.NewSeasonMatchWithKnownOdds("Lens", "Toulouse", "2026/2027", "Ligue 1", time.Date(2026, 11, 21, 19, 45, 0, 0, time.UTC), 11, 2.1, 3.2, 3.3)
}

func TestInMemoryMatchRepository_GetMatch_UnknownMatch(t *testing.T) {
	repo := NewInMemoryMatchRepository()

	match, err := repo.GetMatch("unknown")

	assert.ErrorIs(t, err, ErrMatchNotFound)
	assert.Nil(t, match)
}

func TestInMemoryMatchRepository_ReturnsCopies(t *testing.T) {
	repo := NewInMemoryMatchRepository()
	saved := newTestMatch()
	require.NoError(t, repo.SaveMatch(saved))

	t.Run("changing the saved match afterwards does not change the repository", func(t *testing.T) {
		saved.SetHomeTeamOdds(99)

		stored, err := repo.GetMatch(saved.Id())
		require.NoError(t, err)
		assert.Equal(t, 2.1, stored.GetHomeTeamOdds())
	})

	t.Run("changing a returned match does not change the repository", func(t *testing.T) {
		first, err := repo.GetMatch(saved.Id())
		require.NoError(t, err)
		first.SetHomeTeamOdds(77)
		first.Start()

		second, err := repo.GetMatch(saved.Id())
		require.NoError(t, err)
		assert.Equal(t, 2.1, second.GetHomeTeamOdds())
		assert.False(t, second.IsInProgress())
	})

	t.Run("changing a match returned in a list does not change the repository", func(t *testing.T) {
		all, err := repo.GetMatches()
		require.NoError(t, err)
		all[saved.Id()].SetHomeTeamOdds(55)

		inCompetition, err := repo.GetMatchesByCompetitionAndSeason("Ligue 1", "2026/2027")
		require.NoError(t, err)
		require.Len(t, inCompetition, 1)
		assert.Equal(t, 2.1, inCompetition[0].GetHomeTeamOdds())
	})
}
