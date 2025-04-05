package services

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/rules"
	"time"

	log "github.com/sirupsen/logrus"
)

// GameService is used to really run a game
type GameService struct {
	game    rules.Game
	gameId  string
	repo    GameRepository
	watcher MatchWatcherService
}

func NewGameService(game rules.Game, repo GameRepository) (*GameService, error) {
	gameId, err := repo.SaveGame(game)
	if err != nil {
		return nil, err
	}
	return &GameService{
		game:   game,
		gameId: gameId,
		repo:   repo,
	}, nil
}

// Play returns the winner(s) of the game when it ends
func (g *GameService) Play() ([]models.Player, error) {
	for !g.game.IsFinished() {
		log.Info("Playing game %v", g.gameId)
		updates, err := g.getUpdates()
		if err != nil {
			log.Error("Error getting updates: %v", err)
			return nil, err
		}
		g.handleUpdates(updates)
		g.pause()
	}
	winners := g.game.GetWinner()
	log.Info("Game %v is finished, with winner(s) %v", g.gameId, winners)
	return winners, nil
}

func (g *GameService) handleUpdates(updates map[string]models.Match) {
	for _, match := range updates {
		if match.IsFinished() {
			scores, _, err := g.game.AddFinishedMatch(match)
			if err != nil {
				log.Error("Error adding finished match: %v", err)
				return
			}
			g.handleScoreUpdate(scores)
		} else {
			err := g.game.UpdateMatch(match)
			if err != nil {
				log.Error("Error updating match: %v", err)
				return
			}
		}
	}
}

func (g *GameService) handleScoreUpdate(scores map[models.Player]int) {
	err := g.repo.updateScores(scores)
	if err != nil {
		log.Error("Error updating scores: %v", err)
		return
	}
	for player, score := range scores {
		log.Info("Player %v has earned %v points", player, score)
	}
}

func (g *GameService) getUpdates() (map[string]models.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan MatchWatcherServiceResult)
	go g.watcher.GetUpdates(ctx, done)
	updates := <-done
	select {
	case <-done:
		if updates.Err != nil {
			return nil, updates.Err
		}
		return updates.Value, nil
	case <-ctx.Done():
		log.Error("Timed out and cancelled.")
		return nil, ctx.Err()
	}
}

// pause pauses the updates of the game status for 1 second.
// todo: make it configurable depending if we're running in dev mode or not
func (g *GameService) pause() {
	time.Sleep(1 * time.Second)
}
