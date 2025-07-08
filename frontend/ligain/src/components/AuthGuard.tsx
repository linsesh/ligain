import React, { useEffect } from 'react';
import { View, ActivityIndicator, StyleSheet } from 'react-native';
import { useRouter } from 'expo-router';
import { useAuth } from '../contexts/AuthContext';
import { colors } from '../constants/colors';

interface AuthGuardProps {
  children: React.ReactNode;
}

export const AuthGuard: React.FC<AuthGuardProps> = ({ children }) => {
  const { player, isLoading } = useAuth();
  const router = useRouter();

  console.log('🔐 AuthGuard - isLoading:', isLoading, 'player:', player ? 'exists' : 'null');

  useEffect(() => {
    console.log('🔄 AuthGuard useEffect - isLoading:', isLoading, 'player:', player ? 'exists' : 'null');
    
    if (!isLoading) {
      if (player) {
        console.log('✅ AuthGuard - User authenticated, navigating to /(tabs)');
        // User is authenticated, navigate to main app
        router.replace('/(tabs)');
      } else {
        console.log('❌ AuthGuard - User not authenticated, navigating to /signin');
        // User is not authenticated, navigate to sign in
        router.replace('/signin');
      }
    }
  }, [player, isLoading, router]);

  if (isLoading) {
    console.log('⏳ AuthGuard - Showing loading spinner');
    return (
      <View style={[styles.loadingContainer, { backgroundColor: colors.background }]}>
        <ActivityIndicator size="large" color={colors.link} />
      </View>
    );
  }

  console.log('✅ AuthGuard - Rendering children (allowing navigation to work)');
  return <>{children}</>;
};

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
}); 