package services

import (
	"liguain/backend/models"
)

type BetRepository interface {
	// GetBets returns all bets for a given player
	GetBets(gameId string, player models.Player) ([]models.Bet, error)
	// SaveBet saves or updates a bet and returns the bet id, and an error if saving failed
	SaveBet(bet models.Bet) (string, error)
	// GetBetsForMatch returns all bets for a given match
	GetBetsForMatch(match models.Match) ([]models.Bet, error)
}
