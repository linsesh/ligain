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
	log.Infof("Playing game %v", g.gameId)
	for !g.game.IsFinished() {
		updates, err := g.getUpdates()
		if err != nil {
			log.Errorf("Error getting updates: %v", err)
		} else {
			g.handleUpdates(updates)
		}
	}
	winners := g.game.GetWinner()
	log.Infof("Game %v is finished, with winner(s) %v", g.gameId, winners)
	return winners, nil
}

func (g *GameService) handleUpdates(updates map[string]models.Match) {
	for _, match := range updates {
		log.Infof("Handling update for match %v", match.Id())
		if match.IsFinished() {
			log.Infof("Match %v is finished, handling score update", match.Id())
			g.handleScoreUpdate(match)
		} else {
			log.Infof("Match %v is being updated", match.Id())
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
	err = g.gameRepo.UpdateScores(match, scores)
	if err != nil {
		log.Errorf("Error saving scores to repository: %v", err)
		return
	}
	// Only after successful save, apply the scores
	g.game.ApplyMatchScores(match, scores)
	for player, score := range scores {
		log.Infof("Player %v has earned %v points for match %v", player, score, match.Id())
	}
}

func (g *GameService) getUpdates() (map[string]models.Match, error) {
	ctx, cancel := context.WithTimeout(context.Background(), g.waitTime)
	defer cancel()

	done := make(chan MatchWatcherServiceResult)
	go g.watcher.GetUpdates(ctx, done)
	select {
	case updates := <-done:
		if updates.Err != nil {
			return nil, updates.Err
		}
		return updates.Value, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (g *GameService) updateBet(bet *models.Bet, player models.Player, now time.Time) error {
	err := g.game.AddPlayerBet(&player, bet, now)
	if err != nil {
		log.Errorf("Error adding player bet: %v", err)
		return err
	}
	return nil
}
