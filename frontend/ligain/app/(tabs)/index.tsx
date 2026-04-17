import React, { useState, useEffect, useRef, useMemo } from 'react';
import { View, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert, ScrollView, RefreshControl, KeyboardAvoidingView, Platform, Pressable } from 'react-native';
import { Text } from '../../src/components/ui/Text';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useFocusEffect } from '@react-navigation/native';
import { Swipeable } from 'react-native-gesture-handler';
import { useAuth } from '../../src/contexts/AuthContext';
import { colors } from '../../src/constants/colors';
import { handleGameError } from '../../src/utils/errorMessages';
import { useTranslation } from 'react-i18next';
import { useGames } from '../../src/contexts/GamesContext';
import { useUIEvent } from '../../src/contexts/UIEventContext';
import { API_CONFIG, authenticatedFetch } from '../../src/config/api';
import { getTranslatedGameStatus } from '../../src/utils/gameStatusUtils';
import { GridTag } from '../../src/components/ui/GridTag';
import { PlayerAvatar } from '../../src/components/PlayerAvatar';
import { useGridCellSize } from '../../src/hooks/useGridCellSize';

const RANK_BORDER_COLORS = [
  colors.secondary,
  colors.silver,
  colors.bronze,
];

function getRankBorderColor(rank: number): string {
  return RANK_BORDER_COLORS[rank - 1] ?? colors.textSecondary;
}

function GameCard({ game, onPress, onLeave, disabled }: {
  game: any;
  onPress: () => void;
  onLeave: () => void;
  disabled: boolean;
}) {
  const { player } = useAuth();
  const { t } = useTranslation();
  const cellSize = useGridCellSize();
  const { text: statusText, variant } = getTranslatedGameStatus(game.status || '', t);

  const statusBg = variant === 'warning' ? colors.warning
    : variant === 'success' ? colors.success
    : colors.textSecondary;

  const enrichedPlayers = useMemo(() => {
    const playersMap = new Map(
      (game.players ?? []).map((p: any) => [p.id, p])
    );
    return (game.totalLeaderboard || []).map((p: any, index: number) => ({
      id: p.PlayerID,
      name: p.PlayerName,
      totalScore: p.Points,
      avatarUrl: (playersMap.get(p.PlayerID) as any)?.avatarUrl ?? null,
      rank: index + 1,
    }));
  }, [game.totalLeaderboard, game.players]);

  return (
    <TouchableOpacity
      onPress={onPress}
      activeOpacity={0.7}
      disabled={disabled}
      style={{ backgroundColor: colors.background }}
    >
      {/* Header */}
      <View className="p-4 pb-2">
        <View className="flex-row items-start justify-between">
          <View className="flex-1 mr-2">
            <Text className="font-hk-bold text-foreground text-xl">
              {game.name}
            </Text>
            <Text className="text-foreground-secondary text-sm mt-0.5">
              {game.seasonYear} · {game.competitionName}
            </Text>
          </View>
          <GridTag label={statusText} backgroundColor={statusBg} rounded />
        </View>
      </View>

      {/* Leaderboard rows */}
      {enrichedPlayers.length > 0 && (
        <View className="px-4 pb-3">
          {enrichedPlayers.map((p: any) => {
            const isCurrentPlayer = player?.id === p.id;
            return (
              <View
                key={p.id}
                className="flex-row items-center py-2 border-b border-border"
                style={{ paddingHorizontal: 16, backgroundColor: isCurrentPlayer ? colors.link : undefined, borderRadius: isCurrentPlayer ? 12 : 0 }}
              >
                <View
                  style={{
                    borderWidth: 2,
                    borderColor: getRankBorderColor(p.rank),
                    borderRadius: 999,
                    padding: 1.5,
                  }}
                >
                  <PlayerAvatar
                    player={{ name: p.name, avatarUrl: p.avatarUrl }}
                    displaySize="small"
                  />
                </View>
                <Text className="font-hk-medium flex-1 ml-2 text-sm" style={isCurrentPlayer ? { color: colors.white } : { color: colors.text }}>
                  {p.name}
                </Text>
                <Text className="font-hk-bold text-sm" style={isCurrentPlayer ? { color: colors.white } : { color: colors.text }}>
                  {p.totalScore.toLocaleString()}
                </Text>
              </View>
            );
          })}
        </View>
      )}
    </TouchableOpacity>
  );
}

