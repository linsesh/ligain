import React from 'react';
import { View, Text, StyleSheet, Image } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors } from '../constants/colors';

// Instagram Stories dimensions (9:16 aspect ratio)
const SHARE_WIDTH = 1080;

interface ShareableLeaderboardProps {
  gameName: string;
  period: string;
  players: Array<{
    name: string;
    points: number;
    rank: number;
  }>;
}

const getRankIcon = (rank: number) => {
  switch (rank) {
    case 1:
      return '🥇';
    case 2:
      return '🥈';
    case 3:
      return '🥉';
    default:
      return `${rank}.`;
  }
};

const getRankColor = (rank: number) => {
  switch (rank) {
    case 1:
      return colors.primary; // Gold
    case 2:
      return colors.silver;
    case 3:
      return colors.bronze;
    default:
      return '#cccccc'; // Light gray for better readability on dark background
  }
};

export default function ShareableLeaderboard({
  gameName,
  period,
  players,
}: ShareableLeaderboardProps) {
  const { t } = useTranslation();

  return (
    <View style={styles.container}>
      {/* Header with Ligain branding */}
      <View style={styles.header}>
        <Image 
          source={require('../../assets/images/icon.png')} 
          style={styles.logo}
          resizeMode="contain"
        />
        <Text style={styles.ligainTitle}>Ligain</Text>
      </View>

      {/* Game name and period */}
      <View style={styles.titleContainer}>
        <Text style={styles.gameName}>{gameName}</Text>
        <Text style={styles.period}>{period}</Text>
      </View>

      {/* Leaderboard */}
      <View style={styles.leaderboardContainer}>
        {players.map((player, index) => (
          <View key={index} style={styles.playerRow}>
            <View style={styles.rankContainer}>
              <Text style={styles.rankIcon}>{getRankIcon(player.rank)}</Text>
            </View>
            
            <View style={styles.playerInfo}>
              <Text style={styles.playerName}>{player.name}</Text>
            </View>
            
            <View style={styles.pointsContainer}>
              <Text style={[
                styles.points,
                { color: getRankColor(player.rank) }
              ]}>
                {player.points}
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
    backgroundColor: colors.background,
    padding: 80,
    minHeight: 1200, // Minimum height for Instagram Stories
  },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 60,
  },
  logo: {
    width: 100,
    height: 100,
    marginRight: 30,
  },
  ligainTitle: {
    fontSize: 72,
    fontWeight: 'bold',
    color: colors.primary,
    textAlign: 'center',
  },
  titleContainer: {
    alignItems: 'center',
    marginBottom: 60,
  },
  gameName: {
    fontSize: 48,
    fontWeight: '600',
    color: colors.text,
    textAlign: 'center',
    marginBottom: 15,
  },
  period: {
    fontSize: 36,
    color: colors.textSecondary,
    textAlign: 'center',
  },
  leaderboardContainer: {
    backgroundColor: colors.card,
    padding: 50,
    borderRadius: 30,
    marginBottom: 60,
    flexGrow: 1, // Allow it to grow with content
  },
  playerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 30,
    borderBottomWidth: 2,
    borderBottomColor: colors.border,
  },
  rankContainer: {
    width: 80,
    alignItems: 'center',
    marginRight: 30,
  },
  rankIcon: {
    fontSize: 48,
    fontWeight: 'bold',
  },
  playerInfo: {
    flex: 1,
  },
  playerName: {
    fontSize: 36,
    fontWeight: '600',
    color: colors.text,
  },
  pointsContainer: {
    alignItems: 'center',
    minWidth: 120,
  },
  points: {
    fontSize: 42,
    fontWeight: 'bold',
  },
  pointsLabel: {
    fontSize: 24,
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
