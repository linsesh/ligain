import { createContext, useContext, useState, useEffect, useRef, useCallback, ReactNode } from 'react';
import { SeasonMatch, MatchResult } from '../types/match';
import { useGamesApi } from '../api';
import { useGames } from './GamesContext';
import { translateError } from '../utils/errorMessages';

const TTL_MS = 2 * 60 * 1000;

interface MatchesState {
  incomingMatches: Record<string, MatchResult>;
  incomingByMatchday: Record<number, MatchResult[]>;
  pastMatches: Record<string, MatchResult>;
}

interface MatchesContextType extends MatchesState {
  loading: boolean;
  error: Error | null;
  refresh: () => Promise<void>;
  markBetPlaced(matchId: string, playerId: string, playerName: string, homeGoals: number, awayGoals: number): void;
}

interface CacheEntry extends MatchesState {
  fetchedAt: number;
}

const emptyState: MatchesState = {
  incomingMatches: {},
  incomingByMatchday: {},
  pastMatches: {},
};

const MatchesContext = createContext<MatchesContextType | null>(null);

export function useMatches(): MatchesContextType {
  const ctx = useContext(MatchesContext);
  if (!ctx) throw new Error('useMatches must be used within MatchesProvider');
  return ctx;
}

function processMatchesResponse(data: any): MatchesState {
  const processEntries = (matches: any): Record<string, MatchResult> => {
    const processed: Record<string, MatchResult> = {};
    Object.entries(matches).forEach(([key, value]: [string, any]) => {
      const match = SeasonMatch.fromJSON(value.match);

      const bets = value.bets
        ? Object.entries(value.bets).reduce((acc: Record<string, any>, [, betData]: [string, any]) => {
            acc[betData.playerId] = {
              playerId: betData.playerId,
              playerName: betData.playerName,
              predictedHomeGoals: betData.predictedHomeGoals,
              predictedAwayGoals: betData.predictedAwayGoals,
              isModifiable: () => !match.isFinished() && !match.isInProgress(),
            };
            return acc;
          }, {})
        : null;

      const scores = value.scores
        ? Object.entries(value.scores).reduce((acc: Record<string, any>, [, scoreData]: [string, any]) => {
            acc[scoreData.playerId] = {
              playerId: scoreData.playerId,
              playerName: scoreData.playerName,
              points: scoreData.points,
              baseScore: scoreData.baseScore,
              riskMultiplier: scoreData.riskMultiplier,
              clairvoyantMultiplier: scoreData.clairvoyantMultiplier,
            };
            return acc;
          }, {})
        : null;

      const playerBetStatuses = value.playerBetStatuses
        ? Object.entries(value.playerBetStatuses).reduce((acc: Record<string, any>, [, s]: [string, any]) => {
            acc[s.playerId] = { playerId: s.playerId, playerName: s.playerName, hasBet: s.hasBet };
            return acc;
          }, {})
        : null;

      processed[key] = { match, bets, scores, playerBetStatuses };
    });
    return processed;
  };

  const incomingMatches = processEntries(data.incomingMatches);
  const incomingByMatchday = Object.values(incomingMatches).reduce<Record<number, MatchResult[]>>((acc, mr) => {
    const day = mr.match.getMatchday();
    if (!acc[day]) acc[day] = [];
    acc[day].push(mr);
    return acc;
  }, {});
  const pastMatches = processEntries(data.pastMatches);

  return { incomingMatches, incomingByMatchday, pastMatches };
}

