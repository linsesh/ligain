package models

type ScoreBreakdown struct {
	BaseScore             int     `json:"baseScore"`
	RiskMultiplier        float64 `json:"riskMultiplier"`
	ClairvoyantMultiplier float64 `json:"clairvoyantMultiplier"`
}

type MatchResult struct {
	Match            Match
	Bets             map[string]*Bet
	Scores           map[string]int
	ScoreBreakdowns  map[string]ScoreBreakdown
	PlayerBetStatus  map[string]bool // playerID → hasBet, for non-requesting players
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
	playerBets := make(map[string]*Bet)
	for player, bet := range bets {
		playerBets[player.GetID()] = bet
	}

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

func NewScoredMatchWithBreakdowns(match Match, bets map[Player]*Bet, scores map[Player]int, breakdowns map[Player]ScoreBreakdown) *MatchResult {
	playerBets := make(map[string]*Bet)
	for player, bet := range bets {
		playerBets[player.GetID()] = bet
	}

	playerScores := make(map[string]int)
	for player, score := range scores {
		playerScores[player.GetID()] = score
	}

	playerBreakdowns := make(map[string]ScoreBreakdown)
	for player, bd := range breakdowns {
		playerBreakdowns[player.GetID()] = bd
	}

	return &MatchResult{
		Match:           match,
		Bets:            playerBets,
		Scores:          playerScores,
		ScoreBreakdowns: playerBreakdowns,
	}
}

func (m *MatchResult) IsScored() bool {
	return m.Scores != nil
}
