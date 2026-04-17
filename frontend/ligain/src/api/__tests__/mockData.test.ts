import { getMockMatchesForGame } from '../mockData';

describe('getMockMatchesForGame - allBets transformation', () => {
  describe('past matches', () => {
    it('transforms allBets into scores keyed by playerId', () => {
      const result = getMockMatchesForGame('game-1');
      const pastMatch = Object.values(result.pastMatches)[0] as any;

      expect(pastMatch.scores).toBeDefined();
      expect(pastMatch.scores['mock-player-1']).toMatchObject({
        playerId: 'mock-player-1',
        playerName: expect.any(String),
        points: expect.any(Number),
      });
    });

    it('transforms allBets into bets keyed by playerId', () => {
      const result = getMockMatchesForGame('game-1');
      const pastMatch = Object.values(result.pastMatches)[0] as any;

      expect(pastMatch.bets).toBeDefined();
      expect(pastMatch.bets['mock-player-1']).toMatchObject({
        playerId: 'mock-player-1',
        playerName: expect.any(String),
        predictedHomeGoals: expect.any(Number),
        predictedAwayGoals: expect.any(Number),
      });
    });

    it('maps home prediction to deterministic goals', () => {
      const result = getMockMatchesForGame('game-1');
      // match-19-1: TestPlayer predicted 'home' → deterministicGoals yields [3, 1]
      const match = (result.pastMatches as any)['match-19-1'];
      expect(match.bets['mock-player-1'].predictedHomeGoals).toBe(3);
      expect(match.bets['mock-player-1'].predictedAwayGoals).toBe(1);
    });

    it('maps draw prediction to 1-1 goals', () => {
      const result = getMockMatchesForGame('game-1');
      // match-19-2: TestPlayer predicted 'draw'
      const match = (result.pastMatches as any)['match-19-2'];
      expect(match.bets['mock-player-1'].predictedHomeGoals).toBe(1);
      expect(match.bets['mock-player-1'].predictedAwayGoals).toBe(1);
    });

    it('maps away prediction to 0-2 goals', () => {
      const result = getMockMatchesForGame('game-1');
      // match-19-7: TestPlayer predicted 'away'
      const match = (result.pastMatches as any)['match-19-7'];
      expect(match.bets['mock-player-1'].predictedHomeGoals).toBe(0);
      expect(match.bets['mock-player-1'].predictedAwayGoals).toBe(2);
    });

    it('assigns points when prediction matches result', () => {
      const result = getMockMatchesForGame('game-1');
      // match-19-1: PSG 3-1 OM, TestPlayer predicted 'home' with [3,1] (exact score) → 500 points
      const match = (result.pastMatches as any)['match-19-1'];
      expect(match.scores['mock-player-1'].points).toBe(500);
    });

    it('assigns 0 points when prediction does not match result', () => {
      const result = getMockMatchesForGame('game-1');
      // match-19-4: Nice 0-2 Rennes (away win), TestPlayer predicted 'home' → 0 points
      const match = (result.pastMatches as any)['match-19-4'];
      expect(match.scores['mock-player-1'].points).toBe(0);
    });
  });

  describe('incoming matches', () => {
    it('adds bets entry for current player when bet exists on the match', () => {
      const result = getMockMatchesForGame('game-1');
      // Find an incoming match that has a preset bet
      const matchWithBet = Object.values(result.incomingMatches).find(
        (m: any) => m.bets && m.bets['mock-player-1']
      ) as any;

      expect(matchWithBet).toBeDefined();
      expect(matchWithBet.bets['mock-player-1']).toMatchObject({
        playerId: 'mock-player-1',
        playerName: 'TestPlayer',
      });
    });

    it('leaves bets as undefined when no bet exists on the match', () => {
      const result = getMockMatchesForGame('game-1');
      // Find an incoming match without a preset bet
      const matchWithoutBet = Object.values(result.incomingMatches).find(
        (m: any) => !m.bet && (!m.bets || !m.bets['mock-player-1'])
      ) as any;

      expect(matchWithoutBet).toBeDefined();
    });
  });
});
