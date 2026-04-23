import React, { useState, useEffect, useMemo, useRef } from 'react';
import { View, TouchableOpacity, ScrollView, RefreshControl, Alert } from 'react-native';
import { Text } from '../../src/components/ui/Text';
import * as Haptics from 'expo-haptics';
import * as Clipboard from 'expo-clipboard';
import { Ionicons } from '@expo/vector-icons';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { colors } from '../../src/constants/colors';
import { useTranslation } from 'react-i18next';
import { useGames } from '../../src/contexts/GamesContext';
import { useMatches } from '../../src/contexts/MatchesContext';
import { getTranslatedGameStatus } from '../../src/utils/gameStatusUtils';
import { Picker } from '@react-native-picker/picker';
import { SeasonBanner } from '../../src/components/SeasonBanner';
import { PlayerAvatar } from '../../src/components/PlayerAvatar';
import { GridTag } from '../../src/components/ui/GridTag';
import { useGridCellSize } from '../../src/hooks/useGridCellSize';
import ShareableLeaderboard from '../../src/components/ShareableLeaderboard';
import { captureAndShareWithOptions } from '../../src/utils/shareUtils';
import ViewShot from 'react-native-view-shot';

const RANK_BORDER_COLORS = [
  colors.secondary,
  colors.silver,
  colors.bronze,
];

function getRankBorderColor(rank: number): string {
  return RANK_BORDER_COLORS[rank - 1] ?? colors.textSecondary;
}

function formatMonthLabel(key: string): string {
  try {
    const [y, m] = key.split('-').map(Number);
    const d = new Date(Date.UTC(y, (m || 1) - 1, 1));
    const month = d.toLocaleString(undefined, { month: 'long' });
    return `${month.charAt(0).toUpperCase()}${month.slice(1)} ${y}`;
  } catch {
    return key;
  }
}

