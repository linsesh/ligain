import { useEffect, useRef } from 'react';

/**
 * Automatically submits a bet when both homeGoals and awayGoals are filled.
 * Skips the initial render so pre-filled values from navigation params don't
 * trigger an immediate re-submission of an already-placed bet.
 */
export function useBetAutoSubmit(
  homeGoals: string,
  awayGoals: string,
  onSubmit: (home: number, away: number) => void,
) {
  const isFirstRender = useRef(true);

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false;
      return;
    }
    if (homeGoals === '' || awayGoals === '') return;
    const h = Number(homeGoals);
    const a = Number(awayGoals);
    if (!isNaN(h) && !isNaN(a) && h >= 0 && a >= 0) {
      onSubmit(h, a);
    }
  }, [homeGoals, awayGoals]);
}
