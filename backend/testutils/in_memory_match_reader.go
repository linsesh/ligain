package testutils

import (
	"fmt"
	"ligain/backend/models"
)

// InMemoryMatchReader is a match reader over a fixed set of matches, for tests that build a game
// without a MatchService. Every read returns a copy.
type InMemoryMatchReader struct {
	matches map[string]models.Match
}

func NewInMemoryMatchReader(matches ...models.Match) *InMemoryMatchReader {
	reader := &InMemoryMatchReader{matches: make(map[string]models.Match)}
	for _, m := range matches {
		reader.matches[m.Id()] = m.Clone()
	}
	return reader
}

// NewInMemoryMatchReaderOf is NewInMemoryMatchReader over several lists of matches (for example incoming and past)
func NewInMemoryMatchReaderOf(lists ...[]models.Match) *InMemoryMatchReader {
	var all []models.Match
	for _, list := range lists {
		all = append(all, list...)
	}
	return NewInMemoryMatchReader(all...)
}

func (r *InMemoryMatchReader) GetMatch(id string) (models.Match, error) {
	match, exists := r.matches[id]
	if !exists {
		return nil, fmt.Errorf("match not found: %s", id)
	}
	return match.Clone(), nil
}

func (r *InMemoryMatchReader) GetMatches(ids []string) (map[string]models.Match, error) {
	matches := make(map[string]models.Match, len(ids))
	for _, id := range ids {
		if match, exists := r.matches[id]; exists {
			matches[id] = match.Clone()
		}
	}
	return matches, nil
}
