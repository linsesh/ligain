import React from 'react';
import { View, Text, StyleSheet, Image } from 'react-native';
import { useTranslation } from 'react-i18next';
import { colors } from '../constants/colors';

// Instagram Stories dimensions (9:16 aspect ratio)
const SHARE_WIDTH = 1080;

interface ShareableMatchResultProps {
  homeTeam: string;
  awayTeam: string;
  homeScore: number;
  awayScore: number;
  myHomeScore?: number;
  myAwayScore?: number;
  date: string;
  players: Array<{
    name: string;
    points: number;
    bet?: string;
  }>;
  gameName: string;
}

export default function ShareableMatchResult({
  homeTeam,
  awayTeam,
  homeScore,
  awayScore,
  myHomeScore,
  myAwayScore,
  date,
  players,
  gameName,
}: ShareableMatchResultProps) {
  const { t } = useTranslation();

  // Sort players by points (descending) then by name (ascending)
  const sortedPlayers = [...players].sort((a, b) => {
    // First sort by points (highest first)
    if (b.points !== a.points) {
      return b.points - a.points;
    }
    // If points are equal, sort by name (alphabetical)
    return a.name.localeCompare(b.name);
  });

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

      {/* Game name */}
      <Text style={styles.gameName}>{gameName}</Text>

      {/* Match result */}
      <View style={styles.matchContainer}>
        <View style={styles.teamContainer}>
          <Text style={styles.teamName}>{homeTeam}</Text>
          <Text style={styles.score}>{homeScore}</Text>
        </View>
        
        <View style={styles.vsContainer}>
          <Text style={styles.vsText}>vs</Text>
          <Text style={styles.dateText}>{date}</Text>
        </View>
        
        <View style={styles.teamContainer}>
          <Text style={styles.teamName}>{awayTeam}</Text>
          <Text style={styles.score}>{awayScore}</Text>
        </View>
      </View>

      {/* Your bet */}
      <View style={styles.betSection}>
        <Text style={styles.betTitle}>{t('share.yourBet')}</Text>
        <View style={styles.matchContainer}>
          <View style={styles.teamContainer}>
            <Text style={styles.teamName}>{homeTeam}</Text>
            <Text style={styles.score}>{myHomeScore}</Text>
          </View>
          
          <View style={styles.vsContainer}>
            <Text style={styles.vsText}>vs</Text>
            <Text style={styles.dateText}>{date}</Text>
          </View>
          
          <View style={styles.teamContainer}>
            <Text style={styles.teamName}>{awayTeam}</Text>
            <Text style={styles.score}>{myAwayScore}</Text>
          </View>
        </View>
      </View>

      {/* Players results */}
      <View style={styles.playersContainer}>
        <Text style={styles.playersTitle}>{t('share.playersResults')}</Text>
        {sortedPlayers.map((player, index) => (
          <View key={index} style={styles.playerRow}>
            <View style={styles.playerInfo}>
              <Text style={styles.playerName}>{player.name}</Text>
              {player.bet && (
                <Text style={styles.playerBet}>{t('share.predicted')}: {player.bet}</Text>
              )}
            </View>
            <View style={styles.pointsContainer}>
              <Text style={[
                styles.points,
                player.points > 0 ? styles.positivePoints : 
                player.points < 0 ? styles.negativePoints : styles.zeroPoints
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
  gameName: {
    fontSize: 48,
    fontWeight: '600',
    color: colors.text,
    textAlign: 'center',
    marginBottom: 60,
  },
  matchContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: colors.card,
    padding: 60,
    borderRadius: 30,
    marginBottom: 60,
  },
  teamContainer: {
    flex: 1,
    alignItems: 'center',
  },
  teamName: {
    fontSize: 42,
    fontWeight: '600',
    color: colors.text,
    textAlign: 'center',
    marginBottom: 15,
  },
  score: {
    fontSize: 72,
    fontWeight: 'bold',
    color: colors.primary,
  },
  vsContainer: {
    alignItems: 'center',
    marginHorizontal: 30,
  },
  vsText: {
    fontSize: 36,
    fontWeight: 'bold',
    color: colors.textSecondary,
    marginBottom: 15,
  },
  dateText: {
    fontSize: 28,
    color: colors.textSecondary,
  },
  betSection: {
    marginBottom: 40,
  },
  betTitle: {
    fontSize: 32,
    fontWeight: 'bold',
    color: colors.text,
    textAlign: 'center',
    marginBottom: 20,
  },
  playersContainer: {
    backgroundColor: colors.card,
    padding: 50,
    borderRadius: 30,
    marginBottom: 60,
    flexGrow: 1, // Allow it to grow with content
  },
  playersTitle: {
    fontSize: 42,
    fontWeight: 'bold',
    color: colors.text,
    textAlign: 'center',
    marginBottom: 40,
  },
  playerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 25,
    borderBottomWidth: 2,
    borderBottomColor: colors.border,
  },
  playerInfo: {
    flex: 1,
  },
  playerName: {
    fontSize: 32,
    fontWeight: '600',
    color: colors.text,
    marginBottom: 8,
  },
  playerBet: {
    fontSize: 24,
    color: colors.textSecondary,
  },
  pointsContainer: {
    alignItems: 'center',
  },
  points: {
    fontSize: 36,
    fontWeight: 'bold',
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
