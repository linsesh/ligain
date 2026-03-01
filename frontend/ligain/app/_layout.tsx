import '../global.css';
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
import { ApiProvider } from '../src/api';
import { UpdateRequiredProvider } from '../src/contexts/UpdateRequiredContext';
import { UpdateRequiredModal } from '../src/components/UpdateRequiredModal';
import { GridBackground } from '../src/components/GridBackground';
import { GridFadeProvider } from '../src/contexts/GridFadeContext';
import { PortalHost } from '@rn-primitives/portal';
import { SafeAreaView } from 'react-native-safe-area-context';

export default function Layout() {
  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <GridFadeProvider>
        <GridBackground />
        <I18nextProvider i18n={i18n}>
          <UpdateRequiredProvider>
            <UIEventProvider>
              <ApiProvider>
                <AuthProvider>
                  <TimeServiceProvider service={new RealTimeService()}>
                    <GamesProvider>
                      <AuthGuard>
                        <SafeAreaView style={{ flex: 1, backgroundColor: 'transparent' }} edges={['top']}>
                          <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: 'transparent' } }}>
                            <Stack.Screen name="signin" />
                            <Stack.Screen name="(tabs)" />
                            <Stack.Screen name="about" />
                            <Stack.Screen name="game/[id]" />
                          </Stack>
                        </SafeAreaView>
                      </AuthGuard>
                    </GamesProvider>
                  </TimeServiceProvider>
                </AuthProvider>
              </ApiProvider>
            </UIEventProvider>
            <UpdateRequiredModal />
          </UpdateRequiredProvider>
        </I18nextProvider>
        <PortalHost />
      </GridFadeProvider>
    </GestureHandlerRootView>
  );
}
