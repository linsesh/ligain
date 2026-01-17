import { renderHook, act, waitFor } from '@testing-library/react';
import { useBetSynchronization } from './useBetSynchronization';

// Mock getGameMatches function - defined before the mock
const mockGetGameMatches = jest.fn();

// Mock all the contexts and dependencies
jest.mock('../src/contexts/GamesContext', () => ({
  useGames: jest.fn(),
}));

jest.mock('../src/contexts/AuthContext', () => ({
  useAuth: jest.fn(),
}));

jest.mock('../src/contexts/TimeServiceContext', () => ({
  useTimeService: jest.fn(),
}));

jest.mock('./useMatches', () => ({
  useMatches: jest.fn(),
}));

// Mock the API provider
jest.mock('../src/api', () => ({
  useGamesApi: () => ({
    getGames: jest.fn(),
    getGameMatches: (...args: any[]) => mockGetGameMatches(...args),
    createGame: jest.fn(),
    joinGame: jest.fn(),
    placeBet: jest.fn(),
    leaveGame: jest.fn(),
  }),
}));

jest.mock('../src/config/api', () => ({
  API_CONFIG: { BASE_URL: 'https://test-api.example.com' },
  getAuthenticatedHeaders: jest.fn().mockResolvedValue({
    'X-API-Key': 'test-api-key',
    Authorization: 'Bearer test-token',
  }),
}));

