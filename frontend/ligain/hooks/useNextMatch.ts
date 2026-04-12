import { useMatches } from './useMatches';
import { MatchResult } from '../src/types/match';

interface UseNextMatchResult {
  remainingCount: number;
  nextMatch: MatchResult | null;
}

/**
 * Finds the next unbet upcoming match in the same matchday as the current match.
 *
 * "Unbet" means the current player has no bet recorded for that match.
 * Matches that have already started are excluded.
 * Results are sorted by date ascending so the earliest upcoming match comes first.
 */
export function useNextMatch(
  gameId: string,
  currentMatchId: string,
  matchday: number,
  playerId: string,
): UseNextMatchResult {
  const { incomingByMatchday } = useMatches(gameId);
  const now = new Date();

  const unbetSiblings = (incomingByMatchday[matchday] ?? [])
    .filter(mr =>
      mr.match.id() !== currentMatchId &&
      !mr.match.hasStarted(now) &&
      !(mr.bets && mr.bets[playerId])
    )
    .sort((a, b) => a.match.getDate().getTime() - b.match.getDate().getTime());

  return {
    remainingCount: unbetSiblings.length,
    nextMatch: unbetSiblings[0] ?? null,
  };
}
