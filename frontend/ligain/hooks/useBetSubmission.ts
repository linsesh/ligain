import { useState } from 'react';
import { useGamesApi } from '../src/api';
import { translateError } from '../src/utils/errorMessages';
import { useNotifications } from '../src/hooks/useNotifications';
import { useMatches } from '../src/contexts/MatchesContext';
import { useAuth } from '../src/contexts/AuthContext';

export const useBetSubmission = (gameId: string, onFail?: (matchId: string) => void) => {
  const gamesApi = useGamesApi();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [lastFailedMatchId, setLastFailedMatchId] = useState<string | null>(null);
  const { cancelMatchNotification } = useNotifications();
  const { markBetPlaced } = useMatches();
  const { player } = useAuth();

  const submitBet = async (
    matchId: string,
    homeGoals: number,
    awayGoals: number
  ): Promise<void> => {
    setIsSubmitting(true);
    setError(null);
    setLastFailedMatchId(null);

    try {
      const data = await gamesApi.placeBet(gameId, matchId, homeGoals, awayGoals);
      console.log('Bet saved successfully:', data);

      try {
        if (player) {
          markBetPlaced(matchId, player.id, player.name, homeGoals, awayGoals);
        }
      } catch (updateError) {
        console.warn('Failed to update match context after bet submission:', updateError);
      }

      try {
        await cancelMatchNotification(matchId);
      } catch (notificationError) {
        // Don't fail bet submission if notification cancellation fails
        // Log error but continue with successful bet submission
        console.warn('Failed to cancel notification after bet submission:', notificationError);
      }

      setIsSubmitting(false); // Set to false immediately after success
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to save bet';
      const translatedError = new Error(translateError(errorMessage));
      console.error('Error saving bet:', translatedError);
      setError(translatedError);
      setIsSubmitting(false);
      setLastFailedMatchId(matchId);
      if (onFail) onFail(matchId);
    }
  };

  return {
    submitBet,
    isSubmitting,
    error,
    lastFailedMatchId
  };
}; 