import { useMemo } from 'react';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { View, TouchableOpacity, ScrollView } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Text } from '../../src/components/ui/Text';
import { colors } from '../../src/constants/colors';
import { useGames } from '../../src/contexts/GamesContext';
import { useTranslation } from 'react-i18next';
import { useGridCellSize } from '../../src/hooks/useGridCellSize';
import { computeTeamForm } from '../../src/utils/standings';
import { FormSquares } from '../../src/components/ui/FormSquares';
import { TeamLogo } from '../../src/components/ui/TeamLogo';

type MatchOutcome = 'win' | 'draw' | 'loss';

const OUTCOME_COLORS: Record<MatchOutcome, string> = {
  win: colors.resultWin,
  draw: colors.resultDraw,
  loss: colors.resultLoss,
};

export default function TeamDetailScreen() {
  const { teamName } = useLocalSearchParams<{ teamName: string }>();
  const router = useRouter();
  const { t } = useTranslation();
  const cellSize = useGridCellSize();
  const { leagueStandings, allMatchesForStandings } = useGames();

  const standing = leagueStandings.find(s => s.teamName === teamName);

  const teamForm = useMemo(
    () => computeTeamForm(teamName || '', allMatchesForStandings),
    [teamName, allMatchesForStandings]
  );

  const teamMatches = Object.values(allMatchesForStandings)
    .filter(({ match }) =>
      match.isFinished() &&
      (match.homeTeamName() === teamName || match.awayTeamName() === teamName)
    )
    .sort((a, b) => b.match.getDate().getTime() - a.match.getDate().getTime());

  return (
    <View style={{ flex: 1, backgroundColor: 'transparent' }}>
      {/* Back button */}
      <TouchableOpacity
        onPress={() => router.back()}
        style={{ height: cellSize, justifyContent: 'center', paddingHorizontal: cellSize }}
      >
        <Ionicons name="arrow-back" size={24} color={colors.text} />
      </TouchableOpacity>

      {/* Opaque content zone */}
      <ScrollView
        style={{ backgroundColor: colors.background }}
        contentContainerStyle={{ paddingBottom: 32 }}
      >
        {/* Header: logo + team name + standing badge */}
        <View className="items-center py-6 px-6 gap-2">
          <TeamLogo teamName={teamName || ''} size={72} />
          <Text className="font-hk-bold text-[22px] text-foreground text-center">
            {teamName || ''}
          </Text>
          {standing && (
            <View className="flex-row gap-2 mt-1">
              <View className="bg-border rounded-lg px-3 py-1">
                <Text className="font-hk-bold text-sm text-foreground">
                  {t('team.position', { position: standing.position })}
                </Text>
              </View>
              <View className="rounded-lg px-3 py-1" style={{ backgroundColor: colors.primary }}>
                <Text className="font-hk-bold text-sm" style={{ color: colors.white }}>
                  {t('team.points', { points: standing.points })}
                </Text>
              </View>
            </View>
          )}
          <FormSquares form={teamForm} />
        </View>

        {/* Match history */}
        <View className="px-4 gap-2">
          <Text className="font-hk-bold text-base text-foreground mb-1">
            {t('team.matchHistory')}
          </Text>

          {teamMatches.length === 0 ? (
            <Text className="text-sm text-foreground-secondary text-center py-4">
              {t('team.noMatchHistory')}
            </Text>
          ) : (
            teamMatches.map(({ match }) => {
              const outcome = match.outcomeFor(teamName || '');
              const isHome = match.wasHome(teamName || '');
              const opponentName = isHome ? match.awayTeamName() : match.homeTeamName();
              const opponentDisplay = isHome ? match.awayTeamDisplayName() : match.homeTeamDisplayName();
              const teamGoals = isHome ? match.getHomeGoals() : match.getAwayGoals();
              const opponentGoals = isHome ? match.getAwayGoals() : match.getHomeGoals();

              return (
                <View
                  key={match.id()}
                  className="rounded-xl p-3 gap-1"
                  style={{ backgroundColor: OUTCOME_COLORS[outcome] }}
                >
                  <Text className="text-[11px] text-foreground-secondary">
                    {t('games.matchdayShortPrefix')}{match.getMatchday()} · {isHome ? t('team.home') : t('team.away')}
                  </Text>
                  <View className="flex-row items-center gap-2.5">
                    <TeamLogo teamName={opponentName} size={36} />
                    <Text className="flex-1 text-sm text-foreground">{opponentDisplay}</Text>
                    <Text className="font-hk-bold text-lg text-foreground">
                      {isHome
                        ? `${teamGoals} – ${opponentGoals}`
                        : `${opponentGoals} – ${teamGoals}`}
                    </Text>
                  </View>
                </View>
              );
            })
          )}
        </View>
      </ScrollView>
    </View>
  );
}
