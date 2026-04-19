import React, { useRef, useEffect } from 'react';
import { View, Animated } from 'react-native';
import { MatchesListSkeleton } from './MatchesList';

export function MatchesScreenSkeleton() {
  const opacity = useRef(new Animated.Value(0.4)).current;

  useEffect(() => {
    Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, { toValue: 1, duration: 800, useNativeDriver: true }),
        Animated.timing(opacity, { toValue: 0.4, duration: 800, useNativeDriver: true }),
      ])
    ).start();
  }, [opacity]);

  return (
    <View style={{ flex: 1, backgroundColor: 'transparent' }}>
      {/* Game title + chevron */}
      <Animated.View style={{ opacity, flexDirection: 'row', alignItems: 'center', justifyContent: 'center', marginBottom: 20, gap: 8 }}>
        <View style={{ width: 200, height: 40, borderRadius: 8, backgroundColor: '#ddd' }} />
        <View style={{ width: 24, height: 24, borderRadius: 12, backgroundColor: '#ddd' }} />
      </Animated.View>
      {/* Season banner */}
      <Animated.View style={{ opacity, alignItems: 'center', marginBottom: 16 }}>
        <View style={{ width: 140, height: 28, borderRadius: 12, backgroundColor: '#ddd' }} />
      </Animated.View>
      {/* Matches list skeleton */}
      <View style={{ flex: 1 }}>
        <MatchesListSkeleton />
      </View>
    </View>
  );
}
