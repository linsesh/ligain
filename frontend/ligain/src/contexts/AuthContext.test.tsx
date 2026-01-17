import React from 'react';
import { renderHook, act } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import { ApiProvider } from '../api';
import { getItem, setItem, multiRemove } from '../utils/storage';

// Mock the API module to provide mock implementations
jest.mock('../api/ApiProvider', () => {
  const React = require('react');
  const mockAuthApi = {
    checkAuth: jest.fn(),
    signIn: jest.fn(),
    signInGuest: jest.fn(),
    signOut: jest.fn(),
  };

  const mockGamesApi = {
    getGames: jest.fn(),
    getGameMatches: jest.fn(),
    createGame: jest.fn(),
    joinGame: jest.fn(),
    placeBet: jest.fn(),
    leaveGame: jest.fn(),
  };

  const mockProfileApi = {
    uploadAvatar: jest.fn(),
    deleteAvatar: jest.fn(),
  };

  const ApiContext = React.createContext({ auth: mockAuthApi, games: mockGamesApi, profile: mockProfileApi });

  return {
    ApiProvider: ({ children }: { children: React.ReactNode }) => (
      React.createElement(ApiContext.Provider, { value: { auth: mockAuthApi, games: mockGamesApi, profile: mockProfileApi } }, children)
    ),
    useApi: () => React.useContext(ApiContext),
    useAuthApi: () => React.useContext(ApiContext).auth,
    useGamesApi: () => React.useContext(ApiContext).games,
    useProfileApi: () => React.useContext(ApiContext).profile,
    __mockAuthApi: mockAuthApi,
    __mockGamesApi: mockGamesApi,
    __mockProfileApi: mockProfileApi,
  };
});

jest.mock('../utils/storage', () => ({
  getItem: jest.fn(),
  setItem: jest.fn(),
  multiRemove: jest.fn(),
  isUsingMemoryFallback: jest.fn(() => false),
}));

// Get the mocked auth API for test assertions
const getMockAuthApi = () => {
  return require('../api/ApiProvider').__mockAuthApi;
};

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <ApiProvider>
    <AuthProvider>{children}</AuthProvider>
  </ApiProvider>
);

describe('AuthContext', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (getItem as jest.Mock).mockResolvedValue(null);
    (setItem as jest.Mock).mockResolvedValue(undefined);
    (multiRemove as jest.Mock).mockResolvedValue(undefined);

    // Reset mock API defaults
    const mockAuthApi = getMockAuthApi();
    mockAuthApi.checkAuth.mockResolvedValue(null);
    mockAuthApi.signIn.mockResolvedValue({ token: 'test-token', player: { id: '1', name: 'Test' } });
    mockAuthApi.signInGuest.mockResolvedValue({ token: 'test-token', player: { id: '1', name: 'Test' } });
    mockAuthApi.signOut.mockResolvedValue(undefined);
  });

  describe('signIn', () => {
    it('should handle network errors and show appropriate message', async () => {
      const mockAuthApi = getMockAuthApi();
      // Mock a network error (server unreachable)
      mockAuthApi.signInGuest.mockRejectedValueOnce(new TypeError('fetch failed'));

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
      const mockAuthApi = getMockAuthApi();
      // Mock an API error response
      mockAuthApi.signInGuest.mockRejectedValueOnce(new Error('Invalid credentials'));

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
      const mockAuthApi = getMockAuthApi();
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

      mockAuthApi.signInGuest.mockResolvedValueOnce(mockResponse);

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
      const mockAuthApi = getMockAuthApi();
      // Mock stored token and player data
      (getItem as jest.Mock)
        .mockResolvedValueOnce('stored-token')
        .mockResolvedValueOnce(JSON.stringify({ id: 'test-id', name: 'Test User' }));

      // Mock a network error during token validation
      mockAuthApi.checkAuth.mockRejectedValueOnce(new TypeError('fetch failed'));

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
      const mockAuthApi = getMockAuthApi();
      // Mock stored token
      (getItem as jest.Mock).mockResolvedValueOnce('stored-token');

      // Mock a network error during signout
      mockAuthApi.signOut.mockRejectedValueOnce(new TypeError('fetch failed'));

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
