import React from 'react';
import { View, TextInput, StyleSheet, TouchableOpacity, Image } from 'react-native';
import { Text } from './ui/Text';
import { colors } from '../constants/colors';
import { TeamLogo } from './ui/TeamLogo';
import { useTranslation } from 'react-i18next';
import { FormResult } from '../utils/standings';
import { FormSquares } from './ui/FormSquares';

const DIGIT_IMAGES: Record<string, ReturnType<typeof require>> = {
  '0': require('../../assets/images/0.png'),
  '1': require('../../assets/images/1.png'),
  '2': require('../../assets/images/2.png'),
  '3': require('../../assets/images/3.png'),
  '4': require('../../assets/images/4.png'),
  '5': require('../../assets/images/5.png'),
  '6': require('../../assets/images/6.png'),
  '7': require('../../assets/images/7.png'),
  '8': require('../../assets/images/8.png'),
  '9': require('../../assets/images/9.png'),
};

interface MatchBetCardProps {
  homeTeam: string;
  awayTeam: string;
  homeGoals: string;
  awayGoals: string;
  onHomeGoalsChange: (v: string) => void;
  onAwayGoalsChange: (v: string) => void;
  editable: boolean;
  homeTeamOdds?: number;
  awayTeamOdds?: number;
  drawOdds?: number;
  hasClearFavorite?: boolean;
  favoriteTeam?: string;
  onHomeTeamPress?: () => void;
  onAwayTeamPress?: () => void;
  homeTeamForm?: FormResult[];
  awayTeamForm?: FormResult[];
  showGoodResult?: boolean;
  showBadResult?: boolean;
  actualOutcome?: 'home' | 'draw' | 'away';
  actualHomeGoals?: string;
  actualAwayGoals?: string;
}

function ScoreInput({
  value,
  onChange,
  editable,
}: {
  value: string;
  onChange: (v: string) => void;
  editable: boolean;
}) {
  if (!editable) {
    return (
      <View style={styles.scoreBox}>
        <Text className="font-hk-bold" style={styles.scoreText}>{value}</Text>
      </View>
    );
  }
  return (
    <TextInput
      style={[styles.scoreBox, styles.scoreInput]}
      value={value}
      onChangeText={onChange}
      inputMode="numeric"
      maxLength={2}
      textAlign="center"
      selectTextOnFocus
    />
  );
}

function DigitScore({ homeGoals, awayGoals }: { homeGoals: string; awayGoals: string }) {
  const renderDigits = (n: string) =>
    n.split('').map((d, i) => (
      <Image key={i} source={DIGIT_IMAGES[d] ?? DIGIT_IMAGES['0']} style={{ width: 38, height: 56 }} resizeMode="contain" />
    ));
  return (
    <View style={{ position: 'absolute', top: 56, left: 0, right: 0, flexDirection: 'row', alignItems: 'center', justifyContent: 'center' }} pointerEvents="none">
      {renderDigits(homeGoals)}
      <Text style={{ color: colors.text, fontSize: 16, marginHorizontal: 3 }}>-</Text>
      {renderDigits(awayGoals)}
    </View>
  );
}

function OddsColumn({
  label,
  odds,
  multiplier,
  badResult,
  actualHomeGoals,
  actualAwayGoals,
}: {
  label: string;
  odds: number;
  multiplier?: string;
  badResult?: boolean;
  actualHomeGoals?: string;
  actualAwayGoals?: string;
}) {
  return (
    <View style={[oddsStyles.column, badResult ? { position: 'relative' } : undefined]}>
      {badResult && actualHomeGoals && actualAwayGoals && (
        <DigitScore homeGoals={actualHomeGoals} awayGoals={actualAwayGoals} />
      )}
      <Text style={oddsStyles.label}>{label}</Text>
      <Text style={oddsStyles.value}>{isNaN(odds) ? '-' : odds.toFixed(2)}</Text>
      {multiplier && <Text style={oddsStyles.multiplier}>{multiplier}</Text>}
      {badResult && (
        <Image
          source={require('../../assets/images/bad_result.png')}
          style={[oddsStyles.badResultCircle, !multiplier && { bottom: -17 }]}
          resizeMode="contain"
          pointerEvents="none"
        />
      )}
    </View>
  );
}

function OddsSection({
  homeTeamOdds,
  awayTeamOdds,
  drawOdds,
  hasClearFavorite,
  favoriteTeam,
  homeTeam,
  actualOutcome,
  actualHomeGoals,
  actualAwayGoals,
}: {
  homeTeamOdds: number;
  awayTeamOdds: number;
  drawOdds: number;
  hasClearFavorite: boolean;
  favoriteTeam: string;
  homeTeam: string;
  actualOutcome?: 'home' | 'draw' | 'away';
  actualHomeGoals?: string;
  actualAwayGoals?: string;
}) {
  const { t } = useTranslation();
  const homeIsUnderdog = hasClearFavorite && favoriteTeam !== homeTeam;
  const homeMultiplier = homeIsUnderdog ? '×2' : '×1';
  const awayMultiplier = !homeIsUnderdog ? '×2' : '×1';

  return (
    <View style={[oddsStyles.container, !hasClearFavorite && { marginBottom: 15 }]}>
      <OddsColumn
        label={t('games.homeWin')}
        odds={homeTeamOdds}
        multiplier={hasClearFavorite ? homeMultiplier : undefined}
        badResult={actualOutcome === 'home'}
        actualHomeGoals={actualOutcome === 'home' ? actualHomeGoals : undefined}
        actualAwayGoals={actualOutcome === 'home' ? actualAwayGoals : undefined}
      />
      <OddsColumn
        label={t('games.draw')}
        odds={drawOdds}
        multiplier={hasClearFavorite ? '×1.5' : undefined}
        badResult={actualOutcome === 'draw'}
        actualHomeGoals={actualOutcome === 'draw' ? actualHomeGoals : undefined}
        actualAwayGoals={actualOutcome === 'draw' ? actualAwayGoals : undefined}
      />
      <OddsColumn
        label={t('games.awayWin')}
        odds={awayTeamOdds}
        multiplier={hasClearFavorite ? awayMultiplier : undefined}
        badResult={actualOutcome === 'away'}
        actualHomeGoals={actualOutcome === 'away' ? actualHomeGoals : undefined}
        actualAwayGoals={actualOutcome === 'away' ? actualAwayGoals : undefined}
      />
    </View>
  );
}

