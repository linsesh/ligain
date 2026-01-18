package services

import (
	"context"
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
	// GameUpdateHandler interface methods
	HandleMatchUpdates(updates map[string]models.Match) error
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
	timeFunc       func() time.Time // Function to get current time (for testing)
}

func NewGameService(gameId string, gameRepo repositories.GameRepository, betRepo repositories.BetRepository, gamePlayerRepo repositories.GamePlayerRepository) *GameServiceImpl {
	return &GameServiceImpl{
		gameId:         gameId,
		gameRepo:       gameRepo,
		betRepo:        betRepo,
		gamePlayerRepo: gamePlayerRepo,
		timeFunc:       time.Now, // Default to real time
	}
}

// NewGameServiceWithTime creates a GameService with a custom time function (for testing)
func NewGameServiceWithTime(gameId string, gameRepo repositories.GameRepository, betRepo repositories.BetRepository, gamePlayerRepo repositories.GamePlayerRepository, timeFunc func() time.Time) *GameServiceImpl {
	return &GameServiceImpl{
		gameId:         gameId,
		gameRepo:       gameRepo,
		betRepo:        betRepo,
		gamePlayerRepo: gamePlayerRepo,
		timeFunc:       timeFunc,
	}
}

// getGame always fetches the current game state from the repository
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

// HandleMatchUpdates implements GameUpdateHandler interface
func (g *GameServiceImpl) HandleMatchUpdates(updates map[string]models.Match) error {
	// Get current game state
	game, err := g.getGame()
	if err != nil {
		log.Errorf("Error getting game: %v", err)
		return err
	}

	for _, match := range updates {
		log.Infof("Handling update for match %v", match.Id())
		lastMatchState, err := game.GetMatchById(match.Id())
		if err != nil {
			log.Errorf("Error getting last match state: %v", err)
			return err
		}
		match = g.adjustOdds(match, lastMatchState)
		err = game.UpdateMatch(match)
		if err != nil {
			log.Errorf("Error updating match: %v", err)
			return err
		}
		if match.IsFinished() {
			log.Infof("Match %v is finished with score %d - %d, handling score update", match.Id(), match.GetHomeGoals(), match.GetAwayGoals())
			err = g.handleScoreUpdate(match)
			if err != nil {
				log.Errorf("Error handling score update: %v", err)
				return err
			}
		}
	}

	// Check if game is finished after processing updates
	if game.IsFinished() {
		winners := game.GetWinner()
		log.Infof("Game %v is finished, with winner(s) %v", g.gameId, winners)

		err := g.gameRepo.SaveWithId(g.gameId, game)
		if err != nil {
			log.Errorf("Error saving finished game status: %v", err)
			return err
		}
	}

	return nil
}

// GetGameID implements GameUpdateHandler interface
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
			return err
		}
		log.Infof("Player %v has earned %v points for match %v", player, score, match.Id())
	}

	game.ApplyMatchScores(match, scores)

	// Save the updated game
	err = g.gameRepo.SaveWithId(g.gameId, game)
	if err != nil {
		log.Errorf("Error saving game after score update: %v", err)
		return err
	}

	return nil
}

func (g *GameServiceImpl) UpdatePlayerBet(player models.Player, bet *models.Bet, now time.Time) error {
	game, err := g.getGame()
	if err != nil {
		log.Errorf("Error getting game: %v", err)
		return err
	}

	err = game.CheckPlayerBetValidity(player, bet, now)
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

// adjustOdds ensures that odds are blocked 5minutes before the match starts.
// We use 6 to ensure that the odds are always blocked at least 5 minutes before the match starts.
func (g *GameServiceImpl) adjustOdds(match models.Match, lastMatchState models.Match) models.Match {
	matchDate := match.GetDate()
	now := g.timeFunc()
	if matchDate.Before(now.Add(6 * time.Minute)) {
		match.SetHomeTeamOdds(lastMatchState.GetHomeTeamOdds())
		match.SetAwayTeamOdds(lastMatchState.GetAwayTeamOdds())
		match.SetDrawOdds(lastMatchState.GetDrawOdds())
	}
	return match
}
