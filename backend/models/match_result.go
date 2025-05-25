package models

type MatchResult struct {
	Match  Match
	Bets   map[Player]*Bet
	Scores map[Player]int
}

func NewUnscoredMatch(match Match) *MatchResult {
	return &MatchResult{
		Match:  match,
		Bets:   nil,
		Scores: nil,
	}
}

func NewMatchWithBets(match Match, bets map[Player]*Bet) *MatchResult {
	return &MatchResult{
		Match:  match,
		Bets:   bets,
		Scores: nil,
	}
}

func NewScoredMatch(match Match, bets map[Player]*Bet, scores map[Player]int) *MatchResult {
	return &MatchResult{
		Match:  match,
		Bets:   bets,
		Scores: scores,
	}
}

func (m *MatchResult) IsScored() bool {
	return m.Scores != nil
}
