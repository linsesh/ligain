package repositories

import (
	"liguain/backend/models"
)

type ScoreRepository interface {
	SaveScore(gameId string, betId string, points int) error
	GetScore(gameId string, betId string) (int, error)
	GetScores(gameId string) (map[string]int, error)
	// GetScoresByMatchAndPlayer returns scores organized by match ID and player
	GetScoresByMatchAndPlayer(gameId string) (map[string]map[models.Player]int, error)
}

// ScoreEntry holds all the information for a single score
type ScoreEntry struct {
	GameId  string
	MatchId string
	BetId   string
	Player  models.Player
	Points  int
}

type InMemoryScoreRepository struct {
	cache *Cache[string, ScoreEntry]
}

func NewInMemoryScoreRepository() ScoreRepository {
	return &InMemoryScoreRepository{
		cache: NewCache[string, ScoreEntry](1000), // Adjust capacity as needed
	}
}

func (r *InMemoryScoreRepository) SaveScore(gameId string, betId string, points int) error {
	key := gameId + ":" + betId
	entry := ScoreEntry{
		GameId: gameId,
		BetId:  betId,
		Points: points,
	}
	r.cache.Set(key, entry)
	return nil
}

func (r *InMemoryScoreRepository) GetScore(gameId string, betId string) (int, error) {
	key := gameId + ":" + betId
	entry, err := r.cache.Get(key)
	if err == nil {
		return entry.Points, nil
	}
	return 0, nil
}

func (r *InMemoryScoreRepository) GetScores(gameId string) (map[string]int, error) {
	result := make(map[string]int)
	for _, entry := range r.cache.GetAll() {
		if entry.Value.GameId == gameId {
			result[entry.Value.BetId] = entry.Value.Points
		}
	}
	return result, nil
}

func (r *InMemoryScoreRepository) GetScoresByMatchAndPlayer(gameId string) (map[string]map[models.Player]int, error) {
	playerScores := make(map[string]map[models.Player]int)
	for _, entry := range r.cache.GetAll() {
		if entry.Value.GameId == gameId {
			if _, ok := playerScores[entry.Value.MatchId]; !ok {
				playerScores[entry.Value.MatchId] = make(map[models.Player]int)
			}
			playerScores[entry.Value.MatchId][entry.Value.Player] = entry.Value.Points
		}
	}
	return playerScores, nil
}
