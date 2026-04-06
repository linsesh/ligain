import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react-native';
import MatchesList from '../MatchesList';
import { useBetSynchronization } from '../../../hooks/useBetSynchronization';
import { useBetSubmission } from '../../../hooks/useBetSubmission';
import { useMatches } from '../../../hooks/useMatches';
import { useAuth } from '../../contexts/AuthContext';
import { SeasonMatch } from '../../types/match';

// Mock the hooks - these are our boundaries
jest.mock('../../../hooks/useBetSynchronization');
jest.mock('../../../hooks/useBetSubmission');
jest.mock('../../../hooks/useMatches');
jest.mock('../../contexts/AuthContext');

// Mock notification hooks
jest.mock('../../hooks/useNotifications', () => ({
  useNotifications: jest.fn(() => ({
    preferences: { enabled: false, permissionGranted: false },
    scheduleNotifications: jest.fn(),
    cancelNotifications: jest.fn(),
    scheduleMatchNotification: jest.fn(),
    cancelMatchNotification: jest.fn(),
  })),
}));

// Mock match notifications hook
jest.mock('../../hooks/useMatchNotifications', () => ({
  useMatchNotifications: jest.fn(() => ({})),
}));

// Mock expo-router
jest.mock('expo-router', () => ({
  useRouter: jest.fn(() => ({
    push: jest.fn(),
    back: jest.fn(),
  })),
}));

// Mock expo-sharing
jest.mock('expo-sharing', () => ({
  isAvailableAsync: jest.fn(() => Promise.resolve(true)),
  shareAsync: jest.fn(() => Promise.resolve()),
}));

// Mock react-native-view-shot
jest.mock('react-native-view-shot', () => {
  const React = require('react');
  const { View } = require('react-native');
  const ViewShot = React.forwardRef(({ children }: any, _ref: any) => React.createElement(View, null, children));
  ViewShot.captureRef = jest.fn(() => Promise.resolve('mock-uri'));
  return ViewShot;
});

// Mock teamLogos utility (has asset imports that Jest can't handle)
jest.mock('../../utils/teamLogos', () => ({
  getTeamLogo: jest.fn(() => null),
  isPngLogo: jest.fn(() => false),
}));

// Mock ShareableMatchResult (imports image assets Jest can't parse)
jest.mock('../ShareableMatchResult', () => 'ShareableMatchResult');

// Mock shareUtils
jest.mock('../../utils/shareUtils', () => ({
  shareMatchResult: jest.fn(),
  shareLeaderboard: jest.fn(),
  captureAndShareWithOptions: jest.fn(),
  formatDateForShare: jest.fn(() => '2024-01-15'),
}));

// Mock expo vector icons
jest.mock('@expo/vector-icons', () => ({
  Ionicons: 'Ionicons',
}));

// Mock react-native components - simple string tags for testing
jest.mock('react-native', () => {
  const RN = jest.requireActual('react-native');
  return {
    ...RN,
    View: 'View',
    Text: 'Text',
    TouchableOpacity: 'TouchableOpacity',
    ScrollView: 'ScrollView',
    TextInput: 'TextInput',
    ActivityIndicator: 'ActivityIndicator',
    KeyboardAvoidingView: 'KeyboardAvoidingView',
    RefreshControl: 'RefreshControl',
    Image: 'Image',
    Animated: {
      ...RN.Animated,
      View: 'Animated.View',
    },
    Modal: ({ children, visible }: any) => visible ? children : null,
    Alert: { alert: jest.fn() },
    Keyboard: {
      addListener: jest.fn(() => ({ remove: jest.fn() })),
      dismiss: jest.fn(),
    },
    useWindowDimensions: () => ({ width: 400, height: 800 }),
    Dimensions: { get: () => ({ width: 400, height: 800 }) },
    StyleSheet: RN.StyleSheet,
    Platform: { OS: 'ios', select: (obj: any) => obj.ios || obj.default },
  };
});

// Mock contexts
jest.mock('../../contexts/GamesContext', () => ({
  useGames: jest.fn(() => ({ games: [] })),
}));

jest.mock('../../contexts/TimeServiceContext', () => ({
  useTimeService: jest.fn(() => ({
    now: () => new Date('2024-01-15T10:00:00Z'),
  })),
}));

