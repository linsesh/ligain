import React, { useState, useEffect, useRef, useTransition } from 'react';
import { View, StyleSheet, TouchableOpacity, Alert, ScrollView, RefreshControl, Animated, Dimensions, FlatList } from 'react-native';
import { Text } from './ui/Text';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useMatches } from '../contexts/MatchesContext';
import { useMatchNotifications } from '../hooks/useMatchNotifications';
import { useAuth } from '../contexts/AuthContext';
import { useGames } from '../contexts/GamesContext';
import { useTranslation } from 'react-i18next';
import { formatTime, formatDate } from '../utils/dateUtils';
import { colors } from '../constants/colors';
import StatusTag from './StatusTag';
import ShareableMatchResult from './ShareableMatchResult';
import { captureAndShareWithOptions } from '../utils/shareUtils';
import ViewShot from 'react-native-view-shot';
import { TeamLogo } from './ui/TeamLogo';

function MatchCardSkeleton({ opacity }: { opacity: Animated.Value }) {
  return (
    <Animated.View style={[styles.matchCard, { opacity }]}>
      <View style={styles.bettingContainer}>
        {/* Home team */}
        <View style={styles.teamDisplay}>
          <View style={{ width: 56, height: 56, borderRadius: 28, backgroundColor: '#ddd' }} />
          <View style={{ width: 60, height: 10, borderRadius: 4, backgroundColor: '#ddd' }} />
        </View>
        {/* VS/Score placeholder */}
        <View style={styles.scoreCenterContainer}>
          <View style={{ width: 40, height: 28, borderRadius: 4, backgroundColor: '#ddd' }} />
        </View>
        {/* Away team */}
        <View style={styles.teamDisplay}>
          <View style={{ width: 56, height: 56, borderRadius: 28, backgroundColor: '#ddd' }} />
          <View style={{ width: 60, height: 10, borderRadius: 4, backgroundColor: '#ddd' }} />
        </View>
      </View>
      {/* Tag placeholder */}
      <View style={styles.tagCenterContainer}>
        <View style={{ width: 80, height: 24, borderRadius: 12, backgroundColor: '#ddd' }} />
      </View>
    </Animated.View>
  );
}

export function MatchesListSkeleton() {
  const opacity = useRef(new Animated.Value(0.4)).current;
  const itemWidth = Dimensions.get('window').width / 6;

  useEffect(() => {
    Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, { toValue: 1, duration: 800, useNativeDriver: true }),
        Animated.timing(opacity, { toValue: 0.4, duration: 800, useNativeDriver: true }),
      ])
    ).start();
  }, [opacity]);

  return (
    <View testID="loading-indicator" style={styles.container}>
      {/* Matchday selector skeleton */}
      <Animated.View style={[{ flexDirection: 'row', borderBottomWidth: 1, borderBottomColor: '#e0e0e0', paddingHorizontal: 8, paddingVertical: 8 }, { opacity }]}>
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <View key={i} style={{ width: itemWidth, alignItems: 'center', paddingVertical: 6 }}>
            <View style={{ width: 28, height: 18, borderRadius: 4, backgroundColor: '#ddd' }} />
          </View>
        ))}
      </Animated.View>
      {/* Cards area */}
      <View style={{ backgroundColor: colors.background, flex: 1 }}>
        <View style={[styles.matchesContainer]}>
          <View style={[styles.timeGroup]}>
            <Animated.View style={[styles.timeHeaderContainer, { opacity }]}>
              <View style={{ width: 50, height: 18, borderRadius: 4, backgroundColor: '#ddd', marginBottom: 4 }} />
              <View style={{ width: 120, height: 12, borderRadius: 4, backgroundColor: '#ddd' }} />
            </Animated.View>
            <MatchCardSkeleton opacity={opacity} />
            <MatchCardSkeleton opacity={opacity} />
            <MatchCardSkeleton opacity={opacity} />
          </View>
        </View>
      </View>
    </View>
  );
}

function MatchdayContentSkeleton() {
  const opacity = useRef(new Animated.Value(0.4)).current;

  useEffect(() => {
    Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, { toValue: 1, duration: 800, useNativeDriver: true }),
        Animated.timing(opacity, { toValue: 0.4, duration: 800, useNativeDriver: true }),
      ])
    ).start();
  }, [opacity]);

  return (
    <View style={styles.matchesContainer}>
      <View style={styles.timeGroup}>
        <Animated.View style={[styles.timeHeaderContainer, { opacity }]}>
          <View style={{ width: 50, height: 18, borderRadius: 4, backgroundColor: '#ddd', marginBottom: 4 }} />
          <View style={{ width: 120, height: 12, borderRadius: 4, backgroundColor: '#ddd' }} />
        </Animated.View>
        <MatchCardSkeleton opacity={opacity} />
        <MatchCardSkeleton opacity={opacity} />
        <MatchCardSkeleton opacity={opacity} />
      </View>
    </View>
  );
}

