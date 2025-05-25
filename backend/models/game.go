package models

import "time"

type GameStatus string

const (
	GameStatusNotStarted GameStatus = "not started"
	GameStatusScheduled  GameStatus = "in progress"
	GameStatusFinished   GameStatus = "finished"
)

type Game interface {
	GetSeasonYear() string
	GetCompetitionName() string
	GetGameStatus() GameStatus
	CheckPlayerBetValidity(player Player, bet *Bet, datetime time.Time) error
	AddPlayerBet(player Player, bet *Bet) error
	CalculateMatchScores(match Match) (map[Player]int, error)
	ApplyMatchScores(match Match, scores map[Player]int)
	UpdateMatch(match Match) error
	// GetPastResults returns a map of match id to scored match, with the scores filled
	GetPastResults() map[string]*MatchResult
	// GetIncomingMatches returns a map of match id to scored match
	GetIncomingMatches() map[string]*MatchResult
	GetPlayersPoints() map[Player]int
	IsFinished() bool
	GetWinner() []Player
}
