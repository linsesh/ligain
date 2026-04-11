import { useMemo } from 'react';
import { useLocalSearchParams, useRouter } from 'expo-router';
import { View, TouchableOpacity, ScrollView, Image, StyleSheet } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Text } from '../../src/components/ui/Text';
import { colors } from '../../src/constants/colors';
import { useGames } from '../../src/contexts/GamesContext';
import { getTeamLogo, isPngLogo } from '../../src/utils/teamLogos';
import { useTranslation } from 'react-i18next';
import { useGridCellSize } from '../../src/hooks/useGridCellSize';
import { SeasonMatch } from '../../src/types/match';
import { computeTeamForm } from '../../src/utils/standings';
import { FormSquares } from '../../src/components/ui/FormSquares';

// Result of a finished match from the perspective of a given team
type MatchOutcome = 'win' | 'draw' | 'loss';

function getOutcome(match: SeasonMatch, teamName: string): MatchOutcome {
  if (match.isDraw()) return 'draw';
  const winner = match.getHomeGoals() > match.getAwayGoals()
    ? match.getHomeTeamForLogo()
    : match.getAwayTeamForLogo();
  return winner === teamName ? 'win' : 'loss';
}

const OUTCOME_COLORS: Record<MatchOutcome, string> = {
  win: colors.resultWin,
  draw: colors.resultDraw,
  loss: colors.resultLoss,
};

function TeamLogoDisplay({ teamName, size = 56 }: { teamName: string; size?: number }) {
  const Logo = getTeamLogo(teamName);
  if (!Logo) return <View style={{ width: size, height: size }} />;
  return isPngLogo(Logo) ? (
    <Image source={Logo} style={{ width: size, height: size }} resizeMode="contain" />
  ) : (
    <Logo width={size} height={size} />
  );
}

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

  // Filter finished matches for this team, sorted by matchday descending
  const teamMatches = Object.values(allMatchesForStandings)
    .filter(({ match }) =>
      match.isFinished() &&
      (match.getHomeTeamForLogo() === teamName || match.getAwayTeamForLogo() === teamName)
    )
    .sort((a, b) => b.match.getMatchday() - a.match.getMatchday());

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
        <View style={styles.header}>
          <TeamLogoDisplay teamName={teamName || ''} size={72} />
          <Text className="font-hk-bold" style={styles.teamName}>
            {teamName || ''}
          </Text>
          {standing && (
            <View style={styles.badgeRow}>
              <View style={styles.badge}>
                <Text className="font-hk-bold" style={styles.badgeText}>
                  {t('team.position', { position: standing.position })}
                </Text>
              </View>
              <View style={[styles.badge, { backgroundColor: colors.primary }]}>
                <Text className="font-hk-bold" style={[styles.badgeText, { color: colors.white }]}>
                  {t('team.points', { points: standing.points })}
                </Text>
              </View>
            </View>
          )}
          <FormSquares form={teamForm} />
        </View>

        {/* Match history */}
        <View style={styles.section}>
          <Text className="font-hk-bold" style={styles.sectionTitle}>
            {t('team.matchHistory')}
          </Text>

          {teamMatches.length === 0 ? (
            <Text style={styles.emptyText}>{t('team.noMatchHistory')}</Text>
          ) : (
            teamMatches.map(({ match }) => {
              const isHome = match.getHomeTeamForLogo() === teamName;
              const opponentRaw = isHome
                ? match.getAwayTeamForLogo()
                : match.getHomeTeamForLogo();
              const opponentDisplay = isHome
                ? match.getAwayTeam()
                : match.getHomeTeam();
              const outcome = getOutcome(match, teamName || '');
              const teamGoals = isHome ? match.getHomeGoals() : match.getAwayGoals();
              const opponentGoals = isHome ? match.getAwayGoals() : match.getHomeGoals();

              return (
                <View
                  key={match.id()}
                  style={[styles.matchCard, { backgroundColor: OUTCOME_COLORS[outcome] }]}
                >
                  <Text style={styles.matchdayLabel}>
                    {t('games.matchdayShortPrefix')}{match.getMatchday()} · {isHome ? t('team.home') : t('team.away')}
                  </Text>
                  <View style={styles.matchRow}>
                    <TeamLogoDisplay teamName={opponentRaw} size={36} />
                    <Text style={styles.opponentName}>{opponentDisplay}</Text>
                    <Text className="font-hk-bold" style={styles.score}>
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

const styles = StyleSheet.create({
  header: {
    alignItems: 'center',
    paddingVertical: 24,
    paddingHorizontal: 24,
    gap: 8,
  },
  teamName: {
    fontSize: 22,
    color: colors.text,
    textAlign: 'center',
  },
  badgeRow: {
    flexDirection: 'row',
    gap: 8,
    marginTop: 4,
  },
  badge: {
    backgroundColor: colors.border,
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 4,
  },
  badgeText: {
    fontSize: 14,
    color: colors.text,
  },
  section: {
    paddingHorizontal: 16,
    gap: 8,
  },
  sectionTitle: {
    fontSize: 16,
    color: colors.text,
    marginBottom: 4,
  },
  emptyText: {
    fontSize: 14,
    color: colors.textSecondary,
    textAlign: 'center',
    paddingVertical: 16,
  },
  matchCard: {
    borderRadius: 12,
    padding: 12,
    gap: 4,
  },
  matchdayLabel: {
    fontSize: 11,
    color: colors.textSecondary,
  },
  matchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
  },
  opponentName: {
    flex: 1,
    fontSize: 14,
    color: colors.text,
  },
  score: {
    fontSize: 18,
    color: colors.text,
  },
});
