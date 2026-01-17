/**
 * Shared API mock setup for tests
 *
 * Use this in test files that need to mock the ApiProvider:
 *
 * ```typescript
 * import { createApiMock } from '../src/test-utils/apiMock';
 * jest.mock('../src/api/ApiProvider', () => createApiMock());
 * ```
 */

import React from 'react';

export const createMockAuthApi = () => ({
  checkAuth: jest.fn().mockResolvedValue(null),
  signIn: jest.fn().mockResolvedValue({ token: 'test-token', player: { id: '1', name: 'Test' } }),
  signInGuest: jest.fn().mockResolvedValue({ token: 'test-token', player: { id: '1', name: 'Test' } }),
  signOut: jest.fn().mockResolvedValue(undefined),
});

export const createMockGamesApi = () => ({
  getGames: jest.fn().mockResolvedValue({ games: [] }),
  getGameMatches: jest.fn().mockResolvedValue({ incomingMatches: {}, pastMatches: {} }),
  createGame: jest.fn().mockResolvedValue({ game: { gameId: 'new-game' } }),
  joinGame: jest.fn().mockResolvedValue({ game: { gameId: 'joined-game' } }),
  placeBet: jest.fn().mockResolvedValue({ bet: { matchId: 'match-1', prediction: 'home' } }),
  leaveGame: jest.fn().mockResolvedValue(undefined),
});

export const createMockProfileApi = () => ({
  uploadAvatar: jest.fn().mockResolvedValue({ avatarUrl: 'https://example.com/avatar.jpg' }),
  deleteAvatar: jest.fn().mockResolvedValue(undefined),
});

export const createApiMock = () => {
  const mockAuthApi = createMockAuthApi();
  const mockGamesApi = createMockGamesApi();
  const mockProfileApi = createMockProfileApi();

  const ApiContext = React.createContext({ auth: mockAuthApi, games: mockGamesApi, profile: mockProfileApi });

  return {
    ApiProvider: ({ children }: { children: React.ReactNode }) =>
      React.createElement(ApiContext.Provider, { value: { auth: mockAuthApi, games: mockGamesApi, profile: mockProfileApi } }, children),
    useApi: () => React.useContext(ApiContext),
    useAuthApi: () => React.useContext(ApiContext).auth,
    useGamesApi: () => React.useContext(ApiContext).games,
    useProfileApi: () => React.useContext(ApiContext).profile,
    __mockAuthApi: mockAuthApi,
    __mockGamesApi: mockGamesApi,
    __mockProfileApi: mockProfileApi,
  };
};
