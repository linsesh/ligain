import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert, ScrollView, RefreshControl, KeyboardAvoidingView, Platform, Clipboard } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';

// Local imports
import { useAuth } from '../../src/contexts/AuthContext';
import { API_CONFIG, getAuthenticatedHeaders } from '../../src/config/api';

interface Game {
  gameId: string;
  seasonYear: string;
  competitionName: string;
  name: string;
  status: string;
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

function GamesList() {
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
  const [newGameCode, setNewGameCode] = useState<string | null>(null);
  const [newGameName, setNewGameName] = useState('');

  const fetchGames = async () => {
    try {
      const headers = await getAuthenticatedHeaders();
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/games`, {
        headers,
      });
      
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(`${response.status}: ${errorData.error || 'Unknown error'}`);
      }
      
      const data = await response.json();
      let gamesList = data.games || [];
      // Add hardcoded game if not present
      const hardcodedGameId = '123e4567-e89b-12d3-a456-426614174000';
      if (!(gamesList as Game[]).some((g: Game) => g.gameId === hardcodedGameId)) {
        gamesList = [
          ...gamesList,
          {
            gameId: hardcodedGameId,
            seasonYear: '2025/2026',
            competitionName: 'Ligue 1',
            name: 'Ligue 1 Test Game',
            status: 'in progress',
          },
        ];
      }
      setGames(gamesList);
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
        const errorData = await response.json();
        throw new Error(`${response.status}: ${errorData.error || 'Unknown error'}`);
      }
      
      const data: CreateGameResponse = await response.json();
      setNewGameCode(data.code);
      
      // Close modal and reset form
      setShowCreateModal(false);
      setNewGameName('');
      
      // Refresh games list
      await fetchGames();
      
      Alert.alert(
        'Game Created!',
        `Your game code is: ${data.code}`,
        [
          { text: 'Copy Code', onPress: () => copyToClipboard(data.code) },
          { text: 'OK', onPress: () => setNewGameCode(null) }
        ]
      );
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
        const errorData = await response.json();
        throw new Error(`${response.status}: ${errorData.error || 'Unknown error'}`);
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

  const copyToClipboard = async (text: string) => {
    try {
      await Clipboard.setString(text);
      Alert.alert('Copied!', 'Game code copied to clipboard');
    } catch (err) {
      Alert.alert('Error', 'Failed to copy to clipboard');
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
        <ActivityIndicator size="large" color="#ffd33d" />
      </View>
    );
  }

  return (
    <KeyboardAvoidingView 
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <Text style={styles.title}>My Games</Text>
      
      {/* Guest Testing Banner */}
      {!player?.email && !player?.provider && (
        <View style={styles.guestBanner}>
          <Ionicons name="warning" size={20} color="#FFA500" />
          <Text style={styles.guestBannerText}>
            🧪 Guest Mode - This account is for testing purposes only
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
          <Text style={styles.actionButtonText}>Create Game</Text>
        </TouchableOpacity>
        
        <TouchableOpacity 
          style={[styles.actionButton, styles.joinButton, { marginLeft: 8 }]} 
          onPress={() => setShowJoinModal(true)}
        >
          <Ionicons name="people" size={20} color="#fff" />
          <Text style={styles.actionButtonText}>Join Game</Text>
        </TouchableOpacity>
      </View>

      {/* New Game Code Display */}
      {newGameCode && (
        <View style={styles.codeContainer}>
          <Text style={styles.codeLabel}>Your Game Code:</Text>
          <View style={styles.codeDisplay}>
            <Text style={styles.codeText}>{newGameCode}</Text>
            <TouchableOpacity 
              style={styles.copyButton}
              onPress={() => copyToClipboard(newGameCode)}
            >
              <Ionicons name="copy" size={16} color="#ffd33d" />
            </TouchableOpacity>
          </View>
        </View>
      )}

      <ScrollView 
        style={styles.scrollView}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={['#ffd33d']}
            tintColor="#ffd33d"
            progressBackgroundColor="#25292e"
            progressViewOffset={20}
          />
        }
      >
        {error ? (
          <View style={styles.errorContainer}>
            <Text style={styles.errorText}>Failed to load games: {error}</Text>
            <Text style={styles.refreshHint}>Pull down to refresh</Text>
          </View>
        ) : games.length === 0 ? (
          <View style={styles.emptyContainer}>
            <Ionicons name="game-controller-outline" size={64} color="#666" />
            <Text style={styles.emptyText}>No games yet</Text>
            <Text style={styles.emptySubtext}>Create a game or join one to get started!</Text>
          </View>
        ) : (
          games.map((game) => (
            <TouchableOpacity 
              key={game.gameId} 
              style={styles.gameCard}
              onPress={() => {
                console.log('🎮 Game card pressed, navigating to game:', game.gameId);
                router.push('/(tabs)/games/game/' + game.gameId);
              }}
            >
              <View style={styles.gameHeader}>
                <Text style={styles.gameTitle}>{game.name}</Text>
                <Text style={styles.gameSeason}>{game.seasonYear} • {game.competitionName}</Text>
              </View>
              <View style={styles.gameStatus}>
                <Text style={styles.statusText}>Status: {game.status}</Text>
              </View>
            </TouchableOpacity>
          ))
        )}
      </ScrollView>

      {/* Create Game Modal */}
      {showCreateModal && (
        <KeyboardAvoidingView
          behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
          style={styles.modalOverlay}
        >
          <View style={styles.modal}>
            <Text style={styles.modalTitle}>Create a New Game</Text>
            <Text style={styles.modalSubtitle}>Enter a name for your game</Text>
            
            <TextInput
              style={styles.gameNameInput}
              value={newGameName}
              onChangeText={setNewGameName}
              placeholder="My Awesome Game"
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
                <Text style={styles.cancelButtonText}>Cancel</Text>
              </TouchableOpacity>
              
              <TouchableOpacity 
                style={[styles.modalButton, styles.confirmButton]} 
                onPress={createGame}
                disabled={creatingGame}
              >
                {creatingGame ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text style={styles.confirmButtonText}>Create</Text>
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
            <Text style={styles.modalTitle}>Join a Game</Text>
            <Text style={styles.modalSubtitle}>Enter the 4-letter game code</Text>
            
            <TextInput
              style={styles.codeInput}
              value={joinCode}
              onChangeText={setJoinCode}
              placeholder="ABCD"
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
                <Text style={styles.cancelButtonText}>Cancel</Text>
              </TouchableOpacity>
              
              <TouchableOpacity 
                style={[styles.modalButton, styles.confirmButton]} 
                onPress={joinGame}
                disabled={joiningGame}
              >
                {joiningGame ? (
                  <ActivityIndicator size="small" color="#fff" />
                ) : (
                  <Text style={styles.confirmButtonText}>Join</Text>
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
    backgroundColor: '#2196F3',
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
    color: '#ffd33d',
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
  },
  gameHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
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
    color: '#ffd33d',
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
});
