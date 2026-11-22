package rules

import (
	"fmt"
	"ligain/backend/models"
	"math"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// MatchReader reads matches by ID. Implemented by MatchService.
type MatchReader interface {
	GetMatch(id string) (models.Match, error)
	GetMatches(ids []string) (map[string]models.Match, error)
}

// Game represents a competition between several players, on a specific season and competition
type GameImpl struct {
	seasonCode      string
	competitionCode string
	name            string
	players         []models.Player
	// PlayersPoints is a map of match id to a map of player to points
	playersPoints map[string]map[string]int
	// incomingMatchIDs is a list of match IDs in incoming status (not yet finished)
	incomingMatchIDs []string
	// Bets is a map of match id to a map of player to bet
	bets map[string]map[string]*models.Bet
	// pastMatchIDs is a list of match IDs in past status (finished)
	pastMatchIDs []string
	gameStatus   models.GameStatus
	scorer       Scorer
	scores       map[string]map[string]int
	// matchReader reads matches by ID (MatchService in production or tests)
	matchReader MatchReader
	// mu protects the state of the game: it is read by HTTP requests while match events update it
	mu sync.RWMutex
}

func NewFreshGame(seasonCode, competitionCode, name string, players []models.Player, incomingMatches []models.Match, scorer Scorer, matchReader MatchReader) *GameImpl {
	g := &GameImpl{
		seasonCode:       seasonCode,
		competitionCode:  competitionCode,
		name:             name,
		players:          players,
		gameStatus:       models.GameStatusScheduled,
		playersPoints:    make(map[string]map[string]int),
		incomingMatchIDs: make([]string, 0),
		pastMatchIDs:     make([]string, 0),
		bets:             make(map[string]map[string]*models.Bet),
		scorer:           scorer,
		scores:           make(map[string]map[string]int),
		matchReader:      matchReader,
	}
	for _, match := range incomingMatches {
		g.incomingMatchIDs = append(g.incomingMatchIDs, match.Id())
	}
	return g
}

func NewStartedGame(seasonCode, competitionCode, name string, players []models.Player, incomingMatches []models.Match, pastMatches []models.Match, scorer Scorer, bets map[string]map[string]*models.Bet, scores map[string]map[string]int, matchReader MatchReader) models.Game {
	g := NewFreshGame(seasonCode, competitionCode, name, players, incomingMatches, scorer, matchReader)
	for _, match := range pastMatches {
		g.pastMatchIDs = append(g.pastMatchIDs, match.Id())
	}
	g.bets = bets
	g.scores = scores
	g.playersPoints = scores
	g.gameStatus = models.GameStatusScheduled
	return g
}

// HasIncomingMatch checks if a match ID is in the incoming list (implements models.Game interface)
func (g *GameImpl) HasIncomingMatch(matchID string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.hasIncomingMatchID(matchID)
}

// hasIncomingMatchID checks if a match ID is in the incoming list
// Callers must hold g.mu.
func (g *GameImpl) hasIncomingMatchID(matchID string) bool {
	for _, id := range g.incomingMatchIDs {
		if id == matchID {
			return true
		}
	}
	return false
}

// removeFromIncoming removes a match ID from the incoming list
// Callers must hold g.mu.
func (g *GameImpl) removeFromIncoming(matchID string) {
	for i, id := range g.incomingMatchIDs {
		if id == matchID {
			g.incomingMatchIDs = append(g.incomingMatchIDs[:i], g.incomingMatchIDs[i+1:]...)
			return
		}
	}
}

// addToPast adds a match ID to the past list
// Callers must hold g.mu.
func (g *GameImpl) addToPast(matchID string) {
	g.pastMatchIDs = append(g.pastMatchIDs, matchID)
}

// getIncomingMatches returns all incoming match objects by reading from matchReader
// Callers must hold g.mu.
func (g *GameImpl) getIncomingMatches() (map[string]models.Match, error) {
	return g.matchReader.GetMatches(g.incomingMatchIDs)
}

// getPastMatches returns all past match objects by reading from matchReader
// Callers must hold g.mu.
func (g *GameImpl) getPastMatches() (map[string]models.Match, error) {
	return g.matchReader.GetMatches(g.pastMatchIDs)
}

// containsPlayerByID returns true if a player with the same ID exists in the slice
func containsPlayerByID(players []models.Player, player models.Player) bool {
	for _, p := range players {
		if p.GetID() == player.GetID() {
			return true
		}
	}
	return false
}

func (g *GameImpl) CheckPlayerBetValidity(player models.Player, bet *models.Bet, datetime time.Time) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.hasIncomingMatchID(bet.Match.Id()) {
		return fmt.Errorf("match %v not found", bet.Match.Id())
	}
	if !containsPlayerByID(g.players, player) {
		return fmt.Errorf("player %v not found", player)
	}
	if datetime.After(bet.Match.GetDate()) {
		return fmt.Errorf("too late to bet on match %v", bet.Match.Id())
	}
	return nil
}

