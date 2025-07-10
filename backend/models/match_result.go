package models

type MatchResult struct {
	Match  Match
	Bets   map[string]*Bet
	Scores map[string]int
}

func NewUnscoredMatch(match Match) *MatchResult {
	return &MatchResult{
		Match:  match,
		Bets:   nil,
		Scores: nil,
	}
}

func NewMatchWithBets(match Match, bets map[Player]*Bet) *MatchResult {
	// Convert map[Player]*Bet to map[string]*Bet
	playerBets := make(map[string]*Bet)
	for player, bet := range bets {
		playerBets[player.GetID()] = bet
	}

	return &MatchResult{
		Match:  match,
		Bets:   playerBets,
		Scores: nil,
	}
}

func NewMatchWithBetsWithIDs(match Match, bets map[string]*Bet) *MatchResult {
	return &MatchResult{
		Match:  match,
		Bets:   bets,
		Scores: nil,
	}
}

func NewScoredMatch(match Match, bets map[Player]*Bet, scores map[Player]int) *MatchResult {
	// Convert map[Player]*Bet to map[string]*Bet
	playerBets := make(map[string]*Bet)
	for player, bet := range bets {
		playerBets[player.GetID()] = bet
	}

	// Convert map[Player]int to map[string]int
	playerScores := make(map[string]int)
	for player, score := range scores {
		playerScores[player.GetID()] = score
	}

	return &MatchResult{
		Match:  match,
		Bets:   playerBets,
		Scores: playerScores,
	}
}

func (m *MatchResult) IsScored() bool {
	return m.Scores != nil
}
