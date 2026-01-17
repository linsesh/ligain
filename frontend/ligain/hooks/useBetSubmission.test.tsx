import { renderHook, act } from '@testing-library/react';
import { useBetSubmission } from './useBetSubmission';
import { AuthProvider } from '../src/contexts/AuthContext';
import { ApiProvider } from '../src/api';

// Mock placeBet function - defined before the mock
const mockPlaceBet = jest.fn();

// Mock the API module to provide mock implementations
jest.mock('../src/api/ApiProvider', () => {
  const React = require('react');

  const mockAuthApi = {
    checkAuth: jest.fn().mockResolvedValue(null),
    signIn: jest.fn(),
    signInGuest: jest.fn(),
    signOut: jest.fn(),
  };

  const mockGamesApi = {
    getGames: jest.fn(),
    getGameMatches: jest.fn(),
    createGame: jest.fn(),
    joinGame: jest.fn(),
    placeBet: (...args: any[]) => mockPlaceBet(...args),
    leaveGame: jest.fn(),
  };

  const ApiContext = React.createContext({ auth: mockAuthApi, games: mockGamesApi });

  return {
    ApiProvider: ({ children }: { children: React.ReactNode }) =>
      React.createElement(
        ApiContext.Provider,
        { value: { auth: mockAuthApi, games: mockGamesApi } },
        children
      ),
    useApi: () => React.useContext(ApiContext),
    useAuthApi: () => mockAuthApi,
    useGamesApi: () => mockGamesApi,
  };
});

// Mock the API config
jest.mock('../src/config/api', () => ({
  API_CONFIG: {
    BASE_URL: 'https://test-api.example.com',
    API_KEY: 'test-api-key',
    GAME_ID: 'test-game-id',
  },
  getAuthenticatedHeaders: jest.fn().mockResolvedValue({
    'X-API-Key': 'test-api-key',
    Authorization: 'Bearer test-token',
    'Content-Type': 'application/json',
  }),
}));

// Mock useNotifications hook
const mockCancelMatchNotification = jest.fn().mockResolvedValue(undefined);
jest.mock('../src/hooks/useNotifications', () => ({
  useNotifications: jest.fn(() => ({
    preferences: { enabled: true, permissionGranted: true },
    isLoading: false,
    requestPermissions: jest.fn(),
    setNotificationEnabled: jest.fn(),
    scheduleMatchNotification: jest.fn(),
    cancelMatchNotification: mockCancelMatchNotification,
    cancelAllNotifications: jest.fn(),
    checkPermissionStatus: jest.fn(),
  })),
}));

// Test wrapper component
const TestWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <ApiProvider>
      <AuthProvider>{children}</AuthProvider>
    </ApiProvider>
  );
};

