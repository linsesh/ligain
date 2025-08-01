import i18n from '../i18n';

/**
 * Utility function to convert HTTP status codes to human-readable error messages
 */
export const getHumanReadableError = (status: number, originalError?: string): string => {
  // Check for ngrok errors first
  if (originalError && (originalError.includes('ERR_NGROK_') || originalError.includes('ngrok-error-code'))) {
    return 'ERR_NGROK_TUNNEL_DOWN';
  }
  
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

/**
 * Translates error messages from backend/developer language to human-readable messages
 */
export const translateError = (errorMessage: string): string => {
  // Debug logging to see what error message we're getting
  console.log('🔍 translateError - Input error message:', errorMessage);
  
  // HTTP status codes - check these FIRST to catch server errors before games errors
  if ((errorMessage.startsWith('502') || errorMessage.includes(' 502 ') || errorMessage.endsWith(' 502')) || 
      (errorMessage.startsWith('503') || errorMessage.includes(' 503 ') || errorMessage.endsWith(' 503')) || 
      (errorMessage.startsWith('504') || errorMessage.includes(' 504 ') || errorMessage.endsWith(' 504')) ||
      errorMessage.includes('Bad Gateway') || errorMessage.includes('Service Unavailable') || errorMessage.includes('Gateway Timeout')) {
    console.log('🔍 translateError - Matched 502/503/504 pattern');
    return i18n.t('errors.serviceUnavailable');
  }
  
  if ((errorMessage.startsWith('500') || errorMessage.includes(' 500 ')) || errorMessage.includes('Internal Server Error')) {
    console.log('🔍 translateError - Matched 500 pattern');
    return i18n.t('errors.serverError');
  }
  
  // Handle error formats from useBetSubmission and useMatches for 500
  if (errorMessage.includes('HTTP error! status: 500') || errorMessage.includes('500: ')) {
    console.log('🔍 translateError - Matched HTTP error 500 pattern');
    return i18n.t('errors.serverError');
  }
  
  if ((errorMessage.startsWith('401') || errorMessage.includes(' 401 ')) || errorMessage.includes('Unauthorized')) {
    return i18n.t('errors.authentication');
  }
  
  if ((errorMessage.startsWith('403') || errorMessage.includes(' 403 ')) || errorMessage.includes('Forbidden')) {
    return i18n.t('errors.forbidden');
  }
  
  if ((errorMessage.startsWith('404') || errorMessage.includes(' 404 ')) || errorMessage.includes('Not Found')) {
    return i18n.t('errors.notFound');
  }
  
  if ((errorMessage.startsWith('422') || errorMessage.includes(' 422 ')) || errorMessage.includes('Unprocessable Entity')) {
    return i18n.t('errors.invalidData');
  }
  
  if ((errorMessage.startsWith('429') || errorMessage.includes(' 429 ')) || errorMessage.includes('Too Many Requests')) {
    return i18n.t('errors.tooManyRequests');
  }
  
  if ((errorMessage.startsWith('400') || errorMessage.includes(' 400 ')) || errorMessage.includes('Bad Request')) {
    console.log('🔍 translateError - Matched 400 pattern');
    return i18n.t('errors.invalidData');
  }
  
  // Handle error formats from useBetSubmission and useMatches for 400
  if (errorMessage.includes('HTTP error! status: 400') || errorMessage.includes('400: ')) {
    console.log('🔍 translateError - Matched HTTP error 400 pattern');
    return i18n.t('errors.invalidData');
  }
  
  // Game-specific errors - check these AFTER HTTP status codes
  if (errorMessage.includes('failed to load games') || errorMessage.includes('Failed to fetch games')) {
    console.log('🔍 translateError - Matched failed to load games pattern');
    return i18n.t('errors.failedToLoadGames');
  }
  
  if (errorMessage.includes('failed to load game') || errorMessage.includes('Failed to load game')) {
    return i18n.t('errors.failedToLoadGames');
  }
  
  if (errorMessage.includes('failed to load matches') || errorMessage.includes('Failed to load matches')) {
    return i18n.t('errors.failedToLoadGames');
  }
  
  // Network and server connection errors - check these AFTER games patterns
  if (errorMessage.includes('Network request failed') || 
      errorMessage.includes('TypeError') ||
      errorMessage.includes('NetworkError') ||
      errorMessage.includes('connection') ||
      errorMessage.includes('ECONNREFUSED') ||
      errorMessage.includes('ENOTFOUND') ||
      errorMessage.includes('timeout') ||
      errorMessage.includes('fetch failed') ||
      errorMessage.includes('ERR_CONNECTION_REFUSED') ||
      errorMessage.includes('ERR_NETWORK') ||
      errorMessage.includes('ERR_INTERNET_DISCONNECTED') ||
      errorMessage.includes('ERR_NAME_NOT_RESOLVED')) {
    console.log('🔍 translateError - Matched network error pattern');
    return i18n.t('errors.networkError');
  }
  
  if (errorMessage.includes('server') && errorMessage.includes('not connected')) {
    return i18n.t('errors.serverNotConnected');
  }
  

  
  // Authentication-specific errors
  if (errorMessage.includes('Ligain servers are not available')) {
    return i18n.t('errors.serverUnavailable');
  }
  
  if (errorMessage.includes('Missing authentication context')) {
    return i18n.t('errors.missingAuthContext');
  }
  
  if (errorMessage.includes('display name is required for new users')) {
    return i18n.t('errors.missingAuthContext');
  }
  

  

  
  // Ngrok errors when tunnel is down - check for ngrok error codes
  if (errorMessage.includes('ERR_NGROK_') || 
      errorMessage.includes('ngrok-error-code') ||
      errorMessage === 'ERR_NGROK_TUNNEL_DOWN') {
    console.log('🔍 translateError - Matched ngrok error pattern');
    return i18n.t('errors.serviceUnavailable');
  }
  
  // If no specific pattern matches, return a generic error
  console.log('🔍 translateError - No pattern matched, returning generic error');
  return i18n.t('errors.unknownError');
}; 