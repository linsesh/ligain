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
	playersPoints map[string]map[models.Player]int
	// IncomingMatches is a map of match id to match
	incomingMatches map[string]models.Match
	// Bets is a map of match id to a map of player to bet
	bets map[string]map[models.Player]*models.Bet
	// PastMatches is a map of match id to match
	pastMatches map[string]models.Match
	gameStatus  models.GameStatus
	scorer      Scorer
	scores      map[string]map[models.Player]int
}

func NewFreshGame(seasonCode, competitionCode string, players []models.Player, incomingMatches []models.Match, scorer Scorer) *GameImpl {
	g := &GameImpl{
		seasonCode:      seasonCode,
		competitionCode: competitionCode,
		players:         players,
		gameStatus:      models.GameStatusNotStarted,
		playersPoints:   make(map[string]map[models.Player]int),
		incomingMatches: make(map[string]models.Match),
		pastMatches:     make(map[string]models.Match),
		bets:            make(map[string]map[models.Player]*models.Bet),
		scorer:          scorer,
		scores:          make(map[string]map[models.Player]int),
	}
	for _, match := range incomingMatches {
		g.incomingMatches[match.Id()] = match
	}
	return g
}

func NewStartedGame(seasonCode, competitionCode string, players []models.Player, incomingMatches []models.Match, pastMatches []models.Match, scorer Scorer, bets map[string]map[models.Player]*models.Bet, scores map[string]map[models.Player]int) models.Game {
	g := NewFreshGame(seasonCode, competitionCode, players, incomingMatches, scorer)
	for _, match := range pastMatches {
		g.pastMatches[match.Id()] = match
	}
	g.bets = bets
	g.scores = scores
	// Initialize playersPoints from scores
	for matchId, matchScores := range scores {
		g.playersPoints[matchId] = matchScores
	}
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

func (g *GameImpl) CalculateMatchScores(match models.Match) (map[models.Player]int, error) {
	existingMatch, exists := g.incomingMatches[match.Id()]
	if !exists {
		return nil, fmt.Errorf("match not found")
	}
	if !match.IsFinished() {
		return nil, fmt.Errorf("match is not finished")
	}
	scores := g.scoreMatch(existingMatch)
	return scores, nil
}

func (g *GameImpl) ApplyMatchScores(match models.Match, scores map[models.Player]int) {
	g.updatePlayersPoints(match, scores)
	g.finishMatch(match)
}

// We assume that the match is correct since it's a private method
func (g *GameImpl) finishMatch(match models.Match) {
	existingMatch := g.incomingMatches[match.Id()]
	existingMatch.Finish(match.GetHomeGoals(), match.GetAwayGoals())
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
		g.bets[bet.Match.Id()] = make(map[models.Player]*models.Bet)
	}
	g.bets[bet.Match.Id()][player] = bet
	return nil
}

func (g *GameImpl) scoreMatch(match models.Match) map[models.Player]int {
	bets := g.bets[match.Id()]
	players := make([]models.Player, 0, len(g.players))
	betList := make([]*models.Bet, 0, len(bets))
	for player, bet := range bets {
		players = append(players, player)
		betList = append(betList, bet)
	}
	scores := g.scorer.Score(match, betList)
	scoresMap := make(map[models.Player]int)
	for i, score := range scores {
		scoresMap[players[i]] = score
	}
	return scoresMap
}

func (g *GameImpl) updatePlayersPoints(match models.Match, scores map[models.Player]int) {
	for player, score := range scores {
		if _, exists := g.playersPoints[match.Id()]; !exists {
			g.playersPoints[match.Id()] = make(map[models.Player]int)
		}
		g.playersPoints[match.Id()][player] = score
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
		results[match.Id()] = models.NewScoredMatch(match, g.bets[match.Id()], g.playersPoints[match.Id()])
	}
	return results
}

func (g *GameImpl) GetIncomingMatches() map[string]*models.MatchResult {
	matches := make(map[string]*models.MatchResult)
	for _, match := range g.incomingMatches {
		matches[match.Id()] = models.NewMatchWithBets(match, g.bets[match.Id()])
	}
	return matches
}

func (g *GameImpl) GetPlayersPoints() map[models.Player]int {
	points := make(map[models.Player]int)
	for _, matchPoints := range g.playersPoints {
		for player, score := range matchPoints {
			points[player] += score
		}
	}
	return points
}

func (g *GameImpl) IsFinished() bool {
	return g.gameStatus == models.GameStatusFinished
}

func (g *GameImpl) GetWinner() []models.Player {
	bestScore := 0
	winners := make([]models.Player, 0)
	totalPlayerPoints := g.GetPlayersPoints()
	for player, points := range totalPlayerPoints {
		if points > bestScore {
			bestScore = points
			winners = []models.Player{player}
		} else if points == bestScore {
			winners = append(winners, player)
		}
	}
	return winners
}
