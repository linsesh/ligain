package services

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"time"

	log "github.com/sirupsen/logrus"
)

type GameService interface {
	Play() ([]models.Player, error)
	UpdatePlayerBet(player models.Player, bet *models.Bet, now time.Time) error
	GetPlayerBets(player models.Player) ([]*models.Bet, error)
}

// GameService is used to really run a game
type GameServiceImpl struct {
	game     models.Game
	gameId   string
	gameRepo repositories.GameRepository
	betRepo  repositories.BetRepository
	watcher  MatchWatcherService
	// waitTime is the time we accept to wait for a check of game updates
	waitTime time.Duration
}

func NewGameService(gameId string, game models.Game, gameRepo repositories.GameRepository, betRepo repositories.BetRepository, watcher MatchWatcherService, waitTime time.Duration) *GameServiceImpl {
	return &GameServiceImpl{
		game:     game,
		gameId:   gameId,
		gameRepo: gameRepo,
		betRepo:  betRepo,
		watcher:  watcher,
		waitTime: waitTime,
	}
}

// Play returns the winner(s) of the game when it ends
func (g *GameServiceImpl) Play() ([]models.Player, error) {
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

func (g *GameServiceImpl) handleUpdates(updates map[string]models.Match) {
	for _, match := range updates {
		log.Infof("Handling update for match %v", match.Id())
		err := g.game.UpdateMatch(match)
		if err != nil {
			log.Errorf("Error updating match: %v", err)
		}
		if match.IsFinished() {
			log.Infof("Match %v is finished, handling score update", match.Id())
			g.handleScoreUpdate(match)
		} else {
			log.Infof("Match %v is being updated", match.Id())
		}
	}
}

func (g *GameServiceImpl) handleScoreUpdate(match models.Match) {
	scores, err := g.game.CalculateMatchScores(match)
	if err != nil {
		log.Errorf("Error calculating match scores: %v", err)
		return
	}
	for player, score := range scores {
		err = g.betRepo.SaveScore(g.gameId, match, player, score)
		if err != nil {
			log.Errorf("Error saving score for player %v: %v", player, err)
			continue
		}
		log.Infof("Player %v has earned %v points for match %v", player, score, match.Id())
	}

	g.game.ApplyMatchScores(match, scores)
}

func (g *GameServiceImpl) getUpdates() (map[string]models.Match, error) {
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
func (g *GameServiceImpl) UpdatePlayerBet(player models.Player, bet *models.Bet, now time.Time) error {
	err := g.game.CheckPlayerBetValidity(player, bet, now)
	if err != nil {
		log.Errorf("Error checking player bet validity: %v", err)
		return err
	}
	_, savedBet, err := g.betRepo.SaveBet(g.gameId, bet, player)
	if err != nil {
		log.Errorf("Error saving bet: %v", err)
		return err
	}
	g.game.AddPlayerBet(player, savedBet)
	return nil
}

func (g *GameServiceImpl) GetPlayerBets(player models.Player) ([]*models.Bet, error) {
	return g.betRepo.GetBets(g.gameId, player)
}
