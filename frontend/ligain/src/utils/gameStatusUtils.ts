/**
 * Game Status Translation Utilities
 * 
 * Centralized functions for translating game statuses from backend values
 * to user-friendly localized text across the frontend.
 */

// Backend status values:
// - "not started": Game hasn't started yet
// - "in progress": Game is active (has matches but not all finished) 
// - "finished": All matches in the game are finished

export interface GameStatusInfo {
  text: string;
  variant: 'warning' | 'success' | 'primary' | 'finished' | 'negative';
}

/**
 * Translates a game status string to localized text and appropriate variant
 * @param status - The raw status string from the backend
 * @param t - The translation function from react-i18next
 * @returns Object containing translated text and variant
 */
export function getTranslatedGameStatus(status: string, t: any): GameStatusInfo {
  const statusLower = status.toLowerCase();
  switch (statusLower) {
    case 'in progress':
      return { text: t('games.inProgressTag'), variant: 'warning' };
    case 'finished':
      return { text: t('games.finishedTag'), variant: 'success' };
    case 'not started':
    case 'scheduled':
      return { text: t('games.scheduledTag'), variant: 'primary' };
    default:
      return { text: status, variant: 'warning' }; // fallback to original status if no translation found
  }
}

/**
 * Gets the appropriate variant for a game status
 * @param status - The raw status string from the backend
 * @returns The variant string for styling
 */
export function getGameStatusVariant(status: string): 'warning' | 'success' | 'primary' | 'finished' | 'negative' {
  const statusLower = status.toLowerCase();
  switch (statusLower) {
    case 'in progress':
      return 'warning';
    case 'finished':
      return 'success';
    case 'not started':
    case 'scheduled':
      return 'primary';
    default:
      return 'warning';
  }
} 