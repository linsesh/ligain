import { renderHook } from '@testing-library/react-native';
import { useNextMatch } from '../useNextMatch';
import { MatchResult } from '../../src/types/match';

// ─── Mock useMatches ──────────────────────────────────────────────────────────

let mockIncomingByMatchday: Record<number, MatchResult[]> = {};

jest.mock('../useMatches', () => ({
  useMatches: () => ({ incomingByMatchday: mockIncomingByMatchday }),
}));

// ─── Helpers ──────────────────────────────────────────────────────────────────

const GAME_ID = 'game-1';
const PLAYER_ID = 'player-1';
const MATCHDAY = 5;
const NOW = new Date('2025-04-15T12:00:00Z');

/**
 * Builds a minimal MatchResult stub. Date defaults to 1 hour after NOW (future).
 */
function makeBet(playerId: string) {
  return {
    playerId,
    playerName: 'Test Player',
    predictedHomeGoals: 1,
    predictedAwayGoals: 0,
    isModifiable: () => true,
  };
}

function makeMatchResult(
  id: string,
  opts: {
    date?: Date;
    betPlayerIds?: string[];
    matchday?: number;
  } = {},
): MatchResult {
  const date = opts.date ?? new Date(NOW.getTime() + 60 * 60 * 1000);
  const matchday = opts.matchday ?? MATCHDAY;
  const bets = opts.betPlayerIds
    ? Object.fromEntries(opts.betPlayerIds.map(pid => [pid, makeBet(pid)]))
    : null;
  return {
    match: {
      id: () => id,
      getDate: () => date,
      getMatchday: () => matchday,
      hasStarted: (now: Date) => date <= now,
    } as any,
    bets,
    scores: null,
    playerBetStatuses: null,
  };
}

// ─── Tests ────────────────────────────────────────────────────────────────────

describe('useNextMatch', () => {
  beforeEach(() => {
    mockIncomingByMatchday = {};
    jest.useFakeTimers();
    jest.setSystemTime(NOW);
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('returns zero remaining and null nextMatch when no matches exist for the matchday', () => {
    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(0);
    expect(result.current.nextMatch).toBeNull();
  });

  it('excludes the current match from results', () => {
    const current = makeMatchResult('current-match');
    mockIncomingByMatchday = { [MATCHDAY]: [current] };

    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(0);
    expect(result.current.nextMatch).toBeNull();
  });

  it('excludes matches that have already started', () => {
    const started = makeMatchResult('started-match', {
      date: new Date(NOW.getTime() - 60 * 1000), // 1 minute ago
    });
    mockIncomingByMatchday = { [MATCHDAY]: [started] };

    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(0);
    expect(result.current.nextMatch).toBeNull();
  });

  it('excludes matches where the player already has a bet', () => {
    const alreadyBet = makeMatchResult('bet-match', {
      betPlayerIds: [PLAYER_ID],
    });
    mockIncomingByMatchday = { [MATCHDAY]: [alreadyBet] };

    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(0);
    expect(result.current.nextMatch).toBeNull();
  });

  it('includes a future unbet sibling match and returns it as nextMatch', () => {
    const sibling = makeMatchResult('sibling-match');
    mockIncomingByMatchday = { [MATCHDAY]: [sibling] };

    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(1);
    expect(result.current.nextMatch).toBe(sibling);
  });

  it('returns the earliest upcoming match as nextMatch when multiple are available', () => {
    const late = makeMatchResult('late-match', {
      date: new Date(NOW.getTime() + 3 * 60 * 60 * 1000), // +3h
    });
    const early = makeMatchResult('early-match', {
      date: new Date(NOW.getTime() + 1 * 60 * 60 * 1000), // +1h
    });
    const mid = makeMatchResult('mid-match', {
      date: new Date(NOW.getTime() + 2 * 60 * 60 * 1000), // +2h
    });
    mockIncomingByMatchday = { [MATCHDAY]: [late, early, mid] };

    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(3);
    expect(result.current.nextMatch).toBe(early);
  });

  it('counts only the unbet non-started sibling matches', () => {
    const current = makeMatchResult('current-match');
    const unbet = makeMatchResult('unbet-1');
    const alreadyBet = makeMatchResult('bet-match', {
      betPlayerIds: [PLAYER_ID],
    });
    const started = makeMatchResult('started-match', {
      date: new Date(NOW.getTime() - 1),
    });
    mockIncomingByMatchday = { [MATCHDAY]: [current, unbet, alreadyBet, started] };

    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(1);
    expect(result.current.nextMatch).toBe(unbet);
  });

  it('does not count a bet placed by a different player as the current player having bet', () => {
    const sibling = makeMatchResult('sibling-match', {
      betPlayerIds: ['other-player'],
    });
    mockIncomingByMatchday = { [MATCHDAY]: [sibling] };

    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(1);
    expect(result.current.nextMatch).toBe(sibling);
  });

  it('ignores matches from other matchdays', () => {
    const otherDay = makeMatchResult('other-day-match', { matchday: MATCHDAY + 1 });
    mockIncomingByMatchday = {
      [MATCHDAY]: [],
      [MATCHDAY + 1]: [otherDay],
    };

    const { result } = renderHook(() =>
      useNextMatch(GAME_ID, 'current-match', MATCHDAY, PLAYER_ID),
    );

    expect(result.current.remainingCount).toBe(0);
    expect(result.current.nextMatch).toBeNull();
  });
});
