import { computeMonthlyAndMatchdayScores, computeTotalScores } from './aggregations';

describe('computeMonthlyAndMatchdayScores', () => {
  it('returns empty structures for null/undefined input', () => {
    expect(computeMonthlyAndMatchdayScores(undefined)).toEqual({ perMonthLeaderboard: {}, perMatchdayLeaderboard: {} });
    expect(computeMonthlyAndMatchdayScores(null as any)).toEqual({ perMonthLeaderboard: {}, perMatchdayLeaderboard: {} });
  });

  it('returns empty structures for empty pastMatches object', () => {
    expect(computeMonthlyAndMatchdayScores({} as any)).toEqual({ perMonthLeaderboard: {}, perMatchdayLeaderboard: {} });
  });

  it('aggregates points per month and per matchday and sorts desc by points', () => {
    const pastMatches = {
      m1: {
        match: { date: '2024-08-15T20:00:00Z', matchday: 1 },
        scores: {
          p1: { PlayerID: 'p1', PlayerName: 'Alice', Points: 3 },
          p2: { PlayerID: 'p2', PlayerName: 'Bob', Points: 1 },
        },
      },
      m2: {
        match: { date: '2024-08-20T20:00:00Z', matchday: 1 },
        scores: {
          p1: { PlayerID: 'p1', PlayerName: 'Alice', Points: 0 },
          p2: { PlayerID: 'p2', PlayerName: 'Bob', Points: 2 },
        },
      },
      m3: {
        match: { date: '2024-09-01T20:00:00Z', matchday: 2 },
        scores: {
          p1: { PlayerID: 'p1', PlayerName: 'Alice', Points: 1 },
        },
      },
    } as any;

    const { perMonthLeaderboard, perMatchdayLeaderboard } = computeMonthlyAndMatchdayScores(pastMatches);

    expect(Object.keys(perMonthLeaderboard)).toEqual(['2024-08', '2024-09']);
    // Tied on points (3), then sorted by PlayerName ascending: Alice before Bob
    expect(perMonthLeaderboard['2024-08'].map(x => [x.PlayerID, x.Points])).toEqual([
      ['p1', 3],
      ['p2', 3],
    ]);

    expect(perMonthLeaderboard['2024-09'].map(x => [x.PlayerID, x.Points])).toEqual([
      ['p1', 1],
    ]);

    expect(Object.keys(perMatchdayLeaderboard).map(k => Number(k))).toEqual([1, 2]);
    // Same tie-break: Alice before Bob
    expect(perMatchdayLeaderboard[1].map(x => [x.PlayerID, x.Points])).toEqual([
      ['p1', 3],
      ['p2', 3],
    ]);
    expect(perMatchdayLeaderboard[2].map(x => [x.PlayerID, x.Points])).toEqual([
      ['p1', 1],
    ]);
  });

  it('skips entries without scores or invalid dates', () => {
    const past = {
      a: { match: { date: 'invalid', matchday: 1 }, scores: { p: { PlayerID: 'p', PlayerName: 'X', Points: 2 } } },
      b: { match: { date: '2024-10-10T00:00:00Z', matchday: 3 }, scores: null },
      c: { match: { date: '2024-10-10T00:00:00Z', matchday: 3 }, scores: {} },
    } as any;
    const res = computeMonthlyAndMatchdayScores(past);
    expect(res).toEqual({ perMonthLeaderboard: {}, perMatchdayLeaderboard: {} });
  });

  it('computes total scores across all matches and sorts deterministically', () => {
    const pastMatches = {
      m1: {
        match: { date: '2024-08-15T20:00:00Z', matchday: 1 },
        scores: {
          p1: { PlayerID: 'p1', PlayerName: 'Alice', Points: 3 },
          p2: { PlayerID: 'p2', PlayerName: 'Bob', Points: 1 },
        },
      },
      m2: {
        match: { date: '2024-08-20T20:00:00Z', matchday: 1 },
        scores: {
          p1: { PlayerID: 'p1', PlayerName: 'Alice', Points: 0 },
          p2: { PlayerID: 'p2', PlayerName: 'Bob', Points: 2 },
          p3: { PlayerID: 'p3', PlayerName: 'Charlie', Points: 3 },
        },
      },
    } as any;

    const total = computeTotalScores(pastMatches);
    expect(total.map(x => [x.PlayerID, x.Points])).toEqual([
      ['p1', 3],
      ['p2', 3],
      ['p3', 3],
    ]);
    // Alice should come before Bob by name when tied; Charlie after Bob by name
    expect(total[0].PlayerName).toBe('Alice');
    expect(total[1].PlayerName).toBe('Bob');
    expect(total[2].PlayerName).toBe('Charlie');
  });

  it('handles null/undefined/empty pastMatches for total scores', () => {
    expect(computeTotalScores(undefined as any)).toEqual([]);
    expect(computeTotalScores(null as any)).toEqual([]);
    expect(computeTotalScores({} as any)).toEqual([]);
  });
});


