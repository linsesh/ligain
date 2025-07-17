import { useState } from 'react';
import { API_CONFIG, getAuthenticatedHeaders } from '../src/config/api';

const MAX_RETRIES = 1;
const RETRY_DELAY_MS = 2000; // 2 seconds

export const useBetSubmission = (gameId?: string) => {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const submitBet = async (
    matchId: string,
    homeGoals: number,
    awayGoals: number,
    retryCount = 0
  ): Promise<void> => {
    setIsSubmitting(true);
    setError(null);

    // Use the provided gameId or fall back to the hardcoded one
    const targetGameId = gameId || API_CONFIG.GAME_ID;

    try {
      const headers = await getAuthenticatedHeaders({
        'Content-Type': 'application/json',
      });
      
      const requestBody = {
        matchId,
        predictedHomeGoals: homeGoals,
        predictedAwayGoals: awayGoals
      };
      
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/game/${targetGameId}/bet`, {
        method: 'POST',
        headers,
        body: JSON.stringify(requestBody)
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status} - ${response.statusText}`);
      }

      const data = await response.json();
      console.log('Bet saved successfully:', data);
      setIsSubmitting(false); // Set to false immediately after success
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Failed to save bet');
      console.error('Error saving bet:', error);

      // If we haven't exceeded max retries, try again after delay
      if (retryCount < MAX_RETRIES) {
        console.log(`Retrying bet submission in ${RETRY_DELAY_MS}ms...`);
        setTimeout(() => {
          submitBet(matchId, homeGoals, awayGoals, retryCount + 1);
        }, RETRY_DELAY_MS);
        return;
      }
      setError(error);
      setIsSubmitting(false); // Set to false after max retries
    }
  };

  return {
    submitBet,
    isSubmitting,
    error
  };
}; 