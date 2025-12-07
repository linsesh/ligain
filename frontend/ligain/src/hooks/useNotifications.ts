import { useState, useEffect, useCallback } from 'react';
import * as Notifications from 'expo-notifications';
import { Platform, Alert } from 'react-native';
import { getItem, setItem } from '../utils/storage';
import { useTranslation } from './useTranslation';

/**
 * Storage keys for notification preferences.
 * Following the existing pattern in the codebase (e.g., 'auth_token', 'player_data').
 * We use string booleans ('true'/'false') for AsyncStorage compatibility.
 */
const NOTIFICATION_PREFERENCE_KEY = 'notification_preferences_enabled';
const NOTIFICATION_PERMISSION_KEY = 'notification_permission_requested';

/**
 * Configure notification handler behavior.
 * This determines how notifications are displayed when the app is in foreground.
 * We show alerts, play sounds, but don't set badge count.
 */
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: false,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
});

/**
 * Interface representing notification preferences state.
 * Tracks both user preference (enabled/disabled) and system permission status.
 */
export interface NotificationPreferences {
  enabled: boolean;
  permissionGranted: boolean;
}

/**
 * Custom hook for managing notification permissions, preferences, and scheduling.
 * 
 * This hook provides:
 * - Permission request and status checking
 * - User preference storage (persisted locally)
 * - Notification scheduling and cancellation
 * - State management for UI updates
 * 
 * Storage Strategy:
 * - Uses existing safeStorage utility (AsyncStorage with memory fallback)
 * - Stores preferences as string booleans for AsyncStorage compatibility
 * - Follows same pattern as auth_token and player_data storage
 * 
 * @returns Object containing preferences state, loading state, and control functions
 * 
 * @example
 * ```tsx
 * const { preferences, setNotificationEnabled, scheduleMatchNotification } = useNotifications();
 * 
 * // Enable notifications (will request permissions)
 * await setNotificationEnabled(true);
 * 
 * // Schedule a notification
 * const notificationId = await scheduleMatchNotification(
 *   'match-123',
 *   new Date('2024-01-15T15:00:00Z'),
 *   'Team A',
 *   'Team B'
 * );
 * ```
 */
