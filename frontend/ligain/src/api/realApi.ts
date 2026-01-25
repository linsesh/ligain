/**
 * Real API Implementations
 *
 * These implementations make actual HTTP calls to the backend.
 * They wrap the existing fetch logic from the application.
 */

import { API_CONFIG, getApiHeaders, authenticatedFetch } from '../config/api';
import { Player } from '../contexts/AuthContext';
import {
  AuthApi,
  GamesApi,
  ProfileApi,
  AuthCheckResponse,
  AuthSignInResponse,
  GamesResponse,
  MatchesResponse,
  CreateGameResponse,
  JoinGameResponse,
  BetResponse,
  UploadAvatarResponse,
  AvatarError,
  AvatarErrorCode,
} from './types';

/**
 * Maps player data from backend snake_case to frontend camelCase
 */
function mapPlayerFromBackend(player: Record<string, unknown>): Player {
  return {
    id: player.id as string,
    name: player.name as string,
    email: player.email as string | undefined,
    provider: player.provider as string | undefined,
    provider_id: player.provider_id as string | undefined,
    created_at: player.created_at as string | undefined,
    updated_at: player.updated_at as string | undefined,
    avatarUrl: (player.avatar_url || player.avatar_signed_url || null) as string | null,
  };
}

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
        return { player: mapPlayerFromBackend(data.player) };
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

    const data = await response.json();
    return {
      ...data,
      player: data.player ? mapPlayerFromBackend(data.player) : undefined,
    };
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

    const data = await response.json();
    return {
      ...data,
      player: data.player ? mapPlayerFromBackend(data.player) : undefined,
    };
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

function createAvatarFormData(imageUri: string): FormData {
  const formData = new FormData();
  const uriParts = imageUri.split('/');
  const fileName = uriParts[uriParts.length - 1] || 'avatar.jpg';
  const extension = fileName.split('.').pop()?.toLowerCase() || 'jpg';
  const mimeType = extension === 'png' ? 'image/png' :
                   extension === 'webp' ? 'image/webp' : 'image/jpeg';

  formData.append('avatar', {
    uri: imageUri,
    type: mimeType,
    name: fileName,
  } as unknown as Blob);

  return formData;
}

/**
 * Real Profile API implementation
 * Makes HTTP calls to the backend for profile operations
 */
export class RealProfileApi implements ProfileApi {
  async uploadAvatar(imageUri: string): Promise<UploadAvatarResponse> {
    const response = await authenticatedFetch(
      `${API_CONFIG.BASE_URL}/api/players/me/avatar`,
      { method: 'POST', body: createAvatarFormData(imageUri) }
    );

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const errorCode = errorData.code as AvatarErrorCode;
      const validCodes = ['INVALID_IMAGE', 'FILE_TOO_LARGE', 'IMAGE_TOO_SMALL', 'UPLOAD_FAILED'];
      const code = validCodes.includes(errorCode) ? errorCode : 'UPLOAD_FAILED';
      throw new AvatarError(code, errorData.error || `Upload failed: ${code}`);
    }

    const data = await response.json();
    return { avatarUrl: data.avatar_url };
  }

  async deleteAvatar(): Promise<void> {
    const response = await authenticatedFetch(
      `${API_CONFIG.BASE_URL}/api/players/me/avatar`,
      { method: 'DELETE' }
    );

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `Failed to delete avatar: ${response.status}`);
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

  async placeBet(
    gameId: string,
    matchId: string,
    predictedHomeGoals: number,
    predictedAwayGoals: number
  ): Promise<BetResponse> {
    const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/game/${gameId}/bet`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ matchId, predictedHomeGoals, predictedAwayGoals }),
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
