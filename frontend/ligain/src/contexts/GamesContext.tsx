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

  // Main fetch function - gets games and analyzes matches to find the "best" one
  const fetchGames = useCallback(async () => {
    setLoading(true);
    let didRetry = false;
    let lastError: any = null;
    for (let attempt = 0; attempt < 2; attempt++) {
      try {
        const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games`);
        if (response.status === 401) {
          let errorMsg = '';
          try {
            const errorData = await response.json();
            errorMsg = errorData.error || '';
          } catch (e) {
            errorMsg = '';
          }
          // API key errors
          if (errorMsg === 'API key is required' || errorMsg === 'Invalid API key') {
            setError(t('common.error'));
            setLoading(false);
            return;
          }
          // Token expired: try silent re-auth and retry
          if ((errorMsg === 'Invalid or expired token' || errorMsg.toLowerCase().includes('expired'))) {
            if (!didRetry) {
              await checkAuth();
              didRetry = true;
              continue;
            }
          }
          // Malformed/missing header or player not found: force sign-out and redirect
          if (
            errorMsg === 'Invalid authorization header format' ||
            errorMsg === 'Player not found in context'
          ) {
            await signOut();
            router.replace('/signin');
            setLoading(false);
            return;
          }
          // Fallback: treat as generic error
          setError(t('common.ligainNotAvailable'));
          setLoading(false);
          return;
        }
        if (!response.ok) throw new Error(`Failed to fetch games: ${response.status}`);
        const data = await response.json();
        const gamesData: Game[] = data.games || [];
        const gamesWithMatchInfo: GameWithMatchInfo[] = [];
        for (const game of gamesData) {
          try {
            const matchesResponse = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/game/${game.gameId}/matches`);
            if (matchesResponse.ok) {
              const matchesData = await matchesResponse.json();
              const gameWithInfo = await processGameWithMatches(game, matchesData, player);
              gamesWithMatchInfo.push(gameWithInfo);
            } else {
              gamesWithMatchInfo.push(game);
            }
          } catch (err) {
            gamesWithMatchInfo.push(game);
          }
        }
        setGames(gamesWithMatchInfo);
        const bestGame = determineBestGame(gamesWithMatchInfo);
        setBestGameId(bestGame?.gameId || null);
        setSelectedGameId(bestGame?.gameId || null);
        setError(null);
        setLoading(false);
        return;
      } catch (err) {
        console.log('🔍 GamesContext - Caught error:', err);
        console.log('🔍 GamesContext - Error type:', err instanceof Error ? err.constructor.name : typeof err);
        console.log('🔍 GamesContext - Error message:', err instanceof Error ? err.message : String(err));
        
        lastError = err;
        // Only retry on first 401 with expired token, otherwise break
        if (!(err instanceof Error && err.message.includes('401')) || didRetry) {
          break;
        }
      }
    }
    setError(translateError(lastError instanceof Error ? lastError.message : 'Failed to fetch games'));
    setLoading(false);
  }, [player, timeService, checkAuth, signOut, router]);

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

    return {
      ...game,
      closestUnfinishedMatchday,
    };
  };

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
    try {
      const response = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games/join`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: code.trim().toUpperCase() }),
      });
      
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
      
      await fetchGames(); // Refresh to show new game
    } catch (err) {
      Alert.alert(t('common.error'), err instanceof Error ? err.message : t('errors.failedToJoinGame'));
    }
  };

  // Creates a new game - handles API call and refreshes games list
  const createGame = async (name: string) => {
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
      
      await fetchGames(); // Refresh to show new game
    } catch (err) {
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