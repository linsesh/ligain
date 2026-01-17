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
    homeOdds?: number;
    drawOdds?: number;
    awayOdds?: number;
    matchday: number;
    status: string;
    homeScore?: number;
    awayScore?: number;
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
    prediction: string;
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
   * @param prediction - The prediction (home, draw, away)
   */
  placeBet(gameId: string, matchId: string, prediction: string): Promise<BetResponse>;

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