describe('useBetSubmission', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockCancelMatchNotification.mockClear();
    mockPlaceBet.mockReset();
  });

  it('should submit bet successfully', async () => {
    const mockResponse = {
      bet: {
        matchId: 'match-1',
        predictedHomeGoals: 2,
        predictedAwayGoals: 1,
      },
    };

    mockPlaceBet.mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useBetSubmission('test-game-id'), {
      wrapper: TestWrapper,
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBe(null);

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBe(null);
    expect(mockPlaceBet).toHaveBeenCalledWith('test-game-id', 'match-1', 2, 1);

    // Notification integration: Should cancel notification on successful bet submission
    expect(mockCancelMatchNotification).toHaveBeenCalledWith('match-1');
  });

  it('should handle API errors', async () => {
    mockPlaceBet.mockRejectedValueOnce(new Error('Bad Request'));

    const { result } = renderHook(() => useBetSubmission('test-game-id'), {
      wrapper: TestWrapper,
    });

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    // "Bad Request" gets translated to this specific message
    expect(result.current.error?.message).toBe('Invalid information provided. Please check your details');
  });

  it('should handle network errors', async () => {
    mockPlaceBet.mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useBetSubmission('test-game-id'), {
      wrapper: TestWrapper,
    });

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Something went wrong. Please try again later');
  });

  it('should clear error on new submission', async () => {
    // First submission fails
    mockPlaceBet.mockRejectedValueOnce(new Error('Network error'));
    // Second submission succeeds
    mockPlaceBet.mockResolvedValueOnce({
      bet: { matchId: 'match-1', predictedHomeGoals: 1, predictedAwayGoals: 1 },
    });

    const { result } = renderHook(() => useBetSubmission('test-game-id'), {
      wrapper: TestWrapper,
    });

    // First submission
    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    expect(result.current.error).toBeInstanceOf(Error);

    // Second submission - start it and immediately check if error is cleared
    act(() => {
      result.current.submitBet('match-1', 1, 1);
    });

    // The error should be cleared when starting a new submission
    expect(result.current.error).toBe(null);
  });

  it('should set lastFailedMatchId on error', async () => {
    const mockOnFail = jest.fn();
    mockPlaceBet.mockRejectedValueOnce(new Error('Failed'));

    const { result } = renderHook(() => useBetSubmission('test-game-id', mockOnFail), {
      wrapper: TestWrapper,
    });

    await act(async () => {
      await result.current.submitBet('match-123', 2, 1);
    });

    expect(result.current.lastFailedMatchId).toBe('match-123');
    expect(mockOnFail).toHaveBeenCalledWith('match-123');
  });

  describe('Notification Integration', () => {
    it('should cancel notification when bet is successfully submitted', async () => {
      const mockResponse = {
        bet: {
          matchId: 'match-123',
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        },
      };

      mockPlaceBet.mockResolvedValueOnce(mockResponse);

      const { result } = renderHook(() => useBetSubmission('test-game-id'), {
        wrapper: TestWrapper,
      });

      await act(async () => {
        await result.current.submitBet('match-123', 2, 1);
      });

      // Should cancel notification for the match
      expect(mockCancelMatchNotification).toHaveBeenCalledWith('match-123');
      expect(mockCancelMatchNotification).toHaveBeenCalledTimes(1);
    });

    it('should not cancel notification on failed bet submission', async () => {
      mockPlaceBet.mockRejectedValueOnce(new Error('Network error'));

      const { result } = renderHook(() => useBetSubmission('test-game-id'), {
        wrapper: TestWrapper,
      });

      await act(async () => {
        await result.current.submitBet('match-123', 2, 1);
      });

      // Should not cancel notification if bet submission failed
      expect(mockCancelMatchNotification).not.toHaveBeenCalled();
    });

    it('should not cancel notification on API error', async () => {
      mockPlaceBet.mockRejectedValueOnce(new Error('Invalid data'));

      const { result } = renderHook(() => useBetSubmission('test-game-id'), {
        wrapper: TestWrapper,
      });

      await act(async () => {
        await result.current.submitBet('match-123', 2, 1);
      });

      // Should not cancel notification if bet submission failed
      expect(mockCancelMatchNotification).not.toHaveBeenCalled();
    });

    it('should handle notification cancellation errors gracefully', async () => {
      const mockResponse = {
        bet: {
          matchId: 'match-123',
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        },
      };

      mockPlaceBet.mockResolvedValueOnce(mockResponse);

      // Make notification cancellation fail
      mockCancelMatchNotification.mockRejectedValueOnce(new Error('Cancellation error'));

      const { result } = renderHook(() => useBetSubmission('test-game-id'), {
        wrapper: TestWrapper,
      });

      await act(async () => {
        await result.current.submitBet('match-123', 2, 1);
      });

      // Bet submission should still succeed even if notification cancellation fails
      expect(result.current.isSubmitting).toBe(false);
      expect(result.current.error).toBe(null);
      expect(mockCancelMatchNotification).toHaveBeenCalledWith('match-123');
    });
  });
});
