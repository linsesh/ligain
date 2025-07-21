import React from 'react';
import { View, Text, StyleSheet, ViewStyle } from 'react-native';
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
}

export default function Leaderboard({ players, currentPlayerId, t, showTitle = true, align = 'center', showCurrentPlayerTag = true, containerStyle }: LeaderboardProps) {
  return (
    <View style={[
      styles.leaderboardContainer,
      align === 'left' && { alignItems: 'flex-start' },
      containerStyle,
    ]}>
      {showTitle && (
        <Text style={styles.leaderboardTitle}>{t('games.playerLeaderboard')}</Text>
      )}
      {players.map((playerInfo, index) => (
        <View key={playerInfo.id} style={styles.playerRow}>
          <View style={styles.playerRank}>
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
  leaderboardTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 16,
    textAlign: 'center',
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