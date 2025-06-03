package repositories

import (
	"fmt"
	"liguain/backend/models"
)

const betCacheSize = 5000 // Maximum number of bets to keep in cache

type BetRepository interface {
	// GetBets returns all bets for a given player
	GetBets(gameId string, player models.Player) ([]*models.Bet, error)
	// SaveBet saves or updates a bet and returns the bet id, and an error if saving failed
	SaveBet(gameId string, bet *models.Bet, player models.Player) (string, error)
	// GetBetsForMatch returns all bets and their associated players for a specific match
	GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error)
	// SaveWithId saves a bet with a specific ID
	SaveWithId(gameId string, betId string, bet *models.Bet, player models.Player) error
}

type BetEntry struct {
	GameId  string
	Player  models.Player
	MatchId string
	Bet     *models.Bet
	BetId   string
}

type InMemoryBetRepository struct {
	cache *Cache[string, BetEntry]
}

func NewInMemoryBetRepository() *InMemoryBetRepository {
	return &InMemoryBetRepository{
		cache: NewCache[string, BetEntry](betCacheSize),
	}
}

func (r *InMemoryBetRepository) GetBets(gameId string, player models.Player) ([]*models.Bet, error) {
	var bets []*models.Bet
	for _, entry := range r.cache.GetAll() {
		if entry.Value.GameId == gameId && entry.Value.Player == player {
			bets = append(bets, entry.Value.Bet)
		}
	}
	return bets, nil
}

func (r *InMemoryBetRepository) SaveBet(gameId string, bet *models.Bet, player models.Player) (string, error) {
	matchId := bet.Match.Id()
	betKey := fmt.Sprintf("%s:%s:%s", gameId, player.Name, matchId)
	entry := BetEntry{
		GameId:  gameId,
		Player:  player,
		MatchId: matchId,
		Bet:     bet,
		BetId:   betKey,
	}
	r.cache.Set(betKey, entry)
	return betKey, nil
}

func (r *InMemoryBetRepository) GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error) {
	matchId := match.Id()
	var bets []*models.Bet
	var players []models.Player

	for _, entry := range r.cache.GetAll() {
		if entry.Value.GameId == gameId && entry.Value.MatchId == matchId {
			bets = append(bets, entry.Value.Bet)
			players = append(players, entry.Value.Player)
		}
	}

	return bets, players, nil
}

func (r *InMemoryBetRepository) SaveWithId(gameId string, betId string, bet *models.Bet, player models.Player) error {
	matchId := bet.Match.Id()
	betKey := fmt.Sprintf("%s:%s:%s", gameId, player.Name, matchId)
	entry := BetEntry{
		GameId:  gameId,
		Player:  player,
		MatchId: matchId,
		Bet:     bet,
		BetId:   betId,
	}
	r.cache.Set(betKey, entry)
	return nil
}
