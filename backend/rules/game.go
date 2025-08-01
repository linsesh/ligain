package rules

import (
	"fmt"
	"ligain/backend/models"
	"math"
	"time"
)

// Game represents a competition between several players, on a specific season and competition
type GameImpl struct {
	seasonCode      string
	competitionCode string
	name            string
	players         []models.Player
	// PlayersPoints is a map of match id to a map of player to points
	playersPoints map[string]map[string]int
	// IncomingMatches is a map of match id to match
	incomingMatches map[string]models.Match
	// Bets is a map of match id to a map of player to bet
	bets map[string]map[string]*models.Bet
	// PastMatches is a map of match id to match
	pastMatches map[string]models.Match
	gameStatus  models.GameStatus
	scorer      Scorer
	scores      map[string]map[string]int
}

func NewFreshGame(seasonCode, competitionCode, name string, players []models.Player, incomingMatches []models.Match, scorer Scorer) *GameImpl {
	g := &GameImpl{
		seasonCode:      seasonCode,
		competitionCode: competitionCode,
		name:            name,
		players:         players,
		gameStatus:      models.GameStatusScheduled,
		playersPoints:   make(map[string]map[string]int),
		incomingMatches: make(map[string]models.Match),
		pastMatches:     make(map[string]models.Match),
		bets:            make(map[string]map[string]*models.Bet),
		scorer:          scorer,
		scores:          make(map[string]map[string]int),
	}
	for _, match := range incomingMatches {
		g.incomingMatches[match.Id()] = match
	}
	return g
}

func NewStartedGame(seasonCode, competitionCode, name string, players []models.Player, incomingMatches []models.Match, pastMatches []models.Match, scorer Scorer, bets map[string]map[string]*models.Bet, scores map[string]map[string]int) models.Game {
	g := NewFreshGame(seasonCode, competitionCode, name, players, incomingMatches, scorer)
	for _, match := range pastMatches {
		g.pastMatches[match.Id()] = match
	}
	g.bets = bets
	g.scores = scores
	g.playersPoints = scores
	g.gameStatus = models.GameStatusScheduled
	return g
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
	_, exists := g.incomingMatches[bet.Match.Id()]
	if !exists {
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
	_, exists := g.incomingMatches[match.Id()]
	if !exists {
		return nil, fmt.Errorf("match not found")
	}
	if !match.IsFinished() {
		return nil, fmt.Errorf("match is not finished")
	}
	scores := g.scoreMatch(match)
	return scores, nil
}

func (g *GameImpl) ApplyMatchScores(match models.Match, scores map[string]int) {
	g.updatePlayersPoints(match, scores)
	g.finishMatch(match)
}

// We assume that the match is correct since it's a private method
func (g *GameImpl) finishMatch(match models.Match) {
	delete(g.incomingMatches, match.Id())
	g.pastMatches[match.Id()] = match
	if len(g.incomingMatches) == 0 {
		g.Finish()
	}
}

func (g *GameImpl) UpdateMatch(match models.Match) error {
	_, exists := g.incomingMatches[match.Id()]
	if !exists {
		return fmt.Errorf("match not found")
	}
	g.incomingMatches[match.Id()] = match
	return nil
}

func (g *GameImpl) AddPlayerBet(player models.Player, bet *models.Bet) error {
	_, exists := g.incomingMatches[bet.Match.Id()]
	if !exists {
		return fmt.Errorf("match not found")
	}
	if _, exists := g.bets[bet.Match.Id()]; !exists {
		g.bets[bet.Match.Id()] = make(map[string]*models.Bet)
	}
	g.bets[bet.Match.Id()][player.GetID()] = bet
	return nil
}

func (g *GameImpl) AddPlayer(player models.Player) error {
	// Check if the player is already in the game
	if containsPlayerByID(g.players, player) {
		return fmt.Errorf("player %v is already in the game", player)
	}

	// Add the player to the game
	g.players = append(g.players, player)

	return nil
}

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
	return g.gameStatus
}

func (g *GameImpl) GetName() string {
	return g.name
}

func (g *GameImpl) GetPastResults() map[string]*models.MatchResult {
	results := make(map[string]*models.MatchResult)
	for _, match := range g.pastMatches {
		// Convert map[string]*models.Bet to map[models.Player]*models.Bet for NewScoredMatch
		playerBets := make(map[models.Player]*models.Bet)
		for playerID, bet := range g.bets[match.Id()] {
			// Find the player by ID
			for _, player := range g.players {
				if player.GetID() == playerID {
					playerBets[player] = bet
					break
				}
			}
		}

		// Convert map[string]int to map[models.Player]int for NewScoredMatch
		playerScores := make(map[models.Player]int)
		for playerID, score := range g.playersPoints[match.Id()] {
			// Find the player by ID
			for _, player := range g.players {
				if player.GetID() == playerID {
					playerScores[player] = score
					break
				}
			}
		}

		results[match.Id()] = models.NewScoredMatch(match, playerBets, playerScores)
	}
	return results
}

func (g *GameImpl) GetIncomingMatches(player models.Player) map[string]*models.MatchResult {
	matches := make(map[string]*models.MatchResult)
	for _, match := range g.incomingMatches {
		playerBets := make(map[string]*models.Bet)
		if bets, exists := g.bets[match.Id()]; exists {
			// For future matches, only show current player's bets
			// For in-progress matches, show all players' bets
			if match.IsInProgress() {
				// Show all bets for in-progress matches
				for playerID, bet := range bets {
					playerBets[playerID] = bet
				}
			} else {
				// Only show current player's bet for future matches
				if bet, exists := bets[player.GetID()]; exists {
					playerBets[player.GetID()] = bet
				}
			}
		}
		matches[match.Id()] = models.NewMatchWithBetsWithIDs(match, playerBets)
	}
	return matches
}

// GetIncomingMatchesForTesting returns all incoming matches with all bets (for testing only)
func (g *GameImpl) GetIncomingMatchesForTesting() map[string]*models.MatchResult {
	matches := make(map[string]*models.MatchResult)
	for _, match := range g.incomingMatches {
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
	points := make(map[string]int)
	for _, matchPoints := range g.playersPoints {
		for playerID, score := range matchPoints {
			points[playerID] += score
		}
	}
	return points
}

func (g *GameImpl) IsFinished() bool {
	return g.gameStatus == models.GameStatusFinished
}

func (g *GameImpl) Finish() {
	g.gameStatus = models.GameStatusFinished
}

func (g *GameImpl) GetPlayers() []models.Player {
	return g.players
}

func (g *GameImpl) GetWinner() []models.Player {
	bestScore := math.MinInt32
	winners := make([]models.Player, 0)
	totalPlayerPoints := g.GetPlayersPoints()

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

func (g *GameImpl) GetMatchById(matchId string) (models.Match, error) {
	match, exists := g.incomingMatches[matchId]
	if !exists {
		return nil, fmt.Errorf("match not found")
	}
	return match, nil
}
