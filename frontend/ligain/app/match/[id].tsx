import { useState, useMemo, useEffect } from 'react';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { View, ScrollView, TouchableOpacity, ActivityIndicator, Keyboard, Platform, LayoutAnimation, UIManager } from 'react-native';
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
import { useMatches } from '../../src/contexts/MatchesContext';
import { useNextMatch } from '../../hooks/useNextMatch';
import { useGames } from '../../src/contexts/GamesContext';
import { usePostBetNavigation } from '../../hooks/usePostBetNavigation';
import { useAuth } from '../../src/contexts/AuthContext';
import { computeTeamForm } from '../../src/utils/standings';
import { PlayerBetsBar } from '../../src/components/PlayerBetsBar';
import { FinishedMatchScore } from '../../src/components/FinishedMatchScore';

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
    homeGoals: homeGoalsParam,
    awayGoals: awayGoalsParam,
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
    homeGoals: string;
    awayGoals: string;
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
  const { incomingMatches, pastMatches } = useMatches();
  const { player } = useAuth();
  const playerId = player && typeof player === 'object' ? player.id : '';

  const gamePlayers = games.find(g => g.gameId === gameId)?.players ?? [];
  const incomingMatchResult = id
    ? Object.values(incomingMatches).find(r => r.match.id() === id) ?? null
    : null;
  const matchBetStatuses = incomingMatchResult?.playerBetStatuses ?? null;
  const isInProgress = incomingMatchResult?.match.isInProgress() ?? false;
  const pastMatchResult = id
    ? Object.values(pastMatches).find(r => r.match.id() === id) ?? null
    : null;
  const currentPlayerScore = pastMatchResult?.scores?.[playerId] ?? null;

  const homeForm = useMemo(
    () => homeTeamRaw ? computeTeamForm(homeTeamRaw, allMatchesForStandings) : [],
    [homeTeamRaw, allMatchesForStandings]
  );
  const awayForm = useMemo(
    () => awayTeamRaw ? computeTeamForm(awayTeamRaw, allMatchesForStandings) : [],
    [awayTeamRaw, allMatchesForStandings]
  );

  const isFinished = !!(homeGoalsParam && awayGoalsParam);
  const actualOutcome: 'home' | 'draw' | 'away' | undefined = isFinished && homeGoalsParam && awayGoalsParam
    ? (parseInt(homeGoalsParam) > parseInt(awayGoalsParam) ? 'home'
      : parseInt(homeGoalsParam) < parseInt(awayGoalsParam) ? 'away'
      : 'draw')
    : undefined;
  const [homeGoals, setHomeGoals] = useState(betHomeGoals || (isFinished ? homeGoalsParam : '') || '');
  const [awayGoals, setAwayGoals] = useState(betAwayGoals || (isFinished ? awayGoalsParam : '') || '');
  const [betConfirmed, setBetConfirmed] = useState(!!(betHomeGoals && betAwayGoals));
  const [barVisible, setBarVisible] = useState(!!(betHomeGoals && betAwayGoals));
  const [keyboardHeight, setKeyboardHeight] = useState(0);

  useEffect(() => {
    if (Platform.OS === 'android' && UIManager.setLayoutAnimationEnabledExperimental) {
      UIManager.setLayoutAnimationEnabledExperimental(true);
    }
    const showEvent = Platform.OS === 'ios' ? 'keyboardWillShow' : 'keyboardDidShow';
    const hideEvent = Platform.OS === 'ios' ? 'keyboardWillHide' : 'keyboardDidHide';
    const showSub = Keyboard.addListener(showEvent, (e) => {
      LayoutAnimation.configureNext(LayoutAnimation.Presets.easeInEaseOut);
      setKeyboardHeight(e.endCoordinates.height);
    });
    const hideSub = Keyboard.addListener(hideEvent, () => {
      LayoutAnimation.configureNext(LayoutAnimation.Presets.easeInEaseOut);
      setKeyboardHeight(0);
    });
    return () => {
      showSub.remove();
      hideSub.remove();
    };
  }, []);

  const matchDate = date ? new Date(date) : null;
  const dateLabel = matchDate ? formatMatchHeaderDate(matchDate) : '';
  const timeLabel = matchDate ? formatTime(matchDate) : '';
  const matchdayLabel = matchday ? `${t('games.matchdayShortPrefix')}${matchday}` : '';

  const editable = matchDate ? matchDate > new Date() : false;

  const hasBet = !!(betHomeGoals && betAwayGoals);
  const badgeLabel = isInProgress
    ? t('games.inProgressTag')
    : (!isFinished && !hasBet ? t('games.noBet') : null);
  const badgeColor = isInProgress ? colors.link : colors.primary;

  const homeOdds = homeTeamOdds && !isNaN(parseFloat(homeTeamOdds)) ? parseFloat(homeTeamOdds) : undefined;
  const awayOdds = awayTeamOdds && !isNaN(parseFloat(awayTeamOdds)) ? parseFloat(awayTeamOdds) : undefined;
  const dOdds = drawOdds && !isNaN(parseFloat(drawOdds)) ? parseFloat(drawOdds) : undefined;
  const clearFavorite = hasClearFavorite === 'true';
  const underdogTeam = clearFavorite
    ? (favoriteTeam === homeTeam ? awayTeam || '' : homeTeam || '')
    : '';

  const { remainingCount, nextMatch: nextMatchResult } = useNextMatch(
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

  const { handlePress: handlePostBetPress, isLastMatch } = usePostBetNavigation(
    remainingCount,
    navigateToNextMatch,
  );

  const handlePlaceBet = async (matchId: string, h: number, a: number) => {
    setBarVisible(true);
    try {
      await placeBet(matchId, h, a);
      setBetConfirmed(true);
    } catch {
      setBarVisible(false);
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
      <View style={{ flexDirection: 'row', marginLeft: cellSize, marginRight: cellSize, marginTop: cellSize, justifyContent: 'space-between', alignItems: 'center' }}>
        <View style={{ flexDirection: 'row' }}>
          <GridTag label={dateLabel} />
          <GridTag label={timeLabel} backgroundColor={colors.textSecondary} />
          <GridTag label={matchdayLabel} />
        </View>
        {badgeLabel && (
          <GridTag label={badgeLabel} backgroundColor={badgeColor} rounded />
        )}
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
        <View style={{ backgroundColor: colors.background }}>
          <MatchBetCard
            homeTeam={homeTeam || ''}
            awayTeam={awayTeam || ''}
            homeGoals={homeGoals}
            awayGoals={awayGoals}
            onHomeGoalsChange={(v) => { setHomeGoals(v); setBetConfirmed(false); setBarVisible(false); }}
            onAwayGoalsChange={(v) => { setAwayGoals(v); setBetConfirmed(false); setBarVisible(false); }}
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
            showGoodResult={isFinished && (currentPlayerScore?.baseScore ?? 0) >= 300}
            showBadResult={isFinished && (currentPlayerScore?.points ?? -1) <= 0}
            actualOutcome={actualOutcome}
            actualHomeGoals={homeGoalsParam}
            actualAwayGoals={awayGoalsParam}
          />
        </View>

        {/* Grid gap */}
        <View style={{ height: cellSize }} />

        {/* Zone 2 — secondary content */}
        <View style={{ flex: 1, backgroundColor: colors.background, paddingTop: isFinished ? 0 : 24, paddingBottom: 24 }}>
          {/* Clear favorite info — only for upcoming matches */}
          {clearFavorite && !isFinished && (
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

          {/* Player bets bar — incoming matches */}
          {editable && gamePlayers.length > 0 && (
            <PlayerBetsBar
              players={gamePlayers}
              playerBetStatuses={matchBetStatuses}
              style={{ marginTop: clearFavorite ? 24 : 0 }}
            />
          )}

          {/* In-progress match — show each player's predicted score */}
          {isInProgress && gamePlayers.length > 0 && (
            <PlayerBetsBar
              players={gamePlayers}
              playerBets={incomingMatchResult?.bets ?? null}
              style={{ marginTop: clearFavorite ? 24 : 0 }}
            />
          )}

          {/* Finished match — score breakdown + player scores + leaderboard button */}
          {isFinished && pastMatchResult && (
            <View style={{ marginTop: clearFavorite ? 24 : 0 }}>
              <FinishedMatchScore score={currentPlayerScore} />
              {gamePlayers.length > 0 && (
                <PlayerBetsBar
                  players={gamePlayers}
                  playerScores={pastMatchResult.scores}
                />
              )}
              <View style={{ paddingHorizontal: 24, marginTop: 8 }}>
                <TouchableOpacity
                  onPress={() => router.push({ pathname: '/game/[id]', params: { id: gameId || '' } })}
                  style={{
                    backgroundColor: colors.text,
                    borderRadius: 999,
                    paddingVertical: 16,
                    paddingHorizontal: 40,
                    alignItems: 'center',
                    alignSelf: 'center',
                    minWidth: '70%',
                  }}
                >
                  <Text className="font-hk-bold" style={{ color: colors.white, fontSize: 17 }}>
                    {t('games.viewLeaderboard')}
                  </Text>
                </TouchableOpacity>
              </View>
            </View>
          )}

          {/* Push button to the bottom of the zone */}
          <View style={{ flex: 1 }} />

          {/* Next match / view matches button */}
          {editable && (remainingCount > 0 || betConfirmed) && (
            <View style={{ marginTop: 24, paddingHorizontal: 24 }}>
              <TouchableOpacity
                disabled={isSubmitting}
                onPress={handlePostBetPress}
                style={{
                  backgroundColor: isSubmitting
                    ? colors.disabled
                    : (betConfirmed && !isLastMatch) ? colors.primary : colors.text,
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
                      color: colors.white,
                      fontSize: 17,
                    }}>
                      {isLastMatch ? t('games.viewMatches') : t('games.nextMatch')}
                    </Text>
                    {!betConfirmed && (
                      <Text style={{ color: colors.textSecondary, fontSize: 11, marginTop: 3 }}>
                        {t('games.betNotRegistered')}
                      </Text>
                    )}
                  </>
                )}
              </TouchableOpacity>
              {!isSubmitting && !isLastMatch && (
                <Text style={{ color: colors.textSecondary, fontSize: 12, textAlign: 'center', marginTop: 10 }}>
                  {t('games.remainingMatchesInMatchday', { count: remainingCount, matchday })}
                </Text>
              )}
            </View>
          )}
        </View>
      </ScrollView>

      {/* Floating status bar — sits just above the keyboard */}
      {editable && keyboardHeight > 0 && barVisible && (
        <View style={{
          position: 'absolute',
          bottom: keyboardHeight,
          left: 0,
          right: 0,
          backgroundColor: colors.background,
          paddingHorizontal: 24,
          paddingVertical: 10,
          borderTopWidth: 1,
          borderTopColor: colors.border,
        }}>
          {!betConfirmed ? (
            <ActivityIndicator color={colors.textSecondary} style={{ alignSelf: 'center' }} />
          ) : (
            <>
              <TouchableOpacity
                onPress={() => { Keyboard.dismiss(); handlePostBetPress(); }}
                style={{
                  backgroundColor: isLastMatch ? colors.text : colors.primary,
                  borderRadius: 999,
                  paddingVertical: 16,
                  paddingHorizontal: 40,
                  alignItems: 'center',
                  alignSelf: 'center',
                  minWidth: '70%',
                }}
              >
                <Text className="font-hk-bold" style={{
                  color: colors.white,
                  fontSize: 17,
                }}>
                  {isLastMatch ? t('games.viewMatches') : t('games.nextMatch')}
                </Text>
              </TouchableOpacity>
              {!isLastMatch && (
                <Text style={{ color: colors.textSecondary, fontSize: 12, textAlign: 'center', marginTop: 10 }}>
                  {t('games.remainingMatchesInMatchday', { count: remainingCount, matchday })}
                </Text>
              )}
            </>
          )}
        </View>
      )}
    </View>
  );
}
