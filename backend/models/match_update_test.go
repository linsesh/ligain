package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var providerUpdateKickoff = time.Date(2026, 11, 21, 19, 45, 0, 0, time.UTC)

func scheduledWithOdds(home, away, draw float64) *SeasonMatch {
	return NewSeasonMatchWithKnownOdds("Lens", "Toulouse", "2026/2027", "Ligue 1", providerUpdateKickoff, 11, home, away, draw)
}

func TestSeasonMatch_Clone_IsIndependent(t *testing.T) {
	original := scheduledWithOdds(2.1, 3.2, 3.3)

	clone := original.Clone()
	clone.SetHomeTeamOdds(9)
	clone.Start()

	assert.Equal(t, 2.1, original.GetHomeTeamOdds())
	assert.False(t, original.IsInProgress())
}

func TestSeasonMatch_TryUpdate(t *testing.T) {
	tenMinutesBefore := providerUpdateKickoff.Add(-10 * time.Minute)
	fiveMinutesBefore := providerUpdateKickoff.Add(-5 * time.Minute)

	started := func() *SeasonMatch { m := scheduledWithOdds(4.35, 9, 1.35); m.Start(); return m }
	finished := NewFinishedSeasonMatch("Lens", "Toulouse", 1, 1, "2026/2027", "Ligue 1", providerUpdateKickoff, 11, 21, 56, 1.02)

	tests := []struct {
		name    string
		now     time.Time
		update  Match
		odds    [3]float64 // expected home, away, draw
		status  MatchStatus
		goals   [2]int
		changed bool
	}{
		{"odds follow the provider before the freeze", tenMinutesBefore, scheduledWithOdds(2.2, 3.4, 3.1), [3]float64{2.2, 3.4, 3.1}, MatchStatusScheduled, [2]int{0, 0}, true},
		{"odds reported 5 minutes before kickoff are refused, so nothing changes", fiveMinutesBefore, scheduledWithOdds(2.2, 3.4, 3.1), [3]float64{2.1, 3.2, 3.3}, MatchStatusScheduled, [2]int{0, 0}, false},
		{"odds are frozen when the match starts, status applied", providerUpdateKickoff.Add(time.Minute), started(), [3]float64{2.1, 3.2, 3.3}, MatchStatusStarted, [2]int{0, 0}, true},
		{"odds stay frozen when the match finishes, score applied", providerUpdateKickoff.Add(2 * time.Hour), finished, [3]float64{2.1, 3.2, 3.3}, MatchStatusFinished, [2]int{1, 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := scheduledWithOdds(2.1, 3.2, 3.3)

			result, changed := current.TryUpdate(MatchUpdateOf(tt.update), tt.now)

			assert.Equal(t, tt.odds[0], result.GetHomeTeamOdds())
			assert.Equal(t, tt.odds[1], result.GetAwayTeamOdds())
			assert.Equal(t, tt.odds[2], result.GetDrawOdds())
			assert.Equal(t, tt.status, result.GetStatus())
			assert.Equal(t, tt.goals[0], result.GetHomeGoals())
			assert.Equal(t, tt.goals[1], result.GetAwayGoals())
			assert.Equal(t, tt.changed, changed)
			assert.Equal(t, 2.1, current.GetHomeTeamOdds(), "the current match must not be modified")
		})
	}
}

func TestSeasonMatch_TryUpdate_FirstOddsAreAcceptedEvenInsideTheWindow(t *testing.T) {
	current := NewSeasonMatch("Lens", "Toulouse", "2026/2027", "Ligue 1", providerUpdateKickoff, 11) // no odds yet

	result, _ := current.TryUpdate(MatchUpdateOf(scheduledWithOdds(2.2, 3.4, 3.1)), providerUpdateKickoff.Add(-time.Minute))

	assert.Equal(t, 2.2, result.GetHomeTeamOdds())
}

func TestSeasonMatch_TryUpdate_DateChangeIsAlwaysApplied(t *testing.T) {
	current := scheduledWithOdds(2.1, 3.2, 3.3)
	postponed := scheduledWithOdds(2.2, 3.4, 3.1)
	postponed.Date = providerUpdateKickoff.Add(7 * 24 * time.Hour)

	result, _ := current.TryUpdate(MatchUpdateOf(postponed), providerUpdateKickoff.Add(-time.Minute))

	assert.Equal(t, postponed.Date, result.GetDate())
	assert.Equal(t, 2.2, result.GetHomeTeamOdds(), "the match is no longer about to start, odds follow again")
}
