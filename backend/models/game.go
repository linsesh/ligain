package models

import (
	"time"
)

type GameStatus string

const (
	GameStatusNotStarted GameStatus = "not started"
	GameStatusScheduled  GameStatus = "in progress"
	GameStatusFinished   GameStatus = "finished"
)

// Game represents a game with players, matches, and bets
type Game interface {
	GetIncomingMatches(player Player) map[string]*MatchResult
	GetPastResults() map[string]*MatchResult
	GetSeasonYear() string
	GetCompetitionName() string
	GetGameStatus() GameStatus
	GetName() string
	CheckPlayerBetValidity(player Player, bet *Bet, datetime time.Time) error
	AddPlayerBet(player Player, bet *Bet) error
	AddPlayer(player Player) error
	RemovePlayer(player Player) error
	CalculateMatchScores(match Match) (map[string]int, error)
	ApplyMatchScores(match Match, scores map[string]int)
	UpdateMatch(match Match) error
	GetPlayersPoints() map[string]int
	GetPlayers() []Player
	IsFinished() bool
	GetWinner() []Player
	// GetIncomingMatchesForTesting returns all incoming matches with all bets (for testing only)
	GetIncomingMatchesForTesting() map[string]*MatchResult
	GetMatchById(matchId string) (Match, error)
	Finish()
}
