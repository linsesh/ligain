import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert, ScrollView, RefreshControl, KeyboardAvoidingView, Platform, Animated } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useMatches } from '../../../../hooks/useMatches';
import { useBetSubmission } from '../../../../hooks/useBetSubmission';
import { useAuth } from '../../../../src/contexts/AuthContext';
import { useTranslation } from 'react-i18next';
import { formatTime, formatDate } from '../../../../src/utils/dateUtils';
import { colors } from '../../../../src/constants/colors';

function StatusTag({ text, variant }: { text: string; variant: string }) {
  const baseStyle = [styles.statusTag];
  let variantStyle = null;
  if (variant === 'success') variantStyle = styles.successTag;
  else if (variant === 'warning') variantStyle = styles.inProgressTag;
  else if (variant === 'finished') variantStyle = styles.finishedTag;
  else if (variant === 'primary') variantStyle = styles.primaryTag;
  else if (variant === 'negative') variantStyle = styles.negativeTag;
  return (
    <View style={[...baseStyle, variantStyle]}>
      <Text style={styles.statusTagText}>{text}</Text>
    </View>
  );
}

interface TempScores {
  [key: string]: {
    home?: number;
    away?: number;
  };
}

function TeamInput({ teamName, value, onChange, canModify, isAway = false, onFocus, onBlur, isFavorite }: {
  teamName: string;
  value: string;
  onChange: (value: string) => void;
  canModify: boolean;
  isAway?: boolean;
  onFocus?: () => void;
  onBlur?: () => void;
  isFavorite?: boolean;
}) {
  return (
    <View style={[styles.teamContainer, isAway && styles.awayTeamContainer]}>
      <Text style={[styles.teamName, isAway && styles.awayTeamName]}>
        {teamName}
        {isFavorite && <Text style={{ color: colors.primary }}>⭐</Text>}
      </Text>
      <TextInput
        style={[styles.betInput, !canModify && styles.disabledInput]}
        value={value}
        onChangeText={onChange}
        editable={canModify}
        keyboardType="numeric"
        onFocus={onFocus}
        onBlur={onBlur}
        maxLength={2}
      />
    </View>
  );
}

