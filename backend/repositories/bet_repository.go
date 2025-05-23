package repositories

import (
	"fmt"
	"liguain/backend/models"
)

type BetRepository interface {
	// GetBets returns all bets for a given player
	GetBets(gameId string, player models.Player) ([]*models.Bet, error)
	// SaveBet saves or updates a bet and returns the bet id, and an error if saving failed
	SaveBet(gameId string, bet *models.Bet, player models.Player) (string, error)
	// GetBetsForMatch returns all bets and their associated players for a specific match
	GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error)
}

// InMemoryBetRepository is a simple in-memory implementation of BetRepository
type InMemoryBetRepository struct {
	// bets maps gameId -> player -> matchId -> bet
	bets map[string]map[models.Player]map[string]*models.Bet
}

// NewInMemoryBetRepository creates a new instance of InMemoryBetRepository
func NewInMemoryBetRepository() *InMemoryBetRepository {
	return &InMemoryBetRepository{
		bets: make(map[string]map[models.Player]map[string]*models.Bet),
	}
}

// GetBets returns all bets for a given player in a game
func (r *InMemoryBetRepository) GetBets(gameId string, player models.Player) ([]*models.Bet, error) {
	if gameBets, ok := r.bets[gameId]; ok {
		if playerBets, ok := gameBets[player]; ok {
			bets := make([]*models.Bet, 0, len(playerBets))
			for _, bet := range playerBets {
				bets = append(bets, bet)
			}
			return bets, nil
		}
	}
	return []*models.Bet{}, nil
}

// SaveBet saves or updates a bet and returns the bet id
func (r *InMemoryBetRepository) SaveBet(gameId string, bet *models.Bet, player models.Player) (string, error) {
	matchId := bet.Match.Id()

	// Initialize maps if they don't exist
	if _, ok := r.bets[gameId]; !ok {
		r.bets[gameId] = make(map[models.Player]map[string]*models.Bet)
	}
	if _, ok := r.bets[gameId][player]; !ok {
		r.bets[gameId][player] = make(map[string]*models.Bet)
	}

	// Save or update the bet
	r.bets[gameId][player][matchId] = bet

	// Return a unique identifier for the bet
	betId := fmt.Sprintf("%s-%s-%s", gameId, player.Name, matchId)
	return betId, nil
}

// GetBetsForMatch returns all bets for a specific match
func (r *InMemoryBetRepository) GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error) {
	matchId := match.Id()
	var bets []*models.Bet
	var players []models.Player

	if gameBets, ok := r.bets[gameId]; ok {
		for player, playerBets := range gameBets {
			if bet, ok := playerBets[matchId]; ok {
				bets = append(bets, bet)
				players = append(players, player)
			}
		}
	}

	return bets, players, nil
}
