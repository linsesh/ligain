import { useState } from 'react';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { View, TouchableOpacity, TouchableWithoutFeedback, Keyboard } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../../src/constants/colors';
import { GridTag } from '../../src/components/ui/GridTag';
import { MatchBetCard } from '../../src/components/MatchBetCard';
import { formatMatchHeaderDate, formatTime } from '../../src/utils/dateUtils';
import { useTranslation } from 'react-i18next';
import { useGridCellSize } from '../../src/hooks/useGridCellSize';
import { useBetPlacement } from '../../hooks/useBetPlacement';
import { useBetAutoSubmit } from '../../hooks/useBetAutoSubmit';

export default function MatchDetailScreen() {
  const {
    id,
    gameId,
    matchday,
    date,
    homeTeam,
    awayTeam,
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

  useBetAutoSubmit(
    editable ? homeGoals : '',
    editable ? awayGoals : '',
    (h, a) => placeBet(id, h, a),
  );

  return (
    <TouchableWithoutFeedback onPress={Keyboard.dismiss} accessible={false}>
    <View style={{ flex: 1, backgroundColor: 'transparent' }}>
      {/* Transparent grid zone — back button + header tags */}
      <TouchableOpacity
        onPress={() => router.back()}
        style={{ height: cellSize, justifyContent: 'center', paddingHorizontal: cellSize, marginTop: -cellSize }}
      >
        <Ionicons name="arrow-back" size={24} color={colors.text} />
      </TouchableOpacity>
      <View style={{ flexDirection: 'row', marginLeft: cellSize, marginTop: cellSize }}>
        <GridTag label={dateLabel} />
        <GridTag label={timeLabel} backgroundColor={colors.textSecondary} />
        <GridTag label={matchdayLabel} />
      </View>

      {/* Opaque grey content zone — natural height so grid shows below */}
      <View style={{ backgroundColor: colors.background, paddingTop: 24 }}>
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
        />
      </View>
    </View>
    </TouchableWithoutFeedback>
  );
}
