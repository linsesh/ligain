import React, { useEffect, useRef, useState } from 'react';
import { View, Animated, Easing } from 'react-native';
import { Text } from './ui/Text';
import { colors } from '../constants/colors';
import { useTranslation } from 'react-i18next';
import { SimplifiedScore } from '../types/match';

interface Props {
  score: SimplifiedScore | null;
}

function getTitleKey(baseScore: number | undefined, score: SimplifiedScore | null): string {
  if (!score) return 'games.scoreMissedBet';
  if (baseScore === 500) return 'games.scorePerfect';
  if (baseScore === 400) return 'games.scoreClose';
  if (baseScore === 300) return 'games.scoreGoodResult';
  if (baseScore === 0) return 'games.scoreBadResult';
  if (baseScore === -100) return 'games.scoreMissedBet';
  if (score.points < 0) return 'games.scoreMissedBet';
  return 'games.scoreBadResult';
}

function getBaseScoreLabelKey(baseScore: number): string {
  if (baseScore === 500) return 'games.scorePerfectLabel';
  if (baseScore === 400) return 'games.scoreCloseLabel';
  return 'games.scoreGoodResultLabel';
}

function AnimatedPoints({ points }: { points: number }) {
  const { t } = useTranslation();
  const animValue = useRef(new Animated.Value(0)).current;
  const [displayValue, setDisplayValue] = useState(0);

  useEffect(() => {
    animValue.setValue(0);
    const listener = animValue.addListener(({ value }) => {
      setDisplayValue(Math.round(value));
    });
    Animated.timing(animValue, {
      toValue: points,
      duration: 800,
      easing: Easing.out(Easing.quad),
      useNativeDriver: false,
    }).start();
    return () => animValue.removeListener(listener);
  }, [points, animValue]);

  const isPositive = points > 0;
  const bgColor = isPositive ? '#4a5d23' : colors.formLoss;
  const prefix = displayValue > 0 ? '+ ' : displayValue < 0 ? '- ' : '';
  const label = `${prefix}${Math.abs(displayValue)} ${t('game.points')}`;

  return (
    <View style={{ backgroundColor: bgColor, borderRadius: 12, padding: 16, alignItems: 'center' }}>
      <Text className="font-hk-bold" style={{ color: colors.white, fontSize: 22 }}>
        {label}
      </Text>
    </View>
  );
}

export function FinishedMatchScore({ score }: Props) {
  const { t } = useTranslation();
  const baseScore = score?.baseScore;
  const titleKey = getTitleKey(baseScore, score);
  const hasBreakdown = baseScore !== undefined && score !== null && score.points > 0;
  const totalPoints = score?.points ?? -100;

  return (
    <View style={{ paddingHorizontal: 24, gap: 12 }}>
      {/* Title banner */}
      <View style={{ backgroundColor: colors.black, borderRadius: 12, padding: 16, alignItems: 'center' }}>
        <Text className="font-hk-bold" style={{ color: colors.white, fontSize: 18 }}>
          {t(titleKey)}
        </Text>
      </View>

      {/* Score breakdown row */}
      {hasBreakdown && (
        <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'center', paddingVertical: 8 }}>
          <View style={{ alignItems: 'center', flex: 1 }}>
            <Text className="font-hk-bold" style={{ fontSize: 24, color: colors.text }}>
              {baseScore}
            </Text>
            <Text style={{ fontSize: 11, color: colors.textSecondary, marginTop: 2 }}>
              {t(getBaseScoreLabelKey(baseScore!))}
            </Text>
          </View>

          {(score!.riskMultiplier ?? 1) > 1 && (
            <>
              <Text className="font-hk-bold" style={{ fontSize: 20, color: colors.textSecondary, marginHorizontal: 8 }}>
                x
              </Text>
              <View style={{ alignItems: 'center', flex: 1 }}>
                <Text className="font-hk-bold" style={{ fontSize: 24, color: colors.text }}>
                  {score!.riskMultiplier}
                </Text>
                <Text style={{ fontSize: 11, color: colors.textSecondary, marginTop: 2 }}>
                  {t('games.riskLabel')}
                </Text>
              </View>
            </>
          )}

          {(score!.clairvoyantMultiplier ?? 1) > 1 && (
            <>
              <Text className="font-hk-bold" style={{ fontSize: 20, color: colors.textSecondary, marginHorizontal: 8 }}>
                x
              </Text>
              <View style={{ alignItems: 'center', flex: 1 }}>
                <Text className="font-hk-bold" style={{ fontSize: 24, color: colors.text }}>
                  {score!.clairvoyantMultiplier}
                </Text>
                <Text style={{ fontSize: 11, color: colors.textSecondary, marginTop: 2 }}>
                  {t('games.clairvoyantLabel')}
                </Text>
              </View>
            </>
          )}
        </View>
      )}

      {/* Animated points banner */}
      <AnimatedPoints points={totalPoints} />
    </View>
  );
}