func (g *GameImpl) CalculateMatchScores(match models.Match) (map[string]int, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.hasIncomingMatchID(match.Id()) {
		return nil, fmt.Errorf("match not found")
	}
	if !match.IsFinished() {
		return nil, fmt.Errorf("match is not finished")
	}
	scores := g.scoreMatch(match)
	return scores, nil
}

func (g *GameImpl) ApplyMatchScores(match models.Match, scores map[string]int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.updatePlayersPoints(match, scores)
	g.finishMatch(match)
}

// We assume that the match is correct since it's a private method
// Callers must hold g.mu.
func (g *GameImpl) finishMatch(match models.Match) {
	g.removeFromIncoming(match.Id())
	g.addToPast(match.Id())
	if len(g.incomingMatchIDs) == 0 {
		g.gameStatus = models.GameStatusFinished
	}
}

func (g *GameImpl) AddPlayerBet(player models.Player, bet *models.Bet) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.hasIncomingMatchID(bet.Match.Id()) {
		return fmt.Errorf("match not found")
	}
	if _, exists := g.bets[bet.Match.Id()]; !exists {
		g.bets[bet.Match.Id()] = make(map[string]*models.Bet)
	}
	g.bets[bet.Match.Id()][player.GetID()] = bet
	return nil
}

func (g *GameImpl) AddPlayer(player models.Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	// Check if the player is already in the game
	if containsPlayerByID(g.players, player) {
		return fmt.Errorf("player %v is already in the game", player)
	}

	// Add the player to the game
	g.players = append(g.players, player)

	return nil
}

// Callers must hold g.mu.
func (g *GameImpl) hasPlayerWithID(playerID string) bool {
	for _, p := range g.players {
		if p.GetID() == playerID {
			return true
		}
	}
	return false
}

