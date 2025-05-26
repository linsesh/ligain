import { useState, useEffect, useCallback } from 'react';
import { SeasonMatch, MatchResult, MatchesResponse } from '../src/types/match';

export const API_BASE_URL = __DEV__ 
  ? 'http://192.168.1.37:8080/api'  // Development - local machine
  : 'https://your-production-api.com/api';

export const useMatches = () => {
  const [incomingMatches, setIncomingMatches] = useState<{ [key: string]: MatchResult }>({});
  const [pastMatches, setPastMatches] = useState<{ [key: string]: MatchResult }>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchMatches = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(`${API_BASE_URL}/matches`);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      const data: MatchesResponse = await response.json();
      
      // Process incoming matches
      const processedIncomingMatches: { [key: string]: MatchResult } = {};
      for (const [matchId, matchResult] of Object.entries(data.incomingMatches)) {
        processedIncomingMatches[matchId] = {
          match: SeasonMatch.fromJSON(matchResult.match),
          bets: matchResult.bets,
          scores: matchResult.scores
        };
      }
      setIncomingMatches(processedIncomingMatches);

      // Process past matches
      const processedPastMatches: { [key: string]: MatchResult } = {};
      for (const [matchId, matchResult] of Object.entries(data.pastMatches)) {
        processedPastMatches[matchId] = {
          match: SeasonMatch.fromJSON(matchResult.match),
          bets: matchResult.bets,
          scores: matchResult.scores
        };
      }
      setPastMatches(processedPastMatches);
      
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch matches'));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchMatches();
  }, [fetchMatches]);

  return { incomingMatches, pastMatches, loading, error, refresh: fetchMatches };
}; 