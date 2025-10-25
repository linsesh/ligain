import { Platform, Linking } from 'react-native';
import * as Sharing from 'expo-sharing';
import { captureRef } from 'react-native-view-shot';
import { Alert } from 'react-native';
import { getCurrentLocale } from '../i18n';

export interface ShareOptions {
  title?: string;
  message?: string;
  url?: string;
}

export interface ShareableData {
  gameName: string;
  players: Array<{
    name: string;
    points: number;
    rank?: number;
  }>;
  matchInfo?: {
    homeTeam: string;
    awayTeam: string;
    homeScore: number;
    awayScore: number;
    date: string;
  };
  period?: string;
}

/**
 * Captures a React component as an image using natural sizing
 */
export const captureComponentAsImage = async (
  ref: React.RefObject<any>,
  options: {
    format?: 'png' | 'jpg';
    quality?: number;
  } = {}
): Promise<string> => {
  try {
    const {
      format = 'png',
      quality = 1,
    } = options;

    // Let react-native-view-shot determine the natural size
    const uri = await captureRef(ref, {
      format,
      quality,
      // Remove width/height to let it use natural dimensions
    });

    return uri;
  } catch (error) {
    console.error('Error capturing component as image:', error);
    throw new Error('Failed to capture image');
  }
};

/**
 * Shares content using the native share dialog with Instagram Stories prioritized
 */
export const shareContent = async (
  uri: string,
  options: ShareOptions = {}
): Promise<void> => {
  try {
    const { title = 'Ligain', message = 'Check out my Ligain results!' } = options;

    if (!(await Sharing.isAvailableAsync())) {
      Alert.alert('Sharing not available', 'Sharing is not available on this device');
      return;
    }

    // Use Instagram Stories UTI to prioritize it in the share sheet
    await Sharing.shareAsync(uri, {
      mimeType: 'image/png',
      dialogTitle: title,
      UTI: 'com.instagram.exclusivegram',
    });
  } catch (error) {
    console.error('Error sharing content:', error);
    Alert.alert('Share failed', 'Failed to share. Please try again.');
  }
};

/**
 * Shares content specifically to Instagram Stories using URL scheme
 */
export const shareToInstagramStories = async (
  uri: string,
  options: ShareOptions = {}
): Promise<void> => {
  try {
    const { title = 'Ligain', message = 'Check out my Ligain results!' } = options;

    // Try Instagram Stories URL scheme first (iOS/Android)
    const instagramUrl = `instagram-stories://share?media_uri=${encodeURIComponent(uri)}`;
    
    const canOpen = await Linking.canOpenURL(instagramUrl);
    if (canOpen) {
      await Linking.openURL(instagramUrl);
      return;
    }

    // Fallback to general sharing with Instagram UTI
    if (!(await Sharing.isAvailableAsync())) {
      Alert.alert('Sharing not available', 'Sharing is not available on this device');
      return;
    }

    await Sharing.shareAsync(uri, {
      mimeType: 'image/png',
      dialogTitle: title,
      UTI: 'com.instagram.exclusivegram',
    });
  } catch (error) {
    console.error('Error sharing to Instagram Stories:', error);
    // Final fallback to general sharing
    await shareContent(uri, options);
  }
};

/**
 * Complete share workflow: capture component and share using natural sizing
 */
export const captureAndShare = async (
  ref: React.RefObject<any>,
  options: {
    shareToInstagram?: boolean;
    title?: string;
    message?: string;
    format?: 'png' | 'jpg';
    quality?: number;
  } = {}
): Promise<void> => {
  try {
    const {
      shareToInstagram = false,
      title = 'Ligain',
      message = 'Check out my Ligain results!',
      format = 'png',
      quality = 1,
    } = options;

    // Capture the component as image using natural sizing
    const uri = await captureComponentAsImage(ref, {
      format,
      quality,
    });

    // Share the image
    if (shareToInstagram) {
      await shareToInstagramStories(uri, { title, message });
    } else {
      await shareContent(uri, { title, message });
    }
  } catch (error) {
    console.error('Error in capture and share workflow:', error);
    Alert.alert('Share failed', 'Failed to generate or share image. Please try again.');
  }
};

/**
 * Share with Instagram Stories prioritized in native share dialog using natural sizing
 */
export const captureAndShareWithOptions = async (
  ref: React.RefObject<any>,
  options: {
    title?: string;
    message?: string;
    format?: 'png' | 'jpg';
    quality?: number;
  } = {}
): Promise<void> => {
  try {
    const {
      title = 'Ligain',
      message = 'Check out my Ligain results!',
      format = 'png',
      quality = 1,
    } = options;

    // Capture the component as image using natural sizing
    const uri = await captureComponentAsImage(ref, {
      format,
      quality,
    });

    // Use the enhanced shareContent that prioritizes Instagram Stories
    await shareContent(uri, { title, message });
  } catch (error) {
    console.error('Error in capture and share workflow:', error);
    Alert.alert('Share failed', 'Failed to generate or share image. Please try again.');
  }
};

/**
 * Format date for display in shared content
 */
export const formatDateForShare = (date: Date): string => {
  return date.toLocaleDateString(getCurrentLocale(), {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
};

/**
 * Format time for display in shared content
 */
export const formatTimeForShare = (date: Date): string => {
  return date.toLocaleTimeString(getCurrentLocale(), {
    hour: '2-digit',
    minute: '2-digit',
  });
};
