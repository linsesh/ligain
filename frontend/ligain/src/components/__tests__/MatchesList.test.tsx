import React from 'react';
import { render, screen, waitFor } from '@testing-library/react-native';
import MatchesList from '../MatchesList';
import { useBetPlacement } from '../../../hooks/useBetPlacement';
import { useMatches } from '../../../hooks/useMatches';
import { useAuth } from '../../contexts/AuthContext';

// Mock the hooks - these are our boundaries
jest.mock('../../../hooks/useBetPlacement');
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
const mockUseBetPlacement = useBetPlacement as jest.Mock;
const mockUseMatches = useMatches as jest.Mock;
const mockUseAuth = useAuth as jest.Mock;

// Shared mock functions
const mockPlaceBet = jest.fn();
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

describe('MatchesList - points badge display', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({ player: { id: 'player-1', name: 'Test Player' }, token: 'test-token' });
    mockUseBetPlacement.mockReturnValue({ placeBet: mockPlaceBet, isSubmitting: false, error: null, lastFailedMatchId: null });
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
