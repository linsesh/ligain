import React from 'react';
import { View, StyleSheet } from 'react-native';
import { BlurView } from 'expo-blur';
import { Text } from './ui/Text';
import { PlayerAvatar } from './PlayerAvatar';
import { ScoreBreakdown } from './ScoreBreakdown';
import { colors } from '../constants/colors';
import { useTranslation } from 'react-i18next';

interface Props {
  player: { name: string; avatarUrl?: string | null };
  bet: { predictedHomeGoals: number; predictedAwayGoals: number } | null;
  score: { points: number; baseScore?: number; riskMultiplier?: number; clairvoyantMultiplier?: number };
  homeTeam: string;
  awayTeam: string;
}

export function PlayerBetDetailsOverlay({ player, bet, score, homeTeam, awayTeam }: Props) {
  const { t } = useTranslation();

  return (
    <View style={StyleSheet.absoluteFill} pointerEvents="none">
      <BlurView intensity={90} tint="light" style={StyleSheet.absoluteFill}>
        <View style={{ flex: 1, justifyContent: 'flex-start', alignItems: 'center', paddingHorizontal: 32, paddingTop: '25%' }}>
          {/* Player avatar + name */}
          <PlayerAvatar player={player} displaySize="large" />
          <Text className="font-hk-bold" style={{ fontSize: 18, color: colors.text, marginTop: 12 }}>
            {player.name}
          </Text>

          {/* Bet prediction */}
          {bet ? (
            <View style={{ marginTop: 24, alignItems: 'center' }}>
              <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 16 }}>
                <Text className="font-hk-medium" style={{ fontSize: 14, color: colors.textSecondary, flex: 1, textAlign: 'right' }}>
                  {homeTeam}
                </Text>
                <Text className="font-hk-bold" style={{ fontSize: 32, color: colors.text }}>
                  {bet.predictedHomeGoals} - {bet.predictedAwayGoals}
                </Text>
                <Text className="font-hk-medium" style={{ fontSize: 14, color: colors.textSecondary, flex: 1, textAlign: 'left' }}>
                  {awayTeam}
                </Text>
              </View>
            </View>
          ) : (
            <View style={{ marginTop: 24, alignItems: 'center' }}>
              <Text className="font-hk-medium" style={{ fontSize: 14, color: colors.textSecondary }}>
                {t('games.scoreMissedBet')}
              </Text>
            </View>
          )}

          {/* Score breakdown */}
          <View style={{ marginTop: 24, width: '100%', gap: 12 }}>
            <ScoreBreakdown score={score} animate={false} />
          </View>
        </View>
      </BlurView>
    </View>
  );
}
