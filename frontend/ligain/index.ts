import 'expo-router/entry';
import Constants from 'expo-constants';
import { initCrashlytics } from './src/utils/crashlytics';

// Initialize Crashlytics collection based on environment
const env = (Constants.expoConfig?.extra as any)?.environment ?? 'development';
initCrashlytics(env);