export default function GameOverviewScreen() {
  const { id: gameId } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { player } = useAuth();
  const { t } = useTranslation();
  const { games, loading, error, refresh, selectedGameId } = useGames();
  const { incomingByMatchday } = useMatches();
  const cellSize = useGridCellSize();
  const [refreshing, setRefreshing] = useState(false);
  const [copied, setCopied] = useState(false);
  const [selectedPeriod, setSelectedPeriod] = useState<string>('general');
  const [showPeriodPicker, setShowPeriodPicker] = useState(false);
  const [isSharing, setIsSharing] = useState(false);
  const shareableRef = useRef<ViewShot>(null);

  const gameDetails = games.find((g) => g.gameId === gameId);

  const availableMonths = useMemo(() => {
    if (!gameDetails) return [];
    return Object.keys(gameDetails.perMonthLeaderboard || {}).sort((a, b) => (a < b ? 1 : -1));
  }, [gameDetails?.perMonthLeaderboard]);

  useEffect(() => {
    if (selectedPeriod !== 'general' && !availableMonths.includes(selectedPeriod)) {
      setSelectedPeriod('general');
    }
  }, [availableMonths, selectedPeriod]);

  const sortedPlayers = useMemo(() => {
    if (!gameDetails) return [];
    const source = selectedPeriod === 'general'
      ? (gameDetails.totalLeaderboard || [])
      : ((gameDetails.perMonthLeaderboard?.[selectedPeriod] as any[]) || []);
    return source.map((p: any) => ({ id: p.PlayerID, name: p.PlayerName, totalScore: p.Points }));
  }, [gameDetails?.totalLeaderboard, gameDetails?.perMonthLeaderboard, selectedPeriod]);

  const enrichedPlayers = useMemo(() => {
    const playersMap = new Map(
      (gameDetails?.players ?? []).map((p: any) => [p.id, p])
    );
    return sortedPlayers.map((p, index) => ({
      ...p,
      avatarUrl: playersMap.get(p.id)?.avatarUrl ?? null,
      rank: index + 1,
    }));
  }, [sortedPlayers, gameDetails?.players]);

  // Unbetted matches — only valid when viewing the currently selected game
  const isSelectedGame = gameId === selectedGameId;
  const closestMatchday = gameDetails?.closestUnfinishedMatchday?.matchday;
  const unbettedMatches = useMemo(() => {
    if (!isSelectedGame || !closestMatchday || !incomingByMatchday[closestMatchday]) return [];
    const now = new Date();
    return incomingByMatchday[closestMatchday]
      .filter(mr => !mr.match.hasStarted(now) && !(mr.bets && mr.bets[player?.id ?? '']))
      .sort((a, b) => a.match.getDate().getTime() - b.match.getDate().getTime());
  }, [isSelectedGame, closestMatchday, incomingByMatchday, player?.id]);

  const copyToClipboard = async (text: string) => {
    if (copied) return;
    try {
      await Clipboard.setStringAsync(text);
      setCopied(true);
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      setTimeout(() => setCopied(false), 3000);
    } catch {
      Alert.alert(t('common.error'), t('common.failedToCopyToClipboard'));
    }
  };

  const handleShareLeaderboard = async () => {
    if (isSharing) return;
    setIsSharing(true);
    try {
      await captureAndShareWithOptions(shareableRef, {
        title: t('share.shareTitle'),
        message: t('share.shareTitle'),
      });
    } catch {
      Alert.alert(t('share.shareFailed'), t('share.shareFailed'));
    } finally {
      setIsSharing(false);
    }
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  const navigateToFirstUnbettedMatch = () => {
    const mr = unbettedMatches[0];
    if (!mr) return;
    const m = mr.match;
    const bet = mr.bets?.[player?.id ?? ''];
    router.push({
      pathname: '/match/[id]',
      params: {
        id: m.id(),
        gameId: gameId || '',
        matchday: String(m.getMatchday()),
        date: m.getDate().toISOString(),
        homeTeam: m.homeTeamDisplayName(),
        awayTeam: m.awayTeamDisplayName(),
        homeTeamRaw: m.homeTeamName(),
        awayTeamRaw: m.awayTeamName(),
        betHomeGoals: bet ? String(bet.predictedHomeGoals) : '',
        betAwayGoals: bet ? String(bet.predictedAwayGoals) : '',
        homeTeamOdds: String(m.getHomeTeamOdds()),
        awayTeamOdds: String(m.getAwayTeamOdds()),
        drawOdds: String(m.getDrawOdds()),
        hasClearFavorite: String(m.hasClearFavorite()),
        favoriteTeam: m.getFavoriteTeam() || '',
      },
    });
  };

  const navigateToMatches = () => {
    router.push({
      pathname: '/(tabs)/matches',
      params: { gameId },
    });
  };

  // Loading state
  if (loading && !refreshing) {
    return (
      <View style={{ flex: 1, backgroundColor: 'transparent' }}>
        <TouchableOpacity
          onPress={() => router.back()}
          style={{ height: cellSize, justifyContent: 'center', paddingHorizontal: cellSize }}
        >
          <Ionicons name="arrow-back" size={24} color={colors.text} />
        </TouchableOpacity>
        <View className="flex-1 items-center justify-center">
          <Text className="text-foreground-secondary">{t('common.loading')}</Text>
        </View>
      </View>
    );
  }

  // Error state
  if (error) {
    return (
      <View style={{ flex: 1, backgroundColor: 'transparent' }}>
        <TouchableOpacity
          onPress={() => router.back()}
          style={{ height: cellSize, justifyContent: 'center', paddingHorizontal: cellSize }}
        >
          <Ionicons name="arrow-back" size={24} color={colors.text} />
        </TouchableOpacity>
        <View className="flex-1 items-center justify-center px-5">
          <Text className="text-error text-center mb-5">{error}</Text>
          <TouchableOpacity
            className="bg-primary rounded-full px-6 py-3"
            onPress={refresh}
          >
            <Text className="font-hk-bold text-white">{t('games.retry')}</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  // Not found
  if (!gameDetails) {
    return (
      <View style={{ flex: 1, backgroundColor: 'transparent' }}>
        <TouchableOpacity
          onPress={() => router.back()}
          style={{ height: cellSize, justifyContent: 'center', paddingHorizontal: cellSize }}
        >
          <Ionicons name="arrow-back" size={24} color={colors.text} />
        </TouchableOpacity>
        <View className="flex-1 items-center justify-center">
          <Text className="text-error">{t('games.gameNotFound')}</Text>
        </View>
      </View>
    );
  }

  const hasUnbettedMatches = unbettedMatches.length > 0;

  const { text: statusText, variant: statusVariant } = getTranslatedGameStatus(gameDetails.status || '', t);

  return (
    <View style={{ flex: 1, backgroundColor: 'transparent' }}>
      {/* Back button — grid-aligned like match detail */}
      <TouchableOpacity
        onPress={() => router.back()}
        style={{ height: cellSize, justifyContent: 'center', paddingHorizontal: cellSize }}
      >
        <Ionicons name="arrow-back" size={24} color={colors.text} />
      </TouchableOpacity>

      {/* Title — same layout as matches tab */}
      <View className="items-center justify-center mb-5">
        <Text className="font-hk-extrabold text-center text-4xl">
          {gameDetails.name}
        </Text>
      </View>

      {/* Season banner — same as matches tab */}
      <SeasonBanner
        className="mb-4"
        seasonYear={gameDetails.seasonYear}
        competitionName={gameDetails.competitionName}
      />

      {/* Status badge — centered */}
      <View className="items-center mb-3">
        <GridTag label={statusText} backgroundColor={statusVariant === 'warning' ? colors.warning : statusVariant === 'success' ? colors.success : colors.textSecondary} rounded />
      </View>

      <ScrollView
        style={{ flex: 1 }}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            tintColor={colors.primary}
            progressBackgroundColor={colors.background}
          />
        }
      >
        {/* Grey zone: filter + leaderboard */}
        <View style={{ backgroundColor: colors.background, marginTop: cellSize, paddingBottom: 16 }}>
          {/* Period selector + share button */}
          <View className="flex-row items-center gap-3 px-5 pt-4 pb-2">
            <TouchableOpacity
              onPress={() => setShowPeriodPicker(true)}
              className="flex-row items-center flex-1 rounded-full px-4 py-2.5"
              style={{ borderWidth: 2, borderColor: colors.secondary }}
            >
              <Text className="font-hk-semibold text-foreground flex-1">
                {selectedPeriod === 'general' ? t('games.general') : formatMonthLabel(selectedPeriod)}
              </Text>
              <Ionicons name="chevron-down" size={18} color={colors.text} />
            </TouchableOpacity>
            <TouchableOpacity
              onPress={handleShareLeaderboard}
              disabled={isSharing}
            >
              <Ionicons
                name={isSharing ? 'hourglass-outline' : 'share-social-outline'}
                size={24}
                color={isSharing ? colors.textSecondary : colors.text}
              />
            </TouchableOpacity>
          </View>

          {/* Leaderboard rows */}
          {enrichedPlayers.map((p) => {
            const isCurrentPlayer = player?.id === p.id;
            return (
              <View
                key={p.id}
                className="flex-row items-center py-3 border-b border-border"
                style={{ paddingHorizontal: 20, backgroundColor: isCurrentPlayer ? colors.link : undefined }}
              >
                <View
                  style={{
                    borderWidth: 2.5,
                    borderColor: getRankBorderColor(p.rank),
                    borderRadius: 999,
                    padding: 2,
                  }}
                >
                  <PlayerAvatar
                    player={{ name: p.name, avatarUrl: p.avatarUrl }}
                    displaySize="medium"
                  />
                </View>
                <Text className="font-hk-semibold flex-1 ml-3" style={isCurrentPlayer ? { color: colors.white } : { color: colors.text }}>
                  {p.name}
                </Text>
                <Text className="font-hk-bold text-lg" style={isCurrentPlayer ? { color: colors.white } : { color: colors.text }}>
                  {p.totalScore.toLocaleString()}
                </Text>
              </View>
            );
          })}

        </View>

        {/* Game code card — blue like grand favori */}
        {gameDetails.code && (
          <View className="mx-5 mt-5 mb-6 rounded-2xl px-5 py-4 items-center" style={{ backgroundColor: colors.link }}>
            <Text className="font-hk-medium text-sm mb-3" style={{ color: colors.white, opacity: 0.8 }}>
              {t('games.inviteCodeLabel')}
            </Text>
            <TouchableOpacity
              className="flex-row items-center"
              onPress={() => gameDetails.code && copyToClipboard(gameDetails.code)}
              disabled={copied || !gameDetails.code}
            >
              <Text className="font-hk-bold text-3xl mr-3" style={{ color: colors.white, letterSpacing: 4 }}>
                {gameDetails.code}
              </Text>
              <Ionicons
                name={copied ? 'checkmark-circle' : 'copy-outline'}
                size={24}
                color={colors.white}
              />
            </TouchableOpacity>
          </View>
        )}

        {/* CTA button */}
        <View className="px-6 mb-8 items-center">
          {hasUnbettedMatches ? (
            <>
              <TouchableOpacity
                onPress={navigateToFirstUnbettedMatch}
                className="rounded-full py-4 px-10 items-center self-center"
                style={{ backgroundColor: colors.primary, minWidth: '70%' }}
              >
                <Text className="font-hk-bold text-white text-lg">
                  {t('games.makeMyBets')}
                </Text>
              </TouchableOpacity>
              <Text className="text-foreground-secondary text-xs mt-2 text-center">
                {t('games.remainingThisWeek', { count: unbettedMatches.length })}
              </Text>
            </>
          ) : (
            <TouchableOpacity
              onPress={navigateToMatches}
              className="rounded-full py-4 px-10 items-center self-center"
              style={{ backgroundColor: colors.text, minWidth: '70%' }}
            >
              <Text className="font-hk-bold text-lg" style={{ color: colors.white }}>
                {t('games.viewMatches')}
              </Text>
            </TouchableOpacity>
          )}
        </View>

        {/* Hidden shareable component for image generation */}
        <View style={{ position: 'absolute', left: -9999, top: -9999 }}>
          <ViewShot ref={shareableRef}>
            <ShareableLeaderboard
              gameName={gameDetails.name || 'Ligain Game'}
              period={selectedPeriod === 'general' ? 'General' : formatMonthLabel(selectedPeriod)}
              players={enrichedPlayers.map(p => ({
                name: p.name,
                points: p.totalScore,
                rank: p.rank,
                avatarUrl: p.avatarUrl,
              }))}
            />
          </ViewShot>
        </View>
      </ScrollView>

      {/* Period picker modal */}
      {showPeriodPicker && (
        <View
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.7)',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 10,
          }}
        >
          <View
            className="rounded-xl"
            style={{ backgroundColor: colors.card, width: '80%', maxHeight: '60%' }}
          >
            <View className="flex-row justify-between items-center p-4 border-b border-border">
              <Text className="font-hk-bold text-foreground text-lg">
                {t('games.selectPeriod')}
              </Text>
              <TouchableOpacity onPress={() => setShowPeriodPicker(false)}>
                <Ionicons name="close" size={24} color={colors.text} />
              </TouchableOpacity>
            </View>
            <Picker
              selectedValue={selectedPeriod}
              onValueChange={(itemValue) => {
                setSelectedPeriod(String(itemValue));
                setShowPeriodPicker(false);
              }}
              style={{ color: colors.text, width: '100%' }}
              itemStyle={{ color: colors.text, fontSize: 16 }}
            >
              <Picker.Item label={t('games.general')} value="general" color={colors.text} />
              {availableMonths.map((k) => (
                <Picker.Item
                  key={k}
                  label={formatMonthLabel(k)}
                  value={k}
                  color={colors.text}
                />
              ))}
            </Picker>
          </View>
        </View>
      )}
    </View>
  );
}
