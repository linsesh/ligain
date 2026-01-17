/**
 * Real API Implementations
 *
 * These implementations make actual HTTP calls to the backend.
 * They wrap the existing fetch logic from the application.
 */

import { API_CONFIG, getApiHeaders, authenticatedFetch } from '../config/api';
import {
  AuthApi,
  GamesApi,
  AuthCheckResponse,
  AuthSignInResponse,
  GamesResponse,
  MatchesResponse,
  CreateGameResponse,
  JoinGameResponse,
  BetResponse,
} from './types';

/**
 * Real Auth API implementation
 * Makes HTTP calls to the backend for authentication operations
 */
export class RealAuthApi implements AuthApi {
  async checkAuth(): Promise<AuthCheckResponse | null> {
    try {
      const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/auth/me`);
      if (response.ok) {
        const data = await response.json();
        return { player: data.player };
      }
      return null;
    } catch (error) {
      console.error('RealAuthApi.checkAuth error:', error);
      return null;
    }
  }

  async signIn(
    provider: 'google' | 'apple',
    token: string,
    email: string,
    name: string
  ): Promise<AuthSignInResponse> {
    const response = await fetch(`${API_CONFIG.BASE_URL}/api/auth/signin`, {
      method: 'POST',
      headers: {
        ...getApiHeaders(),
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ provider, token, email, name }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  async signInGuest(name: string): Promise<AuthSignInResponse> {
    const response = await fetch(`${API_CONFIG.BASE_URL}/api/auth/signin/guest`, {
      method: 'POST',
      headers: {
        ...getApiHeaders(),
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ name }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
    }

    return await response.json();
  }

  async signOut(): Promise<void> {
    try {
      await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/auth/signout`, {
        method: 'POST',
      });
    } catch (error) {
      // Ignore network errors during signout
      if (!(error instanceof TypeError && (error as Error).message.includes('fetch'))) {
        throw error;
      }
    }
  }
}

/**
 * Real Games API implementation
 * Makes HTTP calls to the backend for game operations
 */
export class RealGamesApi implements GamesApi {
  async getGames(): Promise<GamesResponse> {
    const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games`);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to fetch games: ${response.status}`);
    }

    const data = await response.json();
    return { games: data.games || [] };
  }

  async getGameMatches(gameId: string): Promise<MatchesResponse> {
    const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/game/${gameId}/matches`);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to fetch matches: ${response.status}`);
    }

    return await response.json();
  }

  async createGame(name: string): Promise<CreateGameResponse> {
    const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        seasonYear: '2025/2026',
        competitionName: 'Ligue 1',
        name: name.trim(),
      }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to create game: ${response.status}`);
    }

    const data = await response.json();
    return { game: data.game };
  }

  async joinGame(code: string): Promise<JoinGameResponse> {
    const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games/join`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code: code.trim().toUpperCase() }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to join game: ${response.status}`);
    }

    const data = await response.json();
    return { game: data.game };
  }

  async placeBet(gameId: string, matchId: string, prediction: string): Promise<BetResponse> {
    const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/game/${gameId}/bet`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ matchId, prediction }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to place bet: ${response.status}`);
    }

    return await response.json();
  }

  async leaveGame(gameId: string): Promise<void> {
    const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/game/${gameId}/leave`, {
      method: 'POST',
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to leave game: ${response.status}`);
    }
  }
}