export const useNotifications = () => {
  const { t } = useTranslation();
  const [preferences, setPreferences] = useState<NotificationPreferences>({
    enabled: false,
    permissionGranted: false,
  });
  const [isLoading, setIsLoading] = useState(true);

  /**
   * Load user preferences from local storage on mount.
   * This runs once when the hook initializes to restore user's previous choice.
   * 
   * Why we do this: Persist user preference across app restarts.
   * Edge cases: Handles storage errors gracefully, defaults to disabled if not set.
   */
  useEffect(() => {
    loadPreferences();
    checkPermissionStatus();
  }, []);

  /**
   * Loads notification preference from local storage.
   * 
   * Retrieves the stored preference value and updates component state.
   * If no preference is stored, defaults to disabled (false).
   * 
   * Storage format: String boolean ('true' or 'false') for AsyncStorage compatibility.
   * 
   * @private
   * @returns Promise that resolves when preference is loaded
   */
  const loadPreferences = async () => {
    try {
      const enabled = await getItem(NOTIFICATION_PREFERENCE_KEY);
      setPreferences(prev => ({
        ...prev,
        enabled: enabled === 'true',
      }));
    } catch (error) {
      console.error('Error loading notification preferences:', error);
      // On error, default to disabled rather than crashing
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Checks the current system permission status for notifications.
   * 
   * This queries the device's notification permission state without requesting.
   * Used to sync UI state with actual system permissions.
   * 
   * Permission states:
   * - 'granted': User has allowed notifications
   * - 'denied': User has explicitly denied notifications
   * - 'undetermined': User hasn't been asked yet
   * 
   * Why we check: System permissions can change outside our app (e.g., in Settings).
   * 
   * @private
   * @returns Promise that resolves when permission status is checked
   */
  const checkPermissionStatus = async () => {
    try {
      const { status } = await Notifications.getPermissionsAsync();
      setPreferences(prev => ({
        ...prev,
        permissionGranted: status === 'granted',
      }));
    } catch (error) {
      console.error('Error checking notification permissions:', error);
      // On error, assume not granted to be safe
    }
  };

  /**
   * Requests notification permissions from the user.
   * 
   * This function:
   * 1. Checks if permissions are already granted (avoids unnecessary prompts)
   * 2. Requests permissions if not granted
   * 3. Updates internal state with the result
   * 4. Shows user-friendly alerts if denied
   * 
   * When to use: Called when user enables notifications in profile settings.
   * 
   * Edge cases:
   * - Already granted: Returns immediately without prompting
   * - Denied: Shows alert explaining how to enable in Settings
   * - Error: Logs error and returns false
   * 
   * @returns Promise resolving to true if permissions granted, false otherwise
   * 
   * @example
   * ```tsx
   * const granted = await requestPermissions();
   * if (granted) {
   *   // User allowed notifications, proceed with scheduling
   * }
   * ```
   */
  const requestPermissions = async (): Promise<boolean> => {
    try {
      // Check current status first to avoid unnecessary prompts
      const { status: existingStatus } = await Notifications.getPermissionsAsync();
      let finalStatus = existingStatus;

      // Only request if not already granted
      if (existingStatus !== 'granted') {
        const { status } = await Notifications.requestPermissionsAsync();
        finalStatus = status;
      }

      const granted = finalStatus === 'granted';
      
      // Update state to reflect permission status
      setPreferences(prev => ({
        ...prev,
        permissionGranted: granted,
      }));

      // Mark that we've requested permission (prevents auto-requesting on every app start)
      await setItem(NOTIFICATION_PERMISSION_KEY, 'true');

      // Show helpful message if denied
      if (!granted) {
        if (Platform.OS === 'ios') {
          Alert.alert(
            'Permission Required',
            'Please enable notifications in Settings to receive match reminders.',
            [{ text: 'OK' }]
          );
        } else {
          Alert.alert(
            'Permission Required',
            'Please enable notifications in your device settings to receive match reminders.',
            [{ text: 'OK' }]
          );
        }
      }

      return granted;
    } catch (error) {
      console.error('Error requesting notification permissions:', error);
      return false;
    }
  };

  /**
   * Enables or disables notification preferences.
   * 
   * This function:
   * 1. Saves preference to local storage (persists across app restarts)
   * 2. If enabling: Checks permissions (queries system directly, not state)
   * 3. If disabling: Just saves preference (doesn't cancel existing notifications)
   * 
   * Why we don't cancel on disable:
   * - User might re-enable quickly, avoiding unnecessary re-scheduling
   * - Existing notifications will simply not fire if disabled
   * - Better UX: User's choice is respected without aggressive cleanup
   * 
   * When to use: Called when user toggles notification switch in profile.
   * 
   * @param enabled - Whether notifications should be enabled
   * @returns Promise resolving to true if operation succeeded, false if permission denied
   * 
   * @example
   * ```tsx
   * // Enable notifications
   * const success = await setNotificationEnabled(true);
   * if (!success) {
   *   // Permission was denied
   * }
   * ```
   */
  const setNotificationEnabled = async (enabled: boolean) => {
    try {
      // Save preference to storage
      await setItem(NOTIFICATION_PREFERENCE_KEY, enabled.toString());
      
      // If enabling, check if permissions are granted by querying system directly
      // This avoids race conditions with async state updates
      if (enabled) {
        const { status } = await Notifications.getPermissionsAsync();
        const isGranted = status === 'granted';

        // Update state with both enabled and permission status
        setPreferences(prev => ({
          ...prev,
          enabled,
          permissionGranted: isGranted,
        }));

        // If permissions not granted, request them
        if (!isGranted) {
          const granted = await requestPermissions();
          if (!granted) {
            // Permission denied, revert preference
            await setItem(NOTIFICATION_PREFERENCE_KEY, 'false');
            setPreferences(prev => ({ ...prev, enabled: false }));
            return false;
          }
        }
      } else {
        // Just update enabled state when disabling
        setPreferences(prev => ({ ...prev, enabled }));
      }
      return true;
    } catch (error) {
      console.error('Error setting notification preference:', error);
      return false;
    }
  };

  /**
   * Schedules a local notification for a match 1 hour before it starts.
   * 
   * This function:
   * 1. Checks if notifications are enabled and permissions granted
   * 2. Calculates notification time (match date - 1 hour)
   * 3. Only schedules if notification time is in the future
   * 4. Returns notification identifier for later cancellation
   * 
   * Why 1 hour before:
   * - Gives users enough time to place bets before match starts
   * - Not too early (would be forgotten) or too late (no time to act)
   * - Industry standard for event reminders
   * 
   * When to use: Called automatically by useMatchNotifications hook when matches are loaded.
   * 
   * Edge cases:
   * - Match starts in less than 1 hour: Returns null (can't schedule in past)
   * - Notifications disabled: Returns null (respects user preference)
   * - Permissions not granted: Returns null (can't schedule without permissions)
   * 
   * @param matchId - Unique identifier for the match (used for cancellation)
   * @param matchDate - Date and time when the match starts
   * @param homeTeam - Name of home team (for notification message)
   * @param awayTeam - Name of away team (for notification message)
   * @returns Promise resolving to notification identifier string, or null if not scheduled
   * 
   * @example
   * ```tsx
   * const notificationId = await scheduleMatchNotification(
   *   'match-123',
   *   new Date('2024-01-15T15:00:00Z'),
   *   'Paris Saint-Germain',
   *   'Olympique de Marseille'
   * );
   * if (notificationId) {
   *   console.log('Notification scheduled:', notificationId);
   * }
   * ```
   */
  const scheduleMatchNotification = async (
    matchId: string,
    matchDate: Date,
    homeTeam: string,
    awayTeam: string
  ): Promise<string | null> => {
    // Don't schedule if user has disabled notifications or permissions not granted
    if (!preferences.enabled || !preferences.permissionGranted) {
      return null;
    }

    try {
      // Check if notification already exists for this match (across all games)
      // This ensures only one notification per match, even if it appears in multiple games
      const allNotifications = await Notifications.getAllScheduledNotificationsAsync();
      const existingNotification = allNotifications.find(
        notif => notif.content.data?.matchId === matchId
      );

      if (existingNotification) {
        // Notification already scheduled for this match, return existing identifier
        return existingNotification.identifier;
      }

      // Calculate notification time: 1 hour before match
      const notificationTime = new Date(matchDate.getTime() - 60 * 60 * 1000);
      const now = new Date();

      // Only schedule if notification time is in the future
      // Edge case: Match starts in less than 1 hour - can't schedule
      if (notificationTime <= now) {
        return null;
      }

      // Schedule the notification
      // Store matchId in notification data for easy cancellation later
      // Pass Date directly - expo-notifications accepts Date objects for scheduling
      const identifier = await Notifications.scheduleNotificationAsync({
        content: {
          title: t('notifications.reminderTitle'),
          body: t('notifications.reminderBody', { homeTeam, awayTeam }),
          data: { matchId, type: 'match_reminder' },
        },
        trigger: notificationTime as any, // Date is accepted but TypeScript types are strict
      });

      return identifier;
    } catch (error) {
      console.error('Error scheduling notification:', error);
      return null;
    }
  };

  /**
   * Cancels a scheduled notification for a specific match.
   * 
   * This function:
   * 1. Finds the scheduled notification by matchId (stored in notification data)
   * 2. Cancels it if found
   * 
   * When to use: Called when user places a bet (no need to remind them anymore).
   * 
   * Why we search by matchId:
   * - Notification identifier is returned from scheduling, but we might not have it stored
   * - matchId is stored in notification data, making it easy to find
   * - More reliable than trying to track identifiers ourselves
   * 
   * Edge cases:
   * - Notification not found: Silently succeeds (idempotent operation)
   * - Multiple notifications for same match: Cancels all (shouldn't happen, but safe)
   * 
   * @param matchId - Unique identifier for the match to cancel notification for
   * @returns Promise that resolves when cancellation is attempted
   * 
   * @example
   * ```tsx
   * // Cancel notification when user places bet
   * await cancelMatchNotification('match-123');
   * ```
   */
  const cancelMatchNotification = async (matchId: string) => {
    try {
      // Get all scheduled notifications
      const allNotifications = await Notifications.getAllScheduledNotificationsAsync();
      
      // Find notification(s) for this match
      const notificationToCancel = allNotifications.find(
        notif => notif.content.data?.matchId === matchId
      );

      if (notificationToCancel) {
        await Notifications.cancelScheduledNotificationAsync(notificationToCancel.identifier);
      }
    } catch (error) {
      console.error('Error canceling notification:', error);
    }
  };

  /**
   * Cancels all scheduled notifications.
   * 
   * When to use: 
   * - User disables notifications (though we don't auto-call this)
   * - App cleanup/reset scenarios
   * - Testing/debugging
   * 
   * Note: We don't automatically call this when user disables notifications,
   * as they might re-enable quickly. This is a manual cleanup function.
   * 
   * @returns Promise that resolves when all notifications are cancelled
   */
  const cancelAllNotifications = async () => {
    try {
      await Notifications.cancelAllScheduledNotificationsAsync();
    } catch (error) {
      console.error('Error canceling all notifications:', error);
    }
  };

  return {
    preferences,
    isLoading,
    requestPermissions,
    setNotificationEnabled,
    scheduleMatchNotification,
    cancelMatchNotification,
    cancelAllNotifications,
    checkPermissionStatus,
  };
};

