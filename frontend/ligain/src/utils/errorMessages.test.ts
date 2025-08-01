import { handleGameError, translateError } from './errorMessages';
import i18n from '../i18n';

// Mock i18n
jest.mock('../i18n', () => ({
  t: jest.fn((key: string) => {
    const translations: { [key: string]: string } = {
      'errors.playerGameLimitReachedTitle': 'Game Limit Reached',
      'errors.playerGameLimitReached': 'You have reached the maximum limit of 5 games. Please leave one of your existing games before creating or joining a new one.',
      'errors.error': 'Error',
      'errors.pleaseEnterGameCode': 'Please enter a game code',
      'errors.gameCodeExpired': 'This game code has expired. Please ask for a new one.',
      'errors.cannotJoinFinishedGame': 'Cannot join a finished game.',
      'errors.unknownError': 'Something went wrong. Please try again later.',
      'errors.failedToLoadGames': 'Unable to load games. Please check your internet connection and try again.',
      'errors.invalidData': 'Invalid information provided. Please check your details',
      'errors.authentication': 'Authentication went wrong, please refresh the page and retry',
      'errors.forbidden': 'Authentication went wrong, please refresh the page and retry',
      'errors.notFound': 'Service not found. Please try again later',
      'errors.tooManyRequests': 'Too many requests. Please wait a moment and try again',
      'errors.serverError': 'Server error. Please try again later',
      'errors.serviceUnavailable': 'Ligain is unavailable at the moment',
      'errors.serverNotConnected': 'Server is not available. Please try again later.',
      'errors.networkError': 'Network error. Please check your connection and try again.',
      'errors.serverUnavailable': 'Ligain servers are not available for now. Please try again later.',
      'errors.missingAuthContext': 'Missing authentication context. Please try again.'
    };
    return translations[key] || key;
  })
}));

describe('handleGameError', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should handle player game limit error', () => {
    const errorMessage = 'player has reached the maximum limit of 5 games';
    const result = handleGameError(errorMessage);
    
    expect(result.title).toBe('Game Limit Reached');
    expect(result.message).toBe('You have reached the maximum limit of 5 games. Please leave one of your existing games before creating or joining a new one.');
  });

  it('should handle invalid game code error', () => {
    const errorMessage = 'invalid game code';
    const result = handleGameError(errorMessage);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('Please enter a game code');
  });

  it('should handle expired game code error', () => {
    const errorMessage = 'game code has expired';
    const result = handleGameError(errorMessage);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('This game code has expired. Please ask for a new one.');
  });

  it('should handle finished game error', () => {
    const errorMessage = 'cannot join a finished game';
    const result = handleGameError(errorMessage);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('Cannot join a finished game.');
  });

  it('should handle unknown errors with fallback', () => {
    const errorMessage = 'some unknown error';
    const result = handleGameError(errorMessage);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('some unknown error');
  });

  it('should handle empty error message with fallback', () => {
    const errorMessage = '';
    const result = handleGameError(errorMessage);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('Something went wrong. Please try again later.');
  });
});

describe('translateError', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should translate failed to load games error', () => {
    const errorMessage = 'failed to load games 400';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Unable to load games. Please check your internet connection and try again.');
  });

  it('should translate Failed to fetch games error', () => {
    const errorMessage = 'Failed to fetch games: 400';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Unable to load games. Please check your internet connection and try again.');
  });

  it('should translate HTTP 400 error', () => {
    const errorMessage = '400 Bad Request';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Invalid information provided. Please check your details');
  });

  it('should translate HTTP 401 error', () => {
    const errorMessage = '401 Unauthorized';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Authentication went wrong, please refresh the page and retry');
  });

  it('should translate HTTP 500 error', () => {
    const errorMessage = '500 Internal Server Error';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Server error. Please try again later');
  });

  it('should translate network error', () => {
    const errorMessage = 'Network request failed';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Network error. Please check your connection and try again.');
  });

  it('should translate TypeError network error', () => {
    const errorMessage = 'TypeError: fetch failed';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Network error. Please check your connection and try again.');
  });

  it('should translate connection refused error', () => {
    const errorMessage = 'ECONNREFUSED';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Network error. Please check your connection and try again.');
  });

  it('should translate ngrok error when tunnel is down', () => {
    const errorMessage = 'ERR_NGROK_3200';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Ligain is unavailable at the moment');
  });

  it('should translate ngrok tunnel down error code', () => {
    const errorMessage = 'ERR_NGROK_TUNNEL_DOWN';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Ligain is unavailable at the moment');
  });

  it('should translate server not connected error', () => {
    const errorMessage = 'server is not connected';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Server is not available. Please try again later.');
  });

  it('should translate unknown error with fallback', () => {
    const errorMessage = 'some unknown error message';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Something went wrong. Please try again later.');
  });

  it('should translate server unavailable error', () => {
    const errorMessage = 'Ligain servers are not available for now. Please try again later.';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Ligain servers are not available for now. Please try again later.');
  });

  it('should translate missing authentication context error', () => {
    const errorMessage = 'Missing authentication context. Please try again.';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Missing authentication context. Please try again.');
  });

  it('should translate display name required error', () => {
    const errorMessage = 'display name is required for new users';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Missing authentication context. Please try again.');
  });

  it('should translate server down error correctly', () => {
    const errorMessage = 'Failed to fetch games: 503';
    const result = translateError(errorMessage);
    
    expect(result).toBe('Ligain is unavailable at the moment');
  });

  it('should not fall into 400 pattern for server errors', () => {
    const errorMessage = 'Failed to fetch games: 400';
    const result = translateError(errorMessage);
    
    // Should return the specific games error, not the generic 400 error
    expect(result).toBe('Unable to load games. Please check your internet connection and try again.');
  });

  it('should handle actual 400 error format from GamesContext', () => {
    const errorMessage = 'Failed to fetch games: 400';
    const result = translateError(errorMessage);
    
    // This should match the games error pattern, not the 400 pattern
    expect(result).toBe('Unable to load games. Please check your internet connection and try again.');
  });

  it('should handle standalone 400 error', () => {
    const errorMessage = '400 Bad Request';
    const result = translateError(errorMessage);
    
    // This should match the 400 pattern
    expect(result).toBe('Invalid information provided. Please check your details');
  });

  it('should handle HTTP error format from useBetSubmission', () => {
    const errorMessage = 'HTTP error! status: 400 - Bad Request';
    const result = translateError(errorMessage);
    
    // This should match the HTTP error 400 pattern
    expect(result).toBe('Invalid information provided. Please check your details');
  });

  it('should handle error format from useMatches', () => {
    const errorMessage = '400: Invalid request';
    const result = translateError(errorMessage);
    
    // This should match the HTTP error 400 pattern
    expect(result).toBe('Invalid information provided. Please check your details');
  });
}); 