import { ExpoConfig, ConfigContext } from 'expo/config';

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: 'Ligain',
  slug: 'ligain',
  version: '1.3.0',
  orientation: 'portrait',
  icon: './assets/images/icon.png',
  userInterfaceStyle: 'light',
  newArchEnabled: false,
  scheme: 'ligain',
  plugins: [
    'expo-router',
    [
      '@react-native-google-signin/google-signin',
      {
        iosUrlScheme: 'com.googleusercontent.apps.628283030166-unsbr5lm16u1fgps9re6bp46u32l2gh2'
      }
    ],
    'expo-apple-authentication',
    'expo-splash-screen',
    [
      'expo-notifications',
      {
        icon: './assets/images/icon.png',
        color: '#25292e',
        sounds: [],
      }
    ],
    [
      'expo-image-picker',
      {
        photosPermission: 'Allow Ligain to access your photos to set your profile picture.',
        cameraPermission: 'Allow Ligain to access your camera to take a profile picture.'
      }
    ],
  ],
  splash: {
    image: './assets/images/splash-icon.png',
    resizeMode: 'contain',
    backgroundColor: '#25292e',
  },
  ios: {
    bundleIdentifier: 'com.ligain.app',
    supportsTablet: true,
    buildNumber: '3',
    infoPlist: {
      CFBundleLocalizations: ['en', 'fr'],
      CFBundleDevelopmentRegion: 'en',
    },
  },
  locales: {
    fr: './src/i18n/locales/ios-fr.json',
  },
  android: {
    package: 'com.ligain.app',
    versionCode: 8,
    adaptiveIcon: {
      foregroundImage: './assets/images/adaptive-icon.png',
      backgroundColor: '#ffffff',
    },
    edgeToEdgeEnabled: true,
    permissions: [
      'android.permission.INTERNET',
      'android.permission.ACCESS_NETWORK_STATE'
    ],
    splash: {
      image: './assets/images/splash-icon.png',
      resizeMode: 'contain',
      backgroundColor: '#25292e',
    },
  },
  web: {
    favicon: './assets/images/favicon.png',
  },
  extra: {
    apiBaseUrl: process.env.API_BASE_URL || 'https://server-dev-09f6d83f-uyqlakruuq-ew.a.run.app',
    apiKey: process.env.API_KEY, // Required - no default value
    environment: process.env.NODE_ENV || 'development',
    appleClientId: process.env.EXPO_PUBLIC_APPLE_CLIENT_ID, // Apple Sign-In client ID
    mockMode: process.env.MOCK_MODE === 'true', // Enable mock API mode for UI development
    eas: {
      projectId: '55e40a95-c826-4c19-9d13-dfe6b906019b'
    }
  },
}); 