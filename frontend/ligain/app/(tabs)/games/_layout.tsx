import React from 'react';
import { Text, TouchableOpacity } from 'react-native';
import { Stack, router } from 'expo-router';
import Ionicons from '@expo/vector-icons/Ionicons';
import { colors } from '../../../src/constants/colors';
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
          headerBackVisible: false,
          headerLeft: () => (
            <TouchableOpacity
              onPress={() => {
                try {
                  router.replace('/(tabs)/index');
                } catch (error) {
                  console.warn('Navigation error:', error);
                  // Fallback navigation
                  router.back();
                }
              }}
              style={{ flexDirection: 'row', alignItems: 'center', paddingHorizontal: 12, paddingVertical: 6 }}
              accessibilityRole="button"
              accessibilityLabel={t('games.back')}
            >
              <Ionicons name="chevron-back" size={22} color="#fff" />
              <Text style={{ color: '#fff', fontSize: 16, marginLeft: 2 }}>{t('games.back')}</Text>
            </TouchableOpacity>
          ),
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
