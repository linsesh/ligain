import React from 'react';
import { View, StyleSheet, Image } from 'react-native';
import { Text } from './ui/Text';
import { useTranslation } from 'react-i18next';
import { colors } from '../constants/colors';
import { getColorForName, getInitials } from './PlayerAvatar';
import { ShareableGridBackground } from './ShareableGridBackground';

const SHARE_WIDTH = 1080;

interface Player {
  name: string;
  points: number;
  rank: number;
  avatarUrl?: string | null;
}

interface ShareableLeaderboardProps {
  gameName?: string;
  period?: string;
  players: Player[];
}

const getRankColor = (rank: number) => {
  switch (rank) {
    case 1: return colors.primary;
    case 2: return colors.silver;
    case 3: return colors.bronze;
    default: return '#666666';
  }
};

const formatPoints = (pts: number): string =>
  new Intl.NumberFormat('fr-FR', { useGrouping: true }).format(pts);

const PODIUM_HEIGHTS = { 1: 160, 2: 130, 3: 110 } as const;

function PodiumAvatar({ player, size }: { player: Player; size: number }) {
  const half = size / 2;
  return (
    <View style={[
      styles.podiumAvatar,
      { width: size, height: size, borderRadius: half, borderColor: getRankColor(player.rank) },
      !player.avatarUrl && { backgroundColor: getColorForName(player.name) },
    ]}>
      {player.avatarUrl ? (
        <Image source={{ uri: player.avatarUrl }} style={{ width: size - 8, height: size - 8, borderRadius: (size - 8) / 2 }} resizeMode="cover" />
      ) : (
        <Text className="font-hk-semibold" style={{ fontSize: size * 0.35, color: '#FFFFFF' }}>
          {getInitials(player.name)}
        </Text>
      )}
    </View>
  );
}

function PodiumBlock({ player, height }: { player: Player; height: number }) {
  const color = getRankColor(player.rank);
  return (
    <View style={styles.podiumColumn}>
      <PodiumAvatar player={player} size={player.rank === 1 ? 100 : 80} />
      <Text className="font-hk-bold" style={[styles.podiumName, { color }]} numberOfLines={1}>
        {player.name}
      </Text>
      <View style={[styles.podiumBlock, { height, backgroundColor: color }]}>
        <Text className="font-hk-extrabold" style={styles.podiumRank}>{player.rank}</Text>
      </View>
      <Text className="font-hk-bold" style={[styles.podiumPoints, { color }]}>
        {formatPoints(player.points)} pts
      </Text>
    </View>
  );
}

function Podium({ players }: { players: Player[] }) {
  const byRank = (r: number) => players.find(p => p.rank === r);
  const first = byRank(1);
  const second = byRank(2);
  const third = byRank(3);

  if (!first) return null;

  return (
    <View style={styles.podiumContainer}>
      <View style={styles.podiumRow}>
        {second ? (
          <PodiumBlock player={second} height={PODIUM_HEIGHTS[2]} />
        ) : (
          <View style={styles.podiumColumn} />
        )}
        <PodiumBlock player={first} height={PODIUM_HEIGHTS[1]} />
        {third ? (
          <PodiumBlock player={third} height={PODIUM_HEIGHTS[3]} />
        ) : (
          <View style={styles.podiumColumn} />
        )}
      </View>
    </View>
  );
}

export default function ShareableLeaderboard({
  players,
}: ShareableLeaderboardProps) {
  const { t } = useTranslation();
  const podiumPlayers = players.filter(p => p.rank <= 3);
  const listPlayers = players.filter(p => p.rank > 3);

  return (
    <View style={styles.container}>
      <ShareableGridBackground height={2400} />

      {/* Header */}
      <View style={styles.header}>
        <Text className="font-hk-extrabold" style={styles.ligainTitle}>LIGAIN</Text>
      </View>

      {/* Leaderboard card (podium + remaining players) */}
      <View style={styles.leaderboardContainer}>
        {podiumPlayers.length > 0 && <Podium players={podiumPlayers} />}

        {listPlayers.length > 0 && listPlayers.map((player, index) => (
          <View key={index} style={[
            styles.playerRow,
            index === listPlayers.length - 1 && { borderBottomWidth: 0 },
          ]}>
            <View style={[
              styles.avatar,
              { backgroundColor: player.avatarUrl ? 'transparent' : getColorForName(player.name) },
            ]}>
              {player.avatarUrl ? (
                <Image source={{ uri: player.avatarUrl }} style={styles.avatarImage} />
              ) : (
                <Text className="font-hk-semibold" style={styles.avatarInitials}>
                  {getInitials(player.name)}
                </Text>
              )}
            </View>

            <View style={styles.playerInfo}>
              <Text className="font-hk-semibold" style={styles.playerName}>{player.name}</Text>
            </View>

            <View style={styles.pointsContainer}>
              <Text className="font-hk-bold" style={styles.points}>
                {formatPoints(player.points)}
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
    backgroundColor: colors.background,
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

  // Podium
  podiumContainer: {
    marginBottom: 40,
    paddingBottom: 40,
    borderBottomWidth: 2,
    borderBottomColor: colors.border,
    alignItems: 'center',
  },
  podiumRow: {
    flexDirection: 'row',
    alignItems: 'flex-end',
    justifyContent: 'center',
    width: '100%',
  },
  podiumColumn: {
    flex: 1,
    alignItems: 'center',
    gap: 10,
  },
  podiumAvatar: {
    borderWidth: 4,
    alignItems: 'center',
    justifyContent: 'center',
    overflow: 'hidden',
  },
  podiumName: {
    fontSize: 28,
    textAlign: 'center',
    maxWidth: 250,
  },
  podiumBlock: {
    width: '80%',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    alignItems: 'center',
    justifyContent: 'center',
  },
  podiumRank: {
    fontSize: 48,
    color: '#FFFFFF',
  },
  podiumPoints: {
    fontSize: 30,
    marginTop: 8,
  },

  // Player list (4th+)
  leaderboardContainer: {
    backgroundColor: colors.surface,
    padding: 50,
    borderRadius: 30,
    marginBottom: 60,
    flexGrow: 1,
  },
  playerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 30,
    borderBottomWidth: 2,
    borderBottomColor: colors.border,
  },
  avatar: {
    width: 74,
    height: 74,
    borderRadius: 37,
    alignItems: 'center',
    justifyContent: 'center',
    overflow: 'hidden',
    marginRight: 25,
  },
  avatarImage: {
    width: 74,
    height: 74,
    borderRadius: 37,
    resizeMode: 'cover',
  },
  avatarInitials: {
    fontSize: 30,
    color: '#FFFFFF',
  },
  playerInfo: {
    flex: 1,
  },
  playerName: {
    fontSize: 36,
    color: colors.text,
  },
  pointsContainer: {
    alignItems: 'center',
    minWidth: 120,
  },
  points: {
    fontSize: 42,
    color: colors.text,
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
