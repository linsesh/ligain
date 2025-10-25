import { renderHook, act } from '@testing-library/react';
import { useBetSynchronization } from './useBetSynchronization';

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

jest.mock('../src/config/api', () => ({
  API_CONFIG: { BASE_URL: 'https://test-api.example.com' },
  getAuthenticatedHeaders: jest.fn().mockResolvedValue({
    'X-API-Key': 'test-api-key',
    'Authorization': 'Bearer test-token',
  }),
}));

// Mock fetch
const mockFetch = jest.fn();
global.fetch = mockFetch;

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

  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({ player: mockPlayer } as any);
    mockUseTimeService.mockReturnValue(mockTimeService as any);
    // Don't set default useMatches here - let each test set it up
    
    // Reset fetch mock completely
    (global.fetch as jest.Mock).mockReset();
    (global.fetch as jest.Mock).mockResolvedValue({
      ok: true,
      json: async () => ({ incomingMatches: {}, pastMatches: {} })
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
    (global.fetch as jest.Mock).mockReset();
  });

  it('should return null when no other games exist', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' }
      ]
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      // Wait for the effect to complete
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    expect(result.current.syncOpportunity).toBeNull();
  });

  it('should return null when no matching season/league games', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2023/2024', competitionName: 'Ligue 1', name: 'Game 2' },
        { gameId: 'game-3', seasonYear: '2024/2025', competitionName: 'Premier League', name: 'Game 3' }
      ]
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    expect(result.current.syncOpportunity).toBeNull();
  });

  it('should return null when player has no bets in other games', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' }
      ]
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    // Mock API responses - no bets in other games
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-1': {
              match: { 
                id: () => 'match-1', 
                getHomeTeam: () => 'Team A', 
                getAwayTeam: () => 'Team B', 
                getMatchday: () => 1, 
                getDate: () => new Date('2024-01-20T15:00:00Z') 
              },
              bets: {}
            }
          },
          pastMatches: {}
        })
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-2': {
              match: { 
                id: () => 'match-2', 
                getHomeTeam: () => 'Team A', 
                getAwayTeam: () => 'Team B', 
                getMatchday: () => 1, 
                getDate: () => new Date('2024-01-20T15:00:00Z') 
              },
              bets: {}
            }
          },
          pastMatches: {}
        })
      });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    expect(result.current.syncOpportunity).toBeNull();
  });

  it('should return null when player has already bet on all matches in current game', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' }
      ]
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': {
          match: { 
            id: () => 'match-1', 
            getHomeTeam: () => 'Team A', 
            getAwayTeam: () => 'Team B', 
            getMatchday: () => 1, 
            getDate: () => new Date('2024-01-20T15:00:00Z') 
          },
          bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 2, predictedAwayGoals: 1 } }
        }
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    // Mock API responses - player has bets in both games
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-1': {
              match: { 
                id: () => 'match-1', 
                getHomeTeam: () => 'Team A', 
                getAwayTeam: () => 'Team B', 
                getMatchday: () => 1, 
                getDate: () => new Date('2024-01-20T15:00:00Z') 
              },
              bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 2, predictedAwayGoals: 1 } }
            }
          },
          pastMatches: {}
        })
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-2': {
              match: { 
                id: () => 'match-2', 
                getHomeTeam: () => 'Team A', 
                getAwayTeam: () => 'Team B', 
                getMatchday: () => 1, 
                getDate: () => new Date('2024-01-20T15:00:00Z') 
              },
              bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 2, predictedAwayGoals: 1 } }
            }
          },
          pastMatches: {}
        })
      });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    expect(result.current.syncOpportunity).toBeNull();
  });

  it('should return correct source game with most future bets', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
        { gameId: 'game-3', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 3' }
      ]
    } as any);

    // Mock current game matches (no bets)
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': {
          match: { 
            id: () => 'match-1', 
            getHomeTeam: () => 'Team A', 
            getAwayTeam: () => 'Team B', 
            getMatchday: () => 1, 
            getDate: () => new Date('2024-01-20T15:00:00Z') 
          },
          bets: {}
        }
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    // Mock API responses for other games
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-2': {
              match: {
                homeTeam: 'Team A',
                awayTeam: 'Team B',
                homeGoals: null,
                awayGoals: null,
                homeTeamOdds: null,
                awayTeamOdds: null,
                drawOdds: null,
                status: 'scheduled',
                seasonCode: '2024/2025',
                competitionCode: 'Ligue 1',
                date: '2024-01-20T15:00:00Z',
                matchday: 1
              },
              bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 2, predictedAwayGoals: 1 } }
            }
          },
          pastMatches: {}
        })
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-3': {
              match: {
                homeTeam: 'Team C',
                awayTeam: 'Team D',
                homeGoals: null,
                awayGoals: null,
                homeTeamOdds: null,
                awayTeamOdds: null,
                drawOdds: null,
                status: 'scheduled',
                seasonCode: '2024/2025',
                competitionCode: 'Ligue 1',
                date: '2024-01-21T15:00:00Z',
                matchday: 2
              },
              bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 1, predictedAwayGoals: 0 } }
            }
          },
          pastMatches: {}
        })
      });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
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
          predictedAwayGoals: 1
        }
      ]
    });
  });

  it('should return first game when multiple games have same bet count', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' },
        { gameId: 'game-3', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 3' }
      ]
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': {
          match: { 
            id: () => 'match-1', 
            getHomeTeam: () => 'Team A', 
            getAwayTeam: () => 'Team B', 
            getMatchday: () => 1, 
            getDate: () => new Date('2024-01-20T15:00:00Z') 
          },
          bets: {}
        }
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    // Mock API responses - both games have 1 bet each
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-1': {
              match: {
                homeTeam: 'Team A',
                awayTeam: 'Team B',
                homeGoals: null,
                awayGoals: null,
                homeTeamOdds: null,
                awayTeamOdds: null,
                drawOdds: null,
                status: 'scheduled',
                seasonCode: '2024/2025',
                competitionCode: 'Ligue 1',
                date: '2024-01-20T15:00:00Z',
                matchday: 1
              },
              bets: {}
            }
          },
          pastMatches: {}
        })
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-2': {
              match: {
                homeTeam: 'Team A',
                awayTeam: 'Team B',
                homeGoals: null,
                awayGoals: null,
                homeTeamOdds: null,
                awayTeamOdds: null,
                drawOdds: null,
                status: 'scheduled',
                seasonCode: '2024/2025',
                competitionCode: 'Ligue 1',
                date: '2024-01-20T15:00:00Z',
                matchday: 1
              },
              bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 2, predictedAwayGoals: 1 } }
            }
          },
          pastMatches: {}
        })
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-3': {
              match: {
                homeTeam: 'Team A',
                awayTeam: 'Team B',
                homeGoals: null,
                awayGoals: null,
                homeTeamOdds: null,
                awayTeamOdds: null,
                drawOdds: null,
                status: 'scheduled',
                seasonCode: '2024/2025',
                competitionCode: 'Ligue 1',
                date: '2024-01-20T15:00:00Z',
                matchday: 1
              },
              bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 1, predictedAwayGoals: 0 } }
            }
          },
          pastMatches: {}
        })
      });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    // Should return game-3 (the one with most bets) since it has more bets than game-2
    expect(result.current.syncOpportunity?.sourceGameId).toBe('game-3');
  });

  it('should correctly identify which matches need syncing', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' }
      ]
    } as any);

    // Mock current game matches (no bets)
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': {
          match: { 
            id: () => 'match-1', 
            getHomeTeam: () => 'Team A', 
            getAwayTeam: () => 'Team B', 
            getMatchday: () => 1, 
            getDate: () => new Date('2024-01-20T15:00:00Z') 
          },
          bets: {}
        },
        'match-2': {
          match: { 
            id: () => 'match-2', 
            getHomeTeam: () => 'Team C', 
            getAwayTeam: () => 'Team D', 
            getMatchday: () => 2, 
            getDate: () => new Date('2024-01-21T15:00:00Z') 
          },
          bets: {}
        }
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    // Mock API responses for other games
    mockFetch
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          incomingMatches: {
            'match-3': {
              match: {
                homeTeam: 'Team A',
                awayTeam: 'Team B',
                homeGoals: null,
                awayGoals: null,
                homeTeamOdds: null,
                awayTeamOdds: null,
                drawOdds: null,
                status: 'scheduled',
                seasonCode: '2024/2025',
                competitionCode: 'Ligue 1',
                date: '2024-01-20T15:00:00Z',
                matchday: 1
              },
              bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 2, predictedAwayGoals: 1 } }
            },
            'match-4': {
              match: {
                homeTeam: 'Team C',
                awayTeam: 'Team D',
                homeGoals: null,
                awayGoals: null,
                homeTeamOdds: null,
                awayTeamOdds: null,
                drawOdds: null,
                status: 'scheduled',
                seasonCode: '2024/2025',
                competitionCode: 'Ligue 1',
                date: '2024-01-21T15:00:00Z',
                matchday: 2
              },
              bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 1, predictedAwayGoals: 0 } }
            }
          },
          pastMatches: {}
        })
      });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    expect(result.current.syncOpportunity?.matchesToSync).toHaveLength(2);
    expect(result.current.syncOpportunity?.matchesToSync).toEqual([
      {
        matchId: 'match-1',
        homeTeam: 'Team A',
        awayTeam: 'Team B',
        matchday: 1,
        predictedHomeGoals: 2,
        predictedAwayGoals: 1
      },
      {
        matchId: 'match-2',
        homeTeam: 'Team C',
        awayTeam: 'Team D',
        matchday: 2,
        predictedHomeGoals: 1,
        predictedAwayGoals: 0
      }
    ]);
  });

  it('should handle API errors gracefully', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' }
      ]
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    // Mock API error
    mockFetch.mockRejectedValueOnce(new Error('API Error'));

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    expect(result.current.syncOpportunity).toBeNull();
    expect(result.current.error).toBe('API Error');
  });

  it('should handle non-200 responses', async () => {
    mockUseGames.mockReturnValue({
      games: [
        { gameId: 'game-1', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 1' },
        { gameId: 'game-2', seasonYear: '2024/2025', competitionName: 'Ligue 1', name: 'Game 2' }
      ]
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    // Mock non-200 response
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: async () => ({ error: 'Server Error' })
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    expect(result.current.syncOpportunity).toBeNull();
    expect(result.current.error).toBe('Failed to fetch matches for game game-2: 500');
  });

  it('should handle invalid match data gracefully', async () => {
    // Mock useMatches to return invalid match data
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': {
          // Missing match property
          bets: {}
        },
        'match-2': {
          match: null, // Null match
          bets: {}
        },
        'match-3': {
          match: {
            // Match without getDate method
            getHomeTeam: () => 'Team A',
            getAwayTeam: () => 'Team B',
            getMatchday: () => 1
            // Missing getDate method
          },
          bets: {}
        }
      },
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn()
    });

    // Mock other games
    mockUseGames.mockReturnValue({
      games: [
        { id: 'game-1', name: 'Game 1' },
        { id: 'game-2', name: 'Game 2' }
      ]
    });

    // Mock successful API response for other game
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({
        incomingMatches: {
          'other-match-1': {
            match: {
              id: 'other-match-1',
              homeTeam: 'Team A',
              awayTeam: 'Team B',
              matchday: 1,
              date: '2024-01-20T15:00:00Z',
              status: 'scheduled'
            },
            bets: { 'player-1': { playerId: 'player-1', predictedHomeGoals: 2, predictedAwayGoals: 1 } }
          }
        },
        pastMatches: {}
      })
    });

    const { result } = renderHook(() => useBetSynchronization('game-1'));

    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 200));
    });

    // Should not crash and should handle invalid data gracefully
    expect(result.current.syncOpportunity).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.loading).toBe(false);
  });
});
