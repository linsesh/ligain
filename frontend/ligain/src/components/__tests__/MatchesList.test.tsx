import React from 'react';
import { render, screen, waitFor } from '@testing-library/react-native';
import MatchesList from '../MatchesList';
import { useMatches } from '../../contexts/MatchesContext';
import { useAuth } from '../../contexts/AuthContext';

jest.mock('../../contexts/MatchesContext');
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
    FlatList: ({ data, renderItem }: any) => {
      const React = require('react');
      return React.createElement('View', null, data?.map((item: any, index: number) =>
        renderItem({ item, index, separators: {} })
      ));
    },
    ActivityIndicator: 'ActivityIndicator',
    RefreshControl: 'RefreshControl',
    Image: 'Image',
    Animated: {
      ...RN.Animated,
      View: 'Animated.View',
    },
    Modal: ({ children, visible }: any) => visible ? children : null,
    Alert: { alert: jest.fn() },
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
        'games.delayedMatch': 'Delayed match',
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
const mockUseMatches = useMatches as jest.Mock;
const mockUseAuth = useAuth as jest.Mock;

// Shared mock functions
const mockRefresh = jest.fn();

// Helper to create a minimal SeasonMatch-like mock object
function makeMatch(status: 'finished' | 'scheduled' | 'in-progress', daysFromNow = 0, matchday = 19) {
  const date = new Date('2024-01-15T10:00:00Z');
  date.setDate(date.getDate() + daysFromNow);
  return {
    id: () => `match-test-${status}-${matchday}`,
    isFinished: () => status === 'finished',
    isInProgress: () => status === 'in-progress',
    hasStarted: () => status !== 'scheduled',
    hasClearFavorite: () => false,
    getFavoriteTeam: () => '',
    homeTeamDisplayName: () => 'Home FC',
    awayTeamDisplayName: () => 'Away FC',
    homeTeamName: () => 'Home FC',
    awayTeamName: () => 'Away FC',
    getHomeGoals: () => 1,
    getAwayGoals: () => 0,
    getHomeTeamOdds: () => 1.8,
    getAwayTeamOdds: () => 2.0,
    getDrawOdds: () => 3.2,
    getDate: () => date,
    getMatchday: () => matchday,
    matchday,
  };
}

describe('MatchesList - points badge display', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({ player: { id: 'player-1', name: 'Test Player' }, token: 'test-token' });
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

describe('MatchesList - matchday selector', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({ player: { id: 'player-1', name: 'Test Player' }, token: 'test-token' });
  });

  it('defaults to the last matchday when no initialMatchday is provided', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': { match: makeMatch('scheduled', 1, 1), bets: null, scores: null, playerBetStatuses: null },
        'match-5': { match: makeMatch('scheduled', 5, 5), bets: null, scores: null, playerBetStatuses: null },
      },
      pastMatches: {},
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" />);

    await waitFor(() => {
      expect(screen.getByText('M5').props.className).toContain('font-hk-bold');
    });
    expect(screen.getByText('M1').props.className).not.toContain('font-hk-bold');
  });

  it('uses initialMatchday when provided and valid', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': { match: makeMatch('scheduled', 1, 1), bets: null, scores: null, playerBetStatuses: null },
        'match-5': { match: makeMatch('scheduled', 5, 5), bets: null, scores: null, playerBetStatuses: null },
      },
      pastMatches: {},
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" initialMatchday={1} />);

    await waitFor(() => {
      expect(screen.getByText('M1').props.className).toContain('font-hk-bold');
    });
    expect(screen.getByText('M5').props.className).not.toContain('font-hk-bold');
  });

  it('highlights activeMatchday in primary color when not selected', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': { match: makeMatch('scheduled', 1, 1), bets: null, scores: null, playerBetStatuses: null },
        'match-3': { match: makeMatch('scheduled', 3, 3), bets: null, scores: null, playerBetStatuses: null },
      },
      pastMatches: {},
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" initialMatchday={1} activeMatchday={3} />);

    await waitFor(() => {
      expect(screen.getByText('M3').props.className).toContain('text-primary');
    });
  });

  it('non-active non-selected tab has secondary color', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-1': { match: makeMatch('scheduled', 1, 1), bets: null, scores: null, playerBetStatuses: null },
        'match-2': { match: makeMatch('scheduled', 2, 2), bets: null, scores: null, playerBetStatuses: null },
        'match-3': { match: makeMatch('scheduled', 3, 3), bets: null, scores: null, playerBetStatuses: null },
      },
      pastMatches: {},
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" initialMatchday={1} activeMatchday={3} />);

    await waitFor(() => {
      expect(screen.getByText('M2').props.className).toContain('text-foreground-secondary');
    });
    expect(screen.getByText('M2').props.className).not.toContain('text-primary');
  });
});

