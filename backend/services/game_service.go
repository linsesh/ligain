package services

import (
	"context"
	"errors"
	"fmt"
	"ligain/backend/models"
	"ligain/backend/repositories"
	"time"

	log "github.com/sirupsen/logrus"
)

type GameService interface {
	GetIncomingMatches(player models.Player) map[string]*models.MatchResult
	GetMatchResults() map[string]*models.MatchResult
	UpdatePlayerBet(player models.Player, bet *models.Bet, now time.Time) error
	GetPlayerBets(player models.Player) ([]*models.Bet, error)
	GetPlayers() []models.Player
	GetGameID() string
	AddPlayer(player models.Player) error
	RemovePlayer(player models.Player) error
}

// GameService is used to really run a game
type GameServiceImpl struct {
	gameId         string
	gameRepo       repositories.GameRepository
	betRepo        repositories.BetRepository
	gamePlayerRepo repositories.GamePlayerRepository
	matchService   MatchService // For reading match data
}

func NewGameService(gameId string, gameRepo repositories.GameRepository, betRepo repositories.BetRepository, gamePlayerRepo repositories.GamePlayerRepository, matchService MatchService) *GameServiceImpl {
	return &GameServiceImpl{
		gameId:         gameId,
		gameRepo:       gameRepo,
		betRepo:        betRepo,
		gamePlayerRepo: gamePlayerRepo,
		matchService:   matchService,
	}
}

func (g *GameServiceImpl) getGame() (models.Game, error) {
	return g.gameRepo.GetGame(g.gameId)
}

func (g *GameServiceImpl) GetMatchResults() map[string]*models.MatchResult {
	game, err := g.getGame()
	if err != nil {
		log.Errorf("Error getting game: %v", err)
		return make(map[string]*models.MatchResult)
	}
	return game.GetPastResults()
}

func (g *GameServiceImpl) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	game, err := g.getGame()
	if err != nil {
		log.Errorf("Error getting game: %v", err)
		return make(map[string]*models.MatchResult)
	}
	return game.GetIncomingMatches(player)
}

// GetGameID returns the game ID
func (g *GameServiceImpl) GetGameID() string {
	return g.gameId
}

func (g *GameServiceImpl) handleScoreUpdate(match models.Match) error {
	game, err := g.getGame()
	if err != nil {
		log.Errorf("Error getting game: %v", err)
		return err
	}

	scores, err := game.CalculateMatchScores(match)
	if err != nil {
		log.Errorf("Error calculating match scores: %v", err)
		return err
	}
	// A player whose score cannot be saved must not prevent the others from being
	// scored, nor keep the finished match open: collect failures and report them at the end.
	var saveErrors []error
	for playerID, score := range scores {
		// Find the player by ID
		var player models.Player
		for _, p := range game.GetPlayers() {
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
			saveErrors = append(saveErrors, fmt.Errorf("player %s: %w", playerID, err))
			delete(scores, playerID)
			continue
		}
		log.Infof("Player %v has earned %v points for match %v", player, score, match.Id())
	}

	game.ApplyMatchScores(match, scores)
	if game.IsFinished() {
		log.Infof("Game %v is finished, with winner(s) %v", g.gameId, game.GetWinner())
	}

	// Save the updated game: it is also what stores the finished status of a game whose last match was just scored
	err = g.gameRepo.SaveWithId(g.gameId, game)
	if err != nil {
		log.Errorf("Error saving game after score update: %v", err)
		return errors.Join(append(saveErrors, err)...)
	}

	return errors.Join(saveErrors...)
}

