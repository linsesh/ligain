import React from 'react';
import { View, StyleSheet, TouchableOpacity, Text } from 'react-native';
import { useRouter } from 'expo-router';
import { useAuth } from '../src/contexts/AuthContext';
import { useTranslation } from '../src/hooks/useTranslation';

export default function NotFoundScreen() {
  const { player } = useAuth();
  const router = useRouter();
  const { t } = useTranslation();

  const handleGoBack = () => {
    if (player) {
      // User is authenticated, go to main app
      router.replace('/(tabs)');
    } else {
      // User is not authenticated, go to sign in
      router.replace('/signin');
    }
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>{t('notFound.title')}</Text>
      <Text style={styles.subtitle}>{t('notFound.subtitle')}</Text>
      <TouchableOpacity style={styles.button} onPress={handleGoBack}>
        <Text style={styles.buttonText}>
          {player ? t('notFound.goToGames') : t('notFound.goToSignIn')}
        </Text>
      </TouchableOpacity>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#25292e',
    justifyContent: 'center',
    alignItems: 'center',
  },

  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 10,
  },

  subtitle: {
    fontSize: 18,
    color: '#888',
    marginBottom: 20,
  },

  button: {
    backgroundColor: '#007bff',
    paddingVertical: 15,
    paddingHorizontal: 30,
    borderRadius: 8,
  },

  buttonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
});