describe('useBetSynchronization', () => {
  const mockPlayer = { id: 'player-1', name: 'Test Player' };
  const mockTimeService = {
    now: jest.fn().mockReturnValue(new Date('2024-01-15T10:00:00Z')),
  };

  // Get the mocked functions
  const mockUseGames = require('../src/contexts/GamesContext').useGames;
  const mockUseAuth = require('../src/contexts/AuthContext').useAuth;
  const mockUseTimeService = require('../src/contexts/TimeServiceContext').useTimeService;
  const mockUseMatches = require('./useMatches').useMatches;

  // Helper to create a match object that looks like what useMatches returns (with methods)
  const createCurrentGameMatch = (
    id: string,
    homeTeam: string,
    awayTeam: string,
    matchday: number,
    date: string,
    playerBet?: { predictedHomeGoals: number; predictedAwayGoals: number }
  ) => ({
    match: {
      id: () => id,
      getHomeTeam: () => homeTeam,
      getAwayTeam: () => awayTeam,
      getMatchday: () => matchday,
      getDate: () => new Date(date),
    },
    bets: playerBet
      ? { 'player-1': { playerId: 'player-1', ...playerBet } }
      : {},
  });

  // Helper to create match data for API response (raw JSON for SeasonMatch.fromJSON)
  const createApiMatchData = (
    homeTeam: string,
    awayTeam: string,
    matchday: number,
    date: string,
    playerBet?: { predictedHomeGoals: number; predictedAwayGoals: number }
  ) => ({
    match: {
      homeTeam,
      awayTeam,
      homeGoals: null,
      awayGoals: null,
      homeTeamOdds: null,
      awayTeamOdds: null,
      drawOdds: null,
      status: 'scheduled',
      seasonCode: '2024/2025',
      competitionCode: 'Ligue 1',
      date,
      matchday,
    },
    bets: playerBet
      ? { 'player-1': { playerId: 'player-1', playerName: 'Test Player', ...playerBet } }
      : {},
  });

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({ player: mockPlayer } as any);
    mockUseTimeService.mockReturnValue(mockTimeService as any);
    mockGetGameMatches.mockReset();
    // Default: return empty matches
    mockGetGameMatches.mockResolvedValue({ incomingMatches: {}, pastMatches: {} });
  });

  afterEach(() => {
    jest.clearAllMocks();
    mockGetGameMatches.mockReset();
  });

  it('should return null when no other games exist', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
      ],
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    // Wait for hook to settle
    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.syncOpportunity).toBeNull();
    // Should not call API since there are no other games
    expect(mockGetGameMatches).not.toHaveBeenCalled();
  });

  it('should return null when no matching season/league games', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2023/2024', competitionName: 'Ligue 1', name: 'Game 2' },
        {
          gameId: 'game-3',
          seasonYear: '2024/2025',
          competitionName: 'Premier League',
          name: 'Game 3',
        },
      ],
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.syncOpportunity).toBeNull();
  });

  it('should return null when player has no bets in other games', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
      ],
    } as any);

    // Current game has a future match without bet
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': createCurrentGameMatch('match-1', 'Team A', 'Team B', 1, '2024-01-20T15:00:00Z'),
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    // Other game has matches but no bets from player
    mockGetGameMatches.mockResolvedValueOnce({
      incomingMatches: {
        'match-1': createApiMatchData('Team A', 'Team B', 1, '2024-01-20T15:00:00Z'),
      },
      pastMatches: {},
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.syncOpportunity).toBeNull();
    expect(mockGetGameMatches).toHaveBeenCalledWith('game-2');
  });

  it('should return null when player has already bet on all matches in current game', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
      ],
    } as any);

    // Current game has a match WITH player's bet already
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': createCurrentGameMatch('match-1', 'Team A', 'Team B', 1, '2024-01-20T15:00:00Z', {
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        }),
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    // Other game also has player's bet on the same match
    mockGetGameMatches.mockResolvedValueOnce({
      incomingMatches: {
        'match-1': createApiMatchData('Team A', 'Team B', 1, '2024-01-20T15:00:00Z', {
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        }),
      },
      pastMatches: {},
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Should be null because player already bet on match-1 in current game
    expect(result.current.syncOpportunity).toBeNull();
  });

  it('should return correct source game with matches to sync', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
      ],
    } as any);

    // Current game has a future match WITHOUT player's bet
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': createCurrentGameMatch('match-1', 'Team A', 'Team B', 1, '2024-01-20T15:00:00Z'),
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    // Other game has the same match WITH player's bet
    // Use mockResolvedValue (not Once) because the hook may call this multiple times
    mockGetGameMatches.mockResolvedValue({
      incomingMatches: {
        'other-match-1': createApiMatchData('Team A', 'Team B', 1, '2024-01-20T15:00:00Z', {
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        }),
      },
      pastMatches: {},
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    // Wait for sync opportunity to be set with the expected data
    await waitFor(() => {
      expect(result.current.syncOpportunity).not.toBeNull();
    });

    expect(result.current.syncOpportunity).toEqual({
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        },
      ],
    });
  });

  it('should correctly identify multiple matches to sync', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
      ],
    } as any);

    // Current game has 2 future matches WITHOUT player's bets
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': createCurrentGameMatch('match-1', 'Team A', 'Team B', 1, '2024-01-20T15:00:00Z'),
        'match-2': createCurrentGameMatch('match-2', 'Team C', 'Team D', 2, '2024-01-21T15:00:00Z'),
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    // Other game has both matches WITH player's bets
    // Use mockResolvedValue (not Once) because the hook may call this multiple times
    mockGetGameMatches.mockResolvedValue({
      incomingMatches: {
        'other-match-1': createApiMatchData('Team A', 'Team B', 1, '2024-01-20T15:00:00Z', {
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        }),
        'other-match-2': createApiMatchData('Team C', 'Team D', 2, '2024-01-21T15:00:00Z', {
          predictedHomeGoals: 1,
          predictedAwayGoals: 0,
        }),
      },
      pastMatches: {},
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    // Wait for sync opportunity to be set with the expected data
    await waitFor(() => {
      expect(result.current.syncOpportunity?.matchesToSync).toHaveLength(2);
    });

    expect(result.current.syncOpportunity?.matchesToSync).toHaveLength(2);
    expect(result.current.syncOpportunity?.matchesToSync).toEqual(
      expect.arrayContaining([
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        },
        {
          matchId: 'match-2',
          homeTeam: 'Team C',
          awayTeam: 'Team D',
          matchday: 2,
          predictedHomeGoals: 1,
          predictedAwayGoals: 0,
        },
      ])
    );
  });

  it('should handle API errors gracefully', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
      ],
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': createCurrentGameMatch('match-1', 'Team A', 'Team B', 1, '2024-01-20T15:00:00Z'),
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    // Mock API error - use mockRejectedValue (not Once) because the hook may call this multiple times
    mockGetGameMatches.mockRejectedValue(new Error('API Error'));

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    // Wait for error to be set
    await waitFor(() => {
      expect(result.current.error).toBe('API Error');
    });

    expect(result.current.syncOpportunity).toBeNull();
    expect(result.current.loading).toBe(false);
  });

  it('should handle invalid match data gracefully', async () => {
    // Mock useMatches to return invalid match data
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': {
          // Missing match property
          bets: {},
        },
        'match-2': {
          match: null, // Null match
          bets: {},
        },
        'match-3': {
          match: {
            // Match without getDate method
            getHomeTeam: () => 'Team A',
            getAwayTeam: () => 'Team B',
            getMatchday: () => 1,
            // Missing getDate method
          },
          bets: {},
        },
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    // Mock other games
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
      ],
    });

    // Mock successful API response for other game
    mockGetGameMatches.mockResolvedValueOnce({
      incomingMatches: {
        'other-match-1': createApiMatchData('Team A', 'Team B', 1, '2024-01-20T15:00:00Z', {
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        }),
      },
      pastMatches: {},
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // Should not crash and should handle invalid data gracefully
    // No future matches in current game (all filtered out due to invalid data)
    // so sync opportunity should be null
    expect(result.current.syncOpportunity).toBeNull();
    expect(result.current.error).toBeNull();
  });

  it('should pick the game with most matches to sync when multiple options exist', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
        { gameId: 'game-3', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 3' },
      ],
    } as any);

    // Current game has 2 future matches WITHOUT bets
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': createCurrentGameMatch('match-1', 'Team A', 'Team B', 1, '2024-01-20T15:00:00Z'),
        'match-2': createCurrentGameMatch('match-2', 'Team C', 'Team D', 2, '2024-01-21T15:00:00Z'),
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });

    // Use mockImplementation to return different data based on gameId
    // This handles the case where the hook may call the API multiple times
    mockGetGameMatches.mockImplementation((gameId: string) => {
      if (gameId === 'game-2') {
        // Game-2 has 1 match with bet
        return Promise.resolve({
          incomingMatches: {
            'g2-match-1': createApiMatchData('Team A', 'Team B', 1, '2024-01-20T15:00:00Z', {
              predictedHomeGoals: 1,
              predictedAwayGoals: 1,
            }),
          },
          pastMatches: {},
        });
      } else if (gameId === 'game-3') {
        // Game-3 has 2 matches with bets (should be selected)
        return Promise.resolve({
          incomingMatches: {
            'g3-match-1': createApiMatchData('Team A', 'Team B', 1, '2024-01-20T15:00:00Z', {
              predictedHomeGoals: 2,
              predictedAwayGoals: 0,
            }),
            'g3-match-2': createApiMatchData('Team C', 'Team D', 2, '2024-01-21T15:00:00Z', {
              predictedHomeGoals: 3,
              predictedAwayGoals: 1,
            }),
          },
          pastMatches: {},
        });
      }
      return Promise.resolve({ incomingMatches: {}, pastMatches: {} });
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    // Wait for sync opportunity to be set with game-3 selected
    await waitFor(() => {
      expect(result.current.syncOpportunity?.sourceGameId).toBe('game-3');
    });

    // Should select game-3 because it has more matches to sync
    expect(result.current.syncOpportunity?.sourceGameId).toBe('game-3');
    expect(result.current.syncOpportunity?.matchesToSync).toHaveLength(2);
  });
});
