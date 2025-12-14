import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import MatchesList from '../../../app/(tabs)/games/game/_MatchesList';
import { useBetSynchronization } from '../../../hooks/useBetSynchronization';
import { useBetSubmission } from '../../../hooks/useBetSubmission';
import { useMatches } from '../../../hooks/useMatches';
import { useAuth } from '../../contexts/AuthContext';

// Mock the hooks
jest.mock('../../../hooks/useBetSynchronization');
jest.mock('../../../hooks/useBetSubmission');
jest.mock('../../../hooks/useMatches');
jest.mock('../../contexts/AuthContext');
jest.mock('react-i18next', () => ({
  useTranslation: jest.fn(() => ({
    t: (key: string) => key,
    i18n: { language: 'en' },
  })),
  initReactI18next: {
    type: '3rdParty',
    init: jest.fn(),
  },
}));
jest.mock('../../hooks/useMatchNotifications', () => ({
  useMatchNotifications: jest.fn(() => ({
    scheduleNotifications: jest.fn(),
    cancelNotifications: jest.fn(),
  })),
}));
jest.mock('../../components/BetSyncModal', () => {
  const React = require('react');
  return {
    BetSyncModal: ({ visible, syncOpportunity, onSynchronize, onNotNow, onRetryFailed, loading, mode = 'initial', syncResult }: any) => {
      console.log('[MOCK] BetSyncModal rendered:', { visible, hasOpportunity: !!syncOpportunity, mode });
      if (!syncOpportunity || !visible) {
        console.log('[MOCK] BetSyncModal returning null:', { syncOpportunity: !!syncOpportunity, visible });
        return null;
      }
      return React.createElement('div', { 'data-testid': 'bet-sync-modal', 'data-mode': mode },
        React.createElement('div', { 'data-testid': `${mode}-message` }, 'Modal Content'),
        React.createElement('button', {
          'data-testid': 'synchronize-button',
          onClick: onSynchronize,
          disabled: loading
        }, loading ? 'Loading...' : 'Synchronize'),
        React.createElement('button', { 'data-testid': 'not-now-button', onClick: onNotNow }, 'Not now'),
        mode === 'partialSuccess' && onRetryFailed && React.createElement('button', {
          'data-testid': 'retry-failed-button',
          onClick: onRetryFailed
        }, 'Retry Failed')
      );
    },
  };
});

// Mock expo-router
jest.mock('expo-router', () => ({
  useRouter: jest.fn(() => ({
    push: jest.fn(),
    replace: jest.fn(),
    back: jest.fn(),
  })),
}));

// Mock expo vector icons
jest.mock('@expo/vector-icons', () => ({
  Ionicons: 'Ionicons',
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
jest.mock('../../components/StatusTag', () => ({
  __esModule: true,
  default: function MockStatusTag() {
    return null;
  },
}));
jest.mock('../../components/ShareableMatchResult', () => ({
  __esModule: true,
  default: function MockShareableMatchResult() {
    return null;
  },
}));
jest.mock('../../utils/shareUtils', () => ({
  captureAndShareWithOptions: jest.fn(),
  formatDateForShare: jest.fn(),
}));
jest.mock('../../utils/dateUtils', () => ({
  formatTime: jest.fn(),
  formatDate: jest.fn(),
}));
jest.mock('../../utils/teamLogos', () => ({
  getTeamLogo: jest.fn(() => {
    // Return a mock React component
    return function MockTeamLogo() {
      return null;
    };
  }),
}));
jest.mock('../../constants/colors', () => ({
  colors: {
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
  }
}));
jest.mock('../../constants/sharedStyles', () => ({
  sharedStyles: {
    shareButton: {
      backgroundColor: '#007AFF',
      padding: 8,
      borderRadius: 4,
    },
  },
}));
jest.mock('react-native-view-shot', () => ({
  __esModule: true,
  default: function MockViewShot({ children }: any) {
    return children;
  },
}));

const mockUseBetSynchronization = useBetSynchronization as jest.MockedFunction<typeof useBetSynchronization>;
const mockUseBetSubmission = useBetSubmission as jest.MockedFunction<typeof useBetSubmission>;
const mockUseMatches = useMatches as jest.MockedFunction<typeof useMatches>;
const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

describe('MatchesList with Bet Synchronization', () => {
  const mockPlayer = { id: 'player-1', name: 'Test Player' };
  const mockSubmitBet = jest.fn();
  const mockRefresh = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();

    mockUseAuth.mockReturnValue({ player: mockPlayer } as any);
    
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

  it('should show modal when sync opportunity exists', async () => {
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

    // Wait for the modal to appear (due to useEffect state updates)
    await waitFor(() => {
      expect(screen.getByTestId('bet-sync-modal')).toBeTruthy();
    });
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
