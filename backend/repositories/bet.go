package repositories

import (
	"errors"
	"fmt"
	"liguain/backend/models"
)

const betCacheSize = 5000 // Maximum number of bets to keep in cache

var ErrScoreNotFound = errors.New("score not found")

type BetRepository interface {
	// GetBets returns all bets for a given player
	GetBets(gameId string, player models.Player) ([]*models.Bet, error)
	// SaveBet saves or updates a bet and returns the bet id and the bet
	SaveBet(gameId string, bet *models.Bet, player models.Player) (string, *models.Bet, error)
	// GetBetsForMatch returns all bets and their associated players for a specific match
	GetBetsForMatch(match models.Match, gameId string) ([]*models.Bet, []models.Player, error)
	// SaveWithId saves a bet with a specific ID
	SaveWithId(gameId string, betId string, bet *models.Bet, player models.Player) error
	// SaveScore saves a score for a bet
	SaveScore(gameId string, match models.Match, player models.Player, points int) error
	// GetScore returns the score for a given bet. Returns ErrScoreNotFound if no score exists.
	GetScore(gameId string, betId string) (int, error)
	// GetScores returns all scores for a game
	GetScores(gameId string) (map[string]int, error)
	// GetScoresByMatchAndPlayer returns scores organized by match ID and player
	GetScoresByMatchAndPlayer(gameId string) (map[string]map[string]int, error)
}

type BetEntry struct {
	GameId  string
	Player  models.Player
	MatchId string
	Bet     *models.Bet
	BetId   string
	Points  *int // Optional score for this bet
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

func (r *InMemoryBetRepository) SaveBet(gameId string, bet *models.Bet, player models.Player) (string, *models.Bet, error) {
	matchId := bet.Match.Id()
	betKey := fmt.Sprintf("%s:%s:%s", gameId, player.GetName(), matchId)
	entry := BetEntry{
		GameId:  gameId,
		Player:  player,
		MatchId: matchId,
		Bet:     bet,
		BetId:   betKey,
	}
	r.cache.Set(betKey, entry)
	return betKey, bet, nil
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
	betKey := fmt.Sprintf("%s:%s:%s", gameId, player.GetName(), matchId)
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

func (r *InMemoryBetRepository) SaveScore(gameId string, match models.Match, player models.Player, points int) error {
	betKey := fmt.Sprintf("%s:%s:%s", gameId, player.GetName(), match.Id())
	entry, err := r.cache.Get(betKey)
	if err != nil {
		return fmt.Errorf("no bet found for match %s and player %s", match.Id(), player.GetName())
	}
	entry.Points = &points
	r.cache.Set(betKey, entry)
	return nil
}

func (r *InMemoryBetRepository) GetScore(gameId string, betId string) (int, error) {
	entry, err := r.cache.Get(betId)
	if err != nil {
		return 0, ErrScoreNotFound
	}
	if entry.Points == nil {
		return 0, ErrScoreNotFound
	}
	return *entry.Points, nil
}

func (r *InMemoryBetRepository) GetScores(gameId string) (map[string]int, error) {
	result := make(map[string]int)
	for _, entry := range r.cache.GetAll() {
		if entry.Value.GameId == gameId && entry.Value.Points != nil {
			result[entry.Value.BetId] = *entry.Value.Points
		}
	}
	return result, nil
}

func (r *InMemoryBetRepository) GetScoresByMatchAndPlayer(gameId string) (map[string]map[string]int, error) {
	playerScores := make(map[string]map[string]int)
	for _, entry := range r.cache.GetAll() {
		if entry.Value.GameId == gameId && entry.Value.Points != nil {
			if _, ok := playerScores[entry.Value.MatchId]; !ok {
				playerScores[entry.Value.MatchId] = make(map[string]int)
			}
			playerScores[entry.Value.MatchId][entry.Value.Player.GetID()] = *entry.Value.Points
		}
	}
	return playerScores, nil
}
