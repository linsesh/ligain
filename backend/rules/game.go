package rules

import (
	"fmt"
	"liguain/backend/models"
	"liguain/backend/utils"
	"slices"
	"time"
)

// Game represents a competition between several players, on a specific season and competition
type GameImpl struct {
	SeasonCode      string
	CompetitionCode string
	Players         []models.Player
	PlayersPoints   map[models.Player]int
	Matches         map[string]models.Match
	GameStatus      models.GameStatus
	Scorer          Scorer
}

func NewGame(seasonCode, competitionCode string, players []models.Player, matches []models.Match, scorer Scorer) models.Game {
	g := &GameImpl{
		SeasonCode:      seasonCode,
		CompetitionCode: competitionCode,
		Players:         players,
		GameStatus:      models.GameStatusNotStarted,
		PlayersPoints:   make(map[models.Player]int),
		Matches:         make(map[string]models.Match),
		Scorer:          scorer,
	}
	for _, match := range matches {
		g.Matches[match.Id()] = match
	}
	return g
}

func (g *GameImpl) CheckPlayerBetValidity(player models.Player, bet *models.Bet, datetime time.Time) error {
	_, exists := g.Matches[bet.Match.Id()]
	if !exists {
		return fmt.Errorf("match %v not found", bet.Match.Id())
	}
	if !slices.Contains(g.Players, player) {
		return fmt.Errorf("player %v not found", player)
	}
	if datetime.After(bet.Match.GetDate()) {
		return fmt.Errorf("too late to bet on match %v", bet.Match.Id())
	}
	return nil
}

func (g *GameImpl) CalculateMatchScores(match models.Match, bets map[models.Player]*models.Bet) (map[models.Player]int, error) {
	existingMatch, exists := g.Matches[match.Id()]
	if !exists {
		return nil, fmt.Errorf("match not found")
	}
	if !match.IsFinished() {
		return nil, fmt.Errorf("match is not finished")
	}
	if existingMatch.IsFinished() {
		return nil, fmt.Errorf("match already finished")
	}
	existingMatch.Finish(match.GetHomeGoals(), match.GetAwayGoals())
	scores := g.scoreMatch(existingMatch, bets)
	return scores, nil
}

func (g *GameImpl) ApplyMatchScores(match models.Match, scores map[models.Player]int) {
	g.updatePlayersPoints(scores)
	g.removeIncomingMatch(match)
}

func (g *GameImpl) UpdateMatch(match models.Match) error {
	_, exists := g.Matches[match.Id()]
	if !exists {
		return fmt.Errorf("match not found")
	}
	g.Matches[match.Id()] = match
	return nil
}

func (g *GameImpl) scoreMatch(match models.Match, bets map[models.Player]*models.Bet) map[models.Player]int {
	players := make([]models.Player, 0, len(bets))
	betList := make([]*models.Bet, 0, len(bets))
	for player, bet := range bets {
		players = append(players, player)
		betList = append(betList, bet)
	}
	scores := g.Scorer.Score(match, betList)
	scoresMap := make(map[models.Player]int)
	for i, score := range scores {
		scoresMap[players[i]] = score
	}
	return scoresMap
}

func (g *GameImpl) updatePlayersPoints(scores map[models.Player]int) {
	for player, score := range scores {
		g.PlayersPoints[player] += score
	}
}

func (g *GameImpl) removeIncomingMatch(match models.Match) {
	delete(g.Matches, match.Id())
	if len(g.Matches) == 0 {
		g.GameStatus = models.GameStatusFinished
	}
}

func (g *GameImpl) GetSeasonYear() string {
	return g.SeasonCode
}

func (g *GameImpl) GetCompetitionName() string {
	return g.CompetitionCode
}

func (g *GameImpl) GetGameStatus() models.GameStatus {
	return g.GameStatus
}

func (g *GameImpl) GetIncomingMatches() []models.Match {
	matches := make([]models.Match, 0)
	for _, match := range utils.MapValues(g.Matches) {
		if !match.IsFinished() {
			matches = append(matches, match)
		}
	}
	return matches
}

func (g *GameImpl) GetPlayersPoints() map[models.Player]int {
	return g.PlayersPoints
}

func (g *GameImpl) IsFinished() bool {
	return g.GameStatus == models.GameStatusFinished
}

func (g *GameImpl) GetWinner() []models.Player {
	bestScore := 0
	winners := make([]models.Player, 0)
	for player, points := range g.PlayersPoints {
		if points > bestScore {
			bestScore = points
			winners = []models.Player{player}
		} else if points == bestScore {
			winners = append(winners, player)
		}
	}
	return winners
}