func (g *GameImpl) RemovePlayer(player models.Player) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, p := range g.players {
		if p.GetID() == player.GetID() {
			g.players = append(g.players[:i], g.players[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("player not found")
}

// Callers must hold g.mu.
func (g *GameImpl) scoreMatch(match models.Match) map[string]int {
	bets := g.bets[match.Id()]

	playerIDs := make([]string, 0, len(bets))
	betList := make([]*models.Bet, 0, len(bets))
	for _, player := range g.players {
		playerIDs = append(playerIDs, player.GetID())
		bet, hasBet := bets[player.GetID()]
		if hasBet {
			betList = append(betList, bet)
		} else {
			betList = append(betList, nil)
		}
	}
	scores := g.scorer.Score(match, betList)
	scoresMap := make(map[string]int)
	for i, score := range scores {
		scoresMap[playerIDs[i]] = score
	}
	return scoresMap
}

// Callers must hold g.mu.
func (g *GameImpl) updatePlayersPoints(match models.Match, scores map[string]int) {
	for playerID, score := range scores {
		if _, exists := g.playersPoints[match.Id()]; !exists {
			g.playersPoints[match.Id()] = make(map[string]int)
		}
		g.playersPoints[match.Id()][playerID] = score
	}
}

func (g *GameImpl) GetSeasonYear() string {
	return g.seasonCode
}

func (g *GameImpl) GetCompetitionName() string {
	return g.competitionCode
}

func (g *GameImpl) GetGameStatus() models.GameStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.gameStatus
}

func (g *GameImpl) GetName() string {
	return g.name
}

func (g *GameImpl) GetPastResults() map[string]*models.MatchResult {
	g.mu.RLock()
	defer g.mu.RUnlock()
	results := make(map[string]*models.MatchResult)
	pastMatches, err := g.getPastMatches()
	if err != nil {
		log.Errorf("Error reading the past matches of game %q: %v", g.name, err)
		return results
	}
	for _, match := range pastMatches {
		playerBets := make(map[models.Player]*models.Bet)
		for playerID, bet := range g.bets[match.Id()] {
			for _, player := range g.players {
				if player.GetID() == playerID {
					playerBets[player] = bet
					break
				}
			}
		}

		playerScores := make(map[models.Player]int)
		for playerID, score := range g.playersPoints[match.Id()] {
			for _, player := range g.players {
				if player.GetID() == playerID {
					playerScores[player] = score
					break
				}
			}
		}

		// Compute breakdowns from scorer
		betList := make([]*models.Bet, 0, len(g.players))
		for _, player := range g.players {
			bet, hasBet := g.bets[match.Id()][player.GetID()]
			if hasBet {
				betList = append(betList, bet)
			} else {
				betList = append(betList, nil)
			}
		}
		scorerBreakdowns := g.scorer.ScoreWithBreakdown(match, betList)
		playerBreakdowns := make(map[models.Player]models.ScoreBreakdown)
		for i, player := range g.players {
			playerBreakdowns[player] = models.ScoreBreakdown{
				BaseScore:             scorerBreakdowns[i].BaseScore,
				RiskMultiplier:        scorerBreakdowns[i].RiskMultiplier,
				ClairvoyantMultiplier: scorerBreakdowns[i].ClairvoyantMultiplier,
			}
		}

		results[match.Id()] = models.NewScoredMatchWithBreakdowns(match, playerBets, playerScores, playerBreakdowns)
	}
	return results
}

func (g *GameImpl) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	g.mu.RLock()
	defer g.mu.RUnlock()
	matches := make(map[string]*models.MatchResult)
	incomingMatches, err := g.getIncomingMatches()
	if err != nil {
		log.Errorf("Error reading the incoming matches of game %q: %v", g.name, err)
		return matches
	}
	for _, match := range incomingMatches {
		playerBets := make(map[string]*models.Bet)

		// Pre-populate all non-requesting players with hasBet: false
		playerBetStatus := make(map[string]bool)
		for _, p := range g.players {
			if p.GetID() != player.GetID() {
				playerBetStatus[p.GetID()] = false
			}
		}

		if bets, exists := g.bets[match.Id()]; exists {
			for playerID, bet := range bets {
				if !g.hasPlayerWithID(playerID) {
					continue
				}
				if playerID == player.GetID() || match.IsInProgress() {
					playerBets[playerID] = bet
				} else {
					playerBetStatus[playerID] = true
				}
			}
		}

		matchResult := models.NewMatchWithBetsWithIDs(match, playerBets)
		matchResult.PlayerBetStatus = playerBetStatus
		matches[match.Id()] = matchResult
	}
	return matches
}

// GetIncomingMatchesForTesting returns all incoming matches with all bets (for testing only)
func (g *GameImpl) GetIncomingMatchesForTesting() map[string]*models.MatchResult {
	g.mu.RLock()
	defer g.mu.RUnlock()
	matches := make(map[string]*models.MatchResult)
	incomingMatches, err := g.getIncomingMatches()
	if err != nil {
		log.Errorf("Error reading the incoming matches of game %q: %v", g.name, err)
		return matches
	}
	for _, match := range incomingMatches {
		playerBets := make(map[string]*models.Bet)
		if bets, exists := g.bets[match.Id()]; exists {
			// Show all bets for testing purposes
			for playerID, bet := range bets {
				playerBets[playerID] = bet
			}
		}
		matches[match.Id()] = models.NewMatchWithBetsWithIDs(match, playerBets)
	}
	return matches
}

func (g *GameImpl) GetPlayersPoints() map[string]int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.totalPlayersPoints()
}

// Callers must hold g.mu.
func (g *GameImpl) totalPlayersPoints() map[string]int {
	points := make(map[string]int)
	for _, matchPoints := range g.playersPoints {
		for playerID, score := range matchPoints {
			points[playerID] += score
		}
	}
	return points
}

func (g *GameImpl) IsFinished() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.gameStatus == models.GameStatusFinished
}

func (g *GameImpl) Finish() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.gameStatus = models.GameStatusFinished
}

func (g *GameImpl) GetPlayers() []models.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return append(make([]models.Player, 0, len(g.players)), g.players...)
}

func (g *GameImpl) GetWinner() []models.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	bestScore := math.MinInt32
	winners := make([]models.Player, 0)
	totalPlayerPoints := g.totalPlayersPoints()

	// Create a map for faster player lookup
	playerMap := make(map[string]models.Player)
	for _, player := range g.players {
		playerMap[player.GetID()] = player
	}

	for playerID, points := range totalPlayerPoints {
		if points > bestScore {
			bestScore = points
			winners = []models.Player{playerMap[playerID]}
		} else if points == bestScore {
			winners = append(winners, playerMap[playerID])
		}
	}
	return winners
}
