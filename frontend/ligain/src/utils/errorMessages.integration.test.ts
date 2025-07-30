import { handleGameError } from './errorMessages';

// Mock i18n for testing
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

describe('Error Handling Integration', () => {
  it('should handle backend player game limit error correctly', () => {
    // This is the exact error message returned by the backend
    const backendError = 'player has reached the maximum limit of 5 games';
    const result = handleGameError(backendError);
    
    expect(result.title).toBe('Game Limit Reached');
    expect(result.message).toBe('You have reached the maximum limit of 5 games. Please leave one of your existing games before creating or joining a new one.');
  });

  it('should handle backend invalid game code error correctly', () => {
    // This is the exact error message returned by the backend
    const backendError = 'invalid game code: game code not found';
    const result = handleGameError(backendError);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('Please enter a game code');
  });

  it('should handle backend expired code error correctly', () => {
    // This is the exact error message returned by the backend
    const backendError = 'game code has expired';
    const result = handleGameError(backendError);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('This game code has expired. Please ask for a new one.');
  });

  it('should handle backend finished game error correctly', () => {
    // This is the exact error message returned by the backend
    const backendError = 'cannot join a finished game';
    const result = handleGameError(backendError);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('Cannot join a finished game.');
  });

  it('should handle other backend errors gracefully', () => {
    // Test with other potential backend errors
    const backendError = 'some other backend error';
    const result = handleGameError(backendError);
    
    expect(result.title).toBe('Error');
    expect(result.message).toBe('some other backend error');
  });
}); 