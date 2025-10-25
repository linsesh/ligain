import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import MatchesList from '../../../app/(tabs)/games/game/_MatchesList';
import { useBetSynchronization } from '../../../hooks/useBetSynchronization';
import { useBetSubmission } from '../../../hooks/useBetSubmission';
import { useMatches } from '../../../hooks/useMatches';
import { useAuth } from '../../contexts/AuthContext';
import { useTranslation } from 'react-i18next';

// Mock the hooks
jest.mock('../../../hooks/useBetSynchronization');
jest.mock('../../../hooks/useBetSubmission');
jest.mock('../../../hooks/useMatches');
jest.mock('../../contexts/AuthContext');
jest.mock('react-i18next');

// Mock expo-router
jest.mock('expo-router', () => ({
  useRouter: jest.fn(() => ({
    push: jest.fn(),
    replace: jest.fn(),
    back: jest.fn(),
  })),
}));

// Mock react-native components that might not be available in test environment
jest.mock('react-native', () => {
  const RN = jest.requireActual('react-native');
  return {
    ...RN,
    Alert: {
      alert: jest.fn(),
    },
  };
});

// Mock other dependencies
jest.mock('../../contexts/GamesContext', () => ({
  useGames: jest.fn(() => ({ games: [] })),
}));

jest.mock('../../contexts/TimeServiceContext', () => ({
  useTimeService: jest.fn(() => ({ now: () => new Date() })),
}));

// Mock other components and utilities
jest.mock('../../components/StatusTag', () => 'StatusTag');
jest.mock('../../components/ShareableMatchResult', () => 'ShareableMatchResult');
jest.mock('../../utils/shareUtils', () => ({
  captureAndShareWithOptions: jest.fn(),
  formatDateForShare: jest.fn(),
}));
jest.mock('../../utils/dateUtils', () => ({
  formatTime: jest.fn(),
  formatDate: jest.fn(),
}));
jest.mock('../../constants/colors', () => ({
  primary: '#007AFF',
  secondary: '#34C759',
  text: '#000000',
  textSecondary: '#666666',
  background: '#FFFFFF',
  card: '#F2F2F7',
  border: '#C6C6C8',
  disabled: '#E5E5EA',
  success: '#34C759',
  error: '#FF3B30',
  warning: '#FF9500',
  loadingBackground: '#F2F2F7',
}));
jest.mock('../../constants/sharedStyles', () => ({
  shareButton: {
    backgroundColor: '#007AFF',
    padding: 8,
    borderRadius: 4,
  },
}));
jest.mock('react-native-view-shot', () => 'ViewShot');

const mockUseBetSynchronization = useBetSynchronization as jest.MockedFunction<typeof useBetSynchronization>;
const mockUseBetSubmission = useBetSubmission as jest.MockedFunction<typeof useBetSubmission>;
const mockUseMatches = useMatches as jest.MockedFunction<typeof useMatches>;
const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;
const mockUseTranslation = useTranslation as jest.MockedFunction<typeof useTranslation>;

// Mock the BetSyncModal component
jest.mock('../../components/BetSyncModal', () => {
  return function MockBetSyncModal({ 
    visible, 
    onSynchronize, 
    onNotNow, 
    onRetryFailed,
    loading, 
    mode = 'initial',
    syncResult 
  }: any) {
    if (!visible) return null;
    
    const getModalContent = () => {
      switch (mode) {
        case 'success':
          return <div data-testid="success-message">All bets synchronized successfully</div>;
        case 'partialSuccess':
          return (
            <div>
              <div data-testid="partial-success-message">Some bets synchronized</div>
              {syncResult?.successful?.length > 0 && (
                <div data-testid="successful-bets">
                  {syncResult.successful.map((bet: any, index: number) => (
                    <div key={index} data-testid={`successful-bet-${index}`}>
                      {bet.homeTeam} vs {bet.awayTeam}
                    </div>
                  ))}
                </div>
              )}
              {syncResult?.failed?.length > 0 && (
                <div data-testid="failed-bets">
                  {syncResult.failed.map((failedBet: any, index: number) => (
                    <div key={index} data-testid={`failed-bet-${index}`}>
                      {failedBet.match.homeTeam} vs {failedBet.match.awayTeam}
                    </div>
                  ))}
                </div>
              )}
            </div>
          );
        case 'failure':
          return <div data-testid="failure-message">All bets failed to synchronize</div>;
        default:
          return <div data-testid="initial-message">Synchronize bets?</div>;
      }
    };

    const getButtons = () => {
      switch (mode) {
        case 'success':
        case 'failure':
          return (
            <button data-testid="close-button" onClick={onNotNow}>
              Close
            </button>
          );
        case 'partialSuccess':
          return (
            <div>
              <button data-testid="close-button" onClick={onNotNow}>
                Close
              </button>
              {onRetryFailed && syncResult?.failed?.length > 0 && (
                <button data-testid="retry-failed-button" onClick={onRetryFailed}>
                  Retry Failed
                </button>
              )}
            </div>
          );
        default:
          return (
            <div>
              <button data-testid="synchronize-button" onClick={onSynchronize} disabled={loading}>
                {loading ? 'Loading...' : 'Synchronize'}
              </button>
              <button data-testid="not-now-button" onClick={onNotNow}>
                Not now
              </button>
            </div>
          );
      }
    };

    return (
      <div data-testid="bet-sync-modal" data-mode={mode}>
        {getModalContent()}
        {getButtons()}
      </div>
    );
  };
});

