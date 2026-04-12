import { useState, useMemo } from 'react';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { View, ScrollView, TouchableOpacity, ActivityIndicator } from 'react-native';
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
import { useNextMatch } from '../../hooks/useNextMatch';
import { useGames } from '../../src/contexts/GamesContext';
import { useAuth } from '../../src/contexts/AuthContext';
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
  const { placeBet, isSubmitting } = useBetPlacement(gameId);
  const { games, allMatchesForStandings } = useGames();
  const { incomingMatches } = useMatches(gameId || '');
  const { player } = useAuth();
  const playerId = typeof player === 'object' ? player.id : '';

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
  const [betConfirmed, setBetConfirmed] = useState(!!(betHomeGoals && betAwayGoals));

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

  const { remainingCount, nextMatch: nextMatchResult } = useNextMatch(
    gameId || '',
    id || '',
    Number(matchday),
    playerId,
  );

  const navigateToNextMatch = () => {
    if (!nextMatchResult) return;
    const m = nextMatchResult.match;
    const bet = nextMatchResult.bets?.[playerId];
    router.push({
      pathname: '/match/[id]',
      params: {
        id: m.id(),
        gameId: gameId || '',
        matchday: String(m.getMatchday()),
        date: m.getDate().toISOString(),
        homeTeam: m.homeTeamDisplayName(),
        awayTeam: m.awayTeamDisplayName(),
        homeTeamRaw: m.homeTeamName(),
        awayTeamRaw: m.awayTeamName(),
        betHomeGoals: bet ? String(bet.predictedHomeGoals) : '',
        betAwayGoals: bet ? String(bet.predictedAwayGoals) : '',
        homeTeamOdds: String(m.getHomeTeamOdds()),
        awayTeamOdds: String(m.getAwayTeamOdds()),
        drawOdds: String(m.getDrawOdds()),
        hasClearFavorite: String(m.hasClearFavorite()),
        favoriteTeam: m.getFavoriteTeam() || '',
      },
    });
  };

  const handlePlaceBet = async (matchId: string, h: number, a: number) => {
    try {
      await placeBet(matchId, h, a);
      setBetConfirmed(true);
    } catch {
      // error state handled by useBetPlacement
    }
  };

  useBetAutoSubmit(
    editable ? homeGoals : '',
    editable ? awayGoals : '',
    (h, a) => handlePlaceBet(id, h, a),
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

      {/* Scrollable content zone */}
      <ScrollView
        style={{ flex: 1 }}
        contentContainerStyle={{ flexGrow: 1 }}
        showsVerticalScrollIndicator={false}
        keyboardShouldPersistTaps="handled"
        keyboardDismissMode="on-drag"
      >
        {/* Zone 1 — match card */}
        <View style={{ backgroundColor: colors.background, paddingTop: 8, paddingBottom: 8 }}>
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

        {/* Grid gap */}
        <View style={{ height: cellSize }} />

        {/* Zone 2 — secondary content */}
        <View style={{ flex: 1, backgroundColor: colors.background, paddingTop: 24, paddingBottom: 24 }}>
          {/* Clear favorite info */}
          {clearFavorite && (
            <View style={{ paddingHorizontal: 24 }}>
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

          {/* Player bets bar */}
          {editable && gamePlayers.length > 0 && (
            <PlayerBetsBar
              players={gamePlayers}
              playerBetStatuses={matchBetStatuses}
              style={{ marginTop: clearFavorite ? 24 : 0 }}
            />
          )}

          {/* Push button to the bottom of the zone */}
          <View style={{ flex: 1 }} />

          {/* Next match button */}
          {editable && remainingCount > 0 && (
            <View style={{ marginTop: 24, paddingHorizontal: 24 }}>
              <TouchableOpacity
                disabled={isSubmitting}
                onPress={navigateToNextMatch}
                style={{
                  backgroundColor: isSubmitting
                    ? colors.disabled
                    : betConfirmed ? colors.primary : colors.text,
                  borderRadius: 999,
                  paddingVertical: 16,
                  paddingHorizontal: 40,
                  alignItems: 'center',
                  alignSelf: 'center',
                  minWidth: '70%',
                }}
              >
                {isSubmitting ? (
                  <ActivityIndicator color={colors.textSecondary} />
                ) : (
                  <>
                    <Text className="font-hk-bold" style={{
                      color: betConfirmed ? colors.text : colors.white,
                      fontSize: 17,
                    }}>
                      {t('games.nextMatch')}
                    </Text>
                    {!betConfirmed && (
                      <Text style={{ color: colors.textSecondary, fontSize: 11, marginTop: 3 }}>
                        {t('games.betNotRegistered')}
                      </Text>
                    )}
                  </>
                )}
              </TouchableOpacity>
              {!isSubmitting && (
                <Text style={{ color: colors.textSecondary, fontSize: 12, textAlign: 'center', marginTop: 10 }}>
                  {t('games.remainingMatchesInMatchday', { count: remainingCount, matchday })}
                </Text>
              )}
            </View>
          )}
        </View>
      </ScrollView>
    </View>
  );
}
