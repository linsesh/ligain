import { Stack } from 'expo-router';
import { useTranslation } from 'react-i18next';

export default function GamesStackLayout() {
  const { t } = useTranslation();
  return (
    <Stack screenOptions={{ headerShown: true }}>
      <Stack.Screen 
        name="game/overview/[id]" 
        options={{ 
          title: t('games.gameOverview'),
          headerShown: true,
          headerBackTitle: t('games.back'),
          headerStyle: {
            backgroundColor: '#25292e',
          },
          headerTintColor: '#fff',
          headerTitleStyle: {
            color: '#fff',
          },
        }} 
      />
    </Stack>
  );
}
