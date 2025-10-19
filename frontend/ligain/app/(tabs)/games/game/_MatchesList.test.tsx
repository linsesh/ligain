import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import MatchesList from './_MatchesList';
import { useBetSynchronization } from '../../../../hooks/useBetSynchronization';
import { useBetSubmission } from '../../../../hooks/useBetSubmission';
import { useMatches } from '../../../../hooks/useMatches';
import { useAuth } from '../../../../src/contexts/AuthContext';
import { useTranslation } from 'react-i18next';

// Mock the hooks
jest.mock('../../../../hooks/useBetSynchronization');
jest.mock('../../../../hooks/useBetSubmission');
jest.mock('../../../../hooks/useMatches');
jest.mock('../../../../src/contexts/AuthContext');
jest.mock('react-i18next');

const mockUseBetSynchronization = useBetSynchronization as jest.MockedFunction<typeof useBetSynchronization>;
const mockUseBetSubmission = useBetSubmission as jest.MockedFunction<typeof useBetSubmission>;
const mockUseMatches = useMatches as jest.MockedFunction<typeof useMatches>;
const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;
const mockUseTranslation = useTranslation as jest.MockedFunction<typeof useTranslation>;

// Mock the BetSyncModal component
jest.mock('../../../../src/components/BetSyncModal', () => {
  return function MockBetSyncModal({ visible, onSynchronize, onNotNow, loading }: any) {
    if (!visible) return null;
    return (
      <div data-testid="bet-sync-modal">
        <button data-testid="synchronize-button" onClick={onSynchronize} disabled={loading}>
          {loading ? 'Loading...' : 'Synchronize'}
        </button>
        <button data-testid="not-now-button" onClick={onNotNow}>
          Not now
        </button>
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

  it('should handle sync errors gracefully', async () => {
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

    // Mock Alert.alert
    const mockAlert = jest.fn();
    jest.doMock('react-native', () => ({
      ...jest.requireActual('react-native'),
      Alert: { alert: mockAlert }
    }));

    render(<MatchesList gameId="game-1" />);
    
    // Click synchronize
    fireEvent.press(screen.getByTestId('synchronize-button'));
    
    // Wait for error handling
    await waitFor(() => {
      expect(mockAlert).toHaveBeenCalledWith('common.error', 'Failed to synchronize bets. Please try again.');
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
    
    // Wait for sync to complete
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
});