function MatchCard({ matchResult, tempScores, expandedMatches, onBetChange, onToggleBetSection, onFocus, onBlur, onDone }: {
  matchResult: any;
  tempScores: TempScores;
  expandedMatches: { [key: string]: boolean };
  onBetChange: (matchId: string, team: 'home' | 'away', value: string) => void;
  onToggleBetSection: (matchId: string) => void;
  onFocus: () => void;
  onBlur: () => void;
  onDone?: () => void;
}) {
  const { player } = useAuth();
  const { t } = useTranslation();
  const now = new Date();
  const isFuture = !matchResult.match.isFinished() && !matchResult.match.isInProgress();
  const userBet = player && matchResult.bets ? matchResult.bets[player.id] : undefined;
  const canModify = isFuture && (userBet?.isModifiable(now) !== false);

  // Tag logic
  let tagText = null;
  let tagVariant = null;
  let hasTag = false;
  if (matchResult.match.isInProgress()) {
    tagText = t('games.inProgressTag');
    tagVariant = 'warning';
    hasTag = true;
  } else if (matchResult.match.isFinished() && player) {
    if (matchResult.scores && matchResult.scores[player.id]) {
      const points = matchResult.scores[player.id].points;
      if (typeof points === 'number' && points > 0 && points < 1000) {
        tagText = `+${points} points`;
        tagVariant = 'success';
        hasTag = true;
      } else if (typeof points === 'number' && points < 0) {
        tagText = `${points} points (${t('games.negativePointsTag')})`;
        tagVariant = 'negative';
        hasTag = true;
      } else {
        tagText = `0 points`;
        tagVariant = 'finished';
        hasTag = true;
      }
    } else {
      // No score entry means no bet was placed, show 0 points
      tagText = `0 points`;
      tagVariant = 'finished';
      hasTag = true;
    }
  }

  // Determine card style: always greyed out if finished
  const cardStyle = [
    styles.matchCard,
    matchResult.match.isFinished() ? styles.finishedMatchCard : null,
    hasTag ? styles.matchCardWithTag : null
  ].filter(Boolean);

  return (
    <View 
      key={matchResult.match.id()} 
      style={cardStyle}
    >
      {/* Status Tag */}
      {tagText && typeof tagVariant === 'string' && (
        <StatusTag text={tagText} variant={tagVariant} />
      )}
      <View style={styles.bettingContainer}>
        {/* Only show the current user's bet for future matches */}
        {isFuture ? (
          <>
            <TeamInput
              teamName={matchResult.match.getHomeTeam()}
              value={tempScores[matchResult.match.id()]?.home?.toString() || ''}
              onChange={(value) => onBetChange(matchResult.match.id(), 'home', value)}
              canModify={canModify}
              onFocus={() => onFocus()}
              onBlur={() => onBlur()}
              isFavorite={matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() === matchResult.match.getHomeTeam()}
            />
            <Text style={styles.vsText}>{t('common.vs')}</Text>
            <TeamInput
              teamName={matchResult.match.getAwayTeam()}
              value={tempScores[matchResult.match.id()]?.away?.toString() || ''}
              onChange={(value) => onBetChange(matchResult.match.id(), 'away', value)}
              canModify={canModify}
              isAway
              onFocus={() => onFocus()}
              onBlur={() => onBlur()}
              isFavorite={matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() === matchResult.match.getAwayTeam()}
            />
          </>
        ) : (
          <>
            <TeamInput
              teamName={matchResult.match.getHomeTeam()}
              value={matchResult.match.getHomeGoals().toString()}
              onChange={() => {}}
              canModify={false}
            />
            <Text style={styles.vsText}>{t('common.vs')}</Text>
            <TeamInput
              teamName={matchResult.match.getAwayTeam()}
              value={matchResult.match.getAwayGoals().toString()}
              onChange={() => {}}
              canModify={false}
              isAway
            />
          </>
        )}
      </View>
      {/* Odds display */}
      <View style={styles.oddsContainer}>
        <View style={styles.oddsRow}>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>1</Text>
            <Text style={styles.oddsValue}>{matchResult.match.getHomeTeamOdds().toFixed(2)}</Text>
            {matchResult.match.hasClearFavorite() && (
              matchResult.match.getFavoriteTeam() === matchResult.match.getHomeTeam()
                ? <Text style={styles.outsiderMark}>x1</Text>
                : <Text style={styles.outsiderMark}>×2</Text>
            )}
          </View>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>N</Text>
            <Text style={styles.oddsValue}>{matchResult.match.getDrawOdds().toFixed(2)}</Text>
            {matchResult.match.hasClearFavorite() && (
              <Text style={styles.outsiderMark}>×1.5</Text>
            )}
          </View>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>2</Text>
            <Text style={styles.oddsValue}>{matchResult.match.getAwayTeamOdds().toFixed(2)}</Text>
            {matchResult.match.hasClearFavorite() && (
              matchResult.match.getFavoriteTeam() === matchResult.match.getAwayTeam()
                ? <Text style={styles.outsiderMark}>x1</Text>
                : <Text style={styles.outsiderMark}>×2</Text>
            )}
          </View>
        </View>
      </View>
      {/* Keep the rest of the card (other players' bets/scores) only for past/in-progress matches */}
      {(!isFuture && (matchResult.match.isFinished() || matchResult.match.isInProgress())) && (
        <>
          <TouchableOpacity 
            style={styles.toggleButton} 
            onPress={() => onToggleBetSection(matchResult.match.id())}
          >
            <Text style={styles.toggleButtonText}>
              {t('games.playersBets')}
            </Text>
            <Ionicons 
              name={expandedMatches[matchResult.match.id()] ? "chevron-up" : "chevron-down"} 
              size={24} 
              color="#333" 
            />
          </TouchableOpacity>
          {expandedMatches[matchResult.match.id()] && (
            <View style={styles.betResultContainer}>
              {matchResult.scores ? (
                <View style={styles.scoresContainer}>
                  {/* Custom order: current user first, then others */}
                  {(() => {
                    const entries = Object.entries(matchResult.scores);
                    let userRow = null;
                    let otherRows: React.ReactNode[] = [];
                    entries.forEach(([playerId, scoreData]) => {
                      const s = scoreData as { playerName: string; points: number };
                      if (player && playerId === player.id) {
                        userRow = (
                          <View key={playerId} style={styles.playerScoreRow}>
                            <Text style={[styles.scoreText, { fontWeight: 'bold' }]}> {s.playerName} (Me): {s.points} points </Text>
                            {matchResult.bets?.[playerId] && (
                              <Text style={styles.betResultText}> ({matchResult.bets[playerId].predictedHomeGoals} - {matchResult.bets[playerId].predictedAwayGoals}) </Text>
                            )}
                          </View>
                        );
                      } else {
                        otherRows.push(
                          <View key={playerId} style={styles.playerScoreRow}>
                            <Text style={styles.scoreText}> {s.playerName}: {s.points} points </Text>
                            {matchResult.bets?.[playerId] && (
                              <Text style={styles.betResultText}> ({matchResult.bets[playerId].predictedHomeGoals} - {matchResult.bets[playerId].predictedAwayGoals}) </Text>
                            )}
                          </View>
                        );
                      }
                    });
                    if (!userRow && player) {
                      userRow = (
                        <View key={player.id} style={styles.playerScoreRow}>
                          <Text style={[styles.scoreText, { fontWeight: 'bold' }]}> {player.name} (Me): <Text style={{ fontStyle: 'italic', color: '#999' }}>No bet</Text> </Text>
                        </View>
                      );
                    }
                    return [userRow, ...otherRows];
                  })()}
                </View>
              ) : matchResult.bets && (
                <View style={styles.scoresContainer}>
                  {(() => {
                    const entries = Object.entries(matchResult.bets);
                    let userRow = null;
                    let otherRows: React.ReactNode[] = [];
                    entries.forEach(([playerId, betData]) => {
                      const b = betData as { playerName: string; predictedHomeGoals: number; predictedAwayGoals: number };
                      if (player && playerId === player.id) {
                        userRow = (
                          <View key={playerId} style={styles.playerScoreRow}>
                            <Text style={[styles.scoreText, { fontWeight: 'bold' }]}> {b.playerName} (Me): </Text>
                            <Text style={styles.betResultText}> ({b.predictedHomeGoals} - {b.predictedAwayGoals}) </Text>
                          </View>
                        );
                      } else {
                        otherRows.push(
                          <View key={playerId} style={styles.playerScoreRow}>
                            <Text style={styles.scoreText}> {b.playerName}: </Text>
                            <Text style={styles.betResultText}> ({b.predictedHomeGoals} - {b.predictedAwayGoals}) </Text>
                          </View>
                        );
                      }
                    });
                    if (!userRow && player) {
                      userRow = (
                        <View key={player.id} style={styles.playerScoreRow}>
                          <Text style={[styles.scoreText, { fontWeight: 'bold' }]}> {player.name} (Me): <Text style={{ fontStyle: 'italic', color: '#999' }}>No bet</Text> </Text>
                        </View>
                      );
                    }
                    return [userRow, ...otherRows];
                  })()}
                </View>
              )}
            </View>
          )}
        </>
      )}
    </View>
  );
}

