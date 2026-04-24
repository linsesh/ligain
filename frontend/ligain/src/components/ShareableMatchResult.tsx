import React from 'react';
import { View, StyleSheet, Image } from 'react-native';
import { Text } from './ui/Text';
import { useTranslation } from 'react-i18next';
import { colors } from '../constants/colors';
import { getColorForName, getInitials } from './PlayerAvatar';
import { ShareableGridBackground } from './ShareableGridBackground';
import { TeamLogo } from './ui/TeamLogo';

const SHARE_WIDTH = 1080;

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

function HandDrawnScore({ homeGoals, awayGoals }: { homeGoals: string; awayGoals: string }) {
  const renderDigits = (n: string) =>
    n.split('').map((d, i) => (
      <Image key={i} source={DIGIT_IMAGES[d] ?? DIGIT_IMAGES['0']} style={{ width: 76, height: 112 }} resizeMode="contain" />
    ));
  return (
    <View style={{ flexDirection: 'row', alignItems: 'center', justifyContent: 'center' }}>
      {renderDigits(homeGoals)}
      <Text style={{ color: colors.text, fontSize: 36, marginHorizontal: 12 }}>-</Text>
      {renderDigits(awayGoals)}
    </View>
  );
}

interface ShareableMatchResultProps {
  homeTeam: string;
  awayTeam: string;
  homeScore: number;
  awayScore: number;
  myHomeScore?: number;
  myAwayScore?: number;
  showGoodResult?: boolean;
  date?: string;
  players: Array<{
    name: string;
    points: number;
    bet?: string;
    avatarUrl?: string | null;
  }>;
  gameName?: string;
}

