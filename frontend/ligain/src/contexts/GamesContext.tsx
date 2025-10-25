/**
 * GamesContext - Centralized Game State Management
 * 
 * WHY THIS FILE EXISTS:
 * Before this context, each screen (matches, games, game overview) was independently 
 * fetching games from the API, causing:
 * - UI flicker: "no games" briefly shown while loading
 * - Inconsistent state: different screens showing different data
 * - Multiple API calls: wasteful network requests
 * - Poor UX: users saw loading states repeatedly
 * 
 * WHAT IT SOLVES:
 * - Centralizes all game data in one shared context
 * - Fetches games once on app start and keeps them in memory
 * - Provides consistent loading states across all screens
 * - Auto-determines the "best" game to show (closest unbet match)
 * - Handles game creation/joining with immediate UI updates
 * - Eliminates the "no games" flicker by showing spinner during loads
 * 
 * HOW IT WORKS:
 * - Provider wraps the app and maintains global games state
 * - useGames() hook gives any component access to games data
 * - Games are fetched once and cached until manual refresh
 * - Smart game selection prioritizes games with upcoming bets
 */

import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import { API_CONFIG, getAuthenticatedHeaders, authenticatedFetch } from '../config/api';
import { useTimeService } from './TimeServiceContext';
import { useAuth } from './AuthContext';
import { useRouter } from 'expo-router';
import { Alert } from 'react-native';
import { useTranslation } from 'react-i18next';
import { handleGameError, translateError } from '../utils/errorMessages';
import { computeMonthlyAndMatchdayScores, computeTotalScores, AggregatedScore } from '../utils/aggregations';

// Basic game data structure returned from the API
// Contains core game information without any derived/calculated fields
export interface Game {
  gameId: string;          // Unique identifier for the game
  seasonYear: string;      // Season like "2024-2025"
  competitionName: string; // Competition like "Ligue 1"
  name: string;           // Display name for the game
  status: string;         // Game status (active, finished, etc.)
  players?: any[];        // List of players in this game
  code?: string;          // Optional game code for joining
}

// Game enhanced with match analysis data for UI decision-making
// Extends base Game with calculated fields to determine the "best" game to show
export interface GameWithMatchInfo extends Game {

  closestUnfinishedMatchday?: {
    matchday: number;
    date: Date;
  };
  // Aggregated statistics prepared for visuals
  perMonthLeaderboard?: Record<string, { PlayerID: string; PlayerName: string; Points: number }[]>;
  perMatchdayLeaderboard?: Record<number, { PlayerID: string; PlayerName: string; Points: number }[]>;
  totalLeaderboard?: AggregatedScore[];
}

// Context API interface that defines what the GamesContext provides to components
// Centralizes all game-related state and operations in one place
interface GamesContextType {
  games: GameWithMatchInfo[];               // All games user has access to
  loading: boolean;                         // Whether games are currently being fetched
  error: string | null;                     // Any error from the last operation
  selectedGameId: string | null;            // Currently selected game ID (user choice)
  setSelectedGameId: (id: string | null) => void; // Function to change selected game
  bestGameId: string | null;                // Auto-determined "best" game to show by default
  refresh: () => void;                      // Manually refresh games from server
  joinGame: (code: string) => Promise<void>; // Join existing game by code
  createGame: (name: string) => Promise<void>; // Create new game
  removeGame: (gameId: string) => void;     // Remove game from local state (after leaving)
}

const GamesContext = createContext<GamesContextType | undefined>(undefined);

// Hook to access games context - throws error if used outside provider
export const useGames = () => {
  const context = useContext(GamesContext);
  if (!context) throw new Error('useGames must be used within a GamesProvider');
  return context;
};

