import React, { useState, useEffect, useRef } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert, ScrollView, RefreshControl, KeyboardAvoidingView, Platform } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useRouter, useLocalSearchParams } from 'expo-router';
import { useFocusEffect } from '@react-navigation/native';
import { Swipeable } from 'react-native-gesture-handler';

// Local imports
import { useAuth } from '../../src/contexts/AuthContext';
import { colors } from '../../src/constants/colors';
import { getHumanReadableError, handleApiError, handleGameError } from '../../src/utils/errorMessages';
import { useTranslation } from 'react-i18next';
import Leaderboard from '../../src/components/Leaderboard';
import { useGames } from '../../src/contexts/GamesContext';
import { useUIEvent } from '../../src/contexts/UIEventContext';
import { API_CONFIG, getAuthenticatedHeaders, authenticatedFetch } from '../../src/config/api';
import { getTranslatedGameStatus } from '../../src/utils/gameStatusUtils';
import StatusTag from '../../src/components/StatusTag';

interface CreateGameResponse {
  gameId: string;
  code: string;
}

interface JoinGameResponse {
  gameId: string;
  seasonYear: string;
  competitionName: string;
  message: string;
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
  const [isAnySwipeActive, setIsAnySwipeActive] = useState<boolean>(false);
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

  // Leave game handler
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
              // Call backend with proper headers and URL
              const res = await authenticatedFetch(`${API_CONFIG.BASE_URL}/api/games/${gameId}/leave`, {
                method: 'DELETE',
              });
              if (!res.ok) {
                const data = await res.json().catch(() => ({}));
                throw new Error(data.error || t('games.failedToLeaveGame', 'Failed to leave game'));
              }
              // Remove game from local state immediately
              removeGame(gameId);
              Alert.alert(t('games.leftGame', 'You have left the game.'));
              // Still refresh to ensure consistency
              refresh();
            } catch (err: any) {
              Alert.alert(t('games.failedToLeaveGame', 'Failed to leave game'), err.message || String(err));
            }
          },
        },
      ]
    );
  };

  // Render right action for swipeable
  const renderRightActions = (gameId: string) => (
    <View style={{
      width: 75,
      height: '100%',
      justifyContent: 'center',
      alignItems: 'center',
      backgroundColor: 'transparent',
    }}>
      <TouchableOpacity
        style={{
          backgroundColor: '#ff3b30',
          justifyContent: 'center',
          alignItems: 'center',
          width: 60,
          height: 60,
          borderRadius: 8,
          shadowColor: '#000',
          shadowOffset: { width: 0, height: 2 },
          shadowOpacity: 0.1,
          shadowRadius: 4,
          elevation: 3,
        }}
        onPress={() => handleLeaveGame(gameId)}
      >
        <Ionicons name="exit-outline" size={20} color="#fff" />
        <Text style={{ 
          color: '#fff', 
          fontWeight: '600', 
          fontSize: 11,
          marginTop: 2
        }}>
          {t('games.leave', 'Leave')}
        </Text>
      </TouchableOpacity>
    </View>
  );

  const handleGamePress = (gameId: string) => {
    if (isAnySwipeActive) {
      return;
    }
    router.push('/(tabs)/games/game/overview/' + gameId);
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  if (loading && !refreshing) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: colors.loadingBackground }}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  if (!games) {
    return <Text style={{color: 'red'}}>{t('games.errorLoadingGames')}</Text>;
  }

  return (
    <KeyboardAvoidingView 
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <Text style={styles.title}>{t('games.myGames')}</Text>
      {!player?.email && !player?.provider && (
        <View style={styles.guestBanner}>
          <Ionicons name="warning" size={20} color="#FFA500" />
          <Text style={styles.guestBannerText}>
            {t('games.guestModeBanner')}
          </Text>
        </View>
      )}
      <ScrollView 
        style={styles.scrollView}
        contentContainerStyle={{ paddingBottom: 140 }}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            tintColor={colors.primary}
            progressBackgroundColor="#25292e"
            progressViewOffset={20}
          />
        }
      >
        {error ? (
          <View style={styles.errorContainer}>
            <Text style={styles.errorText}>{error}</Text>
            <Text style={styles.refreshHint}>{t('games.pullToRefresh')}</Text>
          </View>
        ) : games.length === 0 ? (
          <View style={styles.emptyContainer}>
            <Ionicons name="game-controller-outline" size={64} color="#666" />
            <Text style={styles.emptyText}>{t('games.noGames')}</Text>
            <Text style={styles.emptySubtext}>{t('games.noGamesSubtext')}</Text>
          </View>
        ) : (
          games.map((game) => {

            const { text, variant } = getTranslatedGameStatus(game.status || '', t);
            
            return (
              <Swipeable
                key={game.gameId}
                ref={ref => { swipeableRefs.current[game.gameId] = ref; }}
                renderRightActions={() => renderRightActions(game.gameId)}
                overshootRight={false}
                rightThreshold={40}
                onSwipeableWillOpen={() => {
                  setIsAnySwipeActive(true);
                }}
                onSwipeableClose={() => {
                  setIsAnySwipeActive(false);
                }}
              >
                <View pointerEvents={isAnySwipeActive ? 'none' : 'auto'}>
                  <TouchableOpacity 
                    style={[
                      styles.gameCard,
                      styles.gameCardWithTag
                    ]}
                    onPress={() => handleGamePress(game.gameId)}
                    activeOpacity={0.2}
                  >
                    <StatusTag text={text} variant={variant} style={styles.statusTag} />
                    <View style={styles.headerGroupAbsolute}>
                      <Text style={styles.leagueNamePlain}>{game.name}</Text>
                      <Text style={styles.gameSeasonPlain}>{game.seasonYear} • {game.competitionName}</Text>
                    </View>
                    {game.totalLeaderboard && game.totalLeaderboard.length > 0 && (
                      <Leaderboard
                        players={game.totalLeaderboard.map(p => ({ id: p.PlayerID, name: p.PlayerName, totalScore: p.Points }))}
                        t={t}
                        showTitle={false}
                        align="left"
                        showCurrentPlayerTag={true}
                        currentPlayerId={player?.id}
                        containerStyle={{ marginHorizontal: 0, alignItems: 'flex-start', paddingTop: 0, paddingBottom: 0, paddingLeft: 0, paddingRight: 0 }}
                      />
                    )}
                  </TouchableOpacity>
                </View>
              </Swipeable>
            );
          })
        )}
      </ScrollView>
      {!showActionSheet && !showCreateModal && !showJoinModal && (
        <View style={styles.bottomButtonContainer}>
          <TouchableOpacity
            style={styles.bigJoinButton}
            onPress={() => setShowActionSheet(true)}
            activeOpacity={0.85}
          >
            <Ionicons name="add-circle" size={28} color="#fff" style={{ marginRight: 10 }} />
            <Text style={styles.bigJoinButtonText}>{t('games.joinOrCreate')}</Text>
          </TouchableOpacity>
        </View>
      )}
      {showActionSheet && (
        <View style={styles.actionSheetOverlay}>
          <View style={styles.actionSheetButtonsContainer}>
            <TouchableOpacity
              style={[styles.actionSheetButton, styles.actionSheetJoin]}
              onPress={() => {
                setShowActionSheet(false);
                setShowJoinModal(true);
              }}
            >
              <Ionicons name="people" size={22} color="#fff" style={styles.actionSheetButtonIcon} />
              <Text style={styles.actionSheetButtonText}>{t('games.joinGame')}</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.actionSheetButton, styles.actionSheetCreate]}
              onPress={() => {
                setShowActionSheet(false);
                setShowCreateModal(true);
              }}
            >
              <Ionicons name="add-circle" size={22} color="#fff" style={styles.actionSheetButtonIcon} />
              <Text style={styles.actionSheetButtonText}>{t('games.createGame')}</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={styles.closeTextContainer}
              onPress={() => {
                setShowActionSheet(false);
              }}
            >
              <Text style={styles.closeText}>✕ {t('common.close')}</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}
      {showCreateModal && (
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          style={styles.modalOverlay}
        >
          <View style={styles.modal}>
            <Text style={styles.modalTitle}>{t('games.createNewGame')}</Text>
            <Text style={styles.modalSubtitle}>{t('games.createGameSubtitle')}</Text>
            <TextInput
              style={styles.gameNameInput}
              value={newGameName}
              onChangeText={setNewGameName}
              placeholder={t('games.gameNamePlaceholder')}
              placeholderTextColor="#666"
              maxLength={40}
              autoFocus
            />
            <View style={styles.modalButtons}>
              <TouchableOpacity 
                style={[styles.modalButton, styles.cancelButton]} 
                onPress={() => {
                  setShowCreateModal(false);
                  setNewGameName('');
                }}
              >
                <Text style={styles.cancelButtonText}>{t('common.cancel')}</Text>
              </TouchableOpacity>
              <TouchableOpacity 
                style={[styles.modalButton, styles.confirmCreateButton]} 
                onPress={handleCreateGame}
                disabled={creatingGame}
              >
                {creatingGame ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text style={styles.confirmButtonText}>{t('common.create')}</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      )}
      {showJoinModal && (
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          style={styles.modalOverlay}
        >
          <View style={styles.modal}>
            <Text style={styles.modalTitle}>{t('games.joinGame')}</Text>
            <Text style={styles.modalSubtitle}>{t('games.joinGameSubtitle')}</Text>
            <TextInput
              style={styles.codeInput}
              value={joinCode}
              onChangeText={setJoinCode}
              placeholder={t('games.gameCodePlaceholder')}
              placeholderTextColor="#666"
              maxLength={4}
              autoCapitalize="characters"
              autoFocus
            />
            <View style={styles.modalButtons}>
              <TouchableOpacity 
                style={[styles.modalButton, styles.cancelButton]} 
                onPress={() => {
                  setShowJoinModal(false);
                  setJoinCode('');
                }}
              >
                <Text style={styles.cancelButtonText}>{t('common.cancel')}</Text>
              </TouchableOpacity>
              <TouchableOpacity 
                style={[styles.modalButton, styles.confirmJoinButton]} 
                onPress={handleJoinGame}
                disabled={joiningGame}
              >
                {joiningGame ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text style={styles.confirmButtonText}>{t('common.join')}</Text>
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
  const { player, isLoading: isAuthLoading } = useAuth();
  const { games, loading: isGamesLoading } = useGames();
  const [hasAttemptedInitialRedirect, setHasAttemptedInitialRedirect] = useState(false);
  const router = useRouter();

  // Auto-redirect to matches if user has games (only on initial app load)
  useEffect(() => {
    if (!isAuthLoading && !isGamesLoading && games.length > 0 && !hasAttemptedInitialRedirect) {
      console.log('🔄 Auto-redirecting to matches on initial load - user has games:', games.length);
      setHasAttemptedInitialRedirect(true);
      router.replace('/(tabs)/matches');
    } else if (!isAuthLoading && !isGamesLoading && !hasAttemptedInitialRedirect) {
      // Mark as attempted even if no redirect happened (no games)
      setHasAttemptedInitialRedirect(true);
    }
  }, [isAuthLoading, isGamesLoading, games.length, router, hasAttemptedInitialRedirect]);

  if (isAuthLoading || isGamesLoading) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: colors.loadingBackground }}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  return <GamesList />;
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#25292e',
  },
  scrollView: {
    flex: 1,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    margin: 16,
    color: '#fff',
  },
  guestBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#e65100',
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 8,
    marginHorizontal: 16,
    marginBottom: 16,
  },
  guestBannerText: {
    fontSize: 14,
    color: '#fff',
    marginLeft: 10,
    fontWeight: '600',
    flex: 1,
  },
  actionButtons: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    marginHorizontal: 16,
    marginBottom: 16,
  },
  actionButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    paddingVertical: 12,
    paddingHorizontal: 20,
    borderRadius: 8,
    gap: 8,
    flex: 1,
  },
  createButton: {
    backgroundColor: '#4CAF50',
    marginRight: 8,
  },
  joinButton: {
    backgroundColor: colors.secondary,
    marginLeft: 8,
  },
  actionButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  codeContainer: {
    backgroundColor: '#333',
    padding: 16,
    borderRadius: 8,
    marginHorizontal: 16,
    marginBottom: 16,
    alignItems: 'center',
  },
  codeLabel: {
    fontSize: 16,
    color: '#fff',
    marginBottom: 8,
  },
  codeDisplay: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#444',
    borderRadius: 6,
    padding: 8,
  },
  codeText: {
    fontSize: 24,
    fontWeight: 'bold',
    color: colors.primary,
    marginRight: 8,
  },
  copyButton: {
    padding: 4,
  },
  errorContainer: {
    padding: 20,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 200,
  },
  errorText: {
    color: 'red',
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 20,
  },
  refreshHint: {
    color: '#666',
    fontSize: 14,
    textAlign: 'center',
  },
  emptyContainer: {
    alignItems: 'center',
    padding: 20,
    minHeight: 200,
  },
  emptyText: {
    fontSize: 20,
    color: '#666',
    marginTop: 10,
  },
  emptySubtext: {
    fontSize: 14,
    color: '#999',
    marginTop: 5,
  },
  gameCard: {
    backgroundColor: '#333',
    padding: 16,
    borderRadius: 8,
    marginHorizontal: 16,
    marginBottom: 12,
    position: 'relative',
  },
  gameCardWithTag: {
    paddingTop: 64,
  },
  gameHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  leagueNamePlain: {
    color: '#fff',
    fontSize: 22,
    fontWeight: 'bold',
    textAlign: 'left',
    marginTop: 0,
    marginLeft: 0,
    marginBottom: 0,
  },
  gameSeasonPlain: {
    color: '#999',
    fontSize: 15,
    textAlign: 'left',
    marginLeft: 0,
    marginTop: 0,
    marginBottom: 0,
  },
  gameTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
  },
  gameSeason: {
    fontSize: 14,
    color: '#999',
  },
  gameStatus: {
    backgroundColor: '#444',
    padding: 8,
    borderRadius: 4,
  },
  statusText: {
    fontSize: 14,
    color: colors.primary,
    fontWeight: 'bold',
  },
  modalOverlay: {
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
  modal: {
    backgroundColor: '#25292e',
    borderRadius: 10,
    padding: 20,
    width: '80%',
    alignItems: 'center',
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 5,
  },
  modalSubtitle: {
    fontSize: 14,
    color: '#999',
    marginBottom: 15,
  },
  gameNameInput: {
    borderWidth: 1,
    borderColor: '#666',
    borderRadius: 6,
    padding: 12,
    width: '100%',
    fontSize: 18,
    color: '#fff',
    backgroundColor: '#444',
    textAlign: 'center',
    marginBottom: 20,
  },
  codeInput: {
    borderWidth: 1,
    borderColor: '#666',
    borderRadius: 6,
    padding: 12,
    width: '100%',
    fontSize: 18,
    color: '#fff',
    backgroundColor: '#444',
    textAlign: 'center',
  },
  modalButtons: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    width: '100%',
    marginTop: 20,
  },
  modalButton: {
    paddingVertical: 10,
    paddingHorizontal: 20,
    borderRadius: 6,
  },
  cancelButton: {
    backgroundColor: '#666',
    flex: 1,
    marginRight: 10,
  },
  cancelButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  confirmJoinButton: {
    backgroundColor: colors.secondary,
    flex: 1,
    marginLeft: 10,
  },
  confirmCreateButton: {
    backgroundColor: '#4CAF50',
    flex: 1,
    marginLeft: 10,
  },
  confirmButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  statusTag: {
    position: 'absolute',
    top: 8,
    right: 8,
    zIndex: 1,
  },
  headerGroupAbsolute: {
    position: 'absolute',
    top: 8,
    left: 8,
    width: '80%',
    alignItems: 'flex-start',
    zIndex: 2,
  },
  bottomButtonContainer: {
    position: 'absolute',
    left: 0,
    right: 0,
    bottom: 24,
    alignItems: 'center',
    zIndex: 20,
  },
  bigJoinButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.secondary,
    borderRadius: 32,
    paddingVertical: 18,
    paddingHorizontal: 40,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 6,
    elevation: 4,
  },
  bigJoinButtonText: {
    color: '#fff',
    fontSize: 20,
    fontWeight: 'bold',
  },
  actionSheetOverlay: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0,0,0,0.85)',
    justifyContent: 'flex-end',
    alignItems: 'center',
    zIndex: 30,
  },
  actionSheetButtonsContainer: {
    width: '100%',
    paddingHorizontal: 24,
    paddingBottom: 48,
    alignItems: 'center',
  },
  actionSheetButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    width: '100%',
    paddingVertical: 18,
    borderRadius: 16,
    marginBottom: 16,
  },
  actionSheetButtonIcon: {
    marginRight: 12,
  },
  actionSheetJoin: {
    backgroundColor: colors.secondary,
  },
  actionSheetCreate: {
    backgroundColor: '#4CAF50',
  },
  actionSheetButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  closeTextContainer: {
    marginTop: 8,
    alignItems: 'center',
    width: '100%',
  },
  closeText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
    opacity: 0.7,
    paddingVertical: 8,
  },
});
