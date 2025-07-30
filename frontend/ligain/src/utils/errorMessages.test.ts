import { handleGameError } from './errorMessages';
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
      'errors.unknownError': 'Something went wrong. Please try again later.'
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