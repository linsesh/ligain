import { useState, useEffect } from 'react';
import { SeasonMatch, MatchResult } from '../src/types/match';
import { BetImpl } from '../src/types/bet';
import { useTimeService } from '../src/contexts/TimeServiceContext';

export const API_BASE_URL = 'http://192.168.1.184:8080';

export const useMatches = () => {
  const timeService = useTimeService();
  const [incomingMatches, setIncomingMatches] = useState<{ [key: string]: MatchResult }>({});
  const [pastMatches, setPastMatches] = useState<{ [key: string]: MatchResult }>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchMatches = async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/matches`);
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
          const bets = value.bets ? Object.entries(value.bets).reduce((acc: { [key: string]: BetImpl }, [player, bet]: [string, any]) => {
            acc[player] = new BetImpl(match, bet.predictedHomeGoals, bet.predictedAwayGoals);
            return acc;
          }, {}) : null;
          processed[key] = {
            match,
            bets,
            scores: value.scores
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