// Mock translation hook - return actual English strings
jest.mock('../../hooks/useTranslation', () => ({
  useTranslation: () => ({
    t: (key: string, params?: any) => {
      const translations: Record<string, string> = {
        'betSync.title': 'Synchronize Bets?',
        'betSync.synchronize': 'Synchronize',
        'betSync.notNow': 'Not now',
        'betSync.success.title': 'Synchronization Complete',
        'betSync.success.close': 'Close',
        'betSync.failure.title': 'Synchronization Failed',
        'betSync.failure.message': 'Failed to synchronize any bets. Please try again.',
        'betSync.failure.close': 'Close',
        'betSync.partialSuccess.title': 'Partial Synchronization',
        'betSync.partialSuccess.retryFailed': 'Retry Failed',
        'betSync.partialSuccess.close': 'Close',
        'common.loading': 'Loading...',
        'games.matchday': 'Matchday',
        'games.noBet': 'No prono',
        'games.inProgressTag': 'In progress',
        'games.oddsLegend': 'Odds Legend:',
        'games.clearFavorite': 'Clear Favorite',
        'games.clearFavoriteBonus': 'x1',
        'games.drawBonus': 'Draw Bonus',
        'games.outsiderWinBonus': 'Outsider Win',
        'games.matchdayShortPrefix': 'J',
      };
      let result = translations[key] || key;
      if (params) {
        Object.entries(params).forEach(([k, v]) => {
          result = result.replace(`{{${k}}}`, String(v));
        });
      }
      return result;
    },
  }),
}));

// Mock colors
jest.mock('../../constants/colors', () => ({
  colors: {
    primary: '#007AFF',
    secondary: '#5856D6',
    text: '#000000',
    textSecondary: '#666666',
    background: '#FFFFFF',
    card: '#F2F2F7',
    border: '#C6C6C8',
    success: '#34C759',
    error: '#FF3B30',
    warning: '#FF9500',
  },
}));

// Get mocked functions
const mockUseBetSynchronization = useBetSynchronization as jest.Mock;
const mockUseBetSubmission = useBetSubmission as jest.Mock;
const mockUseMatches = useMatches as jest.Mock;
const mockUseAuth = useAuth as jest.Mock;

// Shared mock functions
const mockSubmitBet = jest.fn();
const mockRefresh = jest.fn();

// Helper to create a minimal SeasonMatch-like mock object
function makeMatch(status: 'finished' | 'scheduled' | 'in-progress', daysFromNow = 0) {
  const date = new Date('2024-01-15T10:00:00Z');
  date.setDate(date.getDate() + daysFromNow);
  return {
    id: () => `match-test-${status}`,
    isFinished: () => status === 'finished',
    isInProgress: () => status === 'in-progress',
    hasStarted: () => status !== 'scheduled',
    hasClearFavorite: () => false,
    getFavoriteTeam: () => '',
    getHomeTeam: () => 'Home FC',
    getAwayTeam: () => 'Away FC',
    getHomeGoals: () => 1,
    getAwayGoals: () => 0,
    getHomeTeamOdds: () => 1.8,
    getAwayTeamOdds: () => 2.0,
    getDrawOdds: () => 3.2,
    getDate: () => date,
    getMatchday: () => 19,
    matchday: 19,
  };
}

