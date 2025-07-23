import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, Platform, ActivityIndicator } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Picker } from '@react-native-picker/picker';
import { useGames } from '../../src/contexts/GamesContext';
import { useTranslation } from 'react-i18next';
import MatchesList from './games/game/_MatchesList';
import { useLocalSearchParams } from 'expo-router';
import { useRouter } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { colors } from '../../src/constants/colors';
import { useUIEvent } from '../../src/contexts/UIEventContext';

export default function MatchesTabScreen() {
  const { t } = useTranslation();
  const { games, selectedGameId, setSelectedGameId, bestGameId, loading } = useGames();
  const { player, isLoading: isAuthLoading } = useAuth();
  const router = useRouter();
  const { setOpenJoinOrCreate } = useUIEvent();
  const [showGamePicker, setShowGamePicker] = useState(false);
  const params = useLocalSearchParams();

  // If a gameId is present in the query params, select it on mount
  useEffect(() => {
    if (params.gameId && games.some(g => g.gameId === params.gameId)) {
      setSelectedGameId(params.gameId as string);
    } else if (!selectedGameId && games.length > 0) {
      // Use the intelligently determined bestGameId instead of just the first game
      setSelectedGameId(bestGameId || games[0].gameId);
    }
  }, [params.gameId, games, setSelectedGameId, selectedGameId, bestGameId]);

  if (isAuthLoading || loading) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: colors.loadingBackground }}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  // If user has no games, show a message and a button to go to Games
  if (games.length === 0) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#25292e' }}>
        <Text style={{ color: '#fff', fontSize: 18, marginBottom: 16 }}>{t('games.noGames')}</Text>
        <TouchableOpacity
          style={{ backgroundColor: colors.secondary, paddingVertical: 14, paddingHorizontal: 32, borderRadius: 999, marginTop: 8 }}
          onPress={() => {
            setOpenJoinOrCreate(true);
            router.replace('/(tabs)/index');
          }}
        >
          <Text style={{ color: '#fff', fontWeight: 'bold', fontSize: 16 }}>{t('games.goToGames')}</Text>
        </TouchableOpacity>
      </View>
    );
  }

  const selectedGame = games.find(g => g.gameId === selectedGameId);

  return (
    <View style={styles.container}>
      {/* Game Selector */}
      <View style={styles.gameSelectionContainer}>
        <TouchableOpacity 
          style={styles.gameSelector}
          onPress={() => setShowGamePicker(true)}
        >
          <Text style={styles.gameSelectorText}>
            {selectedGame ? selectedGame.name : t('games.selectGame')}
          </Text>
          <Ionicons name="chevron-down" size={20} color="#fff" />
        </TouchableOpacity>
        {selectedGame && (
          <Text style={styles.gameInfoText}>
            {selectedGame.seasonYear} • {selectedGame.competitionName}
          </Text>
        )}
      </View>
      {/* Game Picker Modal */}
      {showGamePicker && (
        <View style={styles.pickerOverlay}>
          <View style={styles.pickerContainer}>
            <View style={styles.pickerHeader}>
              <Text style={styles.pickerTitle}>{t('games.selectGame')}</Text>
              <TouchableOpacity onPress={() => setShowGamePicker(false)}>
                <Ionicons name="close" size={24} color="#fff" />
              </TouchableOpacity>
            </View>
            <Picker
              selectedValue={selectedGameId}
              onValueChange={(itemValue) => {
                setSelectedGameId(itemValue);
                setShowGamePicker(false);
              }}
              style={styles.picker}
              itemStyle={styles.pickerItem}
            >
              {games.map((game) => (
                <Picker.Item 
                  key={game.gameId} 
                  label={game.name} 
                  value={game.gameId}
                  color="#fff"
                />
              ))}
            </Picker>
          </View>
        </View>
      )}
      {/* Old MatchesList UI for the selected game */}
      {selectedGameId && (
        <View style={{ flex: 1 }}>
          <MatchesList gameId={selectedGameId} />
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#25292e',
  },
  gameSelectionContainer: {
    paddingHorizontal: 16,
    marginTop: 16,
    marginBottom: 8,
  },
  gameSelector: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: '#333',
    padding: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#444',
  },
  gameSelectorText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
  gameInfoText: {
    color: '#999',
    fontSize: 14,
    marginTop: 4,
    marginLeft: 4,
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
  picker: {
    backgroundColor: '#333',
  },
  pickerItem: {
    color: '#fff',
  },
}); 