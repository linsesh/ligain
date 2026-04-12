import { useState, useMemo } from 'react';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { View, ScrollView, TouchableOpacity } from 'react-native';
import { Text } from '../../src/components/ui/Text';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../../src/constants/colors';
import { GridTag } from '../../src/components/ui/GridTag';
import { MatchBetCard } from '../../src/components/MatchBetCard';
import { formatMatchHeaderDate, formatTime } from '../../src/utils/dateUtils';
import { useTranslation } from 'react-i18next';
import { useGridCellSize } from '../../src/hooks/useGridCellSize';
import { useBetPlacement } from '../../hooks/useBetPlacement';
import { useBetAutoSubmit } from '../../hooks/useBetAutoSubmit';
import { useMatches } from '../../hooks/useMatches';
import { useGames } from '../../src/contexts/GamesContext';
import { computeTeamForm } from '../../src/utils/standings';
import { PlayerBetsBar } from '../../src/components/PlayerBetsBar';

export default function MatchDetailScreen() {
  const {
    id,
    gameId,
    matchday,
    date,
    homeTeam,
    awayTeam,
    homeTeamRaw,
    awayTeamRaw,
    betHomeGoals,
    betAwayGoals,
    homeTeamOdds,
    awayTeamOdds,
    drawOdds,
    hasClearFavorite,
    favoriteTeam,
  } = useLocalSearchParams<{
    id: string;
    gameId: string;
    matchday: string;
    date: string;
    homeTeam: string;
    awayTeam: string;
    homeTeamRaw: string;
    awayTeamRaw: string;
    betHomeGoals: string;
    betAwayGoals: string;
    homeTeamOdds: string;
    awayTeamOdds: string;
    drawOdds: string;
    hasClearFavorite: string;
    favoriteTeam: string;
  }>();

  const router = useRouter();
  const { t } = useTranslation();
  const cellSize = useGridCellSize();
  const { placeBet } = useBetPlacement(gameId);
  const { games, allMatchesForStandings } = useGames();
  const { incomingMatches } = useMatches(gameId || '');

  const gamePlayers = games.find(g => g.gameId === gameId)?.players ?? [];
  const matchBetStatuses = id
    ? Object.values(incomingMatches).find(r => r.match.id() === id)?.playerBetStatuses ?? null
    : null;

  const homeForm = useMemo(
    () => homeTeamRaw ? computeTeamForm(homeTeamRaw, allMatchesForStandings) : [],
    [homeTeamRaw, allMatchesForStandings]
  );
  const awayForm = useMemo(
    () => awayTeamRaw ? computeTeamForm(awayTeamRaw, allMatchesForStandings) : [],
    [awayTeamRaw, allMatchesForStandings]
  );

  const [homeGoals, setHomeGoals] = useState(betHomeGoals || '');
  const [awayGoals, setAwayGoals] = useState(betAwayGoals || '');

  const matchDate = date ? new Date(date) : null;
  const dateLabel = matchDate ? formatMatchHeaderDate(matchDate) : '';
  const timeLabel = matchDate ? formatTime(matchDate) : '';
  const matchdayLabel = matchday ? `${t('games.matchdayShortPrefix')}${matchday}` : '';

  const editable = matchDate ? matchDate > new Date() : false;

  const homeOdds = homeTeamOdds ? parseFloat(homeTeamOdds) : undefined;
  const awayOdds = awayTeamOdds ? parseFloat(awayTeamOdds) : undefined;
  const dOdds = drawOdds ? parseFloat(drawOdds) : undefined;
  const clearFavorite = hasClearFavorite === 'true';
  const underdogTeam = clearFavorite
    ? (favoriteTeam === homeTeam ? awayTeam || '' : homeTeam || '')
    : '';

  useBetAutoSubmit(
    editable ? homeGoals : '',
    editable ? awayGoals : '',
    (h, a) => placeBet(id, h, a),
  );

  return (
    <View style={{ flex: 1, backgroundColor: 'transparent' }}>
      {/* Transparent grid zone — back button + header tags */}
      <TouchableOpacity
        onPress={() => router.back()}
        style={{ height: cellSize, justifyContent: 'center', paddingHorizontal: cellSize }}
      >
        <Ionicons name="arrow-back" size={24} color={colors.text} />
      </TouchableOpacity>
      <View style={{ flexDirection: 'row', marginLeft: cellSize, marginTop: cellSize }}>
        <GridTag label={dateLabel} />
        <GridTag label={timeLabel} backgroundColor={colors.textSecondary} />
        <GridTag label={matchdayLabel} />
      </View>

      <ScrollView showsVerticalScrollIndicator={false} keyboardShouldPersistTaps="handled" keyboardDismissMode="on-drag">
        {/* Opaque grey content zone — natural height so grid shows below */}
        <View style={{ backgroundColor: colors.background }}>
          <MatchBetCard
            homeTeam={homeTeam || ''}
            awayTeam={awayTeam || ''}
            homeGoals={homeGoals}
            awayGoals={awayGoals}
            onHomeGoalsChange={setHomeGoals}
            onAwayGoalsChange={setAwayGoals}
            editable={editable}
            homeTeamOdds={homeOdds}
            awayTeamOdds={awayOdds}
            drawOdds={dOdds}
            hasClearFavorite={clearFavorite}
            favoriteTeam={favoriteTeam || ''}
            homeTeamForm={homeForm}
            awayTeamForm={awayForm}
            onHomeTeamPress={homeTeamRaw ? () => router.push({ pathname: '/team/[teamName]', params: { teamName: homeTeamRaw, gameId: gameId || '' } }) : undefined}
            onAwayTeamPress={awayTeamRaw ? () => router.push({ pathname: '/team/[teamName]', params: { teamName: awayTeamRaw, gameId: gameId || '' } }) : undefined}
          />
        </View>

        {/* Favourite info + player bets bar — share one background zone */}
        {(clearFavorite || (editable && gamePlayers.length > 0)) && (
          <View style={{ backgroundColor: colors.background, marginTop: cellSize }}>
            {clearFavorite && (
              <View style={{ padding: 24 }}>
                <View style={{ backgroundColor: colors.link, borderRadius: 12, padding: 16 }}>
                  <Text className="font-hk-bold" style={{ color: colors.white, fontSize: 22, textAlign: 'center' }}>
                    {t('games.clearFavoriteTeam', { team: favoriteTeam })}
                  </Text>
                </View>
                <View style={{ backgroundColor: colors.border, borderRadius: 12, padding: 16, marginTop: 12 }}>
                  <Text style={{ color: colors.text, fontSize: 12, textAlign: 'center' }}>
                    {t('games.doublePointsHint', { team: underdogTeam })}
                  </Text>
                </View>
              </View>
            )}
            {editable && gamePlayers.length > 0 && (
              <PlayerBetsBar players={gamePlayers} playerBetStatuses={matchBetStatuses} />
            )}
          </View>
        )}
      </ScrollView>
    </View>
  );
}
