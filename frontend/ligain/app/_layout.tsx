import { Stack } from 'expo-router';
import { AuthProvider } from '../src/contexts/AuthContext';
import { TimeServiceProvider } from '../src/contexts/TimeServiceContext';
import { RealTimeService } from '../src/services/timeService';
import { AuthGuard } from '../src/components/AuthGuard';
import { I18nextProvider } from 'react-i18next';
import i18n from '../src/i18n';
import { UIEventProvider } from '../src/contexts/UIEventContext';
import { GamesProvider } from '../src/contexts/GamesContext';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { useEffect } from 'react';
import { LogBox } from 'react-native';

// Enable logging in production builds
if (__DEV__) {
  LogBox.ignoreLogs(['Warning: ...']); // Ignore specific warnings
} else {
  // In production, enable console logging to device logs
  console.log = (...args) => {
    // This will show up in device logs and crash reports
    console.info('[LIGAIN]', ...args);
  };
  console.warn = (...args) => {
    console.info('[LIGAIN WARN]', ...args);
  };
  console.error = (...args) => {
    console.info('[LIGAIN ERROR]', ...args);
  };
}

export default function Layout() {
  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <I18nextProvider i18n={i18n}>
        <UIEventProvider>
          <AuthProvider>
            <TimeServiceProvider service={new RealTimeService()}>
              <GamesProvider>
                <AuthGuard>
                  <Stack screenOptions={{ headerShown: false }}>
                    <Stack.Screen name="signin" />
                    <Stack.Screen name="(tabs)" />
                  </Stack>
                </AuthGuard>
              </GamesProvider>
            </TimeServiceProvider>
          </AuthProvider>
        </UIEventProvider>
      </I18nextProvider>
    </GestureHandlerRootView>
  );
}
