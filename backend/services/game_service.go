package services

import (
	"context"
	"liguain/backend/models"
	"liguain/backend/repositories"
	"time"

	log "github.com/sirupsen/logrus"
)

type GameService interface {
	GetIncomingMatches(player models.Player) map[string]*models.MatchResult
	GetMatchResults() map[string]*models.MatchResult
	Play() ([]models.Player, error)
	UpdatePlayerBet(player models.Player, bet *models.Bet, now time.Time) error
	GetPlayerBets(player models.Player) ([]*models.Bet, error)
	GetPlayers() []models.Player
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

func (g *GameServiceImpl) GetMatchResults() map[string]*models.MatchResult {
	return g.game.GetPastResults()
}

func (g *GameServiceImpl) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	return g.game.GetIncomingMatches(player)
}

// Play returns the winner(s) of the game when it ends
func (g *GameServiceImpl) Play() ([]models.Player, error) {
	log.Infof("Playing game %v", g.gameId)
	for !g.game.IsFinished() {
		updates, err := g.getUpdates()
		if err != nil {
			log.Errorf("Error getting updates: %v", err)
		} else {
			g.HandleUpdates(updates)
		}
	}
	winners := g.game.GetWinner()
	log.Infof("Game %v is finished, with winner(s) %v", g.gameId, winners)
	return winners, nil
}

func (g *GameServiceImpl) HandleUpdates(updates map[string]models.Match) error {
	for _, match := range updates {
		log.Infof("Handling update for match %v", match.Id())
		err := g.game.UpdateMatch(match)
		if err != nil {
			log.Errorf("Error updating match: %v", err)
			return err
		}
		if match.IsFinished() {
			log.Infof("Match %v is finished, handling score update", match.Id())
			err = g.handleScoreUpdate(match)
			if err != nil {
				log.Errorf("Error handling score update: %v", err)
				return err
			}
		} else {
			log.Infof("Match %v is being updated", match.Id())
		}
	}
	return nil
}

func (g *GameServiceImpl) handleScoreUpdate(match models.Match) error {
	scores, err := g.game.CalculateMatchScores(match)
	if err != nil {
		log.Errorf("Error calculating match scores: %v", err)
		return err
	}
	for playerID, score := range scores {
		// Find the player by ID
		var player models.Player
		for _, p := range g.game.GetPlayers() {
			if p.GetID() == playerID {
				player = p
				break
			}
		}
		if player == nil {
			log.Errorf("Player with ID %v not found", playerID)
			continue
		}

		err = g.betRepo.SaveScore(g.gameId, match, player, score)
		if err != nil {
			log.Errorf("Error saving score for player %v: %v", player, err)
			return err
		}
		log.Infof("Player %v has earned %v points for match %v", player, score, match.Id())
	}

	g.game.ApplyMatchScores(match, scores)
	return nil
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

func (g *GameServiceImpl) GetPlayers() []models.Player {
	return g.game.GetPlayers()
}
