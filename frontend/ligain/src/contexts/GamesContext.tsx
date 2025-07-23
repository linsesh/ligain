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
import { API_CONFIG, getAuthenticatedHeaders } from '../config/api';
import { useTimeService } from './TimeServiceContext';
import { useAuth } from './AuthContext';
import { Alert } from 'react-native';
import { useTranslation } from 'react-i18next';

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
  closestUnbetMatch?: {      // Next match where user hasn't placed a bet yet
    matchId: string;         // Match identifier
    date: Date;             // When the match starts
  };
  minUnplayedMatchday?: number; // Earliest matchday with unplayed matches
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
  const { player } = useAuth();
  const [games, setGames] = useState<GameWithMatchInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedGameId, setSelectedGameId] = useState<string | null>(null);
  const [bestGameId, setBestGameId] = useState<string | null>(null);

  // Main fetch function - gets games and analyzes matches to find the "best" one
  const fetchGames = useCallback(async () => {
    setLoading(true);
    try {
      const headers = await getAuthenticatedHeaders();
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/games`, { headers });
      
      if (!response.ok) throw new Error(`Failed to fetch games: ${response.status}`);
      
      const data = await response.json();
      const gamesData: Game[] = data.games || [];
      const gamesWithMatchInfo: GameWithMatchInfo[] = [];
      
      // For each game, fetch its matches to analyze betting opportunities
      for (const game of gamesData) {
        try {
          const matchesResponse = await fetch(`${API_CONFIG.BASE_URL}/api/game/${game.gameId}/matches`, { headers });
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
      
      // Determine which game should be shown first (most urgent betting opportunities)
      const bestGame = determineBestGame(gamesWithMatchInfo);
      setBestGameId(bestGame?.gameId || null);
      setSelectedGameId(bestGame?.gameId || null);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch games');
    } finally {
      setLoading(false);
    }
  }, [player, timeService]);

  useEffect(() => {
    fetchGames();
  }, [fetchGames]);

  // Analyzes a game's matches to find betting opportunities and upcoming matchdays
  const processGameWithMatches = async (game: Game, matchesData: any, player: any): Promise<GameWithMatchInfo> => {
    const now = timeService.now();
    let closestUnbetMatch: GameWithMatchInfo['closestUnbetMatch'] = undefined;
    let minUnplayedMatchday: number | undefined = undefined;
    
    const incomingMatches = matchesData.incomingMatches || {};
    
    for (const [matchId, matchData] of Object.entries(incomingMatches)) {
      const matchDataTyped = matchData as any;
      const match = matchDataTyped.match;
      const matchDate = new Date(match.date);
      
      // Check if player hasn't bet on this match yet
      const hasBet = matchDataTyped.bets && player && matchDataTyped.bets[player.id];
      
      if (!hasBet && matchDate > now) {
        // Find the closest unbet match (most urgent)
        if (!closestUnbetMatch || matchDate < closestUnbetMatch.date) {
          closestUnbetMatch = {
            matchId,
            date: matchDate,
          };
        }
      }
      
      // Track the earliest unplayed matchday
      if (matchDate > now) {
        if (minUnplayedMatchday === undefined || match.matchday < minUnplayedMatchday) {
          minUnplayedMatchday = match.matchday;
        }
      }
    }
    
    return {
      ...game,
      closestUnbetMatch,
      minUnplayedMatchday,
    };
  };

  // Determines the "best" game to show first based on betting urgency
  const determineBestGame = (games: GameWithMatchInfo[]): GameWithMatchInfo | null => {
    if (games.length === 0) return null;
    if (games.length === 1) return games[0];
    
    // Priority: 1) Closest unbet match, 2) Earliest unplayed matchday, 3) Alphabetical
    const sortedGames = [...games].sort((a, b) => {
      // Games with unbet matches come first
      if (a.closestUnbetMatch && !b.closestUnbetMatch) return -1;
      if (!a.closestUnbetMatch && b.closestUnbetMatch) return 1;
      
      // Among games with unbet matches, prioritize by closest date
      if (a.closestUnbetMatch && b.closestUnbetMatch) {
        return a.closestUnbetMatch.date.getTime() - b.closestUnbetMatch.date.getTime();
      }
      
      // Among games without unbet matches, prioritize by earliest matchday
      if (a.minUnplayedMatchday !== undefined && b.minUnplayedMatchday === undefined) return -1;
      if (a.minUnplayedMatchday === undefined && b.minUnplayedMatchday !== undefined) return 1;
      if (a.minUnplayedMatchday !== undefined && b.minUnplayedMatchday !== undefined) {
        return a.minUnplayedMatchday - b.minUnplayedMatchday;
      }
      
      // Fallback to alphabetical order
      return a.name.localeCompare(b.name);
    });
    
    return sortedGames[0];
  };

  // Joins a game by code - handles API call and refreshes games list
  const joinGame = async (code: string) => {
    try {
      const headers = await getAuthenticatedHeaders({ 'Content-Type': 'application/json' });
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/games/join`, {
        method: 'POST',
        headers,
        body: JSON.stringify({ code: code.trim().toUpperCase() }),
      });
      
      if (!response.ok) throw new Error('Failed to join game');
      
      const data = await response.json();
      Alert.alert(t('common.success'), data.message);
      await fetchGames(); // Refresh to show new game
    } catch (err) {
      Alert.alert(t('common.error'), err instanceof Error ? err.message : t('games.failedToJoinGame'));
    }
  };

  // Creates a new game - handles API call and refreshes games list
  const createGame = async (name: string) => {
    try {
      const headers = await getAuthenticatedHeaders({ 'Content-Type': 'application/json' });
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/games`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          seasonYear: '2025/2026',
          competitionName: 'Ligue 1',
          name: name.trim(),
        }),
      });
      
      if (!response.ok) throw new Error('Failed to create game');
      
      await fetchGames(); // Refresh to show new game
    } catch (err) {
      Alert.alert(t('common.error'), err instanceof Error ? err.message : t('games.failedToCreateGame'));
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