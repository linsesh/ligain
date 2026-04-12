import React from 'react';
import { View, FlatList, ViewStyle } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Text } from './ui/Text';
import { PlayerAvatar } from './PlayerAvatar';
import { colors } from '../constants/colors';

interface Player {
  id: string;
  name: string;
  avatarUrl?: string | null;
}

interface PlayerBetsBarProps {
  players: Player[];
  playerBetStatuses: Record<string, { hasBet: boolean }> | null;
  style?: ViewStyle;
}

function PlayerBetsItem({ player, hasBet }: { player: Player; hasBet: boolean }) {
  return (
    <View style={{ alignItems: 'center', width: 64, marginHorizontal: 4 }}>
      <View style={{ position: 'relative' }}>
        <PlayerAvatar player={player} displaySize="medium" />
        {hasBet && (
          <View
            style={{
              position: 'absolute',
              bottom: 0,
              right: 0,
              width: 18,
              height: 18,
              borderRadius: 9,
              backgroundColor: colors.formWin,
              justifyContent: 'center',
              alignItems: 'center',
              borderWidth: 1.5,
              borderColor: colors.background,
            }}
          >
            <Ionicons name="checkmark" size={11} color={colors.white} />
          </View>
        )}
      </View>
      <Text
        numberOfLines={1}
        className="font-hk-medium"
        style={{ fontSize: 11, color: colors.textSecondary, marginTop: 4, textAlign: 'center', width: 64 }}
      >
        {player.name}
      </Text>
    </View>
  );
}

export function PlayerBetsBar({ players, playerBetStatuses, style }: PlayerBetsBarProps) {
  if (players.length === 0) return null;

  return (
    <FlatList
      horizontal
      showsHorizontalScrollIndicator={false}
      data={players}
      keyExtractor={(item) => item.id}
      style={[{ width: '100%' }, style]}
      contentContainerStyle={{ paddingHorizontal: 16, paddingVertical: 16 }}
      renderItem={({ item }) => (
        <PlayerBetsItem
          player={item}
          hasBet={playerBetStatuses?.[item.id]?.hasBet ?? false}
        />
      )}
    />
  );
}
