import { useEffect, useRef } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { useNotifications } from './useNotifications';
import { MatchResult } from '../types/match';

/**
 * Automatically manages match notifications.
 * Schedules notifications for matches without bets (1 hour before start),
 * cancels when bets are placed, and cleans up past matches.
 * 
 * @param incomingMatches - The incoming matches object from useMatches hook
 * @param gameId - The game ID (used for cleanup when game changes)
 */
export const useMatchNotifications = (
  incomingMatches: { [key: string]: MatchResult },
  gameId: string
) => {
  const { player } = useAuth();
  const {
    preferences,
    scheduleMatchNotification,
    cancelMatchNotification,
  } = useNotifications();

  // Track scheduled matches to prevent duplicates
  const scheduledMatchIdsRef = useRef<Set<string>>(new Set());

  /**
   * Checks if the current user has placed a bet for a match.
   * @param matchResult - The match result containing bets information
   * @returns true if user has a bet, false otherwise
   * @private
   */
  const hasUserBet = (matchResult: MatchResult): boolean => {
    if (!player || !matchResult.bets) {
      return false;
    }
    return matchResult.bets[player.id] !== undefined;
  };

  /**
   * Schedules a notification if conditions are met (enabled, no bet, future match, not already scheduled).
   * @param matchId - Unique identifier for the match
   * @param matchResult - The match result containing match and bet information
   * @private
   */
  const scheduleNotificationIfNeeded = async (
    matchId: string,
    matchResult: MatchResult
  ) => {
    if (scheduledMatchIdsRef.current.has(matchId)) {
      return;
    }

    if (hasUserBet(matchResult)) {
      return;
    }

    if (!preferences.enabled || !preferences.permissionGranted) {
      return;
    }

    const match = matchResult.match;
    const matchDate = match.getDate();

    if (!matchDate || matchDate <= new Date()) {
      return;
    }

    const notificationId = await scheduleMatchNotification(
      matchId,
      matchDate,
      match.getHomeTeam(),
      match.getAwayTeam()
    );

    if (notificationId) {
      scheduledMatchIdsRef.current.add(matchId);
    }
  };

  /**
   * Cancels notification if user has placed a bet.
   * @param matchId - Unique identifier for the match
   * @param matchResult - The match result containing match and bet information
   * @private
   */
  const cancelNotificationIfNeeded = async (
    matchId: string,
    matchResult: MatchResult
  ) => {
    if (hasUserBet(matchResult)) {
      await cancelMatchNotification(matchId);
      scheduledMatchIdsRef.current.delete(matchId);
    }
  };

  /**
   * Cleans up notifications for past matches.
   * @param matchId - Unique identifier for the match
   * @param matchResult - The match result containing match information
   * @private
   */
  const cleanupPastMatchNotification = async (
    matchId: string,
    matchResult: MatchResult
  ) => {
    const match = matchResult.match;
    const matchDate = match.getDate();
    const now = new Date();

    if (matchDate && matchDate < now) {
      await cancelMatchNotification(matchId);
      scheduledMatchIdsRef.current.delete(matchId);
    }
  };

  /**
   * Monitors matches and manages notifications.
   * Runs when incomingMatches, preferences, or player changes.
   * For each match: cleans up past matches, cancels if bet exists, schedules if needed.
   */
  useEffect(() => {
    if (!player) {
      return;
    }

    // Process all matches in parallel, but wait for all operations to complete
    // Using Promise.all ensures we wait for all async operations before the effect completes
    Promise.all(
      Object.entries(incomingMatches).map(async ([matchId, matchResult]) => {
        await cleanupPastMatchNotification(matchId, matchResult);
        await cancelNotificationIfNeeded(matchId, matchResult);
        await scheduleNotificationIfNeeded(matchId, matchResult);
      })
    ).catch((error) => {
      // Handle errors gracefully - don't let notification errors break the app
      console.error('Error processing match notifications:', error);
    });
  }, [incomingMatches, preferences.enabled, preferences.permissionGranted, player, scheduleMatchNotification, cancelMatchNotification]);

  /**
   * Resets scheduled match IDs when game changes.
   */
  useEffect(() => {
    scheduledMatchIdsRef.current.clear();
  }, [gameId]);
};

