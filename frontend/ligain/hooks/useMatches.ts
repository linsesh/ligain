import { useState, useEffect } from 'react';
import { SeasonMatch } from '../src/types/match';

const API_BASE_URL = __DEV__ 
  ? 'http://192.168.1.37:8080/api'  // Development - local machine
  : 'https://your-production-api.com/api';  // Production - replace with your actual API URL

export const useMatches = () => {
  const [matches, setMatches] = useState<SeasonMatch[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchMatches = async () => {
      try {
        setLoading(true);
        const response = await fetch(`${API_BASE_URL}/matches`);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        const matches = data.map((match: any) => new SeasonMatch(
          match.homeTeam,
          match.awayTeam,
          match.homeGoals,
          match.awayGoals,
          match.homeTeamOdds,
          match.awayTeamOdds,
          match.drawOdds,
          match.status,
          match.seasonCode,
          match.competitionCode,
          match.date,
          match.matchday
        ));
        setMatches(matches);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch matches'));
      } finally {
        setLoading(false);
      }
    };

    fetchMatches();
  }, []);

  return { matches, loading, error };
}; 