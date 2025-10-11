import React, { useState, useEffect, useMemo } from 'react';
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
import { getTranslatedGameStatus } from '../../../../../src/utils/gameStatusUtils';
import StatusTag from '../../../../../src/components/StatusTag';
import { translateError } from '../../../../../src/utils/errorMessages';
import { Picker } from '@react-native-picker/picker';
import { computeCumulativePointsByMatchday } from '../../../../../src/utils/aggregations';
import CumulativePointsChart from '../../../../../src/components/CumulativePointsChart';
import { ErrorBoundary } from '../../../../../src/components/ErrorBoundary';

export default function GameOverviewScreen() {
  const { id: gameId } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { player } = useAuth();
  const { t } = useTranslation();
  const { games, loading, error, refresh } = useGames();
  const [refreshing, setRefreshing] = useState(false);
  const [copied, setCopied] = useState(false);
  const [selectedPeriod, setSelectedPeriod] = useState<string>('general'); // 'general' or 'YYYY-MM'
  const [showPeriodPicker, setShowPeriodPicker] = useState(false);

  // Add comprehensive logging for debugging
  useEffect(() => {
    console.log('🎮 GameOverviewScreen mounted with gameId:', gameId);
    console.log('🎮 Available games:', games.length);
    console.log('🎮 Player:', player?.id);
    
    return () => {
      console.log('🎮 GameOverviewScreen unmounting for gameId:', gameId);
    };
  }, [gameId, games.length, player?.id]);

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
          <Text style={styles.errorText}>{error}</Text>
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

  const availableMonths = useMemo(() => {
    const keys = Object.keys(gameDetails.perMonthLeaderboard || {});
    return keys.sort((a, b) => (a < b ? 1 : -1)); // Desc by YYYY-MM
  }, [gameDetails.perMonthLeaderboard]);

  // Ensure selectedPeriod is valid when data changes
  useEffect(() => {
    if (selectedPeriod !== 'general' && !availableMonths.includes(selectedPeriod)) {
      setSelectedPeriod('general');
    }
  }, [availableMonths, selectedPeriod]);

  const sortedPlayers = useMemo(() => {
    const source = selectedPeriod === 'general'
      ? (gameDetails.totalLeaderboard || [])
      : ((gameDetails.perMonthLeaderboard?.[selectedPeriod] as any[]) || []);
    return source.map((p: any) => ({ id: p.PlayerID, name: p.PlayerName, totalScore: p.Points }));
  }, [gameDetails.totalLeaderboard, gameDetails.perMonthLeaderboard, selectedPeriod]);

  const getCurrentMonthKey = () => {
    const now = new Date();
    const y = now.getUTCFullYear();
    const m = String(now.getUTCMonth() + 1).padStart(2, '0');
    return `${y}-${m}`;
  };

  const getPreviousMonthKey = () => {
    const now = new Date();
    const y = now.getUTCFullYear();
    const mIndex = now.getUTCMonth();
    const prevDate = new Date(Date.UTC(y, mIndex - 1, 1));
    const py = prevDate.getUTCFullYear();
    const pm = String(prevDate.getUTCMonth() + 1).padStart(2, '0');
    return `${py}-${pm}`;
  };

  const formatMonthLabel = (key: string) => {
    // key is YYYY-MM
    try {
      const [y, m] = key.split('-').map(Number);
      const d = new Date(Date.UTC(y, (m || 1) - 1, 1));
      const month = d.toLocaleString(undefined, { month: 'long' });
      return `${month.charAt(0).toUpperCase()}${month.slice(1)} ${y}`;
    } catch {
      return key;
    }
  };

  const currentMonthKey = getCurrentMonthKey();
  const lastMonthKey = getPreviousMonthKey();
  const currentMonthTop = (gameDetails.perMonthLeaderboard?.[currentMonthKey] || [])[0];
  const lastMonthTop = (gameDetails.perMonthLeaderboard?.[lastMonthKey] || [])[0];

  const cumulativeData = useMemo(() => {
    return computeCumulativePointsByMatchday(gameDetails.perMatchdayLeaderboard || {});
  }, [gameDetails.perMatchdayLeaderboard]);

  return (
    <ErrorBoundary>
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
            {(() => {
              const { text, variant } = getTranslatedGameStatus(gameDetails.status || '', t);
              return <StatusTag text={text} variant={variant} />;
            })()}
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
                {/* Period Selector (matches-like selector) */}
        <View style={styles.periodSelectionContainer}>
          <TouchableOpacity 
            style={styles.periodSelector}
            onPress={() => setShowPeriodPicker(true)}
          >
            <Text style={styles.periodSelectorText}>
              {selectedPeriod === 'general' ? t('games.general') : formatMonthLabel(selectedPeriod)}
            </Text>
            <Ionicons name="chevron-down" size={20} color="#fff" />
          </TouchableOpacity>
        </View>
        {showPeriodPicker && (
          <View style={styles.pickerOverlay}>
            <View style={styles.pickerContainer}>
              <View style={styles.pickerHeader}>
                <Text style={styles.pickerTitle}>{t('games.selectPeriod')}</Text>
                <TouchableOpacity onPress={() => setShowPeriodPicker(false)}>
                  <Ionicons name="close" size={24} color="#fff" />
                </TouchableOpacity>
              </View>
              <Picker
                selectedValue={selectedPeriod}
                onValueChange={(itemValue) => {
                  setSelectedPeriod(String(itemValue));
                  setShowPeriodPicker(false);
                }}
                style={styles.picker}
                itemStyle={styles.pickerItem}
              >
                <Picker.Item label={t('games.general')} value="general" color="#fff" />
                {availableMonths.map((k) => (
                  <Picker.Item 
                    key={k} 
                    label={formatMonthLabel(k)} 
                    value={k}
                    color="#fff"
                  />
                ))}
              </Picker>
            </View>
          </View>
        )}
        <Leaderboard
          players={sortedPlayers}
          currentPlayerId={player?.id}
          t={t}
        />
        {/* Current Month Leader and Last Month Winner cards */}
        {(currentMonthTop && currentMonthTop.Points > 0) && (
          <View style={styles.cardContainer}>
            <Text style={styles.cardTitle}>{t('games.currentMonthLeader')}</Text>
            <View style={styles.cardRow}>
              <Text style={styles.cardPrimary}>{currentMonthTop.PlayerName}</Text>
              <Text style={styles.cardSecondary}>{currentMonthTop.Points} {t('game.points')}</Text>
            </View>
          </View>
        )}
        {(lastMonthTop && lastMonthTop.Points > 0) && (
          <View style={styles.cardContainer}>
            <Text style={styles.cardTitle}>{t('games.lastMonthWinner')}</Text>
            <View style={styles.cardRow}>
              <Text style={styles.cardPrimary}>{lastMonthTop.PlayerName}</Text>
              <Text style={styles.cardSecondary}>{lastMonthTop.Points} {t('game.points')}</Text>
            </View>
          </View>
        )}
        {cumulativeData.series.length > 0 && cumulativeData.matchdays.length > 0 && (
          <View style={styles.cardContainer}>
            <Text style={styles.cardTitle}>{t('games.cumulativePointsByMatchday')}</Text>
            <ErrorBoundary>
              <CumulativePointsChart
                matchdays={cumulativeData.matchdays}
                series={cumulativeData.series.map(s => ({
                  playerId: s.playerId,
                  playerName: s.playerName,
                  values: s.values,
                }))}
              />
            </ErrorBoundary>
          </View>
        )}
        <TouchableOpacity
          style={styles.matchesButton}
          onPress={navigateToMatches}
        >
          <Ionicons name="football" size={24} color="#fff" />
          <Text style={styles.matchesButtonText}>{t('games.viewMatches')}</Text>
        </TouchableOpacity>
        </ScrollView>
      </View>
    </ErrorBoundary>
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
  periodSelectionContainer: {
    paddingHorizontal: 16,
    marginTop: 8,
    marginBottom: 8,
  },
  periodSelector: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: '#333',
    padding: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#444',
  },
  periodSelectorText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  periodInfoText: {
    color: '#999',
    fontSize: 14,
    marginTop: 4,
    marginLeft: 4,
  },
  cardContainer: {
    backgroundColor: '#333',
    padding: 20,
    borderRadius: 12,
    marginHorizontal: 20,
    marginBottom: 20,
  },
  cardTitle: {
    fontSize: 16,
    color: '#fff',
    marginBottom: 12,
    fontWeight: '600',
    textAlign: 'left',
  },
  cardRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  cardPrimary: {
    fontSize: 18,
    color: '#fff',
    fontWeight: '700',
  },
  cardSecondary: {
    fontSize: 16,
    color: '#999',
    fontWeight: '500',
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
  pickerOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.7)',
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 10,
  },
  pickerContainer: {
    backgroundColor: '#25292e',
    borderRadius: 10,
    width: '80%',
    maxHeight: '60%',
  },
  pickerHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#444',
  },
  pickerTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  pickerWrapper: {
    backgroundColor: '#444',
    borderRadius: 8,
  },
  picker: {
    color: '#fff',
    width: '100%',
  },
  pickerItem: {
    color: '#fff',
    fontSize: 16,
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



}); 