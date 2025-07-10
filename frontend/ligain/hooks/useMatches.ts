import { useState, useEffect } from 'react';
import { SeasonMatch, MatchResult } from '../src/types/match';
import { BetImpl } from '../src/types/bet';
import { useTimeService } from '../src/contexts/TimeServiceContext';
import { API_CONFIG, getAuthenticatedHeaders } from '../src/config/api';

export const useMatches = () => {
  const timeService = useTimeService();
  const [incomingMatches, setIncomingMatches] = useState<{ [key: string]: MatchResult }>({});
  const [pastMatches, setPastMatches] = useState<{ [key: string]: MatchResult }>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchMatches = async () => {
    try {
      const headers = await getAuthenticatedHeaders();
      console.log('🔧 useMatches - Using authenticated headers:', {
        hasApiKey: !!headers['X-API-Key'],
        hasAuth: !!headers['Authorization']
      });
      
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/game/${API_CONFIG.GAME_ID}/matches`, {
        headers,
      });
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(`${response.status}: ${errorData.error || 'Unknown error'}`);
      }
      
      const data = await response.json();
      
      // Convert the matches to SeasonMatch objects and bets to BetImpl objects
      const processMatches = (matches: any) => {
        const processed: { [key: string]: MatchResult } = {};
        Object.entries(matches).forEach(([key, value]: [string, any]) => {
          const match = SeasonMatch.fromJSON(value.match);
          
          const bets = value.bets ? Object.entries(value.bets).reduce((acc: { [key: string]: any }, [playerName, betData]: [string, any]) => {
            acc[betData.playerId] = {
              playerId: betData.playerId,
              playerName: betData.playerName,
              predictedHomeGoals: betData.predictedHomeGoals,
              predictedAwayGoals: betData.predictedAwayGoals,
              isModifiable: (now: Date) => {
                // Simple implementation - can be enhanced later
                return !match.isFinished() && !match.isInProgress();
              }
            };
            return acc;
          }, {}) : null;
          
          // Process scores with new structure - use playerId as key
          const scores = value.scores ? Object.entries(value.scores).reduce((acc: { [key: string]: any }, [playerName, scoreData]: [string, any]) => {
            acc[scoreData.playerId] = {
              playerId: scoreData.playerId,
              playerName: scoreData.playerName,
              points: scoreData.points
            };
            return acc;
          }, {}) : null;
          
          processed[key] = {
            match,
            bets,
            scores
          };
        });
        return processed;
      };

      setIncomingMatches(processMatches(data.incomingMatches));
      setPastMatches(processMatches(data.pastMatches));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch matches'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMatches();
  }, []);

  return {
    incomingMatches,
    pastMatches,
    loading,
    error,
    refresh: fetchMatches
  };
}; 