function MatchCard({ matchResult, gameId, isDelayed }: {
  matchResult: any;
  gameId: string;
  isDelayed?: boolean;
}) {
  const matchId = matchResult.match.id();

  const { player } = useAuth();
  const { games } = useGames();
  const { t } = useTranslation();
  const router = useRouter();
  const [isSharing, setIsSharing] = useState(false);
  const shareableRef = useRef<ViewShot>(null);
  const now = new Date();
  const isFuture = !matchResult.match.isFinished() && !matchResult.match.hasStarted(now);
  const userBet = player && matchResult.bets ? matchResult.bets[player.id] : undefined;

  const game = games.find(g => g.gameId === gameId);
  const playersMap = new Map(
    (game?.players ?? []).map((p: any) => [p.id, p])
  );

  const handleShareMatch = async () => {
    if (!matchResult.match.isFinished() || isSharing) return;
    setIsSharing(true);
    try {
      await captureAndShareWithOptions(shareableRef, {
        title: t('share.shareTitle'),
        message: t('share.shareTitle'),
      });
    } catch (error) {
      console.error('Error sharing match:', error);
      Alert.alert(t('share.shareFailed'), t('share.shareFailed'));
    } finally {
      setIsSharing(false);
    }
  };

  // Tag logic
  let tagText: string | null = null;
  let tagVariant: 'warning' | 'success' | 'negative' | 'finished' | 'primary' | 'live' | null = null;
  if (matchResult.match.isInProgress()) {
    tagText = t('games.inProgressTag');
    tagVariant = 'live';
  } else if (matchResult.match.isFinished() && player) {
    if (matchResult.scores && matchResult.scores[player.id]) {
      const points = matchResult.scores[player.id].points;
      if (typeof points === 'number' && points > 0) {
        tagText = `+${points} points`;
        tagVariant = 'success';
      } else {
        tagText = `${points} points`;
        tagVariant = 'negative';
      }
    } else {
      tagText = t('games.noBet');
      tagVariant = 'primary';
    }
  } else if (isFuture && (!matchResult.bets || !matchResult.bets[player?.id ?? ''])) {
    tagText = t('games.noBet');
    tagVariant = 'primary';
  }

  const cardStyle = [
    styles.matchCard,
    matchResult.match.isFinished() ? styles.finishedMatchCard : null,
  ].filter(Boolean);

  return (
    <TouchableOpacity
      style={cardStyle}
      onPress={() => router.push({
        pathname: '/match/[id]',
        params: {
          id: matchResult.match.id(),
          gameId,
          matchday: String(matchResult.match.getMatchday()),
          date: matchResult.match.getDate().toISOString(),
          homeTeam: matchResult.match.homeTeamDisplayName(),
          awayTeam: matchResult.match.awayTeamDisplayName(),
          homeTeamRaw: matchResult.match.homeTeamName(),
          awayTeamRaw: matchResult.match.awayTeamName(),
          betHomeGoals: player && matchResult.bets?.[player.id]
            ? String(matchResult.bets[player.id].predictedHomeGoals)
            : '',
          betAwayGoals: player && matchResult.bets?.[player.id]
            ? String(matchResult.bets[player.id].predictedAwayGoals)
            : '',
          homeGoals: matchResult.match.isFinished() ? String(matchResult.match.getHomeGoals()) : '',
          awayGoals: matchResult.match.isFinished() ? String(matchResult.match.getAwayGoals()) : '',
          homeTeamOdds: String(matchResult.match.getHomeTeamOdds()),
          awayTeamOdds: String(matchResult.match.getAwayTeamOdds()),
          drawOdds: String(matchResult.match.getDrawOdds()),
          hasClearFavorite: String(matchResult.match.hasClearFavorite()),
          favoriteTeam: matchResult.match.getFavoriteTeam(),
        },
      })}
      activeOpacity={0.8}
    >
      {/* Share Button */}
      {matchResult.match.isFinished() && (
        <TouchableOpacity
          style={styles.shareButtonBottomRight}
          onPress={handleShareMatch}
          disabled={isSharing}
        >
          <Ionicons
            name={isSharing ? "hourglass-outline" : "share-outline"}
            size={20}
            color={isSharing ? colors.textSecondary : "#000000"}
          />
        </TouchableOpacity>
      )}

      <View style={styles.bettingContainer}>
        {/* Home team */}
        <View style={styles.teamDisplay}>
          <TeamLogo teamName={matchResult.match.homeTeamDisplayName()} />
          <Text style={styles.teamName}>{matchResult.match.homeTeamDisplayName()}</Text>
        </View>

        {/* Center: VS, predicted score, or actual score */}
        <View style={styles.scoreCenterContainer}>
          {isFuture && !userBet ? (
            <Text className="font-hk-bold" style={styles.vsText}>VS</Text>
          ) : isFuture && userBet ? (
            <Text className="font-hk-bold" style={styles.scoreDisplayText}>
              {userBet.predictedHomeGoals} - {userBet.predictedAwayGoals}
            </Text>
          ) : (
            <Text className="font-hk-bold" style={styles.scoreDisplayText}>
              {matchResult.match.getHomeGoals()} - {matchResult.match.getAwayGoals()}
            </Text>
          )}
        </View>

        {/* Away team */}
        <View style={styles.teamDisplay}>
          <TeamLogo teamName={matchResult.match.awayTeamDisplayName()} />
          <Text style={styles.teamName}>{matchResult.match.awayTeamDisplayName()}</Text>
        </View>
      </View>

      {tagText && typeof tagVariant === 'string' && (
        <View style={styles.tagCenterContainer}>
          <StatusTag text={tagText} variant={tagVariant} style={styles.statusTagWide} />
        </View>
      )}

      {isDelayed && (
        <View style={styles.tagCenterContainer}>
          <StatusTag text={t('games.delayedMatch')} variant="info" style={styles.statusTagWide} />
        </View>
      )}

      {/* Hidden shareable component for image generation */}
      {matchResult.match.isFinished() && (
        <View style={{ position: 'absolute', left: -9999, top: -9999 }}>
          <ViewShot ref={shareableRef}>
            <ShareableMatchResult
              homeTeam={matchResult.match.homeTeamDisplayName()}
              awayTeam={matchResult.match.awayTeamDisplayName()}
              homeScore={matchResult.match.getHomeGoals()}
              awayScore={matchResult.match.getAwayGoals()}
              myHomeScore={player ? matchResult.bets?.[player.id]?.predictedHomeGoals : undefined}
              myAwayScore={player ? matchResult.bets?.[player.id]?.predictedAwayGoals : undefined}
              showGoodResult={player ? (matchResult.scores?.[player.id]?.baseScore ?? 0) >= 300 : false}
              players={Object.entries(matchResult.scores || {}).map(([playerId, scoreData]: [string, any]) => ({
                name: scoreData.playerName,
                points: scoreData.points,
                avatarUrl: playersMap.get(playerId)?.avatarUrl ?? null,
              }))}
            />
          </ViewShot>
        </View>
      )}
    </TouchableOpacity>
  );
}

