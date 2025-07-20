import { renderHook } from '@testing-library/react';
import { useMatches } from './useMatches';
import { MockTimeService } from '../src/test-utils';
import { TimeServiceProvider } from '../src/contexts/TimeServiceContext';
import { AuthProvider } from '../src/contexts/AuthContext';
import { ReactNode } from 'react';

// Mock the API config
jest.mock('../src/config/api', () => ({
  API_CONFIG: {
    BASE_URL: 'https://test-api.example.com',
    API_KEY: 'test-api-key',
    GAME_ID: 'test-game-id',
  },
  getAuthenticatedHeaders: jest.fn().mockResolvedValue({
    'X-API-Key': 'test-api-key',
    'Authorization': 'Bearer test-token',
  }),
}));

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Wrapper component for testing hooks with providers
const wrapper = ({ children }: { children: ReactNode }) => (
  <AuthProvider>
    <TimeServiceProvider service={new MockTimeService()}>
      {children}
    </TimeServiceProvider>
  </AuthProvider>
);

describe('useMatches', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should fetch matches successfully', async () => {
    const mockResponse = {
      incomingMatches: {
        'match-1': {
          match: {
            homeTeam: 'Team A',
            awayTeam: 'Team B',
            homeGoals: 0,
            awayGoals: 0,
            homeTeamOdds: 1.5,
            awayTeamOdds: 2.5,
            drawOdds: 3.0,
            status: 'scheduled',
            seasonCode: '2024',
            competitionCode: 'Premier League',
            date: '2024-01-01T15:00:00Z',
            matchday: 1,
          },
          bets: {
            'player-1': {
              playerId: 'player-1',
              playerName: 'John Doe',
              predictedHomeGoals: 2,
              predictedAwayGoals: 1,
            },
          },
          scores: null,
        },
      },
      pastMatches: {},
    };

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    } as Response);

    const { result } = renderHook(() => useMatches('test-game-id'), { wrapper });

    // Initially loading
    expect(result.current.loading).toBe(true);
    expect(result.current.error).toBe(null);

    // Wait for the fetch to complete
    await new Promise(resolve => setTimeout(resolve, 100));

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe(null);
    expect(result.current.incomingMatches).toHaveProperty('match-1');
    expect(result.current.pastMatches).toEqual({});

    // Verify the match data is processed correctly
    const matchResult = result.current.incomingMatches['match-1'];
    expect(matchResult.match.getHomeTeam()).toBe('Team A');
    expect(matchResult.match.getAwayTeam()).toBe('Team B');
    expect(matchResult.bets).toHaveProperty('player-1');
    expect(matchResult.bets!['player-1'].playerName).toBe('John Doe');
    expect(matchResult.bets!['player-1'].predictedHomeGoals).toBe(2);
    
    // Verify odds are processed correctly
    expect(matchResult.match.getHomeTeamOdds()).toBe(1.5);
    expect(matchResult.match.getAwayTeamOdds()).toBe(2.5);
    expect(matchResult.match.getDrawOdds()).toBe(3.0);
  });

  it('should handle API errors', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      json: async () => ({ error: 'Server error' }),
    } as Response);

    const { result } = renderHook(() => useMatches('test-game-id'), { wrapper });

    await new Promise(resolve => setTimeout(resolve, 100));

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toContain('500');
    expect(result.current.incomingMatches).toEqual({});
    expect(result.current.pastMatches).toEqual({});
  });

  it('should handle network errors', async () => {
    mockFetch.mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useMatches('test-game-id'), { wrapper });

    await new Promise(resolve => setTimeout(resolve, 100));

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Network error');
  });

  it('should process scores correctly', async () => {
    const mockResponse = {
      incomingMatches: {},
      pastMatches: {
        'match-1': {
          match: {
            homeTeam: 'Team A',
            awayTeam: 'Team B',
            homeGoals: 2,
            awayGoals: 1,
            homeTeamOdds: 1.5,
            awayTeamOdds: 2.5,
            drawOdds: 3.0,
            status: 'finished',
            seasonCode: '2024',
            competitionCode: 'Premier League',
            date: '2024-01-01T15:00:00Z',
            matchday: 1,
          },
          bets: {
            'player-1': {
              playerId: 'player-1',
              playerName: 'John Doe',
              predictedHomeGoals: 2,
              predictedAwayGoals: 1,
            },
          },
          scores: {
            'player-1': {
              playerId: 'player-1',
              playerName: 'John Doe',
              points: 3,
            },
          },
        },
      },
    };

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    } as Response);

    const { result } = renderHook(() => useMatches('test-game-id'), { wrapper });

    await new Promise(resolve => setTimeout(resolve, 100));

    const matchResult = result.current.pastMatches['match-1'];
    expect(matchResult.scores).toHaveProperty('player-1');
    expect(matchResult.scores!['player-1'].points).toBe(3);
    expect(matchResult.scores!['player-1'].playerName).toBe('John Doe');
  });

  it('should handle empty responses', async () => {
    const mockResponse = {
      incomingMatches: {},
      pastMatches: {},
    };

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    } as Response);

    const { result } = renderHook(() => useMatches('test-game-id'), { wrapper });

    await new Promise(resolve => setTimeout(resolve, 100));

    expect(result.current.incomingMatches).toEqual({});
    expect(result.current.pastMatches).toEqual({});
    expect(result.current.error).toBe(null);
  });

  it('should call refresh function', async () => {
    const mockResponse = {
      incomingMatches: {},
      pastMatches: {},
    };

    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
    } as Response);

    const { result } = renderHook(() => useMatches('test-game-id'), { wrapper });

    await new Promise(resolve => setTimeout(resolve, 100));

    // Call refresh
    await result.current.refresh();

    // Should have called fetch twice (initial + refresh)
    expect(mockFetch).toHaveBeenCalledTimes(2);
  });
}); 