describe('MatchesList with Bet Synchronization', () => {
  const mockPlayer = { id: 'player-1', name: 'Test Player' };
  const mockSubmitBet = jest.fn();
  const mockRefresh = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    
    mockUseAuth.mockReturnValue({ player: mockPlayer } as any);
    mockUseTranslation.mockReturnValue({ t: (key: string) => key } as any);
    
    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: mockRefresh
    } as any);
    
    mockUseBetSubmission.mockReturnValue({
      submitBet: mockSubmitBet,
      isSubmitting: false,
      error: null
    } as any);
    
    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity: null,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);
  });

  it('should not show modal when no sync opportunity exists', () => {
    render(<MatchesList gameId="game-1" />);
    
    expect(screen.queryByTestId('bet-sync-modal')).toBeNull();
  });

  it('should show modal when sync opportunity exists', () => {
    const syncOpportunity = {
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1
        }
      ]
    };

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);

    render(<MatchesList gameId="game-1" />);
    
    expect(screen.getByTestId('bet-sync-modal')).toBeTruthy();
  });

  it('should not show modal if already shown for this game', () => {
    const syncOpportunity = {
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1
        }
      ]
    };

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);

    const { rerender } = render(<MatchesList gameId="game-1" />);
    
    // Modal should appear first time
    expect(screen.getByTestId('bet-sync-modal')).toBeTruthy();
    
    // Close modal
    fireEvent.press(screen.getByTestId('not-now-button'));
    expect(screen.queryByTestId('bet-sync-modal')).toBeNull();
    
    // Rerender with same game - modal should not appear again
    rerender(<MatchesList gameId="game-1" />);
    expect(screen.queryByTestId('bet-sync-modal')).toBeNull();
  });

  it('should show modal again when switching to different game', () => {
    const syncOpportunity = {
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1
        }
      ]
    };

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);

    const { rerender } = render(<MatchesList gameId="game-1" />);
    
    // Close modal
    fireEvent.press(screen.getByTestId('not-now-button'));
    expect(screen.queryByTestId('bet-sync-modal')).toBeNull();
    
    // Switch to different game - modal should appear again
    rerender(<MatchesList gameId="game-2" />);
    expect(screen.getByTestId('bet-sync-modal')).toBeTruthy();
  });

  it('should trigger bet submissions when synchronize is clicked', async () => {
    const syncOpportunity = {
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1
        },
        {
          matchId: 'match-2',
          homeTeam: 'Team C',
          awayTeam: 'Team D',
          matchday: 2,
          predictedHomeGoals: 1,
          predictedAwayGoals: 0
        }
      ]
    };

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);

    mockSubmitBet.mockResolvedValue(undefined);
    mockRefresh.mockResolvedValue(undefined);

    render(<MatchesList gameId="game-1" />);
    
    // Click synchronize
    fireEvent.press(screen.getByTestId('synchronize-button'));
    
    // Wait for async operations
    await waitFor(() => {
      expect(mockSubmitBet).toHaveBeenCalledTimes(2);
    });
    
    expect(mockSubmitBet).toHaveBeenCalledWith('match-1', 2, 1);
    expect(mockSubmitBet).toHaveBeenCalledWith('match-2', 1, 0);
    expect(mockRefresh).toHaveBeenCalled();
  });

  it('should show failure modal when all bets fail', async () => {
    const syncOpportunity = {
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1
        }
      ]
    };

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);

    mockSubmitBet.mockRejectedValue(new Error('Sync failed'));
    mockRefresh.mockResolvedValue(undefined);

    render(<MatchesList gameId="game-1" />);
    
    // Click synchronize
    fireEvent.press(screen.getByTestId('synchronize-button'));
    
    // Wait for error handling - should show failure modal
    await waitFor(() => {
      expect(screen.getByTestId('bet-sync-modal')).toBeTruthy();
      expect(screen.getByTestId('failure-message')).toBeTruthy();
      expect(screen.getByTestId('close-button')).toBeTruthy();
    });
  });

  it('should show loading state during sync', async () => {
    const syncOpportunity = {
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1
        }
      ]
    };

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);

    // Mock a delayed submitBet
    mockSubmitBet.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));

    render(<MatchesList gameId="game-1" />);
    
    // Click synchronize
    fireEvent.press(screen.getByTestId('synchronize-button'));
    
    // Should show loading state
    expect(screen.getByText('Loading...')).toBeTruthy();
    expect(screen.getByTestId('synchronize-button')).toBeDisabled();
  });

  it('should close modal after successful sync', async () => {
    const syncOpportunity = {
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1
        }
      ]
    };

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);

    mockSubmitBet.mockResolvedValue(undefined);
    mockRefresh.mockResolvedValue(undefined);

    render(<MatchesList gameId="game-1" />);
    
    // Click synchronize
    fireEvent.press(screen.getByTestId('synchronize-button'));
    
    // Wait for sync to complete - modal should close immediately for success
    await waitFor(() => {
      expect(screen.queryByTestId('bet-sync-modal')).toBeNull();
    });
  });

  it('should close modal when not now is clicked', () => {
    const syncOpportunity = {
      sourceGameId: 'game-2',
      sourceGameName: 'Game 2',
      matchesToSync: [
        {
          matchId: 'match-1',
          homeTeam: 'Team A',
          awayTeam: 'Team B',
          matchday: 1,
          predictedHomeGoals: 2,
          predictedAwayGoals: 1
        }
      ]
    };

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity,
      loading: false,
      error: null,
      refetch: jest.fn()
    } as any);

    render(<MatchesList gameId="game-1" />);
    
    // Click not now
    fireEvent.press(screen.getByTestId('not-now-button'));
    
    // Modal should be closed
    expect(screen.queryByTestId('bet-sync-modal')).toBeNull();
  });

  // New tests for improved error handling
  describe('Partial Success Scenarios', () => {
    it('should show partial success modal when some bets succeed and some fail', async () => {
      const syncOpportunity = {
        sourceGameId: 'game-2',
        sourceGameName: 'Game 2',
        matchesToSync: [
          {
            matchId: 'match-1',
            homeTeam: 'Team A',
            awayTeam: 'Team B',
            matchday: 1,
            predictedHomeGoals: 2,
            predictedAwayGoals: 1
          },
          {
            matchId: 'match-2',
            homeTeam: 'Team C',
            awayTeam: 'Team D',
            matchday: 2,
            predictedHomeGoals: 1,
            predictedAwayGoals: 0
          }
        ]
      };

      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity,
        loading: false,
        error: null,
        refetch: jest.fn()
      } as any);

      // First bet succeeds, second fails
      mockSubmitBet
        .mockResolvedValueOnce(undefined) // First call succeeds
        .mockRejectedValueOnce(new Error('Network error')); // Second call fails
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);
      
      // Click synchronize
      fireEvent.press(screen.getByTestId('synchronize-button'));
      
      // Wait for partial success modal
      await waitFor(() => {
        expect(screen.getByTestId('bet-sync-modal')).toBeTruthy();
        expect(screen.getByTestId('partial-success-message')).toBeTruthy();
        expect(screen.getByTestId('successful-bets')).toBeTruthy();
        expect(screen.getByTestId('failed-bets')).toBeTruthy();
        expect(screen.getByTestId('retry-failed-button')).toBeTruthy();
      });

      // Check that successful and failed bets are displayed
      expect(screen.getByTestId('successful-bet-0')).toHaveTextContent('Team A vs Team B');
      expect(screen.getByTestId('failed-bet-0')).toHaveTextContent('Team C vs Team D');
    });

    it('should allow retrying failed bets', async () => {
      const syncOpportunity = {
        sourceGameId: 'game-2',
        sourceGameName: 'Game 2',
        matchesToSync: [
          {
            matchId: 'match-1',
            homeTeam: 'Team A',
            awayTeam: 'Team B',
            matchday: 1,
            predictedHomeGoals: 2,
            predictedAwayGoals: 1
          },
          {
            matchId: 'match-2',
            homeTeam: 'Team C',
            awayTeam: 'Team D',
            matchday: 2,
            predictedHomeGoals: 1,
            predictedAwayGoals: 0
          }
        ]
      };

      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity,
        loading: false,
        error: null,
        refetch: jest.fn()
      } as any);

      // First bet succeeds, second fails
      mockSubmitBet
        .mockResolvedValueOnce(undefined) // First call succeeds
        .mockRejectedValueOnce(new Error('Network error')); // Second call fails
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);
      
      // Click synchronize
      fireEvent.press(screen.getByTestId('synchronize-button'));
      
      // Wait for partial success modal
      await waitFor(() => {
        expect(screen.getByTestId('retry-failed-button')).toBeTruthy();
      });

      // Now retry the failed bet - this time it succeeds
      mockSubmitBet.mockResolvedValueOnce(undefined);
      
      // Click retry failed
      fireEvent.press(screen.getByTestId('retry-failed-button'));
      
      // Wait for retry to complete - modal should close (all successful now)
      await waitFor(() => {
        expect(screen.queryByTestId('bet-sync-modal')).toBeNull();
      });
    });

    it('should show failure modal when all bets fail', async () => {
      const syncOpportunity = {
        sourceGameId: 'game-2',
        sourceGameName: 'Game 2',
        matchesToSync: [
          {
            matchId: 'match-1',
            homeTeam: 'Team A',
            awayTeam: 'Team B',
            matchday: 1,
            predictedHomeGoals: 2,
            predictedAwayGoals: 1
          }
        ]
      };

      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity,
        loading: false,
        error: null,
        refetch: jest.fn()
      } as any);

      mockSubmitBet.mockRejectedValue(new Error('Network error'));
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);
      
      // Click synchronize
      fireEvent.press(screen.getByTestId('synchronize-button'));
      
      // Wait for failure modal
      await waitFor(() => {
        expect(screen.getByTestId('bet-sync-modal')).toBeTruthy();
        expect(screen.getByTestId('failure-message')).toBeTruthy();
        expect(screen.getByTestId('close-button')).toBeTruthy();
        expect(screen.queryByTestId('retry-failed-button')).toBeNull();
      });
    });
  });

  describe('Modal Mode Transitions', () => {
    it('should start in initial mode', () => {
      const syncOpportunity = {
        sourceGameId: 'game-2',
        sourceGameName: 'Game 2',
        matchesToSync: [
          {
            matchId: 'match-1',
            homeTeam: 'Team A',
            awayTeam: 'Team B',
            matchday: 1,
            predictedHomeGoals: 2,
            predictedAwayGoals: 1
          }
        ]
      };

      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity,
        loading: false,
        error: null,
        refetch: jest.fn()
      } as any);

      render(<MatchesList gameId="game-1" />);
      
      const modal = screen.getByTestId('bet-sync-modal');
      expect(modal.props['data-mode']).toBe('initial');
      expect(screen.getByTestId('initial-message')).toBeTruthy();
      expect(screen.getByTestId('synchronize-button')).toBeTruthy();
      expect(screen.getByTestId('not-now-button')).toBeTruthy();
    });

    it('should transition to success mode and close modal', async () => {
      const syncOpportunity = {
        sourceGameId: 'game-2',
        sourceGameName: 'Game 2',
        matchesToSync: [
          {
            matchId: 'match-1',
            homeTeam: 'Team A',
            awayTeam: 'Team B',
            matchday: 1,
            predictedHomeGoals: 2,
            predictedAwayGoals: 1
          }
        ]
      };

      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity,
        loading: false,
        error: null,
        refetch: jest.fn()
      } as any);

      mockSubmitBet.mockResolvedValue(undefined);
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);
      
      // Click synchronize
      fireEvent.press(screen.getByTestId('synchronize-button'));
      
      // Modal should close immediately for success (no success message shown)
      await waitFor(() => {
        expect(screen.queryByTestId('bet-sync-modal')).toBeNull();
      });
    });
  });
});
