import { useState, useEffect } from 'react';
import { BetImpl } from '../src/types/bet';
import { SeasonMatch } from '../src/types/match';

const API_BASE_URL = __DEV__ 
  ? 'http://192.168.1.121:8080/api'  // Development - local machine
  : 'https://your-production-api.com/api';  // Production - replace with your actual API URL

export const useBets = () => {
  const [bets, setBets] = useState<BetImpl[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchBets = async () => {
      try {
        setLoading(true);
        const response = await fetch(`${API_BASE_URL}/bets`);
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json();
        console.log(data);
        const bets = data.map((bet: any) => {
          const matchData = bet.Match;
          const match = new SeasonMatch(
            matchData.homeTeam,
            matchData.awayTeam,
            matchData.homeGoals,
            matchData.awayGoals,
            matchData.homeTeamOdds,
            matchData.awayTeamOdds,
            matchData.drawOdds,
            matchData.status,
            matchData.seasonCode,
            matchData.competitionCode,
            new Date(matchData.date),
            matchData.matchday
          );
          return new BetImpl(
            match,
            bet.PredictedHomeGoals,
            bet.PredictedAwayGoals
          );
        });
        setBets(bets);
        setError(null);
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to fetch matches'));
      } finally {
        setLoading(false);
      }
    };

    fetchBets();
  }, []);

  return { bets, loading, error };
};

export const getTempScoresFromBets = (bets: BetImpl[]) => {
  return bets.reduce((acc: { [key: string]: { home: number; away: number } }, bet) => {
    console.log('Je suis un log');
    console.log(bet);
    acc[bet.match.id()] = { home: bet.predictedHomeGoals, away: bet.predictedAwayGoals };
    return acc;
  }, {});
};
