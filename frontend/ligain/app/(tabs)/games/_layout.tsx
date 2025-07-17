import { Stack } from 'expo-router';

export default function GamesStackLayout() {
  return (
    <Stack screenOptions={{ headerShown: true }}>
      <Stack.Screen 
        name="game/[id]" 
        options={{ 
          title: 'Matches',
          headerShown: true,
          headerBackTitle: 'Back',
          headerStyle: {
            backgroundColor: '#25292e',
          },
          headerTintColor: '#fff',
          headerTitleStyle: {
            color: '#fff',
          },
        }} 
      />
      <Stack.Screen 
        name="game/overview/[id]" 
        options={{ 
          title: 'Game Overview',
          headerShown: true,
          headerBackTitle: 'Back',
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
