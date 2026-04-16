import { useMatches } from '../src/contexts/MatchesContext';
import { MatchResult } from '../src/types/match';

interface UseNextMatchResult {
  remainingCount: number;
  nextMatch: MatchResult | null;
}

export function useNextMatch(
  currentMatchId: string,
  matchday: number,
  playerId: string,
): UseNextMatchResult {
  const { incomingByMatchday } = useMatches();
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
