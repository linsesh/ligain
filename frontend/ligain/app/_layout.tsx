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

export default function Layout() {
  console.log('🏗️ Layout - Rendering main layout');
  
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
