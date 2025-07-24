import React, { useEffect } from 'react';
import { View, ActivityIndicator, StyleSheet } from 'react-native';
import { useRouter, useSegments } from 'expo-router';
import { useAuth } from '../contexts/AuthContext';
import { colors } from '../constants/colors';

interface AuthGuardProps {
  children: React.ReactNode;
}

export const AuthGuard: React.FC<AuthGuardProps> = ({ children }) => {
  const { player, isLoading } = useAuth();
  const router = useRouter();
  const segments = useSegments();


  useEffect(() => {    
    if (!isLoading) {
      const inAuthGroup = segments[0] === '(tabs)';
      
      if (player && !inAuthGroup) {
        console.log('✅ AuthGuard - User authenticated, navigating to /(tabs)');
        // User is authenticated but not in the main app, navigate to main app
        router.replace('/(tabs)');
      } else if (!player && inAuthGroup) {
        console.log('❌ AuthGuard - User not authenticated, navigating to /signin');
        // User is not authenticated but in the main app, navigate to sign in
        router.replace('/signin');
      }
    }
  }, [player, isLoading, segments, router]);

  if (isLoading) {
    return (
      <View style={[styles.loadingContainer, { backgroundColor: colors.loadingBackground }]}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  return <>{children}</>;
};

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
}); 