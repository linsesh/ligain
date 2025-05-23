package repositories

import "liguain/backend/models"

type ScoresRepository interface {
	UpdateScores(gameId string, match models.Match, scores map[models.Player]int) error
}

type InMemoryScoresRepository struct {
	scores map[string]map[string]map[models.Player]int
}

func NewInMemoryScoresRepository() ScoresRepository {
	return &InMemoryScoresRepository{
		scores: make(map[string]map[string]map[models.Player]int),
	}
}

func (r *InMemoryScoresRepository) UpdateScores(gameId string, match models.Match, scores map[models.Player]int) error {
	if r.scores[gameId] == nil {
		r.scores[gameId] = make(map[string]map[models.Player]int)
	}
	r.scores[gameId][match.Id()] = scores
	return nil
}
