import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TouchableOpacity, ScrollView, RefreshControl, Alert } from 'react-native';
import * as Haptics from 'expo-haptics';
import * as Clipboard from 'expo-clipboard';
import { Ionicons } from '@expo/vector-icons';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useAuth } from '../../../../../src/contexts/AuthContext';
import { colors } from '../../../../../src/constants/colors';
import { useTranslation } from 'react-i18next';
import Leaderboard from '../../../../../src/components/Leaderboard';
import { useGames } from '../../../../../src/contexts/GamesContext';

function StatusTag({ text, variant }: { text: string; variant: string }) {
  const baseStyle = [styles.statusTag];
  let variantStyle = null;
  if (variant === 'success') variantStyle = styles.successTag;
  else if (variant === 'warning') variantStyle = styles.inProgressTag;
  else if (variant === 'finished') variantStyle = styles.finishedTag;
  else if (variant === 'primary') variantStyle = styles.primaryTag;
  else if (variant === 'negative') variantStyle = styles.negativeTag;
  return (
    <View style={[...baseStyle, variantStyle]}>
      <Text style={styles.statusTagText}>{text}</Text>
    </View>
  );
}

export default function GameOverviewScreen() {
  const { id: gameId } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { player } = useAuth();
  const { t } = useTranslation();
  const { games, loading, error, refresh } = useGames();
  const [refreshing, setRefreshing] = useState(false);
  const [copied, setCopied] = useState(false);

  const gameDetails = games.find((g) => g.gameId === gameId);

  const copyToClipboard = async (text: string) => {
    if (copied) return;
    try {
      await Clipboard.setStringAsync(text);
      setCopied(true);
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      setTimeout(() => {
        setCopied(false);
      }, 3000);
    } catch (err) {
      Alert.alert(t('common.error'), t('common.failedToCopyToClipboard'));
    }
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  const navigateToMatches = () => {
    router.push({
      pathname: '/(tabs)/matches',
      params: { gameId },
    });
  };

  if (loading && !refreshing) {
    return (
      <View style={[styles.container, { backgroundColor: colors.loadingBackground }]}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.container}>
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{t('games.failedToLoadGame')} {error}</Text>
          <TouchableOpacity style={styles.retryButton} onPress={refresh}>
            <Text style={styles.retryButtonText}>{t('games.retry')}</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  if (!gameDetails) {
    return (
      <View style={styles.container}>
        <Text style={styles.errorText}>{t('games.gameNotFound')}</Text>
      </View>
    );
  }

  const sortedPlayers = [...(gameDetails.players || [])].sort((a, b) => b.totalScore - a.totalScore);

  return (
    <View style={styles.container}>
      <ScrollView
        style={styles.scrollView}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            tintColor={colors.primary}
            progressBackgroundColor="#25292e"
          />
        }
      >
        <View style={styles.gameHeader}>
          <Text style={styles.gameTitle}>{gameDetails.name}</Text>
          <Text style={styles.gameSubtitle}>
            {gameDetails.seasonYear} • {gameDetails.competitionName}
          </Text>
          <View style={styles.statusContainer}>
            {gameDetails.status === 'in progress' && (
              <StatusTag text={t('games.inProgressTag')} variant="warning" />
            )}
            {gameDetails.status === 'finished' && (
              <StatusTag text={t('games.finishedTag')} variant="success" />
            )}
            {gameDetails.status === 'not started' && (
              <StatusTag text={t('games.scheduledTag')} variant="primary" />
            )}
          </View>
        </View>
        {gameDetails.code && (
          <View style={styles.codeContainer}>
            <Text style={styles.codeLabel}>{t('games.gameCode')}</Text>
            <View style={styles.codeDisplay}>
              <Text style={styles.codeText}>{gameDetails.code}</Text>
              <TouchableOpacity
                style={styles.copyButton}
                onPress={() => gameDetails.code && copyToClipboard(gameDetails.code)}
                disabled={copied || !gameDetails.code}
              >
                <Ionicons name={copied ? "checkmark-circle" : "copy"} size={20} color={colors.primary} />
              </TouchableOpacity>
            </View>
          </View>
        )}
        <Leaderboard
          players={sortedPlayers}
          currentPlayerId={player?.id}
          t={t}
        />
        <TouchableOpacity
          style={styles.matchesButton}
          onPress={navigateToMatches}
        >
          <Ionicons name="football" size={24} color="#fff" />
          <Text style={styles.matchesButtonText}>{t('games.viewMatches')}</Text>
        </TouchableOpacity>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#25292e',
  },

  scrollView: {
    flex: 1,
  },
  gameHeader: {
    padding: 20,
    alignItems: 'center',
  },
  gameTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#fff',
    textAlign: 'center',
    marginBottom: 8,
  },
  gameSubtitle: {
    fontSize: 16,
    color: '#999',
    textAlign: 'center',
    marginBottom: 8,
  },
  gameStatus: {
    fontSize: 14,
    color: '#ffd33d',
    fontWeight: 'bold',
    textAlign: 'center',
  },
  statusContainer: {
    alignItems: 'center',
    marginTop: 8,
  },
  codeContainer: {
    backgroundColor: '#333',
    padding: 20,
    borderRadius: 12,
    marginHorizontal: 20,
    marginBottom: 20,
    alignItems: 'center',
  },
  codeLabel: {
    fontSize: 16,
    color: '#fff',
    marginBottom: 12,
    fontWeight: '600',
  },
  codeDisplay: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#444',
    borderRadius: 8,
    padding: 12,
    paddingHorizontal: 16,
  },
  codeText: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#ffd33d',
    marginRight: 12,
    letterSpacing: 2,
  },
  copyButton: {
    padding: 8,
  },
  leaderboardContainer: {
    backgroundColor: '#333',
    borderRadius: 12,
    marginHorizontal: 20,
    marginBottom: 20,
    padding: 20,
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
    backgroundColor: '#4CAF50',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
  },
  currentPlayerText: {
    fontSize: 12,
    color: '#fff',
    fontWeight: 'bold',
  },
  matchesButton: {
    backgroundColor: colors.secondary,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 16,
    paddingHorizontal: 24,
    borderRadius: 12,
    marginHorizontal: 20,
    marginBottom: 20,
    gap: 8,
  },
  matchesButtonText: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  errorText: {
    color: '#ff6b6b',
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 20,
  },
  retryButton: {
    backgroundColor: '#4CAF50',
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
  },
  retryButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  statusTag: {
    paddingHorizontal: 10,
    paddingVertical: 5,
    borderRadius: 8,
    alignSelf: 'flex-start',
  },
  successTag: {
    backgroundColor: '#4CAF50',
  },
  inProgressTag: {
    backgroundColor: '#ffd33d',
  },
  finishedTag: {
    backgroundColor: '#ff6b6b',
  },
  primaryTag: {
    backgroundColor: '#4CAF50',
  },
  negativeTag: {
    backgroundColor: '#ff6b6b',
  },
  statusTagText: {
    fontSize: 12,
    fontWeight: 'bold',
    color: '#fff',
  },


}); 