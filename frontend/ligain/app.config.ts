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
  plugins: ['expo-router'],
  splash: {
    image: './assets/images/splash-icon.png',
    resizeMode: 'contain',
    backgroundColor: '#ffffff',
  },
  ios: {
    bundleIdentifier: 'com.ligain.app',
    supportsTablet: true,
  },
  android: {
    package: 'com.ligain.app',
    adaptiveIcon: {
      foregroundImage: './assets/images/adaptive-icon.png',
      backgroundColor: '#ffffff',
    },
    edgeToEdgeEnabled: true,
  },
  web: {
    favicon: './assets/images/favicon.png',
  },
  extra: {
    // API Configuration - these must be set in environment variables
    apiBaseUrl: process.env.API_BASE_URL || 'https://server-dev-4c7b2bc-uyqlakruuq-ew.a.run.app',
    apiKey: process.env.API_KEY, // Required - no default value
    
    // Environment
    environment: process.env.NODE_ENV || 'development',
  },
}); 