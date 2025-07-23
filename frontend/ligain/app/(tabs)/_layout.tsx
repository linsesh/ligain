
import { Tabs } from 'expo-router';
import Ionicons from '@expo/vector-icons/Ionicons';
import { colors } from '../../src/constants/colors';
import { useTranslation } from 'react-i18next';

export default function TabLayout() {
  const { t } = useTranslation();

  return (
    <Tabs
      initialRouteName="matches"
      screenOptions={{
        tabBarActiveTintColor: colors.primary,
        headerStyle: {
          backgroundColor: '#25292e',
        },
        headerShadowVisible: false,
        headerTintColor: '#fff',
        tabBarStyle: {
          backgroundColor: '#25292e',
        },
      }}
    >
      <Tabs.Screen
        name="matches"
        options={{
          title: t('navigation.matches'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'football' : 'football-outline'} color={color} size={24} />
          ),
        }}
      />
      <Tabs.Screen
        name="index"
        options={{
          title: t('navigation.games'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'game-controller' : 'game-controller-outline'} color={color} size={24} />
          ),
        }}
      />
      <Tabs.Screen
        name="games"
        options={{
          href: null, // Hide from tab bar
          headerShown: false, // Hide the tab header
        }}
      />
      <Tabs.Screen
        name="rules"
        options={{
          title: t('navigation.rules'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'book' : 'book-outline'} color={color} size={24}/>
          ),
        }}
      />
      <Tabs.Screen
        name="profile"
        options={{
          title: t('navigation.profile'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'person' : 'person-outline'} color={color} size={24}/>
          ),
        }}
      />
      <Tabs.Screen
        name="about"
        options={{
          title: t('navigation.about'),
          tabBarIcon: ({ color, focused }) => (
            <Ionicons name={focused ? 'information-circle' : 'information-circle-outline'} color={color} size={24}/>
          ),
        }}
      />
    </Tabs>
  );
}
