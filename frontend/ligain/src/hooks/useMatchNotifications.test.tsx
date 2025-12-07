import { renderHook, act } from '@testing-library/react';
import { useMatchNotifications } from './useMatchNotifications';
import { useMatches } from '../../hooks/useMatches';
import { useAuth } from '../contexts/AuthContext';
import { useNotifications } from './useNotifications';
import { SeasonMatch } from '../types/match';

// Mock dependencies
jest.mock('../../hooks/useMatches');
jest.mock('../contexts/AuthContext');
jest.mock('./useNotifications');

const mockUseMatches = useMatches as jest.MockedFunction<typeof useMatches>;
const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;
const mockUseNotifications = useNotifications as jest.MockedFunction<typeof useNotifications>;

describe('useMatchNotifications', () => {
  const mockPlayer = {
    id: 'player-123',
    name: 'Test Player',
  };

  const createMockMatch = (matchId: string, date: Date, hasBet: boolean = false) => {
    const match = new SeasonMatch(
      'Team A',
      'Team B',
      0,
      0,
      1.5,
      2.5,
      3.0,
      'scheduled',
      '2024',
      'Ligue 1',
      date,
      1
    );

    return {
      match,
      bets: hasBet ? { 
        [mockPlayer.id]: { 
          playerId: mockPlayer.id,
          playerName: mockPlayer.name,
          predictedHomeGoals: 2, 
          predictedAwayGoals: 1,
          isModifiable: (now: Date) => !match.isFinished() && !match.isInProgress() && now < match.getDate(),
        } 
      } : null,
      scores: null,
    };
  };

  beforeEach(() => {
    jest.clearAllMocks();

    // Default mocks
    mockUseAuth.mockReturnValue({
      player: mockPlayer,
      isLoading: false,
      signIn: jest.fn(),
      signOut: jest.fn(),
      checkAuth: jest.fn(),
      setPlayer: jest.fn(),
      showNameModal: false,
      setShowNameModal: jest.fn(),
      authResult: null,
      setAuthResult: jest.fn(),
      selectedProvider: null,
      setSelectedProvider: jest.fn(),
    } as any);

    mockUseNotifications.mockReturnValue({
      preferences: {
        enabled: true,
        permissionGranted: true,
      },
      isLoading: false,
      requestPermissions: jest.fn().mockResolvedValue(true),
      setNotificationEnabled: jest.fn().mockResolvedValue(true),
      scheduleMatchNotification: jest.fn().mockResolvedValue('notification-id'),
      cancelMatchNotification: jest.fn().mockResolvedValue(undefined),
      cancelAllNotifications: jest.fn().mockResolvedValue(undefined),
      checkPermissionStatus: jest.fn().mockResolvedValue(undefined),
    } as any);

    mockUseMatches.mockReturnValue({
      incomingMatches: {},
      pastMatches: {},
      loading: false,
      error: null,
      refresh: jest.fn(),
    });
  });

  describe('Match Monitoring', () => {
    it('should schedule notification for match without bet', async () => {
      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000); // 2 hours from now
      const mockMatchResult = createMockMatch('match-1', futureDate, false);

      mockUseMatches.mockReturnValue({
        incomingMatches: { 'match-1': mockMatchResult },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockScheduleNotification = jest.fn().mockResolvedValue('notif-id');
      mockUseNotifications.mockReturnValue({
        preferences: { enabled: true, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: mockScheduleNotification,
        cancelMatchNotification: jest.fn(),
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      } as any);

      renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      expect(mockScheduleNotification).toHaveBeenCalledWith(
        'match-1',
        futureDate,
        'Team A',
        'Team B'
      );
    });

    it('should not schedule notification if user has bet', async () => {
      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);
      const mockMatchResult = createMockMatch('match-1', futureDate, true); // has bet

      mockUseMatches.mockReturnValue({
        incomingMatches: { 'match-1': mockMatchResult },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockScheduleNotification = jest.fn();
      mockUseNotifications.mockReturnValue({
        preferences: { enabled: true, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: mockScheduleNotification,
        cancelMatchNotification: jest.fn(),
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      } as any);

      renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      expect(mockScheduleNotification).not.toHaveBeenCalled();
    });

    it('should not schedule if notifications disabled', async () => {
      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);
      const mockMatchResult = createMockMatch('match-1', futureDate, false);

      mockUseMatches.mockReturnValue({
        incomingMatches: { 'match-1': mockMatchResult },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockScheduleNotification = jest.fn();
      mockUseNotifications.mockReturnValue({
        preferences: { enabled: false, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: mockScheduleNotification,
        cancelMatchNotification: jest.fn(),
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      } as any);

      renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      expect(mockScheduleNotification).not.toHaveBeenCalled();
    });
  });

  describe('Bet Detection', () => {
    it('should detect user bet correctly', async () => {
      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);
      const mockMatchResult = createMockMatch('match-1', futureDate, true);

      mockUseMatches.mockReturnValue({
        incomingMatches: { 'match-1': mockMatchResult },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockCancelNotification = jest.fn();
      mockUseNotifications.mockReturnValue({
        preferences: { enabled: true, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: jest.fn(),
        cancelMatchNotification: mockCancelNotification,
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      } as any);

      renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Should cancel notification if user has bet
      expect(mockCancelNotification).toHaveBeenCalledWith('match-1');
    });

    it('should handle matches without bets object', async () => {
      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);
      const match = new SeasonMatch(
        'Team A',
        'Team B',
        0,
        0,
        1.5,
        2.5,
        3.0,
        'scheduled',
        '2024',
        'Ligue 1',
        futureDate,
        1
      );
      const mockMatchResult = {
        match,
        bets: null, // No bets object
        scores: null,
      };

      mockUseMatches.mockReturnValue({
        incomingMatches: { 'match-1': mockMatchResult },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockScheduleNotification = jest.fn().mockResolvedValue('notif-id');
      mockUseNotifications.mockReturnValue({
        preferences: { enabled: true, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: mockScheduleNotification,
        cancelMatchNotification: jest.fn(),
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      } as any);

      renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Should schedule since no bets means user hasn't bet
      expect(mockScheduleNotification).toHaveBeenCalled();
    });
  });

  describe('Cleanup', () => {
    it('should cleanup notifications for past matches', async () => {
      const pastDate = new Date(Date.now() - 60 * 60 * 1000); // 1 hour ago
      const mockMatchResult = createMockMatch('match-1', pastDate, false);

      mockUseMatches.mockReturnValue({
        incomingMatches: { 'match-1': mockMatchResult },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockCancelNotification = jest.fn();
      mockUseNotifications.mockReturnValue({
        preferences: { enabled: true, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: jest.fn(),
        cancelMatchNotification: mockCancelNotification,
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      } as any);

      renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Should cancel notification for past match
      expect(mockCancelNotification).toHaveBeenCalledWith('match-1');
    });

    it('should reset scheduled matches when game changes', () => {
      const { rerender } = renderHook(
        ({ gameId }) => useMatchNotifications(gameId),
        { initialProps: { gameId: 'game-1' } }
      );

      // Change game ID
      rerender({ gameId: 'game-2' });

      // The ref should be cleared (we can't directly test this, but the hook should handle it)
      expect(mockUseMatches).toHaveBeenCalledWith('game-2');
    });
  });

  describe('Integration', () => {
    it('should handle multiple matches correctly', async () => {
      const futureDate1 = new Date(Date.now() + 2 * 60 * 60 * 1000);
      const futureDate2 = new Date(Date.now() + 3 * 60 * 60 * 1000);
      const match1 = createMockMatch('match-1', futureDate1, false);
      const match2 = createMockMatch('match-2', futureDate2, true); // has bet

      mockUseMatches.mockReturnValue({
        incomingMatches: {
          'match-1': match1,
          'match-2': match2,
        },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockScheduleNotification = jest.fn().mockResolvedValue('notif-id');
      const mockCancelNotification = jest.fn();
      mockUseNotifications.mockReturnValue({
        preferences: { enabled: true, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: mockScheduleNotification,
        cancelMatchNotification: mockCancelNotification,
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      } as any);

      renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Should schedule for match-1 (no bet)
      expect(mockScheduleNotification).toHaveBeenCalledWith(
        'match-1',
        futureDate1,
        'Team A',
        'Team B'
      );

      // Should cancel for match-2 (has bet)
      expect(mockCancelNotification).toHaveBeenCalledWith('match-2');
    });

    it('should not process if player not logged in', async () => {
      mockUseAuth.mockReturnValue({
        player: null,
        isLoading: false,
        signIn: jest.fn(),
        signOut: jest.fn(),
        checkAuth: jest.fn(),
        setPlayer: jest.fn(),
        showNameModal: false,
        setShowNameModal: jest.fn(),
        authResult: null,
        setAuthResult: jest.fn(),
        selectedProvider: null,
        setSelectedProvider: jest.fn(),
      } as any);

      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);
      const mockMatchResult = createMockMatch('match-1', futureDate, false);

      mockUseMatches.mockReturnValue({
        incomingMatches: { 'match-1': mockMatchResult },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockScheduleNotification = jest.fn();
      mockUseNotifications.mockReturnValue({
        preferences: { enabled: true, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: mockScheduleNotification,
        cancelMatchNotification: jest.fn(),
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      } as any);

      renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Should not schedule if no player
      expect(mockScheduleNotification).not.toHaveBeenCalled();
    });

    it('should handle preference changes', async () => {
      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);
      const mockMatchResult = createMockMatch('match-1', futureDate, false);

      mockUseMatches.mockReturnValue({
        incomingMatches: { 'match-1': mockMatchResult },
        pastMatches: {},
        loading: false,
        error: null,
        refresh: jest.fn(),
      });

      const mockScheduleNotification = jest.fn().mockResolvedValue('notif-id');
      const mockUseNotificationsReturn = {
        preferences: { enabled: false, permissionGranted: true },
        isLoading: false,
        requestPermissions: jest.fn(),
        setNotificationEnabled: jest.fn(),
        scheduleMatchNotification: mockScheduleNotification,
        cancelMatchNotification: jest.fn(),
        cancelAllNotifications: jest.fn(),
        checkPermissionStatus: jest.fn(),
      };

      mockUseNotifications.mockReturnValue(mockUseNotificationsReturn as any);

      const { rerender } = renderHook(() => useMatchNotifications('game-1'));

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Should not schedule when disabled
      expect(mockScheduleNotification).not.toHaveBeenCalled();

      // Enable notifications
      mockUseNotificationsReturn.preferences.enabled = true;
      mockUseNotifications.mockReturnValue(mockUseNotificationsReturn as any);

      rerender();

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 100));
      });

      // Should schedule when enabled
      expect(mockScheduleNotification).toHaveBeenCalled();
    });
  });
});