export default function MatchesList({ gameId }: { gameId: string }) {
  const { incomingMatches, pastMatches, loading: matchesLoading, error: matchesError, refresh } = useMatches(gameId);
  const [tempScores, setTempScores] = useState<TempScores>({});
  const [expandedMatches, setExpandedMatches] = useState<{ [key: string]: boolean }>({});
  const [refreshing, setRefreshing] = useState(false);
  const [editingMatchId, setEditingMatchId] = useState<string | null>(null);
  const [currentMatchday, setCurrentMatchday] = useState<number | null>(null);
  const { submitBet, error: submitError } = useBetSubmission(gameId);
  const { player } = useAuth();
  const { t } = useTranslation();
  const scrollViewRef = React.useRef<ScrollView>(null);

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
      setCurrentMatchday(sortedMatchdays[0]);
    }
  }, [sortedMatchdays, currentMatchday]);

  // Group matches by time within the current matchday
  const getMatchesByTime = (matchday: number) => {
    const matchdayMatches = matchesByMatchday[matchday] || [];
    // Sort by time (ascending)
    const sortedMatches = matchdayMatches.sort((a, b) => 
      a.match.getDate().getTime() - b.match.getDate().getTime()
    );
    // Group by time
    const groupedByTime = sortedMatches.reduce((acc, matchResult) => {
      const timeKey = formatTime(matchResult.match.getDate());
      if (!acc[timeKey]) {
        acc[timeKey] = [];
      }
      acc[timeKey].push(matchResult);
      return acc;
    }, {} as { [key: string]: any[] });
    return groupedByTime;
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  // Initialize tempScores with existing bets
  useEffect(() => {
    const initialTempScores: TempScores = {};
    [...Object.values(incomingMatches), ...Object.values(pastMatches)].forEach((matchResult: any) => {
      if (player && matchResult.bets && matchResult.bets[player.id]) {
        const userBet = matchResult.bets[player.id];
        initialTempScores[matchResult.match.id()] = {
          home: userBet.predictedHomeGoals,
          away: userBet.predictedAwayGoals
        };
      }
    });
    setTempScores(initialTempScores);
  }, [incomingMatches, pastMatches, player]);

  useEffect(() => {
    if (submitError) {
      Alert.alert(
        t('games.betNotSaved'),
        t('games.betNotSavedMessage'),
        [{ text: t('common.ok') }]
      );
    }
  }, [submitError, t]);

  const toggleBetSection = (matchId: string) => {
    setExpandedMatches(prev => ({
      ...prev,
      [matchId]: !prev[matchId]
    }));
  };

  const handleBetChange = async (matchId: string, team: 'home' | 'away', value: string) => {
    const matchResult = matches.find((m: any) => m.match.id() === matchId);
    if (!matchResult) return;
    const currentTempScores = tempScores[matchId] || {};
    const newTempScores = {
      ...currentTempScores,
      [team]: value && value.trim() !== '' ? parseInt(value) : undefined
    };
    setTempScores(prev => ({
      ...prev,
      [matchId]: newTempScores
    }));
    if (newTempScores.home !== undefined && newTempScores.away !== undefined) {
      const homeGoals = Number(newTempScores.home);
      const awayGoals = Number(newTempScores.away);
      if (!isNaN(homeGoals) && !isNaN(awayGoals) && homeGoals >= 0 && awayGoals >= 0) {
        await submitBet(matchId, homeGoals, awayGoals);
        await refresh();
      }
    }
  };

  const handleFocus = (matchId: string) => {
    setEditingMatchId(matchId);
  };

  const handleBlur = () => {
    setEditingMatchId(null);
  };

  const handleDone = () => {
    setEditingMatchId(null);
    Keyboard.dismiss();
  };

  const navigateMatchday = (direction: 'prev' | 'next') => {
    if (!currentMatchday) return;
    const currentIndex = sortedMatchdays.indexOf(currentMatchday);
    if (direction === 'prev' && currentIndex > 0) {
      setCurrentMatchday(sortedMatchdays[currentIndex - 1]);
    } else if (direction === 'next' && currentIndex < sortedMatchdays.length - 1) {
      setCurrentMatchday(sortedMatchdays[currentIndex + 1]);
    }
  };

  // Find the next closest unbet match in the current matchday
  const getNextUnbetMatchId = () => {
    if (!currentMatchday) return null;
    const matchesInDay = matchesByMatchday[currentMatchday] || [];
    const now = new Date();
    // Only consider matches in the future or in progress
    const unbetMatches = matchesInDay.filter(m => {
      const matchDate = m.match.getDate();
      const isFutureOrInProgress = !m.match.isFinished() || matchDate > now;
      const hasBet = player && m.bets && m.bets[player.id];
      return isFutureOrInProgress && !hasBet;
    });
    if (unbetMatches.length === 0) return null;
    // Find the closest in time to now
    unbetMatches.sort((a, b) => a.match.getDate().getTime() - b.match.getDate().getTime());
    return unbetMatches[0].match.id();
  };

  const nextUnbetMatchId = getNextUnbetMatchId();

  // Scroll to the next unbet match
  const scrollToNextMatch = () => {
    if (!nextUnbetMatchId || !scrollViewRef.current) return;
    // Find the y offset of the next match card
    // For simplicity, scroll to top (could be improved with refs per card)
    scrollViewRef.current.scrollTo({ y: 0, animated: true });
  };

  if (matchesLoading && !refreshing) {
    return (
      <View style={[styles.container, { backgroundColor: colors.loadingBackground }]}> 
        <ActivityIndicator testID="loading-indicator" size="large" color={colors.primary} />
      </View>
    );
  }

  const currentMatchdayMatches = currentMatchday ? getMatchesByTime(currentMatchday) : {};
  const sortedTimes = Object.keys(currentMatchdayMatches).sort();

  // Show legend if not editing a match or if the editing match has a clear favorite
  const editingMatch = editingMatchId ? matches.find(m => m.match.id() === editingMatchId) : null;
  const shouldShowLegend = !editingMatchId || (editingMatch && editingMatch.match.hasClearFavorite());

  return (
    <KeyboardAvoidingView 
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <ScrollView 
        ref={scrollViewRef}
        style={styles.scrollView}
        refreshControl={
          <RefreshControl
            refreshing={refreshing}
            onRefresh={onRefresh}
            colors={[colors.primary]}
            tintColor={colors.primary}
            progressBackgroundColor="#25292e"
            progressViewOffset={20}
          />
        }
      >
        {/* Matchday Navigation */}
        <View style={styles.matchdayNavigation}>
          <TouchableOpacity 
            style={[
              styles.navButton, 
              sortedMatchdays.indexOf(currentMatchday!) <= 0 && styles.navButtonDisabled
            ]}
            onPress={() => navigateMatchday('prev')}
            disabled={sortedMatchdays.indexOf(currentMatchday!) <= 0}
            testID="prev-matchday-button"
          >
            <Ionicons name="chevron-back" size={24} color="#fff" />
          </TouchableOpacity>
          <View style={styles.matchdayInfo}>
            <Text style={styles.matchdayTitle}>{t('games.matchday')} {currentMatchday}</Text>
            {currentMatchday && matchesByMatchday[currentMatchday] && matchesByMatchday[currentMatchday].length > 0 && (
              <Text style={styles.matchdayDate}>
                {formatDate(matchesByMatchday[currentMatchday][0].match.getDate())}
              </Text>
            )}
          </View>
          <TouchableOpacity 
            style={[
              styles.navButton, 
              sortedMatchdays.indexOf(currentMatchday!) >= sortedMatchdays.length - 1 && styles.navButtonDisabled
            ]}
            onPress={() => navigateMatchday('next')}
            disabled={sortedMatchdays.indexOf(currentMatchday!) >= sortedMatchdays.length - 1}
            testID="next-matchday-button"
          >
            <Ionicons name="chevron-forward" size={24} color="#fff" />
          </TouchableOpacity>
        </View>
        {/* Matches for current matchday */}
        <View style={styles.matchesContainer}>
          {sortedTimes.map(time => {
            const matchesAtTime = currentMatchdayMatches[time];
            const matchDate = matchesAtTime[0]?.match.getDate();
            const dayDisplay = matchDate ? formatDate(matchDate) : '';
            return (
              <View key={time} style={styles.timeGroup}>
                <View style={styles.timeHeaderContainer}>
                  <Text style={styles.timeHeader}>{time}</Text>
                  <Text style={styles.dayHeader}>{dayDisplay}</Text>
                </View>
                {matchesAtTime.map((matchResult: any) => (
                  <MatchCard
                    key={matchResult.match.id()}
                    matchResult={matchResult}
                    tempScores={tempScores}
                    expandedMatches={expandedMatches}
                    onBetChange={handleBetChange}
                    onToggleBetSection={toggleBetSection}
                    onFocus={() => handleFocus(matchResult.match.id())}
                    onBlur={handleBlur}
                    onDone={editingMatchId === matchResult.match.id() ? handleDone : undefined}
                  />
                ))}
              </View>
            );
          })}
        </View>
        {/* Legend for odds indicators */}
        {shouldShowLegend && (
          <View style={styles.legendContainer}>
            <Text style={styles.legendTitle}>{t('games.oddsLegend')}</Text>
            <View style={styles.legendItem}>
              <Text style={styles.legendStar}>⭐</Text>
              <Text style={styles.legendText}>{t('games.clearFavorite')}</Text>
            </View>
            <View style={styles.legendItem}>
              <Text style={styles.legendMark}>x1</Text>
              <Text style={styles.legendText}>{t('games.clearFavoriteBonus')}</Text>
            </View>
            <View style={styles.legendItem}>
              <Text style={styles.legendMark}>×1.5</Text>
              <Text style={styles.legendText}>{t('games.drawBonus')}</Text>
            </View>
            <View style={styles.legendItem}>
              <Text style={styles.legendMark}>×2</Text>
              <Text style={styles.legendText}>{t('games.outsiderWinBonus')}</Text>
            </View>
          </View>
        )}
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#25292e',
  },
  scrollView: {
    flex: 1,
  },
  matchdayNavigation: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 16,
    paddingVertical: 8,
    backgroundColor: '#333',
    borderBottomWidth: 1,
    borderBottomColor: '#444',
  },
  navButton: {
    padding: 8,
  },
  navButtonDisabled: {
    opacity: 0.5,
  },
  matchdayInfo: {
    flex: 1,
    alignItems: 'center',
  },
  matchdayTitle: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
  matchdayDate: {
    color: '#ccc',
    fontSize: 14,
    marginTop: 4,
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
    fontWeight: 'bold',
    color: '#fff',
  },
  dayHeader: {
    fontSize: 14,
    color: '#ccc',
    marginTop: 2,
  },
  matchCard: {
    backgroundColor: '#f5f5f5',
    padding: 16,
    borderRadius: 8,
    marginBottom: 12,
    position: 'relative',
  },
  matchCardWithTag: {
    paddingTop: 48,
  },
  finishedMatchCard: {
    backgroundColor: '#d3d3d3',
    opacity: 0.6,
  },
  bettingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: 16,
  },
  teamContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
    justifyContent: 'space-between',
  },
  awayTeamContainer: {
    flexDirection: 'row-reverse',
  },
  teamName: {
    fontSize: 14,
    color: '#333',
    flex: 1,
    textAlign: 'left',
  },
  awayTeamName: {
    textAlign: 'right',
  },
  vsText: {
    fontSize: 14,
    color: '#666',
    marginHorizontal: 8,
  },
  betInput: {
    borderWidth: 1,
    borderColor: '#ccc',
    borderRadius: 4,
    padding: 8,
    width: 40,
    textAlign: 'center',
    backgroundColor: '#fff',
  },
  disabledInput: {
    backgroundColor: '#f0f0f0',
    color: '#999',
  },
  resultRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 8,
  },
  resultTeam: {
    fontSize: 14,
    color: '#333',
    flex: 1,
    textAlign: 'center',
  },
  resultScore: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
    marginHorizontal: 8,
  },
  statusTag: {
    position: 'absolute',
    top: 8,
    right: 8,
    backgroundColor: 'rgba(0,0,0,0.7)',
    paddingVertical: 4,
    paddingHorizontal: 8,
    borderRadius: 5,
    zIndex: 1,
  },
  statusTagText: {
    color: '#fff',
    fontSize: 12,
    fontWeight: 'bold',
  },
  successTag: {
    backgroundColor: '#4CAF50',
  },
  inProgressTag: {
    backgroundColor: '#FFC107',
  },
  finishedTag: {
    backgroundColor: '#9E9E9E',
  },
  primaryTag: {
    backgroundColor: colors.primary,
  },
  negativeTag: {
    backgroundColor: '#ff6b6b',
  },
  toggleButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 8,
    paddingHorizontal: 16,
    backgroundColor: '#e8e8e8',
    borderRadius: 4,
    marginTop: 8,
    marginBottom: 4,
  },
  toggleButtonText: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#333',
  },
  betResultContainer: {
    marginTop: 8,
  },
  scoresContainer: {
    backgroundColor: '#f9f9f9',
    padding: 12,
    borderRadius: 4,
  },
  playerScoreRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 4,
  },
  scoreText: {
    fontSize: 14,
    color: '#333',
  },
  betResultText: {
    fontSize: 14,
    color: '#666',
    fontStyle: 'italic',
  },
  legendContainer: {
    marginTop: 24,
    marginHorizontal: 16,
    padding: 16,
    backgroundColor: '#333',
    borderRadius: 8,
  },
  legendTitle: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
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
    color: '#fff',
    fontSize: 14,
    textAlign: 'left',
  },
  doneButton: {
    backgroundColor: colors.primary,
    paddingVertical: 10,
    paddingHorizontal: 20,
    borderRadius: 5,
    alignSelf: 'center',
    marginTop: 10,
  },
  doneButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: 'bold',
  },
  oddsContainer: {
    marginTop: 10,
    paddingVertical: 8,
    paddingHorizontal: 12,
    backgroundColor: '#e0e0e0',
    borderRadius: 4,
  },
  oddsRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'center',
  },
  oddsItem: {
    alignItems: 'center',
  },
  oddsLabel: {
    fontSize: 12,
    color: '#666',
    marginBottom: 4,
  },
  oddsValue: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#333',
  },
  favoriteStar: {
    fontSize: 16,
    color: '#000',
  },
  outsiderMark: {
    fontSize: 14,
    color: '#000',
  },
  drawMark: {
    fontSize: 14,
    color: '#000',
  },
  nextMatchButton: {
    backgroundColor: colors.secondary,
    paddingVertical: 14,
    paddingHorizontal: 32,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    marginVertical: 16,
    marginHorizontal: 32,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 6,
    elevation: 4,
  },
  nextMatchButtonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: 'bold',
  },
}); 