interface MatchesListProps {
  gameId: string;
  initialMatchday?: number;
  activeMatchday?: number;
}

export default function MatchesList({ gameId, initialMatchday, activeMatchday }: MatchesListProps) {
  const { incomingMatches, pastMatches, loading: matchesLoading, refresh } = useMatches();
  useMatchNotifications(incomingMatches, gameId);
  const [refreshing, setRefreshing] = useState(false);
  const [currentMatchday, setCurrentMatchday] = useState<number | null>(initialMatchday ?? null);
  const [isPending, startTransition] = useTransition();
  const { t } = useTranslation();
  const scrollViewRef = React.useRef<ScrollView>(null);
  const itemWidth = Dimensions.get('window').width / 6;
  const matchdaySelectorRef = useRef<FlatList>(null);

  // Combine incoming and past matches
  const matches = [...Object.values(incomingMatches), ...Object.values(pastMatches)];

  // Group matches by matchday
  const matchesByMatchday = matches.reduce((acc, matchResult) => {
    const matchday = matchResult.match.getMatchday();
    if (!acc[matchday]) {
      acc[matchday] = [];
    }
    acc[matchday].push(matchResult);
    return acc;
  }, {} as { [key: number]: any[] });

  // Sort matchdays
  const sortedMatchdays = Object.keys(matchesByMatchday)
    .map(Number)
    .sort((a, b) => a - b);

  // Set initial matchday if not set
  useEffect(() => {
    if (sortedMatchdays.length > 0 && currentMatchday === null) {
      if (initialMatchday && sortedMatchdays.includes(initialMatchday)) {
        setCurrentMatchday(initialMatchday);
      } else {
        setCurrentMatchday(sortedMatchdays[sortedMatchdays.length - 1]);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sortedMatchdays]);

  // Group matches by date and time within the current matchday
  const getMatchesByTime = (matchday: number) => {
    const matchdayMatches = matchesByMatchday[matchday] || [];
    const sortedMatches = matchdayMatches.sort((a, b) =>
      a.match.getDate().getTime() - b.match.getDate().getTime()
    );
    const groupedByDateTime = sortedMatches.reduce((acc, matchResult) => {
      const date = matchResult.match.getDate();
      const dateKey = formatDate(date);
      const timeKey = formatTime(date);
      const dateTimeKey = `${dateKey} - ${timeKey}`;
      if (!acc[dateTimeKey]) {
        acc[dateTimeKey] = [];
      }
      acc[dateTimeKey].push(matchResult);
      return acc;
    }, {} as { [key: string]: any[] });
    return groupedByDateTime;
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  useEffect(() => {
    if (currentMatchday == null || !matchdaySelectorRef.current) return;
    const index = sortedMatchdays.indexOf(currentMatchday);
    if (index >= 0) {
      matchdaySelectorRef.current.scrollToIndex({ index, animated: true, viewPosition: 0.5 });
    }
  }, [currentMatchday, sortedMatchdays]);

  if (matchesLoading && !refreshing) {
    return <MatchesListSkeleton />;
  }

  const matchdaysWithFinishedMatches = new Set<number>(
    matches
      .filter(mr => mr.match.isFinished())
      .map(mr => mr.match.getMatchday())
  );

  const isDelayedMatch = (matchResult: any): boolean => {
    if (matchResult.match.isFinished() || matchResult.match.isInProgress()) return false;
    const md = matchResult.match.getMatchday();
    for (const finishedMd of matchdaysWithFinishedMatches) {
      if (finishedMd > md) return true;
    }
    return false;
  };

  const currentMatchdayMatches = currentMatchday ? getMatchesByTime(currentMatchday) : {};
  const sortedDateTimeKeys = Object.keys(currentMatchdayMatches).sort((a, b) => {
    const aAllFinished = currentMatchdayMatches[a].every((mr: any) => mr.match.isFinished());
    const bAllFinished = currentMatchdayMatches[b].every((mr: any) => mr.match.isFinished());
    if (aAllFinished !== bAllFinished) return aAllFinished ? 1 : -1;

    const dateTimeA = a.split(' - ');
    const dateTimeB = b.split(' - ');
    if (dateTimeA.length !== 2 || dateTimeB.length !== 2) {
      return a.localeCompare(b);
    }
    const dateA = dateTimeA[0];
    const timeA = dateTimeA[1];
    const dateB = dateTimeB[0];
    const timeB = dateTimeB[1];
    if (dateA !== dateB) {
      const dateAObj = new Date(dateA);
      const dateBObj = new Date(dateB);
      return dateAObj.getTime() - dateBObj.getTime();
    }
    const timeANum = parseInt(timeA.replace(':', ''));
    const timeBNum = parseInt(timeB.replace(':', ''));
    return timeANum - timeBNum;
  });
  return (
    <View style={styles.container}>
      <ScrollView
        ref={scrollViewRef}
        style={styles.scrollView}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            tintColor={colors.primary}
            progressBackgroundColor={colors.background}
            progressViewOffset={20}
          />
        }
      >
        {/* Matchday Selector */}
        <FlatList
          ref={matchdaySelectorRef}
          horizontal
          showsHorizontalScrollIndicator={false}
          data={sortedMatchdays}
          keyExtractor={(item) => String(item)}
          className="border-b border-border"
          contentContainerClassName="px-2 py-2"
          getItemLayout={(_, index) => ({
            length: itemWidth,
            offset: itemWidth * index,
            index,
          })}
          onScrollToIndexFailed={(info) => {
            setTimeout(() => {
              matchdaySelectorRef.current?.scrollToIndex({ index: info.index, animated: true, viewPosition: 0.5 });
            }, 100);
          }}
          renderItem={({ item }) => {
            const isSelected = item === currentMatchday;
            const isActive = item === activeMatchday;
            return (
              <TouchableOpacity
                onPress={() => startTransition(() => setCurrentMatchday(item))}
                style={{ width: itemWidth }}
                className="items-center py-1.5"
              >
                <Text className={`text-lg ${isSelected ? `font-hk-bold ${isActive ? 'text-primary' : 'text-foreground'}` : isActive ? 'font-hk-medium text-primary' : 'font-hk-medium text-foreground-secondary'}`}>
                  {t('games.matchdayShortPrefix')}{item}
                </Text>
                <View className={`h-0.5 w-1/2 mt-0.5 rounded-full ${isSelected ? 'bg-primary' : 'bg-transparent'}`} />
              </TouchableOpacity>
            );
          }}
        />
        {/* Background panel: covers grid from below matchday selector */}
        <View style={{ backgroundColor: colors.background, flexGrow: 1 }}>
          {isPending ? (
            <MatchdayContentSkeleton />
          ) : (
            /* Matches for current matchday */
            <View style={styles.matchesContainer}>
              {sortedDateTimeKeys.map((dateTimeKey: string) => {
                const matchesAtTime = currentMatchdayMatches[dateTimeKey];
                const dateTimeParts = dateTimeKey.split(' - ');
                const dateDisplay = dateTimeParts[0] || '';
                const timeDisplay = dateTimeParts[1] || '';
                return (
                  <View key={dateTimeKey} style={styles.timeGroup}>
                    <View style={styles.timeHeaderContainer}>
                      <Text className="font-hk-bold" style={styles.timeHeader}>{timeDisplay}</Text>
                      <Text style={styles.dayHeader}>{dateDisplay}</Text>
                    </View>
                    {matchesAtTime.map((matchResult: any) => (
                      <MatchCard
                        key={matchResult.match.id()}
                        matchResult={matchResult}
                        gameId={gameId}
                        isDelayed={isDelayedMatch(matchResult)}
                      />
                    ))}
                  </View>
                );
              })}
            </View>
          )}
          <View style={{ height: 100 }} />
        </View>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: 'transparent',
  },
  scrollView: {
    flex: 1,
  },
  matchesContainer: {
    marginTop: 16,
  },
  timeGroup: {
    marginHorizontal: 16,
    marginBottom: 12,
  },
  timeHeaderContainer: {
    marginBottom: 8,
  },
  timeHeader: {
    fontSize: 18,
    color: colors.text,
  },
  dayHeader: {
    fontSize: 14,
    color: colors.textSecondary,
    marginTop: 2,
  },
  matchCard: {
    backgroundColor: '#f5f5f5',
    padding: 16,
    borderRadius: 8,
    marginBottom: 12,
    position: 'relative',
  },
  finishedMatchCard: {
    backgroundColor: colors.cardFinished,
  },
  bettingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 8,
  },
  teamDisplay: {
    flex: 1,
    alignItems: 'center',
    gap: 4,
  },
  scoreCenterContainer: {
    alignItems: 'center',
    justifyContent: 'center',
    minWidth: 80,
  },
  vsText: {
    fontSize: 28,
    color: '#666',
    textAlign: 'center',
  },
  scoreDisplayText: {
    fontSize: 28,
    color: colors.text,
    textAlign: 'center',
  },
  teamName: {
    fontSize: 14,
    color: '#333',
    textAlign: 'center',
  },
  tagCenterContainer: {
    alignItems: 'center',
    marginTop: 24,
  },
  statusTagWide: {
    paddingHorizontal: 20,
    paddingVertical: 8,
    alignSelf: 'center',
  },
  shareButtonBottomRight: {
    position: 'absolute',
    bottom: 8,
    right: 8,
    zIndex: 1,
    padding: 4,
  },
  legendContainer: {
    marginTop: 24,
    marginHorizontal: 16,
    padding: 16,
    backgroundColor: colors.card,
    borderRadius: 8,
  },
  legendTitle: {
    color: colors.text,
    fontSize: 16,
    marginBottom: 8,
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  legendStar: {
    fontSize: 18,
    marginRight: 8,
    color: colors.primary,
    width: 32,
    textAlign: 'center',
  },
  legendMark: {
    fontSize: 16,
    marginRight: 8,
    color: colors.primary,
    width: 32,
    textAlign: 'center',
  },
  legendText: {
    color: colors.text,
    fontSize: 14,
    textAlign: 'left',
  },
});
