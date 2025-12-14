import { useState } from 'react';
import { API_CONFIG, getAuthenticatedHeaders } from '../src/config/api';
import { useAuth } from '../src/contexts/AuthContext';
import { translateError } from '../src/utils/errorMessages';
import { useNotifications } from '../src/hooks/useNotifications';

const MAX_RETRIES = 1;
const RETRY_DELAY_MS = 2000; // 2 seconds

export const useBetSubmission = (gameId: string, onFail?: (matchId: string) => void) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [lastFailedMatchId, setLastFailedMatchId] = useState<string | null>(null);
  const { checkAuth } = useAuth();
  // Integration point: Get notification cancellation function
  // This allows us to cancel match reminders when user places a bet
  const { cancelMatchNotification } = useNotifications();

  const submitBet = async (
    matchId: string,
    homeGoals: number,
    awayGoals: number,
    retryCount = 0
  ): Promise<void> => {
    setIsSubmitting(true);
    setError(null);
    setLastFailedMatchId(null);

    try {
      const headers = await getAuthenticatedHeaders({
        'Content-Type': 'application/json',
      });
      const requestBody = {
        matchId,
        predictedHomeGoals: homeGoals,
        predictedAwayGoals: awayGoals
      };
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/game/${gameId}/bet`, {
        method: 'POST',
        headers,
        body: JSON.stringify(requestBody)
      });

      if (!response.ok) {
        let errorMsg = '';
        try {
          const errorData = await response.json();
          errorMsg = errorData.error || '';
        } catch (e) {
          errorMsg = '';
        }
        // 401: Token expired, try silent re-auth and retry
        if (response.status === 401 && (errorMsg === 'Invalid or expired token' || errorMsg.toLowerCase().includes('expired'))) {
          if (retryCount < MAX_RETRIES) {
            await checkAuth();
            setTimeout(() => {
              submitBet(matchId, homeGoals, awayGoals, retryCount + 1);
            }, RETRY_DELAY_MS);
            return;
          }
        }
        throw new Error(`HTTP error! status: ${response.status} - ${errorMsg || response.statusText}`);
      }

      const data = await response.json();
      console.log('Bet saved successfully:', data);
      
      // Notification cancellation: Cancel reminder notification when bet is successfully placed
      // Why we do this: User has already placed bet, no need to remind them anymore
      // When this runs: Only on successful bet submission (not on errors)
      // Integration point: This connects bet submission with notification system
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
      setIsSubmitting(false); // Set to false after max retries
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