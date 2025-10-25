import { useState, useEffect } from 'react';
import { SeasonMatch, MatchResult } from '../src/types/match';
import { BetImpl } from '../src/types/bet';
import { useTimeService } from '../src/contexts/TimeServiceContext';
import { API_CONFIG, getAuthenticatedHeaders } from '../src/config/api';
import { translateError } from '../src/utils/errorMessages';

export const useMatches = (gameId: string) => {
  const timeService = useTimeService();
  const [incomingMatches, setIncomingMatches] = useState<{ [key: string]: MatchResult }>({});
  const [pastMatches, setPastMatches] = useState<{ [key: string]: MatchResult }>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchMatches = async () => {
    try {
      const headers = await getAuthenticatedHeaders();
      
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/game/${gameId}/matches`, {
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

          // Log first bet to see player names
          if (bets && Object.keys(bets).length > 0) {
            const firstBet = Object.values(bets)[0];
          }

          // Process scores with new structure - use playerId as key
          const scores = value.scores ? Object.entries(value.scores).reduce((acc: { [key: string]: any }, [playerName, scoreData]: [string, any]) => {
            acc[scoreData.playerId] = {
              playerId: scoreData.playerId,
              playerName: scoreData.playerName,
              points: scoreData.points
            };
            return acc;
          }, {}) : null;

          // Log first score to see player names
          if (scores && Object.keys(scores).length > 0) {
            const firstScore = Object.values(scores)[0];
          }
          
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
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch matches';
      setError(new Error(translateError(errorMessage)));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMatches();
  }, [gameId]);

  return {
    incomingMatches,
    pastMatches,
    loading,
    error,
    refresh: fetchMatches
  };
}; 