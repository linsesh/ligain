package postgres

import (
	"database/sql"
	"liguain/backend/models"
	"time"
)

// CreateMatchFromDB creates a match from database fields
func CreateMatchFromDB(homeTeamId, awayTeamId, seasonCode, competitionCode string, matchDate time.Time, matchday int, matchStatus string, homeTeamScore, awayTeamScore sql.NullInt32, homeWinOdds, awayWinOdds, drawOdds float64) *models.SeasonMatch {
	match := models.NewSeasonMatch(
		homeTeamId,
		awayTeamId,
		seasonCode,
		competitionCode,
		matchDate,
		matchday,
	)

	// Set odds
	match.SetHomeTeamOdds(homeWinOdds)
	match.SetAwayTeamOdds(awayWinOdds)
	match.SetDrawOdds(drawOdds)

	// Set match status based on database value
	switch models.MatchStatus(matchStatus) {
	case models.MatchStatusFinished:
		if homeTeamScore.Valid && awayTeamScore.Valid {
			match.Finish(int(homeTeamScore.Int32), int(awayTeamScore.Int32))
		}
	case models.MatchStatusStarted:
		match.Start()
	}

	return match
}