describe('MatchesList - match ordering within a matchday', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({ player: { id: 'player-1', name: 'Test Player' }, token: 'test-token' });
  });

  it('displays non-finished matches before finished matches even when the finished match has an earlier kick-off time', async () => {
    // finished match at 14:00 UTC, scheduled match at 16:00 UTC — same matchday
    const finishedMatch = {
      ...makeMatch('finished', 0, 19),
      id: () => 'match-finished-14h',
      getDate: () => new Date('2024-01-15T14:00:00Z'),
      getMatchday: () => 19,
      matchday: 19,
      homeTeamDisplayName: () => 'Finished Home FC',
      homeTeamName: () => 'Finished Home FC',
    };
    const scheduledMatch = {
      ...makeMatch('scheduled', 0, 19),
      id: () => 'match-scheduled-16h',
      getDate: () => new Date('2024-01-15T16:00:00Z'),
      getMatchday: () => 19,
      matchday: 19,
      homeTeamDisplayName: () => 'Scheduled Home FC',
      homeTeamName: () => 'Scheduled Home FC',
    };

    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-scheduled-16h': { match: scheduledMatch, bets: null, scores: null, playerBetStatuses: null },
      },
      pastMatches: {
        'match-finished-14h': { match: finishedMatch, bets: null, scores: { 'player-1': { playerId: 'player-1', playerName: 'Test Player', points: 3 } }, playerBetStatuses: null },
      },
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" initialMatchday={19} />);

    await waitFor(() => {
      expect(screen.getByText('Scheduled Home FC')).toBeTruthy();
      expect(screen.getByText('Finished Home FC')).toBeTruthy();
    });

    // The scheduled match (16:00) must appear before the finished match (14:00)
    // Use the DOM order: the element that comes earlier in the tree has a lower position
    const allTexts = screen.getAllByText(/Home FC/);
    const idxScheduled = allTexts.findIndex(el => el.props.children === 'Scheduled Home FC');
    const idxFinished = allTexts.findIndex(el => el.props.children === 'Finished Home FC');
    expect(idxScheduled).toBeLessThan(idxFinished);
  });
});

describe('MatchesList - delayed match tag', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseAuth.mockReturnValue({ player: { id: 'player-1', name: 'Test Player' }, token: 'test-token' });
  });

  it('shows "Delayed match" tag on a scheduled match when a later matchday has finished matches', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-26': { match: makeMatch('scheduled', 3, 26), bets: null, scores: null, playerBetStatuses: null },
      },
      pastMatches: {
        'match-32': {
          match: makeMatch('finished', 0, 32),
          bets: null,
          scores: { 'player-1': { playerId: 'player-1', playerName: 'Test Player', points: 3 } },
          playerBetStatuses: null,
        },
      },
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" initialMatchday={26} />);
    await waitFor(() => expect(screen.getByText('Delayed match')).toBeTruthy());
  });

  it('does not show "Delayed match" tag when no later matchday has finished matches', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-26': { match: makeMatch('scheduled', 3, 26), bets: null, scores: null, playerBetStatuses: null },
        'match-27': { match: makeMatch('scheduled', 4, 27), bets: null, scores: null, playerBetStatuses: null },
      },
      pastMatches: {},
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" initialMatchday={26} />);
    await waitFor(() => expect(screen.queryByText('Delayed match')).toBeNull());
  });

  it('does not show "Delayed match" tag on a finished match even when later finished matchdays exist', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {
        'match-26': {
          match: makeMatch('finished', 0, 26),
          bets: null,
          scores: { 'player-1': { playerId: 'player-1', playerName: 'Test Player', points: 3 } },
          playerBetStatuses: null,
        },
        'match-30': {
          match: makeMatch('finished', 0, 30),
          bets: null,
          scores: { 'player-1': { playerId: 'player-1', playerName: 'Test Player', points: 3 } },
          playerBetStatuses: null,
        },
      },
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" initialMatchday={26} />);
    await waitFor(() => expect(screen.queryByText('Delayed match')).toBeNull());
  });

  it('does not show "Delayed match" tag on an in-progress match even when later finished matchdays exist', async () => {
    mockUseMatches.mockReturnValue({
      incomingMatches: {
        'match-26': { match: makeMatch('in-progress', 0, 26), bets: null, scores: null, playerBetStatuses: null },
      },
      pastMatches: {
        'match-30': {
          match: makeMatch('finished', 0, 30),
          bets: null,
          scores: { 'player-1': { playerId: 'player-1', playerName: 'Test Player', points: 3 } },
          playerBetStatuses: null,
        },
      },
      loading: false, error: null, refresh: mockRefresh,
    });

    render(<MatchesList gameId="game-1" initialMatchday={26} />);
    await waitFor(() => expect(screen.queryByText('Delayed match')).toBeNull());
  });
});
