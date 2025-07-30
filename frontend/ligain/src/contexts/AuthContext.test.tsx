import React from 'react';
import { renderHook, act } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import { API_CONFIG, getApiHeaders } from '../config/api';
import { getItem, setItem, multiRemove } from '../utils/storage';

// Mock dependencies
jest.mock('../config/api', () => ({
  API_CONFIG: {
    BASE_URL: 'https://test-api.example.com',
    API_KEY: 'test-api-key',
  },
  getApiHeaders: jest.fn(() => ({
    'X-API-Key': 'test-api-key',
  })),
}));

jest.mock('../utils/storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  multiRemove: jest.fn(),
  isUsingMemoryFallback: jest.fn(() => false),
}));

jest.mock('../utils/errorMessages', () => ({
  getHumanReadableError: jest.fn((status: number, error?: string) => {
    switch (status) {
      case 401:
        return 'Authentication went wrong, please refresh the page and retry';
      case 500:
        return 'Server error. Please try again later';
      default:
        return error || `Something went wrong (${status})`;
    }
  }),
  handleApiError: jest.fn(async (response: Response) => {
    const errorData = await response.json();
    const humanReadableError = require('../utils/errorMessages').getHumanReadableError(
      response.status,
      errorData.error
    );
    throw new Error(humanReadableError);
  }),
}));

// Mock fetch globally
const mockFetch = jest.fn();
global.fetch = mockFetch;

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <AuthProvider>{children}</AuthProvider>
);

describe('AuthContext', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (getItem as jest.Mock).mockResolvedValue(null);
    (setItem as jest.Mock).mockResolvedValue(undefined);
    (multiRemove as jest.Mock).mockResolvedValue(undefined);
  });

  describe('signIn', () => {
    it('should handle network errors and show appropriate message', async () => {
      // Mock a network error (server unreachable)
      mockFetch.mockRejectedValueOnce(new TypeError('fetch failed'));

      const { result } = renderHook(() => useAuth(), { wrapper });

      // Wait for initial auth check to complete
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Attempt to sign in
      let error: Error | undefined;
      await act(async () => {
        try {
          await result.current.signIn('guest', '', '', 'Test User');
        } catch (err) {
          error = err as Error;
        }
      });

      // Verify the error message is user-friendly
      expect(error).toBeDefined();
      expect(error?.message).toBe('Ligain servers are not available for now. Please try again later.');
      expect(result.current.isLoading).toBe(false);
    }, 10000);

    it('should handle API errors properly', async () => {
      // Mock an API error response
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        statusText: 'Unauthorized',
        json: async () => ({ error: 'Invalid credentials' }),
      });

      const { result } = renderHook(() => useAuth(), { wrapper });

      // Wait for initial auth check to complete
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Attempt to sign in
      await act(async () => {
        try {
          await result.current.signIn('guest', '', '', 'Test User');
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.isLoading).toBe(false);
    });

    it('should handle successful sign in', async () => {
      // Mock a successful response
      const mockResponse = {
        token: 'test-token',
        player: {
          id: 'test-player-id',
          name: 'Test User',
          email: 'test@example.com',
          provider: 'guest',
        },
      };

      const mockResponseObj = {
        ok: true,
        status: 200,
        statusText: 'OK',
        headers: {
          entries: () => [],
        },
        json: async () => mockResponse,
      };
      mockFetch.mockResolvedValueOnce(mockResponseObj);

      const { result } = renderHook(() => useAuth(), { wrapper });

      // Wait for initial auth check to complete
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Sign in
      await act(async () => {
        await result.current.signIn('guest', '', '', 'Test User');
      });

      expect(result.current.player).toEqual(mockResponse.player);
      expect(result.current.isLoading).toBe(false);
      expect(setItem).toHaveBeenCalledWith('auth_token', mockResponse.token);
      expect(setItem).toHaveBeenCalledWith('player_data', JSON.stringify(mockResponse.player));
    });
  });

  describe('checkAuth', () => {
    it('should handle network errors during token validation gracefully', async () => {
      // Mock stored token and player data
      (getItem as jest.Mock)
        .mockResolvedValueOnce('stored-token')
        .mockResolvedValueOnce(JSON.stringify({ id: 'test-id', name: 'Test User' }));

      // Mock a network error during token validation
      mockFetch.mockRejectedValueOnce(new TypeError('fetch failed'));

      const { result } = renderHook(() => useAuth(), { wrapper });

      // Wait for auth check to complete
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Should not clear storage on network errors, just set player to null temporarily
      expect(result.current.player).toBeNull();
      expect(result.current.isLoading).toBe(false);
      expect(multiRemove).not.toHaveBeenCalled();
    });
  });

  describe('signOut', () => {
    it('should handle network errors during signout gracefully', async () => {
      // Mock stored token
      (getItem as jest.Mock).mockResolvedValueOnce('stored-token');

      // Mock a network error during signout
      mockFetch.mockRejectedValueOnce(new TypeError('fetch failed'));

      const { result } = renderHook(() => useAuth(), { wrapper });

      // Wait for initial auth check to complete
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Sign out
      await act(async () => {
        await result.current.signOut();
      });

      // Should still clear local storage regardless of network errors
      expect(result.current.player).toBeNull();
      expect(multiRemove).toHaveBeenCalledWith(['auth_token', 'player_data']);
    });
  });
}); 