import React from 'react';
import { View, TextInput, Image, StyleSheet } from 'react-native';
import { Text } from './ui/Text';
import { getTeamLogo, isPngLogo } from '../utils/teamLogos';
import { colors } from '../constants/colors';
import { useTranslation } from 'react-i18next';

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
}

function TeamLogo({ teamName }: { teamName: string }) {
  const Logo = getTeamLogo(teamName);
  if (!Logo) return null;
  return isPngLogo(Logo) ? (
    <Image source={Logo} style={{ width: 56, height: 56 }} resizeMode="contain" />
  ) : (
    <Logo width={56} height={56} />
  );
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

function OddsColumn({ label, odds, multiplier }: { label: string; odds: number; multiplier?: string }) {
  return (
    <View style={oddsStyles.column}>
      <Text style={oddsStyles.label}>{label}</Text>
      <Text style={oddsStyles.value}>{odds.toFixed(2)}</Text>
      {multiplier && <Text style={oddsStyles.multiplier}>{multiplier}</Text>}
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
}: {
  homeTeamOdds: number;
  awayTeamOdds: number;
  drawOdds: number;
  hasClearFavorite: boolean;
  favoriteTeam: string;
  homeTeam: string;
}) {
  const { t } = useTranslation();
  const homeIsUnderdog = hasClearFavorite && favoriteTeam !== homeTeam;
  const homeMultiplier = homeIsUnderdog ? '×2' : '×1';
  const awayMultiplier = !homeIsUnderdog ? '×2' : '×1';

  return (
    <View style={oddsStyles.container}>
      <OddsColumn
        label={t('games.homeWin')}
        odds={homeTeamOdds}
        multiplier={hasClearFavorite ? homeMultiplier : undefined}
      />
      <OddsColumn
        label={t('games.draw')}
        odds={drawOdds}
        multiplier={hasClearFavorite ? '×1.5' : undefined}
      />
      <OddsColumn
        label={t('games.awayWin')}
        odds={awayTeamOdds}
        multiplier={hasClearFavorite ? awayMultiplier : undefined}
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
}: MatchBetCardProps) {
  const showOdds =
    homeTeamOdds !== undefined &&
    awayTeamOdds !== undefined &&
    drawOdds !== undefined;

  return (
    <View style={{ paddingBottom: 24 }}>
      <View style={styles.row}>
        {/* Home team */}
        <View style={styles.teamSide}>
          <TeamLogo teamName={homeTeam} />
          <Text style={styles.teamName}>{homeTeam}</Text>
        </View>

        {/* Score inputs */}
        <View style={styles.scoreCenter}>
          <ScoreInput value={homeGoals} onChange={onHomeGoalsChange} editable={editable} />
          <Text className="font-hk-bold" style={styles.vs}>VS</Text>
          <ScoreInput value={awayGoals} onChange={onAwayGoalsChange} editable={editable} />
        </View>

        {/* Away team */}
        <View style={styles.teamSide}>
          <TeamLogo teamName={awayTeam} />
          <Text style={styles.teamName}>{awayTeam}</Text>
        </View>
      </View>

      {showOdds && (
        <OddsSection
          homeTeamOdds={homeTeamOdds!}
          awayTeamOdds={awayTeamOdds!}
          drawOdds={drawOdds!}
          hasClearFavorite={hasClearFavorite ?? false}
          favoriteTeam={favoriteTeam ?? ''}
          homeTeam={homeTeam}
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
    padding: 16,
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
    fontSize: 11,
  },
  value: {
    color: colors.text,
    fontSize: 18,
    fontWeight: 'bold',
  },
  multiplier: {
    color: colors.secondary,
    fontSize: 13,
  },
});
