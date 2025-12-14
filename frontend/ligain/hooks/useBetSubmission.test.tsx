import { renderHook, act } from '@testing-library/react';
import { useBetSubmission } from './useBetSubmission';
import { AuthProvider } from '../src/contexts/AuthContext';

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

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

// Test wrapper component
const TestWrapper: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return <AuthProvider>{children}</AuthProvider>;
};

describe('useBetSubmission', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
    mockCancelMatchNotification.mockClear();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should submit bet successfully', async () => {
    const mockResponse = {
      message: 'Bet saved successfully',
      bet: {
        matchId: 'match-1',
        predictedHomeGoals: 2,
        predictedAwayGoals: 1,
      },
    };

    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => mockResponse,
    } as Response);

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
    expect(mockFetch).toHaveBeenCalledWith(
      'https://test-api.example.com/api/game/test-game-id/bet',
      {
        method: 'POST',
        headers: {
          'X-API-Key': 'test-api-key',
          'Authorization': 'Bearer test-token',
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          matchId: 'match-1',
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        }),
      }
    );
    
    // Notification integration: Should cancel notification on successful bet submission
    expect(mockCancelMatchNotification).toHaveBeenCalledWith('match-1');
  });

  it('should handle API errors', async () => {
    const mockResponse = {
      ok: false,
      status: 400,
      statusText: 'Bad Request',
    };
    mockFetch.mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useBetSubmission('test-game-id'), {
      wrapper: TestWrapper,
    });

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Invalid information provided. Please check your details');
  });

  it('should handle network errors', async () => {
    // Mock network error (no retry for network errors)
    mockFetch.mockRejectedValueOnce(new Error('Network error'));

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

  it('should retry on 401 errors', async () => {
    // First call returns 401, second call succeeds
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      json: async () => ({ error: 'Invalid or expired token' }),
    });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ message: 'Success' }),
    });

    const { result } = renderHook(() => useBetSubmission('test-game-id'), {
      wrapper: TestWrapper,
    });

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    // Fast-forward through the retry delay
    await act(async () => {
      jest.advanceTimersByTime(2000);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBe(null);
  });

  it('should stop retrying after max attempts on 401', async () => {
    // Both calls return 401
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      json: async () => ({ error: 'Invalid or expired token' }),
    });
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 401,
      statusText: 'Unauthorized',
      json: async () => ({ error: 'Invalid or expired token' }),
    });

    const { result } = renderHook(() => useBetSubmission('test-game-id'), {
      wrapper: TestWrapper,
    });

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    // Fast-forward through the retry delay
    await act(async () => {
      jest.advanceTimersByTime(2000);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Authentication went wrong, please refresh the page and retry');
  });

  it('should clear error on new submission', async () => {
    // First submission fails
    mockFetch.mockRejectedValueOnce(new Error('Network error'));
    // Second submission succeeds
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ message: 'Success' }),
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

  describe('Notification Integration', () => {
    it('should cancel notification when bet is successfully submitted', async () => {
      const mockResponse = {
        message: 'Bet saved successfully',
        bet: {
          matchId: 'match-123',
          predictedHomeGoals: 2,
          predictedAwayGoals: 1,
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response);

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
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

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
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        statusText: 'Bad Request',
        json: async () => ({ error: 'Invalid data' }),
      } as Response);

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
        message: 'Bet saved successfully',
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      } as Response);

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