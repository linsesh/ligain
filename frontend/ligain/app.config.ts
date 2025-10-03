import { ExpoConfig, ConfigContext } from 'expo/config';

export default ({ config }: ConfigContext): ExpoConfig => ({
  ...config,
  name: 'Ligain',
  slug: 'ligain',
  version: '1.0.0',
  orientation: 'portrait',
  icon: './assets/images/icon.png',
  userInterfaceStyle: 'light',
  newArchEnabled: true,
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
      '@react-native-firebase/app',
      {
        iosGoogleServicesFile: './ios/GoogleService-Info.plist',
        androidGoogleServicesFile: './android/app/google-services.json'
      }
    ],
    '@react-native-firebase/crashlytics'
  ],
  splash: {
    image: './assets/images/splash-icon.png',
    resizeMode: 'contain',
    backgroundColor: '#25292e',
  },
  ios: {
    bundleIdentifier: 'com.ligain.app',
    supportsTablet: true,
    buildNumber: '28',
  },
  android: {
    package: 'com.ligain.app',
    versionCode: 7,
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
    eas: {
      projectId: '55e40a95-c826-4c19-9d13-dfe6b906019b'
    }
  },
}); 