import React from 'react';
import { renderHook, act } from '@testing-library/react-native';
import { MatchesProvider, useMatches } from '../MatchesContext';

const mockGetGameMatches = jest.fn();

jest.mock('../../api/ApiProvider', () => {
  const React = require('react');

  const mockGamesApi = {
    getGameMatches: (...args: any[]) => mockGetGameMatches(...args),
    getGames: jest.fn(),
    createGame: jest.fn(),
    joinGame: jest.fn(),
    placeBet: jest.fn(),
    leaveGame: jest.fn(),
  };

  const ApiContext = React.createContext({
    auth: {},
    games: mockGamesApi,
    profile: {},
  });

  return {
    ApiProvider: ({ children }: { children: React.ReactNode }) =>
      React.createElement(ApiContext.Provider, {
        value: { auth: {}, games: mockGamesApi, profile: {} },
      }, children),
    useApi: () => React.useContext(ApiContext),
    useAuthApi: () => ({}),
    useGamesApi: () => mockGamesApi,
    useProfileApi: () => ({}),
  };
});

let mockSelectedGameId: string | null = 'game-1';
jest.mock('../GamesContext', () => ({
  useGames: () => ({ selectedGameId: mockSelectedGameId }),
}));

jest.mock('../../utils/errorMessages', () => ({
  translateError: (msg: string) => msg,
}));

let matchCounter = 0;

function makeMatchJSON(opts: {
  homeTeam?: string;
  awayTeam?: string;
  matchday?: number;
  date?: string;
  bets?: Record<string, any> | null;
  playerBetStatuses?: Record<string, any> | null;
} = {}) {
  matchCounter++;
  const homeTeam = opts.homeTeam ?? `Home${matchCounter}`;
  const awayTeam = opts.awayTeam ?? `Away${matchCounter}`;
  const matchday = opts.matchday ?? 5;
  return {
    match: {
      homeTeam,
      awayTeam,
      homeGoals: 0,
      awayGoals: 0,
      homeTeamOdds: 2.0,
      awayTeamOdds: 3.5,
      drawOdds: 3.0,
      status: 'scheduled',
      seasonCode: '2025',
      competitionCode: 'FL1',
      date: opts.date ?? '2025-06-01T15:00:00Z',
      matchday,
    },
    bets: opts.bets ?? null,
    scores: null,
    playerBetStatuses: opts.playerBetStatuses ?? null,
  };
}

function matchId(homeTeam: string, awayTeam: string, matchday: number = 5): string {
  return `FL1-2025-${homeTeam}-${awayTeam}-${matchday}`;
}

function makeApiResponse(incomingMatches: Record<string, any>, pastMatches: Record<string, any> = {}) {
  return { incomingMatches, pastMatches };
}

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <MatchesProvider>{children}</MatchesProvider>
);

describe('MatchesContext - markBetPlaced', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    matchCounter = 0;
    mockSelectedGameId = 'game-1';
  });

  it('updates bets for the given match and player', async () => {
    const matchJSON = makeMatchJSON({ homeTeam: 'PSG', awayTeam: 'Lyon' });
    const key = 'match-1';
    const id = matchId('PSG', 'Lyon');

    mockGetGameMatches.mockResolvedValueOnce(
      makeApiResponse({ [key]: matchJSON })
    );

    const { result } = renderHook(() => useMatches(), { wrapper });

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 50));
    });

    expect(result.current.incomingMatches[key].bets).toBeNull();

    act(() => {
      result.current.markBetPlaced(id, 'player-1', 'Alice', 2, 1);
    });

    const bet = result.current.incomingMatches[key].bets?.['player-1'];
    expect(bet).toBeDefined();
    expect(bet!.playerId).toBe('player-1');
    expect(bet!.playerName).toBe('Alice');
    expect(bet!.predictedHomeGoals).toBe(2);
    expect(bet!.predictedAwayGoals).toBe(1);
  });

  it('updates incomingByMatchday so useNextMatch sees the bet', async () => {
    const day = 5;
    const match1JSON = makeMatchJSON({ homeTeam: 'PSG', awayTeam: 'Lyon', matchday: day });
    const match2JSON = makeMatchJSON({ homeTeam: 'Marseille', awayTeam: 'Lille', matchday: day });
    const id1 = matchId('PSG', 'Lyon', day);
    const id2 = matchId('Marseille', 'Lille', day);

    mockGetGameMatches.mockResolvedValueOnce(
      makeApiResponse({ 'k1': match1JSON, 'k2': match2JSON })
    );

    const { result } = renderHook(() => useMatches(), { wrapper });

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 50));
    });

    act(() => {
      result.current.markBetPlaced(id1, 'player-1', 'Alice', 3, 0);
    });

    const matchdayResults = result.current.incomingByMatchday[day];
    const updatedMatch = matchdayResults.find(mr => mr.match.id() === id1);
    expect(updatedMatch?.bets?.['player-1']).toBeDefined();

    const untouchedMatch = matchdayResults.find(mr => mr.match.id() === id2);
    expect(untouchedMatch?.bets).toBeNull();
  });

  it('updates playerBetStatuses when present', async () => {
    const matchJSON = makeMatchJSON({
      homeTeam: 'PSG',
      awayTeam: 'Lyon',
      playerBetStatuses: {
        'player-1': { playerId: 'player-1', playerName: 'Alice', hasBet: false },
      },
    });
    const key = 'match-1';
    const id = matchId('PSG', 'Lyon');

    mockGetGameMatches.mockResolvedValueOnce(
      makeApiResponse({ [key]: matchJSON })
    );

    const { result } = renderHook(() => useMatches(), { wrapper });

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 50));
    });

    expect(result.current.incomingMatches[key].playerBetStatuses?.['player-1']?.hasBet).toBe(false);

    act(() => {
      result.current.markBetPlaced(id, 'player-1', 'Alice', 1, 1);
    });

    expect(result.current.incomingMatches[key].playerBetStatuses?.['player-1']?.hasBet).toBe(true);
  });

  it('is a no-op when matchId is not found', async () => {
    const matchJSON = makeMatchJSON({ homeTeam: 'PSG', awayTeam: 'Lyon' });
    mockGetGameMatches.mockResolvedValueOnce(
      makeApiResponse({ 'match-1': matchJSON })
    );

    const { result } = renderHook(() => useMatches(), { wrapper });

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 50));
    });

    const stateBefore = result.current.incomingMatches;

    act(() => {
      result.current.markBetPlaced('nonexistent-id', 'player-1', 'Alice', 1, 0);
    });

    expect(result.current.incomingMatches).toEqual(stateBefore);
  });
});
