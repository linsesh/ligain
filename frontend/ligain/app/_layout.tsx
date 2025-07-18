import { Stack } from 'expo-router';
import { AuthProvider } from '../src/contexts/AuthContext';
import { TimeServiceProvider } from '../src/contexts/TimeServiceContext';
import { RealTimeService } from '../src/services/timeService';
import { AuthGuard } from '../src/components/AuthGuard';
import { I18nextProvider } from 'react-i18next';
import i18n from '../src/i18n';

export default function Layout() {
  console.log('🏗️ Layout - Rendering main layout');
  
  return (
    <I18nextProvider i18n={i18n}>
      <AuthProvider>
        <TimeServiceProvider service={new RealTimeService()}>
          <AuthGuard>
            <Stack screenOptions={{ headerShown: false }}>
              <Stack.Screen name="signin" />
              <Stack.Screen name="(tabs)" />
            </Stack>
          </AuthGuard>
        </TimeServiceProvider>
      </AuthProvider>
    </I18nextProvider>
  );
}
