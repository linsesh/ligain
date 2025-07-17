import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TouchableOpacity, ScrollView, RefreshControl, Alert, Clipboard } from 'react-native';
import * as Haptics from 'expo-haptics';
import { Ionicons } from '@expo/vector-icons';
import { useLocalSearchParams, useRouter } from 'expo-router';

// Local imports
import { useAuth } from '../../../../../src/contexts/AuthContext';
import { API_CONFIG, getAuthenticatedHeaders } from '../../../../../src/config/api';

interface PlayerGameInfo {
  id: string;
  name: string;
  totalScore: number;
  scoresByMatch: { [key: string]: number };
}

interface GameDetails {
  gameId: string;
  seasonYear: string;
  competitionName: string;
  name: string;
  status: string;
  players: PlayerGameInfo[];
  code: string;
}

export default function GameOverviewScreen() {
  const { id: gameId } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { player } = useAuth();
  const [gameDetails, setGameDetails] = useState<GameDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const fetchGameDetails = async () => {
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
      const games = data.games || [];
      const currentGame = games.find((g: any) => g.gameId === gameId);
      
      if (!currentGame) {
        throw new Error('Game not found');
      }
      
      setGameDetails(currentGame);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch game details');
    } finally {
      setLoading(false);
    }
  };



  const copyToClipboard = async (text: string) => {
    if (copied) return; // Do nothing if already in check mode
    try {
      await Clipboard.setString(text);
      setCopied(true);
      Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
      // Reset after 3 seconds
      setTimeout(() => {
        setCopied(false);
      }, 3000);
    } catch (err) {
      Alert.alert('Error', 'Failed to copy to clipboard');
    }
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await fetchGameDetails();
    setRefreshing(false);
  }, [gameId]);

  useEffect(() => {
    if (gameId) {
      fetchGameDetails();
    }
  }, [gameId]);

  const navigateToMatches = () => {
    router.push(`/(tabs)/games/game/${gameId}`);
  };









  if (loading && !refreshing) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="large" color="#ffd33d" />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.container}>
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>Failed to load game: {error}</Text>
          <TouchableOpacity style={styles.retryButton} onPress={fetchGameDetails}>
            <Text style={styles.retryButtonText}>Retry</Text>
          </TouchableOpacity>
        </View>
      </View>
    );
  }

  if (!gameDetails) {
    return (
      <View style={styles.container}>
        <Text style={styles.errorText}>Game not found</Text>
      </View>
    );
  }

  // Sort players by total score (descending)
  const sortedPlayers = [...gameDetails.players].sort((a, b) => b.totalScore - a.totalScore);

  return (
    <View style={styles.container}>
      <ScrollView 
        style={styles.scrollView}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={['#ffd33d']}
            tintColor="#ffd33d"
            progressBackgroundColor="#25292e"
          />
        }
      >
        {/* Game Header */}
        <View style={styles.gameHeader}>
          <Text style={styles.gameTitle}>{gameDetails.name}</Text>
          <Text style={styles.gameSubtitle}>
            {gameDetails.seasonYear} • {gameDetails.competitionName}
          </Text>
          <Text style={styles.gameStatus}>Status: {gameDetails.status}</Text>
        </View>

        {/* Game Code Section */}
        {gameDetails.code && (
          <View style={styles.codeContainer}>
            <Text style={styles.codeLabel}>Game Code:</Text>
            <View style={styles.codeDisplay}>
              <Text style={styles.codeText}>{gameDetails.code}</Text>
              <TouchableOpacity 
                style={styles.copyButton}
                onPress={() => copyToClipboard(gameDetails.code)}
                disabled={copied}
              >
                <Ionicons name={copied ? "checkmark-circle" : "copy"} size={20} color="#ffd33d" />
              </TouchableOpacity>
            </View>
          </View>
        )}

        {/* Player Leaderboard */}
        <View style={styles.leaderboardContainer}>
          <Text style={styles.leaderboardTitle}>Player Leaderboard</Text>
          {sortedPlayers.map((playerInfo, index) => (
            <View key={playerInfo.id} style={styles.playerRow}>
              <View style={styles.playerRank}>
                <Text style={styles.rankText}>{index + 1}</Text>
              </View>
              <View style={styles.playerInfo}>
                <Text style={styles.playerName}>{playerInfo.name}</Text>
                <Text style={styles.playerScore}>{playerInfo.totalScore} points</Text>
              </View>
              {playerInfo.id === player?.id && (
                <View style={styles.currentPlayerIndicator}>
                  <Text style={styles.currentPlayerText}>You</Text>
                </View>
              )}
            </View>
          ))}
        </View>

        {/* Navigation Button */}
        <TouchableOpacity 
          style={styles.matchesButton}
          onPress={navigateToMatches}
        >
          <Ionicons name="football" size={24} color="#fff" />
          <Text style={styles.matchesButtonText}>View Matches</Text>
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
    backgroundColor: '#4CAF50',
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