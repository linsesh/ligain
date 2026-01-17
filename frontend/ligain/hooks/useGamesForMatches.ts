import { useState, useEffect } from 'react';
import { useGamesApi } from '../src/api';
import { useTimeService } from '../src/contexts/TimeServiceContext';
import { useAuth } from '../src/contexts/AuthContext';
import { translateError } from '../src/utils/errorMessages';

interface Game {
  gameId: string;
  seasonYear: string;
  competitionName: string;
  name: string;
  status: string;
  players?: any[];
}

interface GameWithMatchInfo extends Game {
  closestUnbetMatch?: {
    matchId: string;
    date: Date;
    matchday: number;
    homeTeam: string;
    awayTeam: string;
  };
  minUnplayedMatchday?: number;
}

export const useGamesForMatches = () => {
  const gamesApi = useGamesApi();
  const timeService = useTimeService();
  const { player } = useAuth();
  const [games, setGames] = useState<GameWithMatchInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedGameId, setSelectedGameId] = useState<string | null>(null);
  const [bestGameId, setBestGameId] = useState<string | null>(null);

  const fetchGames = async () => {
    try {
      const data = await gamesApi.getGames();
      const gamesData: Game[] = data.games || [];

      // Process each game to find the best one
      const gamesWithMatchInfo: GameWithMatchInfo[] = [];

      for (const game of gamesData) {
        try {
          // Fetch matches for this game
          const matchesData = await gamesApi.getGameMatches(game.gameId);
          const gameWithInfo = await processGameWithMatches(game, matchesData, player);
          gamesWithMatchInfo.push(gameWithInfo);
        } catch (err) {
          console.warn(`Failed to fetch matches for game ${game.gameId}:`, err);
          gamesWithMatchInfo.push(game);
        }
      }
      
      setGames(gamesWithMatchInfo);
      
      // Determine the best game to show
      const bestGame = determineBestGame(gamesWithMatchInfo);
      setBestGameId(bestGame?.gameId || null);
      setSelectedGameId(bestGame?.gameId || null);
      
      setError(null);
    } catch (err) {
      setError(translateError(err instanceof Error ? err.message : 'Failed to fetch games'));
    } finally {
      setLoading(false);
    }
  };

  const processGameWithMatches = async (game: Game, matchesData: any, player: any): Promise<GameWithMatchInfo> => {
    const now = timeService.now();
    
    let closestUnbetMatch: GameWithMatchInfo['closestUnbetMatch'] = undefined;
    let minUnplayedMatchday: number | undefined = undefined;
    
    // Process incoming matches to find closest unbet match
    const incomingMatches = matchesData.incomingMatches || {};
    for (const [matchId, matchData] of Object.entries(incomingMatches)) {
      const matchDataTyped = matchData as any;
      const match = matchDataTyped.match;
      const matchDate = new Date(match.date);
      
      // Check if player hasn't bet on this match
      const hasBet = matchDataTyped.bets && player && matchDataTyped.bets[player.id];
      
      if (!hasBet && matchDate > now) {
        if (!closestUnbetMatch || matchDate < closestUnbetMatch.date) {
          closestUnbetMatch = {
            matchId,
            date: matchDate,
            matchday: match.matchday,
            homeTeam: match.homeTeam,
            awayTeam: match.awayTeam,
          };
        }
      }
      
      // Track minimum unplayed matchday
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

  const determineBestGame = (games: GameWithMatchInfo[]): GameWithMatchInfo | null => {
    if (games.length === 0) return null;
    if (games.length === 1) return games[0];
    
    // Sort games by priority:
    // 1. Games with closest unbet match (closest date first)
    // 2. Games with unplayed matchdays (lowest matchday first)
    // 3. Alphabetically by name
    
    const sortedGames = [...games].sort((a, b) => {
      // First priority: games with closest unbet match
      if (a.closestUnbetMatch && !b.closestUnbetMatch) return -1;
      if (!a.closestUnbetMatch && b.closestUnbetMatch) return 1;
      if (a.closestUnbetMatch && b.closestUnbetMatch) {
        return a.closestUnbetMatch.date.getTime() - b.closestUnbetMatch.date.getTime();
      }
      
      // Second priority: games with unplayed matchdays
      if (a.minUnplayedMatchday !== undefined && b.minUnplayedMatchday === undefined) return -1;
      if (a.minUnplayedMatchday === undefined && b.minUnplayedMatchday !== undefined) return 1;
      if (a.minUnplayedMatchday !== undefined && b.minUnplayedMatchday !== undefined) {
        return a.minUnplayedMatchday - b.minUnplayedMatchday;
      }
      
      // Third priority: alphabetical by name
      return a.name.localeCompare(b.name);
    });
    
    return sortedGames[0];
  };

  useEffect(() => {
    fetchGames();
  }, []);

  return {
    games,
    selectedGameId,
    bestGameId,
    loading,
    error,
    setSelectedGameId,
    refresh: fetchGames,
  };
}; 