export const GamesProvider = ({ children }: { children: React.ReactNode }) => {
  const { t } = useTranslation();
  const timeService = useTimeService();
  const { player, checkAuth, signOut } = useAuth();
  const [games, setGames] = useState<GameWithMatchInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedGameId, setSelectedGameId] = useState<string | null>(null);
  const [bestGameId, setBestGameId] = useState<string | null>(null);
  const router = useRouter();
  const [isMounted, setIsMounted] = useState(true);

  // Track if component is mounted to prevent state updates after unmount
  useEffect(() => {
    setIsMounted(true);
    return () => setIsMounted(false);
  }, []);

  // Small helper: handle 401 cases consistently
  const handle401 = useCallback(async (response: Response) => {
    let errorMsg = '';
    try {
      const errorData = await response.json();
      errorMsg = errorData.error || '';
    } catch (e) {
      errorMsg = '';
    }
    if (errorMsg === 'API key is required' || errorMsg === 'Invalid API key') {
      setError(t('common.error'));
      setLoading(false);
      return { handled: true, retry: false } as const;
    }
    if ((errorMsg === 'Invalid or expired token' || errorMsg.toLowerCase().includes('expired'))) {
      return { handled: false, retry: true } as const;
    }
    if (
      errorMsg === 'Invalid authorization header format' ||
      errorMsg === 'Player not found in context'
    ) {
      await signOut();
      router.replace('/signin');
      setLoading(false);
      return { handled: true, retry: false } as const;
    }
    setError(t('common.ligainNotAvailable'));
    setLoading(false);
    return { handled: true, retry: false } as const;
  }, [router, signOut, t]);

  // Fetch list of games for the player
  const fetchPlayerGames = useCallback(async (): Promise<Game[]> => {
    const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games`);
    if (response.status === 401) {
      const res = await handle401(response);
      if (res.retry) throw new Error('401-retry');
      if (res.handled) throw new Error('401-handled');
    }
    if (!response.ok) throw new Error(`Failed to fetch games: ${response.status}`);
    const data = await response.json();
    return (data.games || []) as Game[];
  }, [handle401]);

  // enrichGameWithMatches declared after processGameWithMatches to avoid TDZ issues
  let enrichGameWithMatches: (game: Game) => Promise<GameWithMatchInfo>;

  // Main fetch function - gets games and analyzes matches to find the "best" one
  const fetchGames = useCallback(async () => {
    if (!isMounted) return;

    setLoading(true);
    let didRetry = false;
    let lastError: any = null;
    for (let attempt = 0; attempt < 2; attempt++) {
      try {
        if (!isMounted) return; // Check again before async operations

        const gamesData: Game[] = await fetchPlayerGames();
        if (!isMounted) return; // Check after async operation
        
        const gamesWithMatchInfo: GameWithMatchInfo[] = [];
        for (const game of gamesData) {
          if (!isMounted) return; // Check in loop
          const enriched = await enrichGameWithMatches(game);
          gamesWithMatchInfo.push(enriched);
        }
        
        if (!isMounted) return; // Final check before state updates
        
        setGames(gamesWithMatchInfo);
        const bestGame = determineBestGame(gamesWithMatchInfo);
        setBestGameId(bestGame?.gameId || null);
        setSelectedGameId(bestGame?.gameId || null);
        setError(null);
        setLoading(false);
        return;
      } catch (err) {
        if (!isMounted) return; // Check after error
        
        console.log('🔍 GamesContext - Caught error:', err);
        console.log('🔍 GamesContext - Error type:', err instanceof Error ? err.constructor.name : typeof err);
        console.log('🔍 GamesContext - Error message:', err instanceof Error ? err.message : String(err));
        
        lastError = err;
        // Attempt silent reauth when flagged for retry
        if (err instanceof Error && err.message === '401-retry' && !didRetry) {
          await checkAuth();
          didRetry = true;
          continue;
        }
        // Only retry once; otherwise break
        if (!(err instanceof Error && err.message.includes('401')) || didRetry) {
          break;
        }
      }
    }
    
    if (!isMounted) return; // Check before final state updates
    
    setError(translateError(lastError instanceof Error ? lastError.message : 'Failed to fetch games'));
    setLoading(false);
  }, [player, timeService, checkAuth, fetchPlayerGames, isMounted]);

  useEffect(() => {
    if (player) {
      fetchGames();
    }
  }, [fetchGames, player]);

  // Analyzes a game's matches to find betting opportunities and upcoming matchdays
  const processGameWithMatches = async (game: Game, matchesData: any, player: any): Promise<GameWithMatchInfo> => {
    const now = timeService.now();
    // Track soonest unfinished match per matchday
    const unfinishedMatchdays: Record<number, Date> = {};
    let closestUnfinishedMatchday: { matchday: number; date: Date } | undefined = undefined;

    const incomingMatches = matchesData.incomingMatches || {};
    const pastMatches = matchesData.pastMatches || {};

    for (const [matchId, matchData] of Object.entries(incomingMatches)) {
      const matchDataTyped = matchData as any;
      const match = matchDataTyped.match;
      const matchDate = new Date(match.date);
      const isFinished = match.status === 'finished' || match.status === 'complete';
      // Track soonest unfinished match per matchday
      if (!isFinished && matchDate > now) {
        if (
          !(match.matchday in unfinishedMatchdays) ||
          matchDate < unfinishedMatchdays[match.matchday]
        ) {
          unfinishedMatchdays[match.matchday] = matchDate;
        }
      }
    }

    // Find the matchday with the soonest unfinished match in the future
    for (const [matchdayStr, date] of Object.entries(unfinishedMatchdays)) {
      const matchday = Number(matchdayStr);
      if (!closestUnfinishedMatchday || date < closestUnfinishedMatchday.date) {
        closestUnfinishedMatchday = { matchday, date };
      }
    }

    const { perMonthLeaderboard, perMatchdayLeaderboard } = computeMonthlyAndMatchdayScores(pastMatches);
    const totalLeaderboard = computeTotalScores(pastMatches);

    return {
      ...game,
      closestUnfinishedMatchday,
      perMonthLeaderboard,
      perMatchdayLeaderboard,
      totalLeaderboard,
    };
  };

  // Now that processGameWithMatches is defined, bind enrichGameWithMatches
  enrichGameWithMatches = useCallback(async (game: Game): Promise<GameWithMatchInfo> => {
    if (!isMounted) {
      return {
        ...game,
        perMonthLeaderboard: {},
        perMatchdayLeaderboard: {},
        totalLeaderboard: [],
      } as GameWithMatchInfo;
    }
    
    try {
      const matchesResponse = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/game/${game.gameId}/matches`);
      if (!isMounted) return {
        ...game,
        perMonthLeaderboard: {},
        perMatchdayLeaderboard: {},
        totalLeaderboard: [],
      } as GameWithMatchInfo;
      
      if (!matchesResponse.ok) {
        return {
          ...game,
          perMonthLeaderboard: {},
          perMatchdayLeaderboard: {},
          totalLeaderboard: [],
        } as GameWithMatchInfo;
      }
      const matchesData = await matchesResponse.json();
      if (!isMounted) return {
        ...game,
        perMonthLeaderboard: {},
        perMatchdayLeaderboard: {},
        totalLeaderboard: [],
      } as GameWithMatchInfo;
      
      return await processGameWithMatches(game, matchesData, player);
    } catch {
      return {
        ...game,
        perMonthLeaderboard: {},
        perMatchdayLeaderboard: {},
        totalLeaderboard: [],
      } as GameWithMatchInfo;
    }
  }, [player, isMounted]);

  // Determines the "best" game to show first based on betting urgency
  const determineBestGame = (games: GameWithMatchInfo[]): GameWithMatchInfo | null => {
    if (games.length === 0) return null;
    if (games.length === 1) return games[0];

    // Priority: 1) Closest unfinished matchday, 2) Alphabetical
    const sortedGames = [...games].sort((a, b) => {
      // 1. Closest unfinished matchday
      if (a.closestUnfinishedMatchday && !b.closestUnfinishedMatchday) return -1;
      if (!a.closestUnfinishedMatchday && b.closestUnfinishedMatchday) return 1;
      if (a.closestUnfinishedMatchday && b.closestUnfinishedMatchday) {
        return a.closestUnfinishedMatchday.date.getTime() - b.closestUnfinishedMatchday.date.getTime();
      }
      // 2. Fallback to alphabetical order
      return a.name.localeCompare(b.name);
    });

    return sortedGames[0];
  };

  // Joins a game by code - handles API call and refreshes games list
  const joinGame = async (code: string) => {
    if (!isMounted) return;
    
    try {
      const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: code.trim().toUpperCase() }),
      });
      
      if (!isMounted) return; // Check after async operation
      
      if (!response.ok) {
        let errorMessage = '';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || '';
        } catch (e) {
          errorMessage = '';
        }
        
        // Use the game error handling utility
        const { title, message } = handleGameError(errorMessage);
        throw new Error(message);
      }
      
      if (!isMounted) return; // Check before refresh
      await fetchGames(); // Refresh to show new game
    } catch (err) {
      if (!isMounted) return; // Check before showing alert
      Alert.alert(t('common.error'), err instanceof Error ? err.message : t('errors.failedToJoinGame'));
    }
  };

  // Creates a new game - handles API call and refreshes games list
  const createGame = async (name: string) => {
    if (!isMounted) return;
    
    try {
      const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          seasonYear: '2025/2026',
          competitionName: 'Ligue 1',
          name: name.trim(),
        }),
      });
      
      if (!isMounted) return; // Check after async operation
      
      if (!response.ok) {
        let errorMessage = '';
        try {
          const errorData = await response.json();
          errorMessage = errorData.error || '';
        } catch (e) {
          errorMessage = '';
        }
        
        // Use the game error handling utility
        const { title, message } = handleGameError(errorMessage);
        throw new Error(message);
      }
      
      if (!isMounted) return; // Check before refresh
      await fetchGames(); // Refresh to show new game
    } catch (err) {
      if (!isMounted) return; // Check before showing alert
      Alert.alert(t('common.error'), err instanceof Error ? err.message : t('errors.failedToCreateGame'));
    }
  };

  // Removes a game from the local state by gameId
  const removeGame = (gameId: string) => {
    setGames(prevGames => prevGames.filter(game => game.gameId !== gameId));
    if (selectedGameId === gameId) {
      setSelectedGameId(null);
    }
    if (bestGameId === gameId) {
      setBestGameId(null);
    }
  };

  return (
    <GamesContext.Provider
      value={{
        games,
        loading,
        error,
        selectedGameId,
        setSelectedGameId,
        bestGameId,
        refresh: fetchGames,
        joinGame,
        createGame,
        removeGame,
      }}
    >
      {children}
    </GamesContext.Provider>
  );
}; 