export function MatchBetCard({
  homeTeam,
  awayTeam,
  homeGoals,
  awayGoals,
  onHomeGoalsChange,
  onAwayGoalsChange,
  editable,
  homeTeamOdds,
  awayTeamOdds,
  drawOdds,
  hasClearFavorite,
  favoriteTeam,
  onHomeTeamPress,
  onAwayTeamPress,
  homeTeamForm,
  awayTeamForm,
  showGoodResult,
  showBadResult,
  actualOutcome,
  actualHomeGoals,
  actualAwayGoals,
}: MatchBetCardProps) {
  const showOdds =
    homeTeamOdds !== undefined &&
    awayTeamOdds !== undefined &&
    drawOdds !== undefined;

  return (
    <View style={{ paddingBottom: 24 }}>
      <View style={styles.row}>
        {/* Home team */}
        <TouchableOpacity
          style={styles.teamSide}
          onPress={onHomeTeamPress}
          disabled={!onHomeTeamPress}
          activeOpacity={onHomeTeamPress ? 0.7 : 1}
        >
          <TeamLogo teamName={homeTeam} />
          <Text style={styles.teamName}>{homeTeam}</Text>
          {homeTeamForm && <FormSquares form={homeTeamForm} />}
        </TouchableOpacity>

        {/* Score inputs */}
        <View style={styles.scoreCenter}>
          <ScoreInput value={homeGoals} onChange={onHomeGoalsChange} editable={editable} />
          <Text className="font-hk-bold" style={styles.vs}>VS</Text>
          <ScoreInput value={awayGoals} onChange={onAwayGoalsChange} editable={editable} />
          {showGoodResult && (
            <Image
              source={require('../../assets/images/good_result.png')}
              style={styles.goodResultCircle}
              resizeMode="contain"
              pointerEvents="none"
            />
          )}
        </View>

        {/* Away team */}
        <TouchableOpacity
          style={styles.teamSide}
          onPress={onAwayTeamPress}
          disabled={!onAwayTeamPress}
          activeOpacity={onAwayTeamPress ? 0.7 : 1}
        >
          <TeamLogo teamName={awayTeam} />
          <Text style={styles.teamName}>{awayTeam}</Text>
          {awayTeamForm && <FormSquares form={awayTeamForm} />}
        </TouchableOpacity>
      </View>

      {showOdds && (
        <OddsSection
          homeTeamOdds={homeTeamOdds!}
          awayTeamOdds={awayTeamOdds!}
          drawOdds={drawOdds!}
          hasClearFavorite={hasClearFavorite ?? false}
          favoriteTeam={favoriteTeam ?? ''}
          homeTeam={homeTeam}
          actualOutcome={showBadResult ? actualOutcome : undefined}
          actualHomeGoals={showBadResult ? actualHomeGoals : undefined}
          actualAwayGoals={showBadResult ? actualAwayGoals : undefined}
        />
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 24,
    gap: 8,
  },
  teamSide: {
    flex: 1,
    alignItems: 'center',
    gap: 8,
  },
  teamName: {
    fontSize: 14,
    color: colors.text,
    textAlign: 'center',
  },
  scoreCenter: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    position: 'relative',
  },
  goodResultCircle: {
    position: 'absolute',
    width: 180,
    height: 90,
    alignSelf: 'center',
    left: -16,
    right: -16,
    top: -24,
    pointerEvents: 'none',
  },
  scoreBox: {
    width: 56,
    height: 56,
    borderRadius: 8,
    backgroundColor: colors.white,
    alignItems: 'center',
    justifyContent: 'center',
  },
  scoreText: {
    fontSize: 28,
    color: colors.text,
  },
  scoreInput: {
    fontSize: 28,
    fontWeight: 'bold',
    color: colors.text,
    padding: 0,
  },
  vs: {
    fontSize: 20,
    color: colors.textSecondary,
  },
});

const oddsStyles = StyleSheet.create({
  container: {
    backgroundColor: colors.border,
    borderRadius: 12,
    padding: 10,
    flexDirection: 'row',
    marginHorizontal: 24,
  },
  column: {
    flex: 1,
    alignItems: 'center',
    gap: 4,
  },
  label: {
    color: colors.textSecondary,
    fontSize: 10,
  },
  value: {
    color: colors.text,
    fontSize: 15,
    fontWeight: 'bold',
  },
  multiplier: {
    color: colors.secondary,
    fontSize: 11,
  },
  badResultCircle: {
    position: 'absolute',
    width: 110,
    height: 74,
    alignSelf: 'center',
    left: -2,
    right: -12,
    bottom: -2,
    pointerEvents: 'none',
  },
});