describe('MatchesList with Bet Synchronization', () => {
  beforeEach(() => {
    jest.clearAllMocks();

    // Setup default mocks
    mockUseAuth.mockReturnValue({
      player: { id: 'player-1', name: 'Test Player' },
      token: 'test-token',
    });

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: mockRefresh,
    });

    mockUseBetSubmission.mockReturnValue({
      submitBet: mockSubmitBet,
      loading: false,
      error: null,
    });

    mockUseBetSynchronization.mockReturnValue({
      syncOpportunity: null,
      loading: false,
      error: null,
      refetch: jest.fn(),
    });
  });

  describe('Modal visibility', () => {
    it('does not show modal when no sync opportunity exists', () => {
      render(<MatchesList gameId="game-1" />);

      expect(screen.queryByText('Synchronize Bets?')).toBeNull();
    });

    it('shows modal when sync opportunity exists', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      render(<MatchesList gameId="game-1" />);

      await waitFor(() => {
        expect(screen.getByText('Synchronize Bets?')).toBeTruthy();
        expect(screen.getByText('Synchronize')).toBeTruthy();
        expect(screen.getByText('Not now')).toBeTruthy();
      });
    });

    it('closes modal when "Not now" is pressed', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      render(<MatchesList gameId="game-1" />);

      // Modal should be visible
      await waitFor(() => {
        expect(screen.getByText('Synchronize Bets?')).toBeTruthy();
      });

      // Press "Not now"
      fireEvent.press(screen.getByText('Not now'));

      // Modal should close
      expect(screen.queryByText('Synchronize Bets?')).toBeNull();
    });

    it('does not show modal again after dismissing for same game', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      const { rerender } = render(<MatchesList gameId="game-1" />);

      // Dismiss modal
      await waitFor(() => {
        expect(screen.getByText('Not now')).toBeTruthy();
      });
      fireEvent.press(screen.getByText('Not now'));

      // Rerender same game - modal should not appear
      rerender(<MatchesList gameId="game-1" />);
      expect(screen.queryByText('Synchronize Bets?')).toBeNull();
    });

    it('shows modal again when switching to different game', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      const { rerender } = render(<MatchesList gameId="game-1" />);

      // Dismiss modal
      await waitFor(() => {
        expect(screen.getByText('Not now')).toBeTruthy();
      });
      fireEvent.press(screen.getByText('Not now'));

      // Switch to different game - modal should appear again
      rerender(<MatchesList gameId="game-3" />);

      await waitFor(() => {
        expect(screen.getByText('Synchronize Bets?')).toBeTruthy();
      });
    });
  });

  describe('Synchronization flow', () => {
    it('submits bets when "Synchronize" is pressed', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
            { matchId: 'match-2', homeTeam: 'Team C', awayTeam: 'Team D', matchday: 2, predictedHomeGoals: 1, predictedAwayGoals: 0 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      mockSubmitBet.mockResolvedValue(undefined);
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);

      await waitFor(() => {
        expect(screen.getByText('Synchronize')).toBeTruthy();
      });

      fireEvent.press(screen.getByText('Synchronize'));

      await waitFor(() => {
        expect(mockSubmitBet).toHaveBeenCalledTimes(2);
      });

      expect(mockSubmitBet).toHaveBeenCalledWith('match-1', 2, 1);
      expect(mockSubmitBet).toHaveBeenCalledWith('match-2', 1, 0);
      expect(mockRefresh).toHaveBeenCalled();
    });

    it('shows loading state during sync', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      // Make submitBet hang to test loading state
      mockSubmitBet.mockImplementation(() => new Promise(() => {}));

      render(<MatchesList gameId="game-1" />);

      await waitFor(() => {
        expect(screen.getByText('Synchronize')).toBeTruthy();
      });

      fireEvent.press(screen.getByText('Synchronize'));

      await waitFor(() => {
        expect(screen.getByText('Loading...')).toBeTruthy();
      });
    });

    it('closes modal after successful sync', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      mockSubmitBet.mockResolvedValue(undefined);
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);

      await waitFor(() => {
        expect(screen.getByText('Synchronize')).toBeTruthy();
      });

      fireEvent.press(screen.getByText('Synchronize'));

      await waitFor(() => {
        expect(screen.queryByText('Synchronize Bets?')).toBeNull();
      });
    });
  });

  describe('Error handling', () => {
    it('shows failure modal when all bets fail', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      mockSubmitBet.mockRejectedValue(new Error('Network error'));
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);

      await waitFor(() => {
        expect(screen.getByText('Synchronize')).toBeTruthy();
      });

      fireEvent.press(screen.getByText('Synchronize'));

      await waitFor(() => {
        expect(screen.getByText('Synchronization Failed')).toBeTruthy();
      });
    });

    it('shows partial success when some bets fail', async () => {
      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
            { matchId: 'match-2', homeTeam: 'Team C', awayTeam: 'Team D', matchday: 2, predictedHomeGoals: 1, predictedAwayGoals: 0 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      // First succeeds, second fails
      mockSubmitBet
        .mockResolvedValueOnce(undefined)
        .mockRejectedValueOnce(new Error('Network error'));
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);

      await waitFor(() => {
        expect(screen.getByText('Synchronize')).toBeTruthy();
      });

      fireEvent.press(screen.getByText('Synchronize'));

      await waitFor(() => {
        expect(screen.getByText('Partial Synchronization')).toBeTruthy();
        expect(screen.getByText('Retry Failed')).toBeTruthy();
      });
    });

    it('allows retrying failed bets', async () => {

      mockUseBetSynchronization.mockReturnValue({
        syncOpportunity: {
          sourceGameId: 'game-2',
          sourceGameName: 'Game 2',
          matchesToSync: [
            { matchId: 'match-1', homeTeam: 'Team A', awayTeam: 'Team B', matchday: 1, predictedHomeGoals: 2, predictedAwayGoals: 1 },
            { matchId: 'match-2', homeTeam: 'Team C', awayTeam: 'Team D', matchday: 2, predictedHomeGoals: 1, predictedAwayGoals: 0 },
          ],
        },
        loading: false,
        error: null,
        refetch: jest.fn(),
      });

      // First sync: first succeeds, second fails
      mockSubmitBet
        .mockResolvedValueOnce(undefined)
        .mockRejectedValueOnce(new Error('Network error'));
      mockRefresh.mockResolvedValue(undefined);

      render(<MatchesList gameId="game-1" />);

      await waitFor(() => {
        expect(screen.getByText('Synchronize')).toBeTruthy();
      });

      fireEvent.press(screen.getByText('Synchronize'));

      await waitFor(() => {
        expect(screen.getByText('Retry Failed')).toBeTruthy();
      });

      // Retry succeeds
      mockSubmitBet.mockResolvedValueOnce(undefined);

      fireEvent.press(screen.getByText('Retry Failed'));

      // Modal should close after successful retry
      await waitFor(() => {
        expect(screen.queryByText('Partial Synchronization')).toBeNull();
      });
    });
  });
});

