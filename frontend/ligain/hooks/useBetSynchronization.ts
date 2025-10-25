import { useState, useEffect, useCallback, useRef } from 'react';
import { useGames } from '../src/contexts/GamesContext';
import { useAuth } from '../src/contexts/AuthContext';
import { useTimeService } from '../src/contexts/TimeServiceContext';
import { useMatches } from './useMatches';
import { API_CONFIG, getAuthenticatedHeaders } from '../src/config/api';
import { SeasonMatch } from '../src/types/match';

export interface SyncOpportunity {
  sourceGameId: string;
  sourceGameName: string;
  matchesToSync: Array<{
    matchId: string;
    homeTeam: string;
    awayTeam: string;
    matchday: number;
    predictedHomeGoals: number;
    predictedAwayGoals: number;
  }>;
}

export const useBetSynchronization = (currentGameId: string) => {
  const { games } = useGames();
  const { player } = useAuth();
  const timeService = useTimeService();
  const [syncOpportunity, setSyncOpportunity] = useState<SyncOpportunity | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Get current game's matches using the existing useMatches hook
  const { incomingMatches: currentIncomingMatches, pastMatches: currentPastMatches } = useMatches(currentGameId);

  const getFutureMatches = (incomingMatches: any, pastMatches: any) => {
    const allMatches = [...Object.values(incomingMatches), ...Object.values(pastMatches)];
    return allMatches.filter((matchResult: any) => {
      // Add null/undefined checks to prevent runtime errors
      if (!matchResult || !matchResult.match || typeof matchResult.match.getDate !== 'function') {
        console.warn('Invalid match result found:', matchResult);
        return false;
      }
      return new Date(matchResult.match.getDate()) > timeService.now();
    });
  };

  const findMatchingCurrentMatch = (
    otherMatchResult: any, 
    currentGameFutureMatches: any[]
  ) => {
    return currentGameFutureMatches.find((currentMatchResult: any) => {
      const currentMatch = currentMatchResult.match;
      const otherMatch = otherMatchResult.match;
      return currentMatch.getHomeTeam() === otherMatch.getHomeTeam() && 
             currentMatch.getAwayTeam() === otherMatch.getAwayTeam() && 
             currentMatch.getMatchday() === otherMatch.getMatchday();
    });
  };

  const canSyncMatch = (currentMatchResult: any, playerId: string) => {
    return currentMatchResult && !currentMatchResult.bets?.[playerId];
  };

  const createMatchToSync = (currentMatchResult: any, playerBet: any) => {
    const currentMatch = currentMatchResult.match;
    return {
      matchId: currentMatch.id(),
      homeTeam: currentMatch.getHomeTeam(),
      awayTeam: currentMatch.getAwayTeam(),
      matchday: currentMatch.getMatchday(),
      predictedHomeGoals: playerBet.predictedHomeGoals,
      predictedAwayGoals: playerBet.predictedAwayGoals
    };
  };

  const fetchOtherGameMatches = async (gameId: string) => {
    const headers = await getAuthenticatedHeaders();
    const response = await fetch(`${API_CONFIG.BASE_URL}/api/game/${gameId}/matches`, {
      headers,
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch matches for game ${gameId}: ${response.status}`);
    }

    const data = await response.json();
    const allMatches = [...Object.values(data.incomingMatches || {}), ...Object.values(data.pastMatches || {})];
    
    return allMatches.map((matchData: any) => {
      // Process the match data the same way useMatches does
      const match = SeasonMatch.fromJSON(matchData.match);
      
      const bets = matchData.bets ? Object.entries(matchData.bets).reduce((acc: { [key: string]: any }, [playerName, betData]: [string, any]) => {
        acc[betData.playerId] = {
          playerId: betData.playerId,
          playerName: betData.playerName,
          predictedHomeGoals: betData.predictedHomeGoals,
          predictedAwayGoals: betData.predictedAwayGoals,
        };
        return acc;
      }, {}) : {};
      
      return {
        match,
        bets
      };
    });
  };

  const findMatchesToSyncForGame = async (
    otherGame: any,
    currentGameFutureMatches: any[]
  ) => {
    try {
      const otherGameMatches = await fetchOtherGameMatches(otherGame.gameId);
      const otherGameFutureMatches = otherGameMatches.filter((matchResult: any) => {
        // Add null/undefined checks to prevent runtime errors
        if (!matchResult || !matchResult.match || typeof matchResult.match.getDate !== 'function') {
          console.warn('Invalid match result found in other game:', matchResult);
          return false;
        }
        return new Date(matchResult.match.getDate()) > timeService.now();
      });
      
      const matchesToSync: any[] = [];

      for (const otherMatchResult of otherGameFutureMatches) {
        if (!player?.id) continue;
        
        const playerBet = otherMatchResult.bets?.[player.id];
        if (!playerBet) continue;

        const currentMatchResult = findMatchingCurrentMatch(otherMatchResult, currentGameFutureMatches);
        
        if (canSyncMatch(currentMatchResult, player.id)) {
          matchesToSync.push(createMatchToSync(currentMatchResult, playerBet));
        }
      }

      return matchesToSync;
    } catch (err) {
      // Re-throw the error so it can be caught by the main function
      throw err;
    }
  };

  const validateSyncPrerequisites = () => {
    if (!player || !currentGameId || games.length === 0) {
      return { isValid: false, currentGame: null, otherGames: [] };
    }

    const currentGame = games.find(g => g.gameId === currentGameId);
    if (!currentGame) {
      return { isValid: false, currentGame: null, otherGames: [] };
    }

    const otherGames = games.filter(g => 
      g.gameId !== currentGameId && 
      g.seasonYear === currentGame.seasonYear && 
      g.competitionName === currentGame.competitionName
    );

    return { isValid: otherGames.length > 0, currentGame, otherGames };
  };

  const findBestSourceGame = async (
    otherGames: any[],
    currentGameFutureMatches: any[]
  ) => {
    let bestSourceGame: { game: any; matchesToSync: any[] } | null = null;

    for (const otherGame of otherGames) {
      const matchesToSync = await findMatchesToSyncForGame(otherGame, currentGameFutureMatches);
      
      if (matchesToSync.length > 0 && 
          (!bestSourceGame || matchesToSync.length > bestSourceGame.matchesToSync.length)) {
        bestSourceGame = {
          game: otherGame,
          matchesToSync
        };
      }
    }

    return bestSourceGame;
  };

  const findSyncOpportunity = useCallback(async () => {
    const { isValid, currentGame, otherGames } = validateSyncPrerequisites();
    
    if (!isValid) {
      setSyncOpportunity(null);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // Get current game's future matches
      const currentGameFutureMatches = getFutureMatches(currentIncomingMatches, currentPastMatches);

      // Find the best source game with most matches to sync
      const bestSourceGame = await findBestSourceGame(otherGames, currentGameFutureMatches);

      if (bestSourceGame) {
        setSyncOpportunity({
          sourceGameId: bestSourceGame.game.gameId,
          sourceGameName: bestSourceGame.game.name,
          matchesToSync: bestSourceGame.matchesToSync
        });
      } else {
        setSyncOpportunity(null);
      }
    } catch (err) {
      console.error('Error finding sync opportunity:', err);
      setError(err instanceof Error ? err.message : 'Failed to check for sync opportunities');
      setSyncOpportunity(null);
    } finally {
      setLoading(false);
    }
  }, [currentGameId, games, player, timeService, currentIncomingMatches, currentPastMatches]);

  // Use a ref to track if we've already run the sync check for the current game
  const hasRunSyncCheck = useRef(false);
  const lastGameId = useRef<string | null>(null);

  useEffect(() => {
    // Only run sync check when:
    // 1. Game ID changes (new game selected)
    // 2. We haven't run the check yet for this game
    // 3. Essential dependencies change (games, player, timeService)
    if (currentGameId !== lastGameId.current || !hasRunSyncCheck.current) {
      lastGameId.current = currentGameId;
      hasRunSyncCheck.current = true;
      findSyncOpportunity();
    }
  }, [currentGameId, games, player, timeService, findSyncOpportunity]);

  return {
    syncOpportunity,
    loading,
    error,
    refetch: findSyncOpportunity
  };
};
