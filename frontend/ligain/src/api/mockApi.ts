/**
 * Mock API Implementations
 *
 * These implementations return mock data for UI development.
 * They simulate API delays and maintain state in memory.
 */

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
import {
  MOCK_CURRENT_PLAYER,
  MOCK_GAMES,
  MOCK_AUTH_TOKEN,
  getMockMatchesForGame,
} from './mockData';
import { Game } from '../contexts/GamesContext';

// Simulates network delay for realistic UX
const simulateDelay = (ms: number = 300): Promise<void> =>
  new Promise((resolve) => setTimeout(resolve, ms));

// For testing error scenarios
let mockUploadError: AvatarErrorCode | null = null;

/**
 * Set a mock error for the next avatar upload (for testing)
 */
export function setMockUploadError(error: AvatarErrorCode | null): void {
  mockUploadError = error;
}

/**
 * Mock Auth API implementation
 * Auto-authenticates and returns mock player data
 */
export class MockAuthApi implements AuthApi {
  async checkAuth(): Promise<AuthCheckResponse | null> {
    await simulateDelay(200);
    console.log('[MockAuthApi] checkAuth - returning mock player');
    return { player: MOCK_CURRENT_PLAYER };
  }

  async signIn(
    provider: 'google' | 'apple',
    _token: string,
    _email: string,
    _name: string
  ): Promise<AuthSignInResponse> {
    await simulateDelay(500);
    console.log(`[MockAuthApi] signIn with ${provider} - returning mock response`);
    return {
      token: MOCK_AUTH_TOKEN,
      player: MOCK_CURRENT_PLAYER,
    };
  }

  async signInGuest(name: string): Promise<AuthSignInResponse> {
    await simulateDelay(500);
    console.log(`[MockAuthApi] signInGuest as "${name}" - returning mock response`);
    return {
      token: MOCK_AUTH_TOKEN,
      player: {
        ...MOCK_CURRENT_PLAYER,
        name,
        provider: 'guest',
      },
    };
  }

  async signOut(): Promise<void> {
    await simulateDelay(200);
    console.log('[MockAuthApi] signOut');
  }
}

/**
 * Mock Profile API implementation
 * Simulates profile operations with mock data
 */
export class MockProfileApi implements ProfileApi {
  async uploadAvatar(imageUri: string): Promise<UploadAvatarResponse> {
    await simulateDelay(800); // Simulate upload time
    console.log('[MockProfileApi] uploadAvatar:', imageUri);

    // Check for mock error
    if (mockUploadError) {
      const error = mockUploadError;
      mockUploadError = null; // Reset after use
      throw new AvatarError(error, `Mock error: ${error}`);
    }

    // Update mock player's avatar
    MOCK_CURRENT_PLAYER.avatarUrl = imageUri;

    return { avatarUrl: imageUri };
  }

  async deleteAvatar(): Promise<void> {
    await simulateDelay(300);
    console.log('[MockProfileApi] deleteAvatar');

    // Clear mock player's avatar
    MOCK_CURRENT_PLAYER.avatarUrl = null;
  }
}

/**
 * Mock Games API implementation
 * Returns mock games and matches, stores bets in memory
 */
export class MockGamesApi implements GamesApi {
  // In-memory state for session persistence
  private games: Game[] = [...MOCK_GAMES];
  private bets: Map<string, Map<string, { predictedHomeGoals: number; predictedAwayGoals: number }>> =
    new Map(); // gameId -> (matchId -> bet)
  private nextGameIndex: number = MOCK_GAMES.length + 1;

  async getGames(): Promise<GamesResponse> {
    await simulateDelay(300);
    console.log('[MockGamesApi] getGames - returning', this.games.length, 'games');
    return { games: this.games };
  }

  async getGameMatches(gameId: string): Promise<MatchesResponse> {
    await simulateDelay(200);
    console.log('[MockGamesApi] getGameMatches for', gameId);

    const baseMatches = getMockMatchesForGame(gameId);
    const gameBets = this.bets.get(gameId);

    // If we have session bets, merge them into the matches
    if (gameBets && gameBets.size > 0) {
      const updatedIncoming = { ...baseMatches.incomingMatches };
      for (const [matchId, bet] of gameBets) {
        if (updatedIncoming[matchId]) {
          updatedIncoming[matchId] = {
            ...updatedIncoming[matchId],
            bet: {
              prediction: `${bet.predictedHomeGoals}-${bet.predictedAwayGoals}`,
            },
          };
        }
      }
      return {
        incomingMatches: updatedIncoming,
        pastMatches: baseMatches.pastMatches,
      };
    }

    return baseMatches;
  }

  async createGame(name: string): Promise<CreateGameResponse> {
    await simulateDelay(400);
    console.log('[MockGamesApi] createGame:', name);

    const newGame: Game = {
      gameId: `game-${this.nextGameIndex++}`,
      name: name.trim(),
      seasonYear: '2025/2026',
      competitionName: 'Ligue 1',
      status: 'active',
      code: this.generateCode(),
      players: [MOCK_CURRENT_PLAYER],
    };

    this.games.push(newGame);
    return { game: newGame };
  }

  async joinGame(code: string): Promise<JoinGameResponse> {
    await simulateDelay(400);
    console.log('[MockGamesApi] joinGame with code:', code);

    const game = this.games.find((g) => g.code === code.toUpperCase());
    if (!game) {
      throw new Error('Game not found');
    }

    // Check if already a member
    if (game.players?.some((p) => p.id === MOCK_CURRENT_PLAYER.id)) {
      throw new Error('Already a member of this game');
    }

    // Add player to game
    game.players = [...(game.players || []), MOCK_CURRENT_PLAYER];
    return { game };
  }

  async placeBet(
    gameId: string,
    matchId: string,
    predictedHomeGoals: number,
    predictedAwayGoals: number
  ): Promise<BetResponse> {
    await simulateDelay(200);
    console.log('[MockGamesApi] placeBet:', { gameId, matchId, predictedHomeGoals, predictedAwayGoals });

    // Store bet in memory
    if (!this.bets.has(gameId)) {
      this.bets.set(gameId, new Map());
    }
    this.bets.get(gameId)!.set(matchId, { predictedHomeGoals, predictedAwayGoals });

    return {
      bet: { matchId, predictedHomeGoals, predictedAwayGoals },
    };
  }

  async leaveGame(gameId: string): Promise<void> {
    await simulateDelay(300);
    console.log('[MockGamesApi] leaveGame:', gameId);

    // Remove game from list
    this.games = this.games.filter((g) => g.gameId !== gameId);
    // Clear bets for this game
    this.bets.delete(gameId);
  }

  private generateCode(): string {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    let code = '';
    for (let i = 0; i < 4; i++) {
      code += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return code;
  }
}
