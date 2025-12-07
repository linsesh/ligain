import { renderHook, act } from '@testing-library/react';
import { Alert } from 'react-native';
import { useNotifications } from './useNotifications';
import * as Notifications from 'expo-notifications';
import { getItem, setItem } from '../utils/storage';

// Mock dependencies
jest.mock('expo-notifications');
jest.mock('../utils/storage');
jest.mock('./useTranslation', () => ({
  useTranslation: () => ({
    t: (key: string, params?: any) => {
      const translations: { [key: string]: string } = {
        'notifications.reminderTitle': 'Match Reminder',
        'notifications.reminderBody': '{{homeTeam}} vs {{awayTeam}} starts in 1 hour! Don\'t forget to place your bet.',
      };
      let translation = translations[key] || key;
      if (params) {
        Object.keys(params).forEach(param => {
          translation = translation.replace(new RegExp(`{{${param}}}`, 'g'), params[param]);
        });
      }
      return translation;
    },
  }),
}));
jest.mock('react-native', () => {
  const RN = jest.requireActual('react-native');
  return {
    ...RN,
    Alert: {
      alert: jest.fn(),
    },
    Platform: {
      OS: 'ios',
    },
  };
});

const mockNotifications = Notifications as jest.Mocked<typeof Notifications>;
const mockGetItem = getItem as jest.MockedFunction<typeof getItem>;
const mockSetItem = setItem as jest.MockedFunction<typeof setItem>;
const mockAlert = Alert.alert as jest.MockedFunction<typeof Alert.alert>;

