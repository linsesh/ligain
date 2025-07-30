import i18n from '../i18n';

/**
 * Utility function to convert HTTP status codes to human-readable error messages
 */
export const getHumanReadableError = (status: number, originalError?: string): string => {
  switch (status) {
    case 401:
      return i18n.t('errors.authentication');
    case 403:
      return i18n.t('errors.forbidden');
    case 404:
      return i18n.t('errors.notFound');
    case 422:
      return i18n.t('errors.invalidData');
    case 429:
      return i18n.t('errors.tooManyRequests');
    case 500:
      return i18n.t('errors.serverError');
    case 502:
    case 503:
    case 504:
      return i18n.t('errors.serviceUnavailable');
    default:
      // For unknown status codes, fall back to a generic error message
      // without exposing the HTTP status code to users
      return originalError || i18n.t('errors.unknownError');
  }
};

/**
 * Handles a fetch Response, throwing a human-readable error if not ok
 */
export async function handleApiError(response: Response): Promise<never> {
  let errorData;
  try {
    errorData = await response.json();
  } catch (parseError) {
    errorData = { error: `HTTP ${response.status}: ${response.statusText}` };
  }
  const humanReadableError = getHumanReadableError(response.status, errorData.error);
  throw new Error(humanReadableError);
}

/**
 * Handles game-related errors with specific translations and user-friendly messages
 */
export const handleGameError = (errorMessage: string): { title: string; message: string } => {
  // Check for specific game-related errors
  if (errorMessage.includes('player has reached the maximum limit of 5 games')) {
    return {
      title: i18n.t('errors.playerGameLimitReachedTitle'),
      message: i18n.t('errors.playerGameLimitReached')
    };
  }
  
  if (errorMessage.includes('invalid game code')) {
    return {
      title: i18n.t('errors.error'),
      message: i18n.t('errors.pleaseEnterGameCode')
    };
  }
  
  if (errorMessage.includes('game code has expired')) {
    return {
      title: i18n.t('errors.error'),
      message: i18n.t('errors.gameCodeExpired', 'This game code has expired. Please ask for a new one.')
    };
  }
  
  if (errorMessage.includes('cannot join a finished game')) {
    return {
      title: i18n.t('errors.error'),
      message: i18n.t('errors.cannotJoinFinishedGame', 'Cannot join a finished game.')
    };
  }
  
  // Default fallback
  return {
    title: i18n.t('errors.error'),
    message: errorMessage || i18n.t('errors.unknownError')
  };
}; 