import React, { useState } from 'react';
import { View, FlatList, ViewStyle, Pressable } from 'react-native';
import * as Haptics from 'expo-haptics';
import { Portal } from '@rn-primitives/portal';
import { Ionicons } from '@expo/vector-icons';
import { Text } from './ui/Text';
import { PlayerAvatar } from './PlayerAvatar';
import { PlayerBetDetailsOverlay } from './PlayerBetDetailsOverlay';
import { colors } from '../constants/colors';

interface Player {
  id: string;
  name: string;
  avatarUrl?: string | null;
}

interface PlayerBetsBarProps {
  players: Player[];
  playerBetStatuses?: Record<string, { hasBet: boolean }> | null;
  playerScores?: Record<string, { points: number; baseScore?: number; riskMultiplier?: number; clairvoyantMultiplier?: number }> | null;
  playerBets?: Record<string, { predictedHomeGoals: number; predictedAwayGoals: number }> | null;
  homeTeam?: string;
  awayTeam?: string;
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

function PlayerBetPredictionItem({ player, homeGoals, awayGoals }: { player: Player; homeGoals: number | null; awayGoals: number | null }) {
  const label = homeGoals !== null && awayGoals !== null ? `${homeGoals} - ${awayGoals}` : 'x';

  return (
    <View style={{ alignItems: 'center', width: 64, marginHorizontal: 4 }}>
      <PlayerAvatar player={player} displaySize="medium" />
      <Text
        numberOfLines={1}
        className="font-hk-bold"
        style={{ fontSize: 12, color: colors.text, marginTop: 4, textAlign: 'center', width: 64 }}
      >
        {label}
      </Text>
      <Text
        numberOfLines={1}
        className="font-hk-medium"
        style={{ fontSize: 10, color: colors.textSecondary, textAlign: 'center', width: 64 }}
      >
        {player.name}
      </Text>
    </View>
  );
}

function PlayerScoreItem({ player, points }: { player: Player; points: number }) {
  const pointsColor = points > 0 ? colors.formWin : points < 0 ? colors.formLoss : colors.textSecondary;
  const pointsLabel = points > 0 ? `+${points}` : String(points);

  return (
    <View style={{ alignItems: 'center', width: 64, marginHorizontal: 4 }}>
      <PlayerAvatar player={player} displaySize="medium" />
      <Text
        numberOfLines={1}
        className="font-hk-bold"
        style={{ fontSize: 12, color: pointsColor, marginTop: 4, textAlign: 'center', width: 64 }}
      >
        {pointsLabel}
      </Text>
      <Text
        numberOfLines={1}
        className="font-hk-medium"
        style={{ fontSize: 10, color: colors.textSecondary, textAlign: 'center', width: 64 }}
      >
        {player.name}
      </Text>
    </View>
  );
}

export function PlayerBetsBar({ players, playerBetStatuses, playerScores, playerBets, homeTeam, awayTeam, style }: PlayerBetsBarProps) {
  const [selectedPlayerId, setSelectedPlayerId] = useState<string | null>(null);

  if (players.length === 0) return null;

  const isScoresMode = !!playerScores;
  const isBetsMode = !!playerBets;
  const sortedPlayers = isScoresMode
    ? [...players].sort((a, b) => (playerScores![b.id]?.points ?? -Infinity) - (playerScores![a.id]?.points ?? -Infinity))
    : players;

  const selectedPlayer = selectedPlayerId ? sortedPlayers.find(p => p.id === selectedPlayerId) : null;

  return (
    <>
      <FlatList
        horizontal
        showsHorizontalScrollIndicator={false}
        data={sortedPlayers}
        keyExtractor={(item) => item.id}
        style={[{ width: '100%' }, style]}
        contentContainerStyle={{ paddingHorizontal: 16, paddingVertical: 16, flexGrow: 1, justifyContent: 'center' }}
        renderItem={({ item }) =>
          isScoresMode ? (
            <Pressable
              onLongPress={() => { Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium); setSelectedPlayerId(item.id); }}
              onPressOut={() => setSelectedPlayerId(null)}
              delayLongPress={300}
            >
              <PlayerScoreItem
                player={item}
                points={playerScores![item.id]?.points ?? -100}
              />
            </Pressable>
          ) : isBetsMode ? (
            <PlayerBetPredictionItem
              player={item}
              homeGoals={playerBets![item.id]?.predictedHomeGoals ?? null}
              awayGoals={playerBets![item.id]?.predictedAwayGoals ?? null}
            />
          ) : (
            <PlayerBetsItem
              player={item}
              hasBet={playerBetStatuses?.[item.id]?.hasBet ?? false}
            />
          )
        }
      />
      {isScoresMode && selectedPlayer && (
        <Portal name="player-bet-details">
          <PlayerBetDetailsOverlay
            player={selectedPlayer}
            bet={playerBets?.[selectedPlayer.id] ?? null}
            score={playerScores![selectedPlayer.id] ?? { points: -100 }}
            homeTeam={homeTeam ?? ''}
            awayTeam={awayTeam ?? ''}
          />
        </Portal>
      )}
    </>
  );
}
