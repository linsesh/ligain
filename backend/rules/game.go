package rules

import (
	"fmt"
	"liguain/backend/models"
	"liguain/backend/utils"
)

// GameStatus represents the current status of a game
type GameStatus string

const (
	GameStatusNotStarted GameStatus = "not started"
	GameStatusScheduled  GameStatus = "in progress"
	GameStatusFinished   GameStatus = "finished"
)

type Game interface {
	GetSeasonYear() string
	GetCompetitionName() string
	GetGameStatus() GameStatus
	AddPlayerBet(player *models.Player, bet *models.Bet)
	AddFinishedMatch(match *models.Match) (map[models.Player]int, error)
	GetMatchBets(match *models.Match) map[models.Player]*models.Bet
	GetIncomingMatches() []*models.Match
	GetPlayersPoints() map[models.Player]int
}

// Game represents a competition between several players, on a specific season and competition
type GameImpl struct {
	SeasonCode         string
	CompetitionCode    string
	Players            []models.Player
	PlayerBetsPerMatch map[string]map[models.Player]*models.Bet
	PlayersPoints      map[models.Player]int
	Matches            map[string]*models.Match
	GameStatus         GameStatus
	Scorer             Scorer
}

func NewGame(seasonCode, competitionCode string, players []models.Player, matches []*models.Match, scorer Scorer) *GameImpl {
	g := &GameImpl{
		SeasonCode:         seasonCode,
		CompetitionCode:    competitionCode,
		Players:            players,
		GameStatus:         GameStatusNotStarted,
		PlayerBetsPerMatch: make(map[string]map[models.Player]*models.Bet),
		PlayersPoints:      make(map[models.Player]int),
		Matches:            make(map[string]*models.Match),
		Scorer:             scorer,
	}
	for _, match := range matches {
		g.Matches[match.Id()] = match
	}
	return g
}

func (g *GameImpl) AddPlayerBet(player *models.Player, bet *models.Bet) {
	g.PlayerBetsPerMatch[bet.Match.Id()][*player] = bet
}

func (g *GameImpl) GetIncomingMatches() []*models.Match {
	matches := make([]*models.Match, 0)
	for _, match := range utils.MapValues(g.Matches) {
		if !match.IsFinished() {
			matches = append(matches, match)
		}
	}
	return matches
}

func (g *GameImpl) AddFinishedMatch(match *models.Match) (map[models.Player]int, error) {
	_, exists := g.Matches[match.Id()]
	if !exists {
		return g.PlayersPoints, fmt.Errorf("match not found")
	}
	if !match.IsFinished() {
		return g.PlayersPoints, fmt.Errorf("match is not finished")
	}
	g.Matches[match.Id()] = match
	scores := g.scoreMatch(match)
	g.updatePlayersPoints(scores)
	return g.PlayersPoints, nil
}

func (g *GameImpl) scoreMatch(match *models.Match) map[models.Player]int {
	matchBets := g.PlayerBetsPerMatch[match.Id()]
	players, bets := utils.MapKeysValues(matchBets)
	scores := g.Scorer.Score(match, bets)
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

func (g *GameImpl) GetSeasonCode() string {
	return g.SeasonCode
}

func (g *GameImpl) GetCompetitionCode() string {
	return g.CompetitionCode
}

func (g *GameImpl) GetGameStatus() GameStatus {
	return g.GameStatus
}

func (g *GameImpl) GetMatchBets(match *models.Match) map[models.Player]*models.Bet {
	return g.PlayerBetsPerMatch[match.Id()]
}

func (g *GameImpl) GetPlayersPoints() map[models.Player]int {
	return g.PlayersPoints
}
