import React from 'react';
import { View, Text, StyleSheet, ViewStyle, TouchableOpacity } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../constants/colors';

interface PlayerGameInfo {
  id: string;
  name: string;
  totalScore: number;
}

interface LeaderboardProps {
  players: PlayerGameInfo[];
  currentPlayerId?: string;
  t: (key: string) => string;
  showTitle?: boolean;
  align?: 'left' | 'center';
  showCurrentPlayerTag?: boolean;
  containerStyle?: ViewStyle;
  onShare?: () => void;
  isSharing?: boolean;
}

// Helper function to get rank-based background colors
const getRankBackgroundColor = (index: number) => {
  switch (index) {
    case 0: // 1st place
      return { backgroundColor: colors.primary }; // Gold/Yellow
    case 1: // 2nd place
      return { backgroundColor: colors.silver }; // Silver
    case 2: // 3rd place
      return { backgroundColor: colors.bronze }; // Bronze
    default: // 4th place onwards
      return { backgroundColor: '#666666' }; // Neutral grey
  }
};

export default function Leaderboard({ players, currentPlayerId, t, showTitle = true, align = 'center', showCurrentPlayerTag = true, containerStyle, onShare, isSharing }: LeaderboardProps) {
  return (
    <View style={[
      styles.leaderboardContainer,
      align === 'left' && { alignItems: 'flex-start' },
      containerStyle,
    ]}>
      {showTitle && (
        <View style={styles.titleContainer}>
          <Text style={styles.leaderboardTitle}>{t('games.playerLeaderboard')}</Text>
          {onShare && (
            <TouchableOpacity
              style={styles.shareButton}
              onPress={onShare}
              disabled={isSharing}
            >
              <Ionicons
                name={isSharing ? "hourglass-outline" : "share-outline"}
                size={20}
                color={isSharing ? colors.textSecondary : colors.primary}
              />
            </TouchableOpacity>
          )}
        </View>
      )}
      {players.map((playerInfo, index) => (
        <View key={playerInfo.id} style={styles.playerRow}>
          <View style={[styles.playerRank, getRankBackgroundColor(index)]}>
            <Text style={styles.rankText}>{index + 1}</Text>
          </View>
          <View style={styles.playerInfo}>
            <Text style={styles.playerName}>{playerInfo.name}</Text>
            <Text style={styles.playerScore}>{playerInfo.totalScore} {t('game.points')}</Text>
          </View>
          {showCurrentPlayerTag && playerInfo.id === currentPlayerId && (
            <View style={styles.currentPlayerIndicator}>
              <Text style={styles.currentPlayerText}>{t('game.currentPlayer')}</Text>
            </View>
          )}
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  leaderboardContainer: {
    backgroundColor: '#333',
    borderRadius: 12,
    marginHorizontal: 20,
    marginBottom: 20,
    padding: 20,
    alignItems: 'center',
  },
  titleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    width: '100%',
    marginBottom: 16,
  },
  leaderboardTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
    textAlign: 'left',
  },
  shareButton: {
    padding: 6,
    borderRadius: 6,
    backgroundColor: '#444',
    alignItems: 'center',
    justifyContent: 'center',
    width: 40,
    height: 32,
  },
  playerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#444',
  },
  playerRank: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: '#ffd33d',
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: 16,
  },
  rankText: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#25292e',
  },
  playerInfo: {
    flex: 1,
  },
  playerName: {
    fontSize: 16,
    color: '#fff',
    fontWeight: '600',
  },
  playerScore: {
    fontSize: 14,
    color: '#999',
    marginTop: 2,
  },
  currentPlayerIndicator: {
    backgroundColor: colors.secondary,
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
    marginLeft: 8,
  },
  currentPlayerText: {
    fontSize: 12,
    color: '#fff',
    fontWeight: 'bold',
  },
}); 