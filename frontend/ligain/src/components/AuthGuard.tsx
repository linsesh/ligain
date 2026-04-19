import React, { useEffect } from 'react';
import { useRouter, useSegments } from 'expo-router';
import { useAuth } from '../contexts/AuthContext';
import { MatchesScreenSkeleton } from './MatchesScreenSkeleton';

interface AuthGuardProps {
  children: React.ReactNode;
}

export const AuthGuard: React.FC<AuthGuardProps> = ({ children }) => {
  const { player, isLoading } = useAuth();
  const router = useRouter();
  const segments = useSegments();


  useEffect(() => {    
    if (!isLoading) {
      const inAllowedGroup = segments[0] === '(tabs)' || segments[0] === 'about' || segments[0] === 'game' || segments[0] === 'match' || segments[0] === 'team';

      if (player && !inAllowedGroup) {
        console.log('✅ AuthGuard - User authenticated, navigating to /(tabs)');
        // User is authenticated but not in the main app, navigate to main app
        router.replace('/(tabs)');
      } else if (!player && inAllowedGroup) {
        console.log('❌ AuthGuard - User not authenticated, navigating to /signin');
        // User is not authenticated but in the main app, navigate to sign in
        router.replace('/signin');
      }
    }
  }, [player, isLoading, segments, router]);

  if (isLoading) {
    return <MatchesScreenSkeleton />;
  }

  return <>{children}</>;
};

