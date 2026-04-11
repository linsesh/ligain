import { SeasonMatch, MatchResult } from '../types/match';
import { computeLeagueStandings, resolveCurrentMatchday, computeFingerprint } from './standings';

// Helper to build a minimal MatchResult for testing
function makeMatch(opts: {
  homeTeam: string;
  awayTeam: string;
  homeGoals: number;
  awayGoals: number;
  status: 'scheduled' | 'in-progress' | 'finished';
  matchday: number;
  date: string;
}): MatchResult {
  return {
    match: SeasonMatch.fromJSON({
      homeTeam: opts.homeTeam,
      awayTeam: opts.awayTeam,
      homeGoals: opts.homeGoals,
      awayGoals: opts.awayGoals,
      homeTeamOdds: 1.5,
      awayTeamOdds: 2.5,
      drawOdds: 3.0,
      status: opts.status,
      seasonCode: '2024-2025',
      competitionCode: 'L1',
      date: opts.date,
      matchday: opts.matchday,
    }),
    bets: null,
    scores: null,
    playerBetStatuses: null,
  };
}

// ─── computeLeagueStandings ───────────────────────────────────────────────────

describe('computeLeagueStandings', () => {
  it('returns empty array for empty input', () => {
    expect(computeLeagueStandings({})).toEqual([]);
  });

  it('ignores non-finished matches (scheduled, in-progress)', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 2, awayGoals: 0, status: 'scheduled', matchday: 1, date: '2025-01-01T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 1, awayGoals: 1, status: 'in-progress', matchday: 1, date: '2025-01-01T20:00:00Z' }),
    };
    expect(computeLeagueStandings(matches)).toEqual([]);
  });

  it('correctly accumulates a win: 3 pts for winner, 0 for loser', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'PSG', awayTeam: 'Lyon', homeGoals: 2, awayGoals: 1, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
    };
    const standings = computeLeagueStandings(matches);
    const psg = standings.find(s => s.teamName === 'PSG')!;
    const lyon = standings.find(s => s.teamName === 'Lyon')!;

    expect(psg.won).toBe(1);
    expect(psg.drawn).toBe(0);
    expect(psg.lost).toBe(0);
    expect(psg.points).toBe(3);
    expect(lyon.won).toBe(0);
    expect(lyon.lost).toBe(1);
    expect(lyon.points).toBe(0);
  });

  it('correctly accumulates a draw: 1 pt each', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 1, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
    };
    const standings = computeLeagueStandings(matches);
    const a = standings.find(s => s.teamName === 'A')!;
    const b = standings.find(s => s.teamName === 'B')!;

    expect(a.drawn).toBe(1);
    expect(a.points).toBe(1);
    expect(b.drawn).toBe(1);
    expect(b.points).toBe(1);
  });

  it('computes goal difference = goalsFor - goalsAgainst', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 3, awayGoals: 1, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
    };
    const standings = computeLeagueStandings(matches);
    const a = standings.find(s => s.teamName === 'A')!;
    const b = standings.find(s => s.teamName === 'B')!;

    expect(a.goalsFor).toBe(3);
    expect(a.goalsAgainst).toBe(1);
    expect(a.goalDifference).toBe(2);
    expect(b.goalsFor).toBe(1);
    expect(b.goalsAgainst).toBe(3);
    expect(b.goalDifference).toBe(-2);
  });

  it('accumulates correctly when a team appears as both home and away', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 2, awayGoals: 0, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'A', homeGoals: 0, awayGoals: 1, status: 'finished', matchday: 2, date: '2025-01-08T20:00:00Z' }),
    };
    const standings = computeLeagueStandings(matches);
    const a = standings.find(s => s.teamName === 'A')!;

    expect(a.played).toBe(2);
    expect(a.won).toBe(2);
    expect(a.points).toBe(6);
    expect(a.goalsFor).toBe(3); // 2 home + 1 away
    expect(a.goalsAgainst).toBe(0);
  });

  it('sorts by points desc, then goal difference desc, then goalsFor desc, then teamName asc', () => {
    // A: 6 pts, GD +6, GF 6
    // B: 6 pts, GD +4, GF 5
    // C: 6 pts, GD +4, GF 6
    // D: 0 pts
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'D', homeGoals: 3, awayGoals: 0, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'A', awayTeam: 'D', homeGoals: 3, awayGoals: 0, status: 'finished', matchday: 2, date: '2025-01-08T20:00:00Z' }),
      m3: makeMatch({ homeTeam: 'B', awayTeam: 'D', homeGoals: 3, awayGoals: 0, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
      m4: makeMatch({ homeTeam: 'B', awayTeam: 'D', homeGoals: 2, awayGoals: 1, status: 'finished', matchday: 2, date: '2025-01-08T20:00:00Z' }),
      m5: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 3, awayGoals: 0, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
      m6: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 3, awayGoals: 2, status: 'finished', matchday: 2, date: '2025-01-08T20:00:00Z' }),
    };
    const standings = computeLeagueStandings(matches);
    const names = standings.map(s => s.teamName);

    // A: 6pts, GD+6, GF6
    // C: 6pts, GD+4, GF6  → GD tie with B, but GF wins over B
    // B: 6pts, GD+4, GF5
    // D: 0pts
    expect(names).toEqual(['A', 'C', 'B', 'D']);
  });

  it('assigns 1-based position correctly', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
    };
    const standings = computeLeagueStandings(matches);
    const a = standings.find(s => s.teamName === 'A')!;
    const b = standings.find(s => s.teamName === 'B')!;

    expect(a.position).toBe(1);
    expect(b.position).toBe(2);
  });

  it('breaks ties by teamName asc when all metrics equal', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'Zebra', awayTeam: 'Apple', homeGoals: 1, awayGoals: 1, status: 'finished', matchday: 1, date: '2025-01-01T20:00:00Z' }),
    };
    const standings = computeLeagueStandings(matches);

    expect(standings[0].teamName).toBe('Apple');
    expect(standings[1].teamName).toBe('Zebra');
  });
});

// ─── resolveCurrentMatchday ───────────────────────────────────────────────────

describe('resolveCurrentMatchday', () => {
  const NOW = new Date('2025-04-11T21:00:00Z');

  it('returns undefined for empty input', () => {
    expect(resolveCurrentMatchday({}, NOW)).toBeUndefined();
  });

  it('returns the matchday of the closest future match when nothing started recently', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 0, awayGoals: 0, status: 'scheduled', matchday: 5, date: '2025-04-12T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 0, awayGoals: 0, status: 'scheduled', matchday: 6, date: '2025-04-19T20:00:00Z' }),
    };
    expect(resolveCurrentMatchday(matches, NOW)).toBe(5);
  });

  it('returns the matchday of a match started less than 4h ago', () => {
    // Started 1h ago
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'in-progress', matchday: 3, date: '2025-04-11T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 0, awayGoals: 0, status: 'scheduled', matchday: 4, date: '2025-04-19T20:00:00Z' }),
    };
    expect(resolveCurrentMatchday(matches, NOW)).toBe(3);
  });

  it('ignores a match started more than 4h ago (falls back to closest future)', () => {
    // Started 5h ago
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 2, awayGoals: 1, status: 'finished', matchday: 2, date: '2025-04-11T16:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 0, awayGoals: 0, status: 'scheduled', matchday: 3, date: '2025-04-12T20:00:00Z' }),
    };
    expect(resolveCurrentMatchday(matches, NOW)).toBe(3);
  });

  it('when two matchdays both have recent matches, returns the one with the most recently started match', () => {
    // Matchday 7: last started at 20:30 → 30 min ago
    // Matchday 8: last started at 20:00 → 1h ago
    const now = new Date('2025-04-11T21:00:00Z');
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 0, awayGoals: 0, status: 'in-progress', matchday: 7, date: '2025-04-11T20:00:00Z' }), // 1h ago
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 0, awayGoals: 0, status: 'in-progress', matchday: 7, date: '2025-04-11T20:30:00Z' }), // 30 min ago — latest in md7
      m3: makeMatch({ homeTeam: 'E', awayTeam: 'F', homeGoals: 0, awayGoals: 0, status: 'in-progress', matchday: 8, date: '2025-04-11T20:00:00Z' }), // 1h ago — only one in md8
    };
    // md7's most recent start = 20:30; md8's most recent start = 20:00 → md7 wins
    expect(resolveCurrentMatchday(matches, now)).toBe(7);
  });

  it('returns undefined when all matches are in the past and beyond 4h', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: 1, date: '2025-04-10T15:00:00Z' }),
    };
    expect(resolveCurrentMatchday(matches, NOW)).toBeUndefined();
  });
});

// ─── computeFingerprint ───────────────────────────────────────────────────────

describe('computeFingerprint', () => {
  const MD = 5; // current matchday under test

  it('returns empty string when currentMatchday is undefined', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: 5, date: '2025-01-01T20:00:00Z' }),
    };
    expect(computeFingerprint(matches, undefined)).toBe('');
  });

  it('returns empty string when no finished matches exist in the given matchday', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 0, awayGoals: 0, status: 'scheduled', matchday: MD, date: '2025-01-01T20:00:00Z' }),
    };
    expect(computeFingerprint(matches, MD)).toBe('');
  });

  it('returns the same string when called twice with identical data', () => {
    const matches = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: MD, date: '2025-01-01T20:00:00Z' }),
    };
    expect(computeFingerprint(matches, MD)).toBe(computeFingerprint(matches, MD));
  });

  it('returns a different string when a new match becomes finished in the current matchday', () => {
    const before = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: MD, date: '2025-01-01T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 0, awayGoals: 0, status: 'in-progress', matchday: MD, date: '2025-01-01T20:00:00Z' }),
    };
    const after = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: MD, date: '2025-01-01T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 2, awayGoals: 1, status: 'finished', matchday: MD, date: '2025-01-01T20:00:00Z' }),
    };
    expect(computeFingerprint(before, MD)).not.toBe(computeFingerprint(after, MD));
  });

  it('is NOT affected by a match finishing in a different (past) matchday', () => {
    const before = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'in-progress', matchday: MD - 1, date: '2025-01-01T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 0, awayGoals: 0, status: 'scheduled', matchday: MD, date: '2025-01-08T20:00:00Z' }),
    };
    const after = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: MD - 1, date: '2025-01-01T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 0, awayGoals: 0, status: 'scheduled', matchday: MD, date: '2025-01-08T20:00:00Z' }),
    };
    expect(computeFingerprint(before, MD)).toBe(computeFingerprint(after, MD));
  });

  it('is NOT affected by a match transitioning from scheduled to in-progress in the current matchday', () => {
    const before = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 0, awayGoals: 0, status: 'scheduled', matchday: MD, date: '2025-01-01T20:00:00Z' }),
    };
    const after = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 0, awayGoals: 0, status: 'in-progress', matchday: MD, date: '2025-01-01T20:00:00Z' }),
    };
    expect(computeFingerprint(before, MD)).toBe(computeFingerprint(after, MD));
  });

  it('is order-independent (same matches in different object key order → same fingerprint)', () => {
    const order1 = {
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: MD, date: '2025-01-01T20:00:00Z' }),
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 2, awayGoals: 1, status: 'finished', matchday: MD, date: '2025-01-01T20:00:00Z' }),
    };
    const order2 = {
      m2: makeMatch({ homeTeam: 'C', awayTeam: 'D', homeGoals: 2, awayGoals: 1, status: 'finished', matchday: MD, date: '2025-01-01T20:00:00Z' }),
      m1: makeMatch({ homeTeam: 'A', awayTeam: 'B', homeGoals: 1, awayGoals: 0, status: 'finished', matchday: MD, date: '2025-01-01T20:00:00Z' }),
    };
    expect(computeFingerprint(order1, MD)).toBe(computeFingerprint(order2, MD));
  });
});
