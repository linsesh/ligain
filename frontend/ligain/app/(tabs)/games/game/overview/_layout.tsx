import { Stack } from 'expo-router';

export default function GameOverviewLayout() {
  return (
    <Stack
      screenOptions={{
        headerStyle: {
          backgroundColor: '#25292e',
        },
        headerTintColor: '#fff',
        headerTitleStyle: {
          color: '#fff',
        },
      }}
    >
      <Stack.Screen
        name="[id]"
        options={{
          title: 'Game Overview',
        }}
      />
    </Stack>
  );
} 