describe('useNotifications', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    // Default mock implementations
    mockGetItem.mockResolvedValue(null);
    mockSetItem.mockResolvedValue(undefined);
    const defaultPermissionResponse = { 
      status: 'undetermined' as Notifications.PermissionStatus,
      granted: false,
      expires: 'never' as const,
      canAskAgain: true,
    };
    mockNotifications.getPermissionsAsync.mockResolvedValue(defaultPermissionResponse);
    mockNotifications.requestPermissionsAsync.mockResolvedValue(defaultPermissionResponse);
    mockNotifications.scheduleNotificationAsync.mockResolvedValue('notification-id-1');
    mockNotifications.getAllScheduledNotificationsAsync.mockResolvedValue([]);
    mockNotifications.cancelScheduledNotificationAsync.mockResolvedValue();
    mockNotifications.cancelAllScheduledNotificationsAsync.mockResolvedValue();
  });

  describe('Initialization', () => {
    it('should load preferences from storage on mount', async () => {
      mockGetItem.mockResolvedValueOnce('true'); // notification_preferences_enabled

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      expect(mockGetItem).toHaveBeenCalledWith('notification_preferences_enabled');
      expect(result.current.preferences.enabled).toBe(true);
    });

    it('should default to disabled if no preference stored', async () => {
      mockGetItem.mockResolvedValueOnce(null);

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      expect(result.current.preferences.enabled).toBe(false);
    });

    it('should check permission status on mount', async () => {
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      expect(mockNotifications.getPermissionsAsync).toHaveBeenCalled();
      expect(result.current.preferences.permissionGranted).toBe(true);
    });
  });

  describe('Permission Management', () => {
    it('should request permissions when called', async () => {
      mockNotifications.requestPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        const granted = await result.current.requestPermissions();
        expect(granted).toBe(true);
      });

      expect(mockNotifications.requestPermissionsAsync).toHaveBeenCalled();
      expect(mockSetItem).toHaveBeenCalledWith('notification_permission_requested', 'true');
      expect(result.current.preferences.permissionGranted).toBe(true);
    });

    it('should not request if already granted', async () => {
      const grantedResponse = { 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      };
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce(grantedResponse);

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      // Mock should return granted status when checked again
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce(grantedResponse);

      await act(async () => {
        const granted = await result.current.requestPermissions();
        expect(granted).toBe(true);
      });

      // Should check but not request again since already granted
      expect(mockNotifications.getPermissionsAsync).toHaveBeenCalled();
      expect(mockNotifications.requestPermissionsAsync).not.toHaveBeenCalled();
    });

    it('should show alert when permission denied on iOS', async () => {
      mockNotifications.requestPermissionsAsync.mockResolvedValueOnce({ 
        status: 'denied' as Notifications.PermissionStatus,
        granted: false,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await result.current.requestPermissions();
      });

      expect(mockAlert).toHaveBeenCalledWith(
        'Permission Required',
        'Please enable notifications in Settings to receive match reminders.',
        [{ text: 'OK' }]
      );
      expect(result.current.preferences.permissionGranted).toBe(false);
    });

    it('should handle permission request errors', async () => {
      mockNotifications.requestPermissionsAsync.mockRejectedValueOnce(new Error('Permission error'));

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        const granted = await result.current.requestPermissions();
        expect(granted).toBe(false);
      });
    });
  });

  describe('Preference Management', () => {
    it('should enable notifications and request permissions', async () => {
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'undetermined' as Notifications.PermissionStatus,
        granted: false,
        expires: 'never' as const,
        canAskAgain: true,
      });
      mockNotifications.requestPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      await act(async () => {
        const success = await result.current.setNotificationEnabled(true);
        expect(success).toBe(true);
      });

      expect(mockSetItem).toHaveBeenCalledWith('notification_preferences_enabled', 'true');
      expect(mockNotifications.requestPermissionsAsync).toHaveBeenCalled();
      expect(result.current.preferences.enabled).toBe(true);
    });

    it('should disable notifications without canceling', async () => {
      mockGetItem.mockResolvedValueOnce('true'); // Start with enabled

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      await act(async () => {
        const success = await result.current.setNotificationEnabled(false);
        expect(success).toBe(true);
      });

      expect(mockSetItem).toHaveBeenCalledWith('notification_preferences_enabled', 'false');
      expect(result.current.preferences.enabled).toBe(false);
      // Should not cancel notifications when disabling
      expect(mockNotifications.cancelAllScheduledNotificationsAsync).not.toHaveBeenCalled();
    });

    it('should revert preference if permission denied when enabling', async () => {
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'undetermined' as Notifications.PermissionStatus,
        granted: false,
        expires: 'never' as const,
        canAskAgain: true,
      });
      mockNotifications.requestPermissionsAsync.mockResolvedValueOnce({ 
        status: 'denied' as Notifications.PermissionStatus,
        granted: false,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      await act(async () => {
        const success = await result.current.setNotificationEnabled(true);
        expect(success).toBe(false);
      });

      // Should revert to disabled
      expect(mockSetItem).toHaveBeenCalledWith('notification_preferences_enabled', 'false');
      expect(result.current.preferences.enabled).toBe(false);
    });

    it('should handle storage errors gracefully', async () => {
      mockSetItem.mockRejectedValueOnce(new Error('Storage error'));

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      await act(async () => {
        const success = await result.current.setNotificationEnabled(true);
        expect(success).toBe(false);
      });
    });
  });

  describe('Notification Scheduling', () => {
    it('should schedule notification for future match', async () => {
      mockGetItem.mockResolvedValueOnce('true'); // notifications enabled
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000); // 2 hours from now

      await act(async () => {
        const notificationId = await result.current.scheduleMatchNotification(
          'match-123',
          futureDate,
          'Team A',
          'Team B'
        );
        expect(notificationId).toBeTruthy();
      });

      expect(mockNotifications.scheduleNotificationAsync).toHaveBeenCalledWith(
        expect.objectContaining({
          content: expect.objectContaining({
            title: 'Match Reminder',
            body: 'Team A vs Team B starts in 1 hour! Don\'t forget to place your bet.',
            data: { matchId: 'match-123', type: 'match_reminder' },
          }),
          trigger: expect.anything(), // Date or DateTrigger object
        })
      );
    });

    it('should not schedule if notifications disabled', async () => {
      mockGetItem.mockResolvedValueOnce('false'); // notifications disabled

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);

      await act(async () => {
        const notificationId = await result.current.scheduleMatchNotification(
          'match-123',
          futureDate,
          'Team A',
          'Team B'
        );
        expect(notificationId).toBeNull();
      });

      expect(mockNotifications.scheduleNotificationAsync).not.toHaveBeenCalled();
    });

    it('should not schedule if permissions not granted', async () => {
      mockGetItem.mockResolvedValueOnce('true');
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'denied' as Notifications.PermissionStatus,
        granted: false,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);

      await act(async () => {
        const notificationId = await result.current.scheduleMatchNotification(
          'match-123',
          futureDate,
          'Team A',
          'Team B'
        );
        expect(notificationId).toBeNull();
      });

      expect(mockNotifications.scheduleNotificationAsync).not.toHaveBeenCalled();
    });

    it('should not schedule for past dates', async () => {
      mockGetItem.mockResolvedValueOnce('true');
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      const pastDate = new Date(Date.now() - 60 * 60 * 1000); // 1 hour ago

      await act(async () => {
        const notificationId = await result.current.scheduleMatchNotification(
          'match-123',
          pastDate,
          'Team A',
          'Team B'
        );
        expect(notificationId).toBeNull();
      });

      expect(mockNotifications.scheduleNotificationAsync).not.toHaveBeenCalled();
    });

    it('should not create duplicate notification if already scheduled for same match', async () => {
      mockGetItem.mockResolvedValueOnce('true'); // notifications enabled
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000); // 2 hours from now
      const matchId = 'match-123';
      
      // First call: no existing notification
      mockNotifications.getAllScheduledNotificationsAsync.mockResolvedValueOnce([]);
      mockNotifications.scheduleNotificationAsync.mockResolvedValueOnce('notification-id-1');

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      // First call - should schedule
      await act(async () => {
        const notificationId = await result.current.scheduleMatchNotification(
          matchId,
          futureDate,
          'Team A',
          'Team B'
        );
        expect(notificationId).toBe('notification-id-1');
      });

      expect(mockNotifications.scheduleNotificationAsync).toHaveBeenCalledTimes(1);

      // Second call for same match: existing notification found
      const existingNotification: Notifications.NotificationRequest = {
        identifier: 'notification-id-1',
        content: {
          title: 'Match Reminder',
          subtitle: null,
          body: 'Team A vs Team B starts in 1 hour! Don\'t forget to place your bet.',
          data: { matchId, type: 'match_reminder' },
          categoryIdentifier: null,
          sound: null,
        },
        trigger: null, // Trigger type doesn't matter for this test - we only check matchId in data
      } as Notifications.NotificationRequest;
      mockNotifications.getAllScheduledNotificationsAsync.mockResolvedValueOnce([existingNotification]);

      // Second call - should return existing notification, not create new one
      await act(async () => {
        const notificationId = await result.current.scheduleMatchNotification(
          matchId,
          futureDate,
          'Team A',
          'Team B'
        );
        expect(notificationId).toBe('notification-id-1');
      });

      // Should still only have been called once (no new notification created)
      expect(mockNotifications.scheduleNotificationAsync).toHaveBeenCalledTimes(1);
      expect(mockNotifications.getAllScheduledNotificationsAsync).toHaveBeenCalledTimes(2);
    });

    it('should not schedule if match starts in less than 1 hour', async () => {
      mockGetItem.mockResolvedValueOnce('true');
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      // Match starts in 30 minutes (less than 1 hour)
      const soonDate = new Date(Date.now() + 30 * 60 * 1000);

      await act(async () => {
        const notificationId = await result.current.scheduleMatchNotification(
          'match-123',
          soonDate,
          'Team A',
          'Team B'
        );
        expect(notificationId).toBeNull();
      });

      expect(mockNotifications.scheduleNotificationAsync).not.toHaveBeenCalled();
    });

    it('should handle scheduling errors gracefully', async () => {
      mockGetItem.mockResolvedValueOnce('true');
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });
      mockNotifications.scheduleNotificationAsync.mockRejectedValueOnce(new Error('Scheduling error'));

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      const futureDate = new Date(Date.now() + 2 * 60 * 60 * 1000);

      await act(async () => {
        const notificationId = await result.current.scheduleMatchNotification(
          'match-123',
          futureDate,
          'Team A',
          'Team B'
        );
        expect(notificationId).toBeNull();
      });
    });
  });

  describe('Notification Cancellation', () => {
    it('should cancel notification for specific match', async () => {
      const mockNotification: Notifications.NotificationRequest = {
        identifier: 'notif-1',
        content: {
          title: 'Test',
          subtitle: null,
          body: 'Test body',
          data: { matchId: 'match-123' },
          categoryIdentifier: null,
          sound: null,
        },
        trigger: null,
      };

      mockNotifications.getAllScheduledNotificationsAsync.mockResolvedValueOnce([mockNotification]);

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await result.current.cancelMatchNotification('match-123');
      });

      expect(mockNotifications.getAllScheduledNotificationsAsync).toHaveBeenCalled();
      expect(mockNotifications.cancelScheduledNotificationAsync).toHaveBeenCalledWith('notif-1');
    });

    it('should handle cancellation when notification not found', async () => {
      mockNotifications.getAllScheduledNotificationsAsync.mockResolvedValueOnce([]);

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await result.current.cancelMatchNotification('match-123');
      });

      expect(mockNotifications.cancelScheduledNotificationAsync).not.toHaveBeenCalled();
    });

    it('should cancel all notifications', async () => {
      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await result.current.cancelAllNotifications();
      });

      expect(mockNotifications.cancelAllScheduledNotificationsAsync).toHaveBeenCalled();
    });

    it('should handle cancellation errors gracefully', async () => {
      mockNotifications.getAllScheduledNotificationsAsync.mockRejectedValueOnce(new Error('Cancellation error'));

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await result.current.cancelMatchNotification('match-123');
      });

      // Should not throw
      expect(result.current).toBeDefined();
    });
  });

  describe('Permission Status Checking', () => {
    it('should update permission status when checked', async () => {
      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      // Mock granted permission for the check
      mockNotifications.getPermissionsAsync.mockResolvedValueOnce({ 
        status: 'granted' as Notifications.PermissionStatus,
        granted: true,
        expires: 'never' as const,
        canAskAgain: false,
      });

      await act(async () => {
        await result.current.checkPermissionStatus();
      });

      expect(mockNotifications.getPermissionsAsync).toHaveBeenCalled();
      expect(result.current.preferences.permissionGranted).toBe(true);
    });

    it('should handle permission check errors', async () => {
      mockNotifications.getPermissionsAsync.mockRejectedValueOnce(new Error('Permission check error'));

      const { result } = renderHook(() => useNotifications());

      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      await act(async () => {
        await result.current.checkPermissionStatus();
      });

      // Should not crash, permission remains false
      expect(result.current.preferences.permissionGranted).toBe(false);
    });
  });
});