function GamesList() {
  const { t } = useTranslation();
  const { player } = useAuth();
  const router = useRouter();
  const { games, loading, error, joinGame, createGame, refresh, removeGame } = useGames();
  const [refreshing, setRefreshing] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [joinCode, setJoinCode] = useState('');
  const [creatingGame, setCreatingGame] = useState(false);
  const [joiningGame, setJoiningGame] = useState(false);
  const [newGameName, setNewGameName] = useState('');
  const [showActionSheet, setShowActionSheet] = useState(false);
  const { openJoinOrCreate, setOpenJoinOrCreate } = useUIEvent();
  const [isAnySwipeActive, setIsAnySwipeActive] = useState(false);
  const swipeableRefs = useRef<{ [key: string]: any }>({});

  useFocusEffect(
    React.useCallback(() => {
      if (openJoinOrCreate) {
        setShowActionSheet(true);
        setOpenJoinOrCreate(false);
      }
    }, [openJoinOrCreate, setOpenJoinOrCreate])
  );

  const handleCreateGame = async () => {
    if (!newGameName.trim()) {
      Alert.alert(t('errors.error'), t('errors.pleaseEnterGameName'));
      return;
    }
    setCreatingGame(true);
    try {
      await createGame(newGameName.trim());
      setShowCreateModal(false);
      setNewGameName('');
      refresh();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : t('errors.failedToCreateGame');
      const { title, message } = handleGameError(errorMessage);
      Alert.alert(title, message);
    } finally {
      setCreatingGame(false);
    }
  };

  const handleJoinGame = async () => {
    if (!joinCode.trim()) {
      Alert.alert(t('errors.error'), t('errors.pleaseEnterGameCode'));
      return;
    }
    setJoiningGame(true);
    try {
      await joinGame(joinCode.trim());
      setJoinCode('');
      setShowJoinModal(false);
      refresh();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : t('errors.failedToJoinGame');
      const { title, message } = handleGameError(errorMessage);
      Alert.alert(title, message);
    } finally {
      setJoiningGame(false);
    }
  };

  const handleLeaveGame = async (gameId: string) => {
    Alert.alert(
      t('games.leaveGameTitle', 'Leave Game'),
      t('games.leaveGameConfirm', 'Are you sure you want to leave this game?'),
      [
        { text: t('common.cancel', 'Cancel'), style: 'cancel' },
        {
          text: t('games.leave', 'Leave'),
          style: 'destructive',
          onPress: async () => {
            try {
              const res = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games/${gameId}/leave`, {
                method: 'DELETE',
              });
              if (!res.ok) {
                const data = await res.json().catch(() => ({}));
                throw new Error(data.error || t('games.failedToLeaveGame', 'Failed to leave game'));
              }
              removeGame(gameId);
              Alert.alert(t('games.leftGame', 'You have left the game.'));
              refresh();
            } catch (err: any) {
              Alert.alert(t('games.failedToLeaveGame', 'Failed to leave game'), err.message || String(err));
            }
          },
        },
      ]
    );
  };

  const renderRightActions = (gameId: string) => (
    <View className="w-[75] h-full justify-center items-center" style={{ backgroundColor: 'transparent' }}>
      <TouchableOpacity
        className="justify-center items-center w-[60] h-[60] rounded-lg"
        style={{ backgroundColor: '#ff3b30' }}
        onPress={() => handleLeaveGame(gameId)}
      >
        <Ionicons name="exit-outline" size={20} color="#fff" />
        <Text className="font-hk-semibold text-foreground mt-0.5" style={{ fontSize: 11 }}>
          {t('games.leave', 'Leave')}
        </Text>
      </TouchableOpacity>
    </View>
  );

  const handleGamePress = (gameId: string) => {
    if (isAnySwipeActive) return;
    router.push('/game/' + gameId);
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  if (loading && !refreshing) {
    return (
      <View className="flex-1 justify-center items-center">
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  if (!games) {
    return <Text className="text-error">{t('games.errorLoadingGames')}</Text>;
  }

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={{ flex: 1, backgroundColor: 'transparent' }}
    >
      <Text className="font-hk-extrabold text-foreground text-3xl mx-4 my-4">
        {t('games.myGames')}
      </Text>

      {!player?.email && !player?.provider && (
        <View className="flex-row items-center rounded-lg mx-4 mb-4 px-4 py-3" style={{ backgroundColor: colors.warning }}>
          <Ionicons name="warning" size={20} color="#FFA500" />
          <Text className="font-hk-semibold text-foreground text-sm ml-2.5 flex-1">
            {t('games.guestModeBanner')}
          </Text>
        </View>
      )}

      <ScrollView
        style={{ flex: 1 }}
        contentContainerStyle={{ paddingBottom: 140 }}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            tintColor={colors.primary}
            progressBackgroundColor={colors.background}
            progressViewOffset={20}
          />
        }
      >
        {error ? (
          <View className="items-center justify-center p-5" style={{ minHeight: 200 }}>
            <Text className="text-error text-center text-base mb-5">{error}</Text>
            <Text className="text-foreground-secondary text-center text-sm">{t('games.pullToRefresh')}</Text>
          </View>
        ) : games.length === 0 ? (
          <View className="items-center p-5" style={{ minHeight: 200 }}>
            <Ionicons name="game-controller-outline" size={64} color={colors.textSecondary} />
            <Text className="text-foreground-secondary text-xl mt-2.5">{t('games.noGames')}</Text>
            <Text className="text-foreground-secondary text-sm mt-1">{t('games.noGamesSubtext')}</Text>
          </View>
        ) : (
          games.map((game) => (
            <View key={game.gameId} className="mx-4 mb-3 rounded-2xl overflow-hidden">
              <Swipeable
                ref={ref => { swipeableRefs.current[game.gameId] = ref; }}
                renderRightActions={() => renderRightActions(game.gameId)}
                overshootRight={false}
                rightThreshold={40}
                onSwipeableWillOpen={() => setIsAnySwipeActive(true)}
                onSwipeableClose={() => setIsAnySwipeActive(false)}
              >
                <GameCard
                  game={game}
                  onPress={() => handleGamePress(game.gameId)}
                  onLeave={() => handleLeaveGame(game.gameId)}
                  disabled={isAnySwipeActive}
                />
              </Swipeable>
            </View>
          ))
        )}
      </ScrollView>

      {/* Bottom join button */}
      {!showActionSheet && !showCreateModal && !showJoinModal && (
        <View className="absolute left-0 right-0 bottom-6 items-center" style={{ zIndex: 20 }}>
          <TouchableOpacity
            className="flex-row items-center justify-center rounded-2xl py-6 px-10"
            style={{
              backgroundColor: colors.primary,
              shadowColor: '#000',
              shadowOffset: { width: 0, height: 2 },
              shadowOpacity: 0.2,
              shadowRadius: 6,
              elevation: 4,
            }}
            onPress={() => setShowActionSheet(true)}
            activeOpacity={0.85}
          >
            <Ionicons name="add-circle" size={28} color="#fff" style={{ marginRight: 10 }} />
            <Text className="font-hk-bold text-white text-xl">{t('games.joinOrCreate')}</Text>
          </TouchableOpacity>
        </View>
      )}

      {/* Action sheet overlay */}
      {showActionSheet && (
        <Pressable
          onPress={() => setShowActionSheet(false)}
          style={{
            position: 'absolute',
            top: 0, left: 0, right: 0, bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.85)',
            justifyContent: 'flex-end',
            alignItems: 'center',
            zIndex: 30,
          }}
        >
          <View className="w-full px-6 pb-12 items-center">
            <TouchableOpacity onPress={() => setShowActionSheet(false)} className="self-end mb-4">
              <View className="bg-white/20 rounded-full p-2">
                <Ionicons name="close" size={20} color="#fff" />
              </View>
            </TouchableOpacity>
            <TouchableOpacity
              className="flex-row items-center justify-center w-full py-6 rounded-2xl mb-4"
              style={{ backgroundColor: colors.primary }}
              onPress={() => { setShowActionSheet(false); setShowJoinModal(true); }}
            >
              <Ionicons name="people" size={28} color="#fff" style={{ marginRight: 12 }} />
              <Text className="font-hk-bold text-white text-xl">{t('games.joinGame')}</Text>
            </TouchableOpacity>
            <TouchableOpacity
              className="flex-row items-center justify-center w-full py-6 rounded-2xl"
              style={{ backgroundColor: colors.link }}
              onPress={() => { setShowActionSheet(false); setShowCreateModal(true); }}
            >
              <Ionicons name="add-circle" size={28} color="#fff" style={{ marginRight: 12 }} />
              <Text className="font-hk-bold text-white text-xl">{t('games.createGame')}</Text>
            </TouchableOpacity>
          </View>
        </Pressable>
      )}

      {/* Create modal */}
      {showCreateModal && (
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          style={{
            position: 'absolute',
            top: 0, left: 0, right: 0, bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.7)',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 10,
          }}
        >
          <View className="rounded-xl p-5 items-center" style={{ backgroundColor: colors.card, width: '80%' }}>
            <Text className="font-hk-bold text-foreground text-xl mb-1">{t('games.createNewGame')}</Text>
            <Text className="text-foreground-secondary text-sm mb-4">{t('games.createGameSubtitle')}</Text>
            <TextInput
              className="w-full text-lg text-center mb-5 rounded-md p-3"
              style={{ borderWidth: 1, borderColor: colors.border, color: colors.text, backgroundColor: colors.border }}
              value={newGameName}
              onChangeText={setNewGameName}
              placeholder={t('games.gameNamePlaceholder')}
              placeholderTextColor={colors.textSecondary}
              maxLength={40}
              autoFocus
            />
            <View className="flex-row justify-around w-full">
              <TouchableOpacity
                className="flex-1 mr-2.5 rounded-md py-2.5 px-5"
                style={{ backgroundColor: colors.textSecondary }}
                onPress={() => { setShowCreateModal(false); setNewGameName(''); }}
              >
                <Text className="font-hk-bold text-foreground text-base text-center">{t('common.cancel')}</Text>
              </TouchableOpacity>
              <TouchableOpacity
                className="flex-1 ml-2.5 rounded-md py-2.5 px-5"
                style={{ backgroundColor: colors.success }}
                onPress={handleCreateGame}
                disabled={creatingGame}
              >
                {creatingGame ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text className="font-hk-bold text-foreground text-base text-center">{t('common.create')}</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      )}

      {/* Join modal */}
      {showJoinModal && (
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          style={{
            position: 'absolute',
            top: 0, left: 0, right: 0, bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.7)',
            justifyContent: 'center',
            alignItems: 'center',
            zIndex: 10,
          }}
        >
          <View className="rounded-xl p-5 items-center" style={{ backgroundColor: colors.card, width: '80%' }}>
            <Text className="font-hk-bold text-foreground text-xl mb-1">{t('games.joinGame')}</Text>
            <Text className="text-foreground-secondary text-sm mb-4">{t('games.joinGameSubtitle')}</Text>
            <TextInput
              className="w-full text-lg text-center rounded-md p-3"
              style={{ borderWidth: 1, borderColor: colors.border, color: colors.text, backgroundColor: colors.border }}
              value={joinCode}
              onChangeText={setJoinCode}
              placeholder={t('games.gameCodePlaceholder')}
              placeholderTextColor={colors.textSecondary}
              maxLength={4}
              autoCapitalize="characters"
              autoFocus
            />
            <View className="flex-row justify-around w-full mt-5">
              <TouchableOpacity
                className="flex-1 mr-2.5 rounded-md py-2.5 px-5"
                style={{ backgroundColor: colors.textSecondary }}
                onPress={() => { setShowJoinModal(false); setJoinCode(''); }}
              >
                <Text className="font-hk-bold text-foreground text-base text-center">{t('common.cancel')}</Text>
              </TouchableOpacity>
              <TouchableOpacity
                className="flex-1 ml-2.5 rounded-md py-2.5 px-5"
                style={{ backgroundColor: colors.secondary }}
                onPress={handleJoinGame}
                disabled={joiningGame}
              >
                {joiningGame ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text className="font-hk-bold text-foreground text-base text-center">{t('common.join')}</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      )}
    </KeyboardAvoidingView>
  );
}

export default function TabOneScreen() {
  const { isLoading: isAuthLoading } = useAuth();
  const { games, loading: isGamesLoading } = useGames();
  const router = useRouter();

  useEffect(() => {
    if (isAuthLoading || isGamesLoading) return;
    if (hasAttemptedInitialRedirectGlobal) return;
    hasAttemptedInitialRedirectGlobal = true;
    if (games.length > 0) {
      router.replace('/(tabs)/matches');
    }
  }, [isAuthLoading, isGamesLoading, games.length, router]);

  if (isAuthLoading || isGamesLoading) {
    return (
      <View className="flex-1 justify-center items-center">
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  return <GamesList />;
}

let hasAttemptedInitialRedirectGlobal = false;