describe('MatchesList - points badge display', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({ player: { id: 'player-1', name: 'Test Player' }, token: 'test-token' });
    mockUseBetSubmission.mockReturnValue({ submitBet: jest.fn(), loading: false, error: null });
    mockUseBetSynchronization.mockReturnValue({ syncOpportunity: null, loading: false, error: null, refetch: jest.fn() });
  });

  it('shows positive points badge in green for a won finished match', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {
        'match-1': {
          match: makeMatch('finished'),
          bets: null,
          scores: { 'player-1': { playerId: 'player-1', playerName: 'Test Player', points: 3 } },
          playerBetStatuses: null,
        },
      },
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" />);
    await waitFor(() => expect(screen.getByText('+3 points')).toBeTruthy());
  });

  it('shows 0 points badge in red for a finished match with wrong prediction', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {
        'match-1': {
          match: makeMatch('finished'),
          bets: null,
          scores: { 'player-1': { playerId: 'player-1', playerName: 'Test Player', points: 0 } },
          playerBetStatuses: null,
        },
      },
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" />);
    await waitFor(() => expect(screen.getByText('0 points')).toBeTruthy());
  });

  it('shows "No prono" badge when finished match has no score entry for player', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {
        'match-1': {
          match: makeMatch('finished'),
          bets: null,
          scores: null,
          playerBetStatuses: null,
        },
      },
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" />);
    await waitFor(() => expect(screen.getByText('No prono')).toBeTruthy());
  });

  it('shows "No prono" badge for a future match with no bet placed', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': {
          match: makeMatch('scheduled', 3),
          bets: null,
          scores: null,
          playerBetStatuses: null,
        },
      },
      pastMatches: {},
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" />);
    await waitFor(() => expect(screen.getByText('No prono')).toBeTruthy());
  });

  it('does not show "No prono" badge for a future match where the player has already bet', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': {
          match: makeMatch('scheduled', 3),
          bets: {
            'player-1': { playerId: 'player-1', playerName: 'Test Player', predictedHomeGoals: 2, predictedAwayGoals: 1, isModifiable: () => true },
          },
          scores: null,
          playerBetStatuses: null,
        },
      },
      pastMatches: {},
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" />);
    await waitFor(() => expect(screen.queryByText('No prono')).toBeNull());
  });
});
