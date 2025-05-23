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
	CalculateMatchScores(match Match, bets map[Player]*Bet) (map[Player]int, error)
	ApplyMatchScores(match Match, scores map[Player]int)
	UpdateMatch(match Match) error
	GetIncomingMatches() []Match
	GetPlayersPoints() map[Player]int
	IsFinished() bool
	GetWinner() []Player
}
