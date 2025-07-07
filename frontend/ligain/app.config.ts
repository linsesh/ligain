import { ExpoConfig, ConfigContext } from 'expo/config';
import * as dotenv from 'dotenv';
import * as path from 'path';

// Load .env file explicitly
const result = dotenv.config();
//console.log('Dotenv result:', result);
//console.log('API_KEY loaded:', !!process.env.API_KEY);
//console.log('API_BASE_URL loaded:', !!process.env.API_BASE_URL);

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
    apiBaseUrl: process.env.API_BASE_URL || 'https://server-dev-5be58e3-uyqlakruuq-ew.a.run.app',
    apiKey: process.env.API_KEY, // Required - no default value
    environment: process.env.NODE_ENV || 'development',
  },
}); 