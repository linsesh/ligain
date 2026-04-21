import React, { useState, useEffect } from 'react';
import { View, StyleSheet, TouchableOpacity } from 'react-native';
import { Text } from '../../src/components/ui/Text';
import { Ionicons } from '@expo/vector-icons';
import { Picker } from '@react-native-picker/picker';
import { useGames } from '../../src/contexts/GamesContext';
import { useTranslation } from 'react-i18next';
import MatchesList from '../../src/components/MatchesList';
import { SeasonBanner } from '../../src/components/SeasonBanner';
import { useLocalSearchParams } from 'expo-router';
import { useRouter } from 'expo-router';
import { useAuth } from '../../src/contexts/AuthContext';
import { colors } from '../../src/constants/colors';
import { useUIEvent } from '../../src/contexts/UIEventContext';
import { MatchesScreenSkeleton } from '../../src/components/MatchesScreenSkeleton';

export default function MatchesTabScreen() {
  const { t } = useTranslation();
  const { games, selectedGameId, setSelectedGameId, bestGameId, loading } = useGames();
  const { isLoading: isAuthLoading } = useAuth();
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
    return <MatchesScreenSkeleton />;
  }

  // If user has no games, show a message and a button to go to Games
  if (games.length === 0) {
    return (
      <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
        <Text style={{ color: colors.text, fontSize: 18, marginBottom: 16 }}>{t('games.noGames')}</Text>
        <TouchableOpacity
          style={{ backgroundColor: colors.secondary, paddingVertical: 14, paddingHorizontal: 32, borderRadius: 999, marginTop: 8 }}
          onPress={() => {
            setOpenJoinOrCreate(true);
            router.replace('/(tabs)/index');
          }}
        >
          <Text className="font-hk-bold" style={{ color: colors.text, fontSize: 16 }}>{t('games.goToGames')}</Text>
        </TouchableOpacity>
      </View>
    );
  }

  const selectedGame = games.find(g => g.gameId === selectedGameId);
  const initialMatchday = selectedGame?.closestUnfinishedMatchday?.matchday || undefined;

  return (
    <View style={styles.container}>
      {/* Game title — tappable to open game picker */}
      <TouchableOpacity
        onPress={() => setShowGamePicker(true)}
        className="flex-row items-center justify-center mb-5 gap-2"
      >
        <Text className="font-hk-extrabold text-center text-4xl">
          {selectedGame ? (selectedGame.name.length > 12 ? selectedGame.name.slice(0, 12) + '…' : selectedGame.name) : t('games.selectGame')}
        </Text>
        <Ionicons name="chevron-down" size={24} color={colors.text} />
      </TouchableOpacity>

      {/* Season banner */}
      {selectedGame && (
        <SeasonBanner className="mb-4"
          seasonYear={selectedGame.seasonYear}
          competitionName={selectedGame.competitionName}
        />
      )}

      {/* Match list */}
      {selectedGameId && (
        <View style={{ flex: 1 }}>
          <MatchesList
            key={selectedGameId}
            gameId={selectedGameId}
            initialMatchday={initialMatchday}
            activeMatchday={selectedGame?.closestUnfinishedMatchday?.matchday}
          />
        </View>
      )}

      {/* Game Picker Modal */}
      {showGamePicker && (
        <View style={styles.pickerOverlay}>
          <View style={styles.pickerContainer}>
            <View style={styles.pickerHeader}>
              <Text className="font-hk-bold" style={styles.pickerTitle}>{t('games.selectGame')}</Text>
              <TouchableOpacity onPress={() => setShowGamePicker(false)}>
                <Ionicons name="close" size={24} color={colors.text} />
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

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: 'transparent',
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
    backgroundColor: colors.card,
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
    borderBottomColor: colors.border,
  },
  pickerTitle: {
    color: colors.text,
    fontSize: 18,
  },
  picker: {
    backgroundColor: colors.card,
  },
  pickerItem: {
    color: colors.text,
    fontFamily: 'HKGroteskWide-Regular',
  },
}); 