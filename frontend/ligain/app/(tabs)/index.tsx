import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert, ScrollView, RefreshControl, KeyboardAvoidingView, Platform } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';

// Local imports
import { useAuth } from '../../src/contexts/AuthContext';
import { API_CONFIG, getAuthenticatedHeaders } from '../../src/config/api';
import { colors } from '../../src/constants/colors';
import { getHumanReadableError, handleApiError } from '../../src/utils/errorMessages';
import { useTranslation } from 'react-i18next';
import Leaderboard from '../../src/components/Leaderboard';

interface Game {
  gameId: string;
  seasonYear: string;
  competitionName: string;
  name: string;
  status: string;
  players?: any[];
}

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

// Reusable StatusTag component (factorized from match card)
function StatusTag({ text, variant }: { text: string; variant: string }) {
  const baseStyle = [styles.statusTag];
  let variantStyle = null;
  if (variant === 'success') variantStyle = styles.successTag;
  else if (variant === 'warning') variantStyle = styles.inProgressTag;
  else if (variant === 'finished') variantStyle = styles.finishedTag;
  else if (variant === 'primary') variantStyle = styles.primaryTag;
  return (
    <View style={[...baseStyle, variantStyle]}>
      <Text style={styles.statusTagText}>{text}</Text>
    </View>
  );
}

function GamesList() {
  const { t } = useTranslation();
  const { player } = useAuth();
  const router = useRouter();
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshing, setRefreshing] = useState(false);
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [joinCode, setJoinCode] = useState('');
  const [creatingGame, setCreatingGame] = useState(false);
  const [joiningGame, setJoiningGame] = useState(false);
  const [newGameName, setNewGameName] = useState('');

  const fetchGames = async () => {
    try {
      const headers = await getAuthenticatedHeaders();
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/games`, {
        headers,
      });
      
      if (!response.ok) {
        await handleApiError(response);
      }
      
      const data = await response.json();
      setGames(data.games || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch games');
    } finally {
      setLoading(false);
    }
  };

  const createGame = async () => {
    if (!newGameName.trim()) {
      Alert.alert('Error', 'Please enter a game name');
      return;
    }
    setCreatingGame(true);
    try {
      const headers = await getAuthenticatedHeaders({
        'Content-Type': 'application/json',
      });
      
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/games`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          seasonYear: '2025/2026',
          competitionName: 'Ligue 1',
          name: newGameName.trim(),
        }),
      });
      
      if (!response.ok) {
        await handleApiError(response);
      }
      
      const data: CreateGameResponse = await response.json();
      
      // Close modal and reset form
      setShowCreateModal(false);
      setNewGameName('');
      
      // Refresh games list
      await fetchGames();
      
      // Navigate to the new game overview page
      router.push(`/(tabs)/games/game/overview/${data.gameId}`);
    } catch (err) {
      Alert.alert('Error', err instanceof Error ? err.message : 'Failed to create game');
    } finally {
      setCreatingGame(false);
    }
  };

  const joinGame = async () => {
    if (!joinCode.trim()) {
      Alert.alert('Error', 'Please enter a game code');
      return;
    }
    
    setJoiningGame(true);
    try {
      const headers = await getAuthenticatedHeaders({
        'Content-Type': 'application/json',
      });
      
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/games/join`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          code: joinCode.trim().toUpperCase()
        }),
      });
      
      if (!response.ok) {
        await handleApiError(response);
      }
      
      const data: JoinGameResponse = await response.json();
      setJoinCode('');
      setShowJoinModal(false);
      
      // Refresh games list
      await fetchGames();
      
      Alert.alert('Success', data.message);
    } catch (err) {
      Alert.alert('Error', err instanceof Error ? err.message : 'Failed to join game');
    } finally {
      setJoiningGame(false);
    }
  };



  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await fetchGames();
    setRefreshing(false);
  }, []);

  useEffect(() => {
    fetchGames();
  }, []);

  if (loading && !refreshing) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  return (
    <KeyboardAvoidingView 
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <Text style={styles.title}>{t('games.myGames')}</Text>
      
      {/* Guest Testing Banner */}
      {!player?.email && !player?.provider && (
        <View style={styles.guestBanner}>
          <Ionicons name="warning" size={20} color="#FFA500" />
          <Text style={styles.guestBannerText}>
            {t('games.guestModeBanner')}
          </Text>
        </View>
      )}
      
      {/* Action Buttons */}
      <View style={styles.actionButtons}>
        <TouchableOpacity 
          style={[styles.actionButton, styles.createButton, { marginRight: 8 }]} 
          onPress={() => setShowCreateModal(true)}
        >
          <Ionicons name="add-circle" size={20} color="#fff" />
          <Text style={styles.actionButtonText}>{t('games.createGame')}</Text>
        </TouchableOpacity>
        
        <TouchableOpacity 
          style={[styles.actionButton, styles.joinButton, { marginLeft: 8 }]} 
          onPress={() => setShowJoinModal(true)}
        >
          <Ionicons name="people" size={20} color="#fff" />
          <Text style={styles.actionButtonText}>{t('games.joinGame')}</Text>
        </TouchableOpacity>
      </View>



      <ScrollView 
        style={styles.scrollView}
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
            <Text style={styles.errorText}>{t('games.errorLoadingGames')} {error}</Text>
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
            console.log('Game status:', game.status);
            const status = (game.status || '').toLowerCase();
            return (
              <TouchableOpacity 
                key={game.gameId} 
                style={[
                  styles.gameCard,
                  styles.gameCardWithTag
                ]}
                onPress={() => {
                  console.log('🎮 Game card pressed, navigating to game overview:', game.gameId);
                  router.push('/(tabs)/games/game/overview/' + game.gameId);
                }}
              >
                <StatusTag text={String(game.status)} variant="warning" />
                <View style={styles.headerGroupAbsolute}>
                  <Text style={styles.leagueNamePlain}>{game.name}</Text>
                  <Text style={styles.gameSeasonPlain}>{game.seasonYear} • {game.competitionName}</Text>
                </View>
                {/* If game.players exists, render the leaderboard */}
                {game.players && Array.isArray(game.players) && (
                  <Leaderboard
                    players={game.players}
                    t={t}
                    showTitle={false}
                    align="left"
                    showCurrentPlayerTag={true}
                    currentPlayerId={player?.id}
                    containerStyle={{ marginHorizontal: 0, alignItems: 'flex-start', paddingTop: 0, paddingBottom: 0, paddingLeft: 0, paddingRight: 0 }}
                  />
                )}
              </TouchableOpacity>
            );
          })
        )}
      </ScrollView>

      {/* Create Game Modal */}
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
                style={[styles.modalButton, styles.confirmButton]} 
                onPress={createGame}
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

      {/* Join Game Modal */}
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
                style={[styles.modalButton, styles.confirmButton]} 
                onPress={joinGame}
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
  console.log('🏠 TabOneScreen - Rendering games screen');
  
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
    backgroundColor: colors.primary,
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
  confirmButton: {
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
    backgroundColor: 'rgba(0,0,0,0.7)',
    paddingVertical: 4,
    paddingHorizontal: 8,
    borderRadius: 5,
    zIndex: 1,
  },
  statusTagText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: 'bold',
  },
  successTag: {
    backgroundColor: '#4CAF50',
  },
  inProgressTag: {
    backgroundColor: '#FFC107',
  },
  finishedTag: {
    backgroundColor: '#9E9E9E',
  },
  primaryTag: {
    backgroundColor: colors.primary,
  },
  headerGroupAbsolute: {
    position: 'absolute',
    top: 8,
    left: 8,
    width: '80%',
    alignItems: 'flex-start',
    zIndex: 2,
  },
});