export default function ShareableMatchResult({
  homeTeam,
  awayTeam,
  homeScore,
  awayScore,
  myHomeScore,
  myAwayScore,
  showGoodResult,
  players,
}: ShareableMatchResultProps) {
  const { t } = useTranslation();

  const sortedPlayers = [...players].sort((a, b) => {
    if (b.points !== a.points) return b.points - a.points;
    return a.name.localeCompare(b.name);
  });

  const hasBet = myHomeScore !== undefined && myAwayScore !== undefined;

  return (
    <View style={styles.container}>
      <ShareableGridBackground height={2400} />

      {/* Header */}
      <View style={styles.header}>
        <Text className="font-hk-extrabold" style={styles.ligainTitle}>LIGAIN</Text>
      </View>

      {/* Match card */}
      <View style={styles.matchCard}>
        {/* Teams + Score on same row */}
        <View style={styles.matchRow}>
          <View style={styles.teamSide}>
            <TeamLogo teamName={homeTeam} size={180} />
            <Text className="font-hk-semibold" style={styles.teamName}>{homeTeam}</Text>
          </View>

          <View style={styles.scoreCenter}>
            <View style={styles.scoreRow}>
              <View style={styles.scoreBox}>
                <Text className="font-hk-bold" style={styles.scoreText}>{homeScore}</Text>
              </View>
              <Text className="font-hk-bold" style={styles.vsText}>VS</Text>
              <View style={styles.scoreBox}>
                <Text className="font-hk-bold" style={styles.scoreText}>{awayScore}</Text>
              </View>
              {showGoodResult && (
                <Image
                  source={require('../../assets/images/good_result.png')}
                  style={styles.goodResultOverlay}
                  resizeMode="contain"
                />
              )}
            </View>
          </View>

          <View style={styles.teamSide}>
            <TeamLogo teamName={awayTeam} size={180} />
            <Text className="font-hk-semibold" style={styles.teamName}>{awayTeam}</Text>
          </View>
        </View>

        {/* My bet as hand-drawn digits — only when not a good result */}
        {hasBet && !showGoodResult && (
          <View style={styles.betSection}>
            <Text className="font-hk-semibold" style={styles.betLabel}>{t('share.yourBet')}</Text>
            <HandDrawnScore homeGoals={String(myHomeScore)} awayGoals={String(myAwayScore)} />
          </View>
        )}
      </View>

      {/* Players results */}
      <View style={styles.playersContainer}>
        <Text className="font-hk-bold" style={styles.playersTitle}>{t('share.playersResults')}</Text>
        {sortedPlayers.map((player, index) => (
          <View key={index} style={[
            styles.playerRow,
            index === sortedPlayers.length - 1 && { borderBottomWidth: 0 },
          ]}>
            <View style={[
              styles.playerAvatar,
              { backgroundColor: player.avatarUrl ? 'transparent' : getColorForName(player.name) },
            ]}>
              {player.avatarUrl ? (
                <Image source={{ uri: player.avatarUrl }} style={styles.playerAvatarImage} />
              ) : (
                <Text className="font-hk-semibold" style={styles.playerAvatarInitials}>
                  {getInitials(player.name)}
                </Text>
              )}
            </View>
            <View style={styles.playerInfo}>
              <Text className="font-hk-semibold" style={styles.playerName}>{player.name}</Text>
            </View>
            <View style={styles.pointsContainer}>
              <Text className="font-hk-bold" style={[
                styles.points,
                player.points > 0 ? styles.positivePoints :
                player.points < 0 ? styles.negativePoints : styles.zeroPoints,
              ]}>
                {player.points > 0 ? '+' : ''}{player.points}
              </Text>
              <Text style={styles.pointsLabel}>{t('share.points')}</Text>
            </View>
          </View>
        ))}
      </View>

      {/* Footer */}
      <View style={styles.footer}>
        <Text style={styles.footerText}>{t('share.downloadLigain')}</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    width: SHARE_WIDTH,
    padding: 80,
    minHeight: 1200,
  },
  header: {
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 60,
  },
  ligainTitle: {
    fontSize: 96,
    color: colors.primary,
    textAlign: 'center',
  },

  // Match card
  matchCard: {
    backgroundColor: colors.surface,
    borderRadius: 30,
    padding: 50,
    marginBottom: 50,
  },
  matchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  teamSide: {
    flex: 1,
    alignItems: 'center',
    gap: 12,
  },
  teamName: {
    fontSize: 32,
    color: colors.text,
    textAlign: 'center',
  },
  scoreCenter: {
    alignItems: 'center',
    paddingHorizontal: 10,
  },
  scoreRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 20,
    position: 'relative',
  },
  scoreBox: {
    width: 120,
    height: 120,
    borderRadius: 20,
    backgroundColor: colors.white,
    alignItems: 'center',
    justifyContent: 'center',
  },
  scoreText: {
    fontSize: 64,
    color: colors.text,
  },
  vsText: {
    fontSize: 40,
    color: colors.textSecondary,
  },
  goodResultOverlay: {
    position: 'absolute',
    width: 360,
    height: 180,
    alignSelf: 'center',
    left: 10,
    right: -50,
    top: -45,
  },

  // Bet
  betSection: {
    alignItems: 'center',
    marginTop: 30,
    gap: 12,
  },
  betLabel: {
    fontSize: 30,
    color: colors.textSecondary,
  },

  // Players
  playersContainer: {
    backgroundColor: colors.surface,
    padding: 50,
    borderRadius: 30,
    marginBottom: 60,
    flexGrow: 1,
  },
  playersTitle: {
    fontSize: 42,
    color: colors.text,
    textAlign: 'center',
    marginBottom: 40,
  },
  playerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 25,
    borderBottomWidth: 2,
    borderBottomColor: colors.border,
  },
  playerAvatar: {
    width: 60,
    height: 60,
    borderRadius: 30,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 20,
    overflow: 'hidden',
  },
  playerAvatarImage: {
    width: 60,
    height: 60,
    borderRadius: 30,
    resizeMode: 'cover',
  },
  playerAvatarInitials: {
    fontSize: 24,
    color: '#FFFFFF',
  },
  playerInfo: {
    flex: 1,
  },
  playerName: {
    fontSize: 32,
    color: colors.text,
  },
  pointsContainer: {
    alignItems: 'center',
  },
  points: {
    fontSize: 36,
  },
  positivePoints: {
    color: colors.success,
  },
  negativePoints: {
    color: colors.error,
  },
  zeroPoints: {
    color: colors.textSecondary,
  },
  pointsLabel: {
    fontSize: 20,
    color: colors.textSecondary,
    marginTop: 4,
  },
  footer: {
    alignItems: 'center',
  },
  footerText: {
    fontSize: 28,
    color: colors.textSecondary,
    textAlign: 'center',
  },
});
