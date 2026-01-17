/**
 * API Interface Definitions
 *
 * These interfaces define the contract for all API operations.
 * Both real and mock implementations must conform to these interfaces.
 * The rest of the application depends only on these abstractions,
 * enabling dependency injection without knowing the concrete implementation.
 */

import { Player } from '../contexts/AuthContext';
import { Game } from '../contexts/GamesContext';

// Auth API Response Types
export interface AuthCheckResponse {
  player: Player;
}

// Avatar API Types
export interface UploadAvatarResponse {
  avatarUrl: string;
}

export type AvatarErrorCode = 'INVALID_IMAGE' | 'FILE_TOO_LARGE' | 'IMAGE_TOO_SMALL' | 'UPLOAD_FAILED';

export class AvatarError extends Error {
  code: AvatarErrorCode;

  constructor(code: AvatarErrorCode, message: string) {
    super(message);
    this.code = code;
    this.name = 'AvatarError';
  }
}

export interface AuthSignInResponse {
  token: string;
  player: Player;
  status?: 'need_display_name';
  suggestedName?: string;
  error?: string;
}

export interface SignInResult {
  needDisplayName?: boolean;
  suggestedName?: string;
  error?: string;
}

// Games API Response Types
export interface MatchData {
  match: {
    id: string;
    date: string;
    homeTeam: string;
    awayTeam: string;
    homeTeamOdds?: number;
    drawOdds?: number;
    awayTeamOdds?: number;
    matchday: number;
    status: string;
    homeGoals?: number;
    awayGoals?: number;
  };
  bet?: {
    prediction: string;
  };
  allBets?: {
    playerId: string;
    playerName: string;
    prediction: string;
    points?: number;
  }[];
}

export interface MatchesResponse {
  incomingMatches: Record<string, MatchData>;
  pastMatches: Record<string, MatchData>;
}

export interface GamesResponse {
  games: Game[];
}

export interface CreateGameResponse {
  game: Game;
}

export interface JoinGameResponse {
  game: Game;
}

export interface BetResponse {
  bet: {
    matchId: string;
    predictedHomeGoals: number;
    predictedAwayGoals: number;
  };
}

/**
 * Auth API Interface
 *
 * Handles authentication operations including:
 * - Checking current authentication state
 * - Signing in with OAuth providers or as a guest
 * - Signing out
 */
export interface AuthApi {
  /**
   * Check if the user is currently authenticated
   * @returns The authenticated player or null if not authenticated
   */
  checkAuth(): Promise<AuthCheckResponse | null>;

  /**
   * Sign in with an OAuth provider (Google/Apple)
   * @param provider - The OAuth provider
   * @param token - The OAuth token from the provider
   * @param email - User's email
   * @param name - User's display name
   */
  signIn(
    provider: 'google' | 'apple',
    token: string,
    email: string,
    name: string
  ): Promise<AuthSignInResponse>;

  /**
   * Sign in as a guest user
   * @param name - Display name for the guest
   */
  signInGuest(name: string): Promise<AuthSignInResponse>;

  /**
   * Sign out the current user
   */
  signOut(): Promise<void>;

  /**
   * Upload a new avatar image
   * @param imageUri - The local URI of the image to upload
   * @returns The uploaded avatar URL
   * @throws AvatarError if validation or upload fails
   */
  uploadAvatar(imageUri: string): Promise<UploadAvatarResponse>;

  /**
   * Delete the current user's avatar
   */
  deleteAvatar(): Promise<void>;
}

/**
 * Games API Interface
 *
 * Handles all game-related operations including:
 * - Fetching games the player belongs to
 * - Fetching matches for a game
 * - Creating and joining games
 * - Placing bets
 * - Leaving games
 */
export interface GamesApi {
  /**
   * Get all games the current player belongs to
   */
  getGames(): Promise<GamesResponse>;

  /**
   * Get all matches for a specific game
   * @param gameId - The game to fetch matches for
   */
  getGameMatches(gameId: string): Promise<MatchesResponse>;

  /**
   * Create a new game
   * @param name - Display name for the game
   */
  createGame(name: string): Promise<CreateGameResponse>;

  /**
   * Join an existing game by code
   * @param code - The game's join code
   */
  joinGame(code: string): Promise<JoinGameResponse>;

  /**
   * Place a bet on a match
   * @param gameId - The game the match belongs to
   * @param matchId - The match to bet on
   * @param predictedHomeGoals - The predicted home team goals
   * @param predictedAwayGoals - The predicted away team goals
   */
  placeBet(
    gameId: string,
    matchId: string,
    predictedHomeGoals: number,
    predictedAwayGoals: number
  ): Promise<BetResponse>;

  /**
   * Leave a game
   * @param gameId - The game to leave
   */
  leaveGame(gameId: string): Promise<void>;
}

/**
 * Combined API interface for dependency injection
 */
export interface Api {
  auth: AuthApi;
  games: GamesApi;
}
