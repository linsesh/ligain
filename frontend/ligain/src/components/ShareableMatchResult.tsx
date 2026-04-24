import React from 'react';
import { View, StyleSheet, Image } from 'react-native';
import { Text } from './ui/Text';
import { useTranslation } from 'react-i18next';
import { colors } from '../constants/colors';
import { getShareTeamLogo } from '../utils/teamLogos';
import { getColorForName, getInitials } from './PlayerAvatar';
import { ShareableGridBackground } from './ShareableGridBackground';

// Instagram Stories dimensions (9:16 aspect ratio)
const SHARE_WIDTH = 1080;

function TeamLogoInline({ teamName, size }: { teamName: string; size: number }) {
  const logo = getShareTeamLogo(teamName);
  if (!logo) return null;
  return <Image source={logo} style={{ width: size, height: size }} resizeMode="contain" />;
}

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
  date,
  players,
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
      <ShareableGridBackground height={2400} />

      {/* Header */}
      <View style={styles.header}>
        <Text className="font-hk-extrabold" style={styles.ligainTitle}>LIGAIN</Text>
      </View>

      {/* Match result */}
      <View style={styles.matchContainer}>
        <View style={styles.teamContainer}>
          <TeamLogoInline teamName={homeTeam} size={90} />
          <Text className="font-hk-semibold" style={styles.teamName}>{homeTeam}</Text>
          <Text className="font-hk-bold" style={styles.score}>{homeScore}</Text>
        </View>

        <View style={styles.vsContainer}>
          <Text className="font-hk-bold" style={styles.vsText}>vs</Text>
          <Text style={styles.dateText}>{date}</Text>
        </View>

        <View style={styles.teamContainer}>
          <TeamLogoInline teamName={awayTeam} size={90} />
          <Text className="font-hk-semibold" style={styles.teamName}>{awayTeam}</Text>
          <Text className="font-hk-bold" style={styles.score}>{awayScore}</Text>
        </View>
      </View>

      {/* Your bet */}
      <View style={styles.betSection}>
        <Text className="font-hk-bold" style={styles.betTitle}>{t('share.yourBet')}</Text>
        <View style={styles.matchContainer}>
          <View style={styles.teamContainer}>
            <TeamLogoInline teamName={homeTeam} size={70} />
            <Text style={styles.teamName}>{homeTeam}</Text>
            <Text style={styles.score}>{myHomeScore}</Text>
          </View>

          <View style={styles.vsContainer}>
            <Text className="font-hk-bold" style={styles.vsText}>vs</Text>
            <Text style={styles.dateText}>{date}</Text>
          </View>

          <View style={styles.teamContainer}>
            <TeamLogoInline teamName={awayTeam} size={70} />
            <Text style={styles.teamName}>{awayTeam}</Text>
            <Text style={styles.score}>{myAwayScore}</Text>
          </View>
        </View>
      </View>

      {/* Players results */}
      <View style={styles.playersContainer}>
        <Text className="font-hk-bold" style={styles.playersTitle}>{t('share.playersResults')}</Text>
        {sortedPlayers.map((player, index) => (
          <View key={index} style={styles.playerRow}>
            <View style={[
              styles.playerAvatar,
              { backgroundColor: player.avatarUrl ? 'transparent' : getColorForName(player.name) }
            ]}>
              {player.avatarUrl ? (
                <Image
                  source={{ uri: player.avatarUrl }}
                  style={styles.playerAvatarImage}
                />
              ) : (
                <Text className="font-hk-semibold" style={styles.playerAvatarInitials}>
                  {getInitials(player.name)}
                </Text>
              )}
            </View>
            <View style={styles.playerInfo}>
              <Text className="font-hk-semibold" style={styles.playerName}>{player.name}</Text>
              {player.bet && (
                <Text style={styles.playerBet}>{t('share.predicted')}: {player.bet}</Text>
              )}
            </View>
            <View style={styles.pointsContainer}>
              <Text className="font-hk-bold" style={[
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
    gap: 10,
  },
  teamName: {
    fontSize: 42,
    color: colors.text,
    textAlign: 'center',
  },
  score: {
    fontSize: 72,
    color: colors.primary,
  },
  vsContainer: {
    alignItems: 'center',
    marginHorizontal: 30,
  },
  vsText: {
    fontSize: 36,
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
  playerAvatar: {
    width: 50,
    height: 50,
    borderRadius: 25,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 20,
    overflow: 'hidden',
  },
  playerAvatarImage: {
    width: 50,
    height: 50,
    borderRadius: 25,
    resizeMode: 'cover',
  },
  playerAvatarInitials: {
    fontSize: 20,
    color: '#FFFFFF',
  },
  playerInfo: {
    flex: 1,
  },
  playerName: {
    fontSize: 32,
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