export function MatchesProvider({ children }: { children: ReactNode }) {
  const gamesApi = useGamesApi();
  const { selectedGameId } = useGames();

  const cacheRef = useRef<Record<string, CacheEntry>>({});
  const fetchPromiseRef = useRef<Record<string, Promise<void>>>({});
  const mountedRef = useRef(true);

  const [matchesState, setMatchesState] = useState<MatchesState>(emptyState);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    mountedRef.current = true;
    return () => { mountedRef.current = false; };
  }, []);

  const applyCache = useCallback((gameId: string) => {
    const entry = cacheRef.current[gameId];
    if (entry) {
      setMatchesState({
        incomingMatches: entry.incomingMatches,
        incomingByMatchday: entry.incomingByMatchday,
        pastMatches: entry.pastMatches,
      });
    }
  }, []);

  const fetchForGame = useCallback(async (gameId: string, showLoading: boolean) => {
    if (gameId in fetchPromiseRef.current) {
      return fetchPromiseRef.current[gameId];
    }

    const promise = (async () => {
      if (showLoading) setLoading(true);
      try {
        const data = await gamesApi.getGameMatches(gameId);
        if (!mountedRef.current) return;

        const processed = processMatchesResponse(data);
        cacheRef.current[gameId] = { ...processed, fetchedAt: Date.now() };
        setMatchesState(processed);
        setError(null);
      } catch (err) {
        if (!mountedRef.current) return;
        const msg = err instanceof Error ? err.message : 'Failed to fetch matches';
        setError(new Error(translateError(msg)));
      } finally {
        if (mountedRef.current && showLoading) setLoading(false);
        delete fetchPromiseRef.current[gameId];
      }
    })();

    fetchPromiseRef.current[gameId] = promise;
    return promise;
  }, [gamesApi]);

  useEffect(() => {
    if (!selectedGameId) {
      setMatchesState(emptyState);
      return;
    }

    const cached = cacheRef.current[selectedGameId];
    if (cached) {
      applyCache(selectedGameId);
      if (Date.now() - cached.fetchedAt > TTL_MS) {
        fetchForGame(selectedGameId, false);
      }
    } else {
      setLoading(true);
      fetchForGame(selectedGameId, true);
    }
  }, [selectedGameId, applyCache, fetchForGame]);

  const refresh = useCallback(async () => {
    if (!selectedGameId) return;
    delete cacheRef.current[selectedGameId];
    await fetchForGame(selectedGameId, true);
  }, [selectedGameId, fetchForGame]);

  const markBetPlaced = useCallback((matchId: string, playerId: string, playerName: string, homeGoals: number, awayGoals: number) => {
    setMatchesState(prev => {
      const matchKey = Object.keys(prev.incomingMatches).find(
        key => prev.incomingMatches[key].match.id() === matchId
      );
      if (!matchKey) return prev;

      const mr = prev.incomingMatches[matchKey];
      const newBet = {
        playerId,
        playerName,
        predictedHomeGoals: homeGoals,
        predictedAwayGoals: awayGoals,
        isModifiable: () => true,
      };

      const updatedBets = { ...(mr.bets ?? {}), [playerId]: newBet };
      const updatedStatuses = mr.playerBetStatuses
        ? {
            ...mr.playerBetStatuses,
            ...(mr.playerBetStatuses[playerId]
              ? { [playerId]: { ...mr.playerBetStatuses[playerId], hasBet: true } }
              : {}),
          }
        : null;

      const updatedMr = { ...mr, bets: updatedBets, playerBetStatuses: updatedStatuses };
      const updatedIncoming = { ...prev.incomingMatches, [matchKey]: updatedMr };

      const incomingByMatchday = Object.values(updatedIncoming).reduce<Record<number, MatchResult[]>>((acc, m) => {
        const day = m.match.getMatchday();
        if (!acc[day]) acc[day] = [];
        acc[day].push(m);
        return acc;
      }, {});

      const newState = { ...prev, incomingMatches: updatedIncoming, incomingByMatchday };

      if (selectedGameId && cacheRef.current[selectedGameId]) {
        cacheRef.current[selectedGameId] = {
          ...newState,
          fetchedAt: cacheRef.current[selectedGameId].fetchedAt,
        };
      }

      return newState;
    });
  }, [selectedGameId]);

  return (
    <MatchesContext.Provider value={{ ...matchesState, loading, error, refresh, markBetPlaced }}>
      {children}
    </MatchesContext.Provider>
  );
}
