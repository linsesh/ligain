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
	game     rules.Game
	gameId   string
	gameRepo GameRepository
	betRepo  BetRepository
	watcher  MatchWatcherService
	// waitTime is the time we accept to wait for a check of game updates
	waitTime time.Duration
}

func NewGameService(game rules.Game, gameRepo GameRepository, betRepo BetRepository, watcher MatchWatcherService, waitTime time.Duration) (*GameService, error) {
	gameId, err := gameRepo.SaveGame(game)
	if err != nil {
		return nil, err
	}
	return &GameService{
		game:     game,
		gameId:   gameId,
		gameRepo: gameRepo,
		betRepo:  betRepo,
		watcher:  watcher,
		waitTime: waitTime,
	}, nil
}

// Play returns the winner(s) of the game when it ends
func (g *GameService) Play() ([]models.Player, error) {
	for !g.game.IsFinished() {
		log.Infof("Playing game %v", g.gameId)
		updates, err := g.getUpdates()
		if err != nil {
			log.Errorf("Error getting updates: %v", err)
			return nil, err
		}
		g.handleUpdates(updates)
	}
	winners := g.game.GetWinner()
	log.Infof("Game %v is finished, with winner(s) %v", g.gameId, winners)
	return winners, nil
}

func (g *GameService) handleUpdates(updates map[string]models.Match) {
	for _, match := range updates {
		if match.IsFinished() {
			g.handleScoreUpdate(match)
		} else {
			err := g.game.UpdateMatch(match)
			if err != nil {
				log.Errorf("Error updating match: %v", err)
			}
		}
	}
}

func (g *GameService) handleScoreUpdate(match models.Match) {
	scores, err := g.game.CalculateMatchScores(match)
	if err != nil {
		log.Errorf("Error calculating match scores: %v", err)
		return
	}

	// Save to repository first
	err = g.gameRepo.UpdateScores(match, scores)
	if err != nil {
		log.Errorf("Error saving scores to repository: %v", err)
		return
	}

	// Only after successful save, apply the scores
	err = g.game.ApplyMatchScores(scores)
	if err != nil {
		log.Errorf("Error applying scores: %v", err)
		return
	}
	for player, score := range scores {
		log.Infof("Player %v has earned %v points", player, score)
	}
}

func (g *GameService) getUpdates() (map[string]models.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.waitTime)
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

func (g *GameService) updateBet(bet *models.Bet, player models.Player) error {
	err := g.game.AddPlayerBet(&player, bet, time.Now())
	if err != nil {
		log.Errorf("Error adding player bet: %v", err)
		return err
	}
	return nil
}
