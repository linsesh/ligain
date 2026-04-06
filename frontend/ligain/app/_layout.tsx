import '../global.css';
import { Stack } from 'expo-router';
import { useFonts } from 'expo-font';
import * as SplashScreen from 'expo-splash-screen';
import { useEffect } from 'react';
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
import { PortalHost } from '@rn-primitives/portal';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { View } from 'react-native';
import { useGridCellSize } from '../src/hooks/useGridCellSize';

SplashScreen.preventAutoHideAsync();

export default function Layout() {
  const [fontsLoaded, fontError] = useFonts({
    'HKGroteskWide-Light': require('../assets/fonts/hkgroteskwide-light.otf'),
    'HKGroteskWide-Regular': require('../assets/fonts/hkgroteskwide-regular.otf'),
    'HKGroteskWide-Medium': require('../assets/fonts/hkgroteskwide-medium.otf'),
    'HKGroteskWide-SemiBold': require('../assets/fonts/hkgroteskwide-semibold.otf'),
    'HKGroteskWide-Bold': require('../assets/fonts/hkgroteskwide-bold.otf'),
    'HKGroteskWide-ExtraBold': require('../assets/fonts/hkgroteskwide-extrabold.otf'),
    'HKGroteskWide-Black': require('../assets/fonts/hkgroteskwide-black.otf'),
  });

  useEffect(() => {
    if (fontsLoaded || fontError) {
      SplashScreen.hideAsync();
    }
  }, [fontsLoaded, fontError]);

  const insets = useSafeAreaInsets();
  const cellSize = useGridCellSize();
  const alignedTop = Math.ceil(insets.top / cellSize) * cellSize;

  if (!fontsLoaded && !fontError) return null;

  return (
    <GestureHandlerRootView style={{ flex: 1 }}>
      <GridBackground />
      <I18nextProvider i18n={i18n}>
          <UpdateRequiredProvider>
            <UIEventProvider>
              <ApiProvider>
                <AuthProvider>
                  <TimeServiceProvider service={new RealTimeService()}>
                    <GamesProvider>
                      <AuthGuard>
                        <View style={{ flex: 1, backgroundColor: 'transparent', paddingTop: alignedTop }}>
                          <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: 'transparent' } }}>
                            <Stack.Screen name="signin" />
                            <Stack.Screen name="(tabs)" />
                            <Stack.Screen name="about" />
                            <Stack.Screen name="game/[id]" />
                            <Stack.Screen name="match/[id]" />
                          </Stack>
                        </View>
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
    </GestureHandlerRootView>
  );
}
