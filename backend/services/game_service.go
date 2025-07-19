package services

import (
	"liguain/backend/models"
	"liguain/backend/repositories"
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
}

// GameService is used to really run a game
type GameServiceImpl struct {
	game     models.Game
	gameId   string
	gameRepo repositories.GameRepository
	betRepo  repositories.BetRepository
	timeFunc func() time.Time // Function to get current time (for testing)
}

func NewGameService(gameId string, game models.Game, gameRepo repositories.GameRepository, betRepo repositories.BetRepository) *GameServiceImpl {
	return &GameServiceImpl{
		game:     game,
		gameId:   gameId,
		gameRepo: gameRepo,
		betRepo:  betRepo,
		timeFunc: time.Now, // Default to real time
	}
}

// NewGameServiceWithTime creates a GameService with a custom time function (for testing)
func NewGameServiceWithTime(gameId string, game models.Game, gameRepo repositories.GameRepository, betRepo repositories.BetRepository, timeFunc func() time.Time) *GameServiceImpl {
	return &GameServiceImpl{
		game:     game,
		gameId:   gameId,
		gameRepo: gameRepo,
		betRepo:  betRepo,
		timeFunc: timeFunc,
	}
}

func (g *GameServiceImpl) GetMatchResults() map[string]*models.MatchResult {
	return g.game.GetPastResults()
}

func (g *GameServiceImpl) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	return g.game.GetIncomingMatches(player)
}

// HandleMatchUpdates implements GameUpdateHandler interface
func (g *GameServiceImpl) HandleMatchUpdates(updates map[string]models.Match) error {
	log.Infof("Game %v received %d match updates", g.gameId, len(updates))

	for _, match := range updates {
		log.Infof("Handling update for match %v", match.Id())
		lastMatchState, err := g.game.GetMatchById(match.Id())
		if err != nil {
			log.Errorf("Error getting last match state: %v", err)
			return err
		}
		match = g.adjustOdds(match, lastMatchState)
		err = g.game.UpdateMatch(match)
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

	// Check if game is finished after processing updates
	if g.game.IsFinished() {
		winners := g.game.GetWinner()
		log.Infof("Game %v is finished, with winner(s) %v", g.gameId, winners)
	}

	return nil
}

// GetGameID implements GameUpdateHandler interface
func (g *GameServiceImpl) GetGameID() string {
	return g.gameId
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

func (g *GameServiceImpl) AddPlayer(player models.Player) error {
	return g.game.AddPlayer(player)
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
