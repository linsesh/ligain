package repositories

import (
	"fmt"
	"liguain/backend/models"
)

const scoresCacheSize = 1000 // Maximum number of score entries to keep in cache

type ScoresRepository interface {
	UpdateScores(gameId string, match models.Match, scores map[models.Player]int) error
}

type ScoresEntry struct {
	GameId  string
	MatchId string
	Scores  map[models.Player]int
}

type InMemoryScoresRepository struct {
	cache *Cache[string, ScoresEntry]
}

func NewInMemoryScoresRepository() ScoresRepository {
	return &InMemoryScoresRepository{
		cache: NewCache[string, ScoresEntry](scoresCacheSize),
	}
}

func (r *InMemoryScoresRepository) UpdateScores(gameId string, match models.Match, scores map[models.Player]int) error {
	key := fmt.Sprintf("%s:%s", gameId, match.Id())
	entry := ScoresEntry{
		GameId:  gameId,
		MatchId: match.Id(),
		Scores:  scores,
	}
	r.cache.Set(key, entry)
	return nil
}
