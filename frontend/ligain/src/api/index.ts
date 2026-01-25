/**
 * API Module Exports
 *
 * Central export point for the API layer.
 * Import from '@api' or '../api' in the application.
 */

// Types
export type {
  Api,
  AuthApi,
  GamesApi,
  ProfileApi,
  AuthCheckResponse,
  AuthSignInResponse,
  SignInResult,
  MatchData,
  MatchesResponse,
  GamesResponse,
  CreateGameResponse,
  JoinGameResponse,
  BetResponse,
} from './types';

// Provider and hooks
export { ApiProvider, useApi, useAuthApi, useGamesApi, useProfileApi } from './ApiProvider';

// Real implementations (for direct use if needed)
export { RealAuthApi, RealGamesApi, RealProfileApi, mapPlayerFromBackend } from './realApi';

// Mock implementations (for testing)
export { MockAuthApi, MockGamesApi, MockProfileApi } from './mockApi';
