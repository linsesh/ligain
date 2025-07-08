package rules

import (
	"fmt"
	"liguain/backend/models"
	"slices"
	"time"
)

// Game represents a competition between several players, on a specific season and competition
type GameImpl struct {
	seasonCode      string
	competitionCode string
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

func NewFreshGame(seasonCode, competitionCode string, players []models.Player, incomingMatches []models.Match, scorer Scorer) *GameImpl {
	g := &GameImpl{
		seasonCode:      seasonCode,
		competitionCode: competitionCode,
		players:         players,
		gameStatus:      models.GameStatusNotStarted,
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

func NewStartedGame(seasonCode, competitionCode string, players []models.Player, incomingMatches []models.Match, pastMatches []models.Match, scorer Scorer, bets map[string]map[string]*models.Bet, scores map[string]map[string]int) models.Game {
	g := NewFreshGame(seasonCode, competitionCode, players, incomingMatches, scorer)
	for _, match := range pastMatches {
		g.pastMatches[match.Id()] = match
	}
	g.bets = bets
	g.scores = scores
	g.playersPoints = scores
	g.gameStatus = models.GameStatusScheduled
	return g
}

func (g *GameImpl) CheckPlayerBetValidity(player models.Player, bet *models.Bet, datetime time.Time) error {
	_, exists := g.incomingMatches[bet.Match.Id()]
	if !exists {
		return fmt.Errorf("match %v not found", bet.Match.Id())
	}
	if !slices.Contains(g.players, player) {
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
		g.gameStatus = models.GameStatusFinished
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

func (g *GameImpl) scoreMatch(match models.Match) map[string]int {
	bets := g.bets[match.Id()]
	if bets == nil {
		return make(map[string]int)
	}

	playerIDs := make([]string, 0, len(bets))
	betList := make([]*models.Bet, 0, len(bets))
	for playerID, bet := range bets {
		playerIDs = append(playerIDs, playerID)
		betList = append(betList, bet)
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
		playerBets := make(map[models.Player]*models.Bet)
		if bets, exists := g.bets[match.Id()]; exists {
			for playerID, bet := range bets {
				// Find the player by ID
				for _, p := range g.players {
					if p.GetID() == playerID && p.GetName() == player.GetName() {
						playerBets[p] = bet
						break
					}
				}
			}
		}
		matches[match.Id()] = models.NewMatchWithBets(match, playerBets)
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

func (g *GameImpl) GetPlayers() []models.Player {
	return g.players
}

func (g *GameImpl) GetWinner() []models.Player {
	bestScore := 0
	winners := make([]models.Player, 0)
	totalPlayerPoints := g.GetPlayersPoints()
	for playerID, points := range totalPlayerPoints {
		if points > bestScore {
			bestScore = points
			// Find the player by ID
			for _, player := range g.players {
				if player.GetID() == playerID {
					winners = []models.Player{player}
					break
				}
			}
		} else if points == bestScore {
			// Find the player by ID
			for _, player := range g.players {
				if player.GetID() == playerID {
					winners = append(winners, player)
					break
				}
			}
		}
	}
	return winners
}