func (g *GameServiceImpl) UpdatePlayerBet(player models.Player, bet *models.Bet, now time.Time) error {
	game, err := g.getGame()
	if err != nil {
		log.Errorf("Error getting game: %v", err)
		return err
	}

	// Fetch the match from MatchService to validate the bet
	match, err := g.matchService.GetMatch(bet.MatchID)
	if err != nil {
		log.Errorf("Error getting match %s: %v", bet.MatchID, err)
		return err
	}

	err = game.CheckPlayerBetValidity(player, bet, match, now)
	if err != nil {
		log.Errorf("Error checking player bet validity: %v", err)
		return err
	}
	_, savedBet, err := g.betRepo.SaveBet(g.gameId, bet, player)
	if err != nil {
		log.Errorf("Error saving bet: %v", err)
		return err
	}
	game.AddPlayerBet(player, savedBet)

	// Save the updated game
	err = g.gameRepo.SaveWithId(g.gameId, game)
	if err != nil {
		log.Errorf("Error saving game after bet update: %v", err)
		return err
	}

	return nil
}

func (g *GameServiceImpl) GetPlayerBets(player models.Player) ([]*models.Bet, error) {
	return g.betRepo.GetBets(g.gameId, player)
}

func (g *GameServiceImpl) GetPlayers() []models.Player {
	// Fetch players directly from the repository to ensure we get the latest data
	// This is important because player names can be updated, and we don't want to return stale cached data
	players, err := g.gamePlayerRepo.GetPlayersInGame(context.Background(), g.gameId)
	if err != nil {
		log.Errorf("Error getting players for game %s: %v", g.gameId, err)
		return []models.Player{}
	}
	return players
}

func (g *GameServiceImpl) AddPlayer(player models.Player) error {
	game, err := g.getGame()
	if err != nil {
		log.Errorf("Error getting game: %v", err)
		return err
	}

	err = game.AddPlayer(player)
	if err != nil {
		return err
	}

	// Save the updated game
	err = g.gameRepo.SaveWithId(g.gameId, game)
	if err != nil {
		log.Errorf("Error saving game after adding player: %v", err)
		return err
	}

	return nil
}

func (g *GameServiceImpl) RemovePlayer(player models.Player) error {
	game, err := g.getGame()
	if err != nil {
		log.Errorf("Error getting game: %v", err)
		return err
	}
	err = game.RemovePlayer(player)
	if err != nil {
		log.Errorf("Error removing player: %v", err)
		return err
	}

	// Save the updated game
	err = g.gameRepo.SaveWithId(g.gameId, game)
	if err != nil {
		log.Errorf("Error saving game after removing player: %v", err)
		return err
	}

	return nil
}

// OnMatchesUpdated implements MatchUpdateListener interface.
// It processes updates only for matches in this game's incoming list.
// - Ignores matches not in incoming list (no error, no scoring, no DB write)
// - Is idempotent for matches already in past (no rescoring)
// - Continues processing after errors, collecting all errors
func (g *GameServiceImpl) OnMatchesUpdated(matchIDs []string) error {
	if g.matchService == nil {
		log.Warnf("OnMatchesUpdated called but matchService is nil for game %s", g.gameId)
		return nil
	}

	var updateErrors []error

	for _, matchID := range matchIDs {
		// Get current game state to check if this match is incoming
		game, err := g.getGame()
		if err != nil {
			log.Errorf("Error getting game: %v", err)
			updateErrors = append(updateErrors, fmt.Errorf("match %s: get game: %w", matchID, err))
			continue
		}

		// Skip matches not in this game's incoming list (idempotent for already-past matches too)
		if !game.HasIncomingMatch(matchID) {
			log.Debugf("Match %s not in game %s incoming list, skipping", matchID, g.gameId)
			continue
		}

		// Get the match from MatchService
		match, err := g.matchService.GetMatch(matchID)
		if err != nil {
			log.Errorf("Error getting match %s from MatchService: %v", matchID, err)
			updateErrors = append(updateErrors, fmt.Errorf("match %s: get match: %w", matchID, err))
			continue
		}

		// Only score finished matches
		if !match.IsFinished() {
			log.Debugf("Match %s not finished yet, skipping", matchID)
			continue
		}

		log.Infof("Handling update for finished match %s", matchID)
		err = g.handleScoreUpdate(match)
		if err != nil {
			log.Errorf("Error handling score update for match %s: %v", matchID, err)
			updateErrors = append(updateErrors, fmt.Errorf("match %s: handle score update: %w", matchID, err))
			continue
		}
	}

	return errors.Join(updateErrors...)
}
