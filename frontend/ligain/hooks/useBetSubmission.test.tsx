import { renderHook, act } from '@testing-library/react';
import { useBetSubmission } from './useBetSubmission';

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

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

describe('useBetSubmission', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.useFakeTimers();
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

    const { result } = renderHook(() => useBetSubmission('test-game-id'));

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
  });

  it('should handle API errors', async () => {
    const mockResponse = {
      ok: false,
      status: 400,
      statusText: 'Bad Request',
    };
    // Mock both the initial call and the retry call
    mockFetch.mockResolvedValueOnce(mockResponse);
    mockFetch.mockResolvedValueOnce(mockResponse);

    const { result } = renderHook(() => useBetSubmission('test-game-id'));

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    // Fast-forward through the retry delay
    await act(async () => {
      jest.advanceTimersByTime(2000);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toContain('400');
  });

  it('should handle network errors', async () => {
    // Mock both the initial call and the retry call
    mockFetch.mockRejectedValueOnce(new Error('Network error'));
    mockFetch.mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useBetSubmission('test-game-id'));

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    // Fast-forward through the retry delay
    await act(async () => {
      jest.advanceTimersByTime(2000);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    // The hook preserves the original error message, not the generic one
    expect(result.current.error?.message).toBe('Network error');
  });

  it('should retry on failure', async () => {
    // First call fails, second call succeeds
    mockFetch.mockRejectedValueOnce(new Error('Network error'));
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ message: 'Success' }),
    });

    const { result } = renderHook(() => useBetSubmission('test-game-id'));

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

  it('should stop retrying after max attempts', async () => {
    // Both calls fail
    mockFetch.mockRejectedValueOnce(new Error('Network error'));
    mockFetch.mockRejectedValueOnce(new Error('Network error'));

    const { result } = renderHook(() => useBetSubmission('test-game-id'));

    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    // Fast-forward through the retry delay
    await act(async () => {
      jest.advanceTimersByTime(2000);
    });

    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeInstanceOf(Error);
    expect(result.current.error?.message).toBe('Network error');
  });

  it('should set isSubmitting state correctly', async () => {
    // Mock a slow response to catch the submitting state
    mockFetch.mockImplementation(() => 
      new Promise(resolve => 
        setTimeout(() => 
          resolve({
            ok: true,
            json: async () => ({ message: 'Success' }),
          }), 100)
      )
    );

    const { result } = renderHook(() => useBetSubmission('test-game-id'));

    // Start the submission
    const submitPromise = result.current.submitBet('match-1', 2, 1);

    // Fast-forward to catch the submitting state
    await act(async () => {
      jest.advanceTimersByTime(10);
    });

    // Should be submitting immediately after starting
    expect(result.current.isSubmitting).toBe(true);

    // Fast-forward to complete the request
    await act(async () => {
      jest.advanceTimersByTime(100);
      await submitPromise;
    });

    // Should be done after completion
    expect(result.current.isSubmitting).toBe(false);
  });

  it('should clear error on new submission', async () => {
    // First submission fails
    mockFetch.mockRejectedValueOnce(new Error('Network error'));
    mockFetch.mockRejectedValueOnce(new Error('Network error'));
    // Second submission succeeds
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ message: 'Success' }),
    });

    const { result } = renderHook(() => useBetSubmission('test-game-id'));

    // First submission
    await act(async () => {
      await result.current.submitBet('match-1', 2, 1);
    });

    // Fast-forward through the retry delay
    await act(async () => {
      jest.advanceTimersByTime(2000);
    });

    expect(result.current.error).toBeInstanceOf(Error);

    // Second submission - start it and immediately check if error is cleared
    act(() => {
      result.current.submitBet('match-1', 1, 1);
    });
    
    // The error should be cleared when starting a new submission
    expect(result.current.error).toBe(null);
  });
}); 