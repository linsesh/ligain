import React from 'react';
import { View } from 'react-native';
import { SimplifiedScore } from '../types/match';
import { ScoreBreakdown } from './ScoreBreakdown';

interface Props {
  score: SimplifiedScore | null;
}

export function FinishedMatchScore({ score }: Props) {
  return (
    <View style={{ paddingHorizontal: 24, gap: 12 }}>
      <ScoreBreakdown score={score} />
    </View>
  );
}
