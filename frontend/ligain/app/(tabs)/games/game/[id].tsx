import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert, ScrollView, RefreshControl, KeyboardAvoidingView, Platform } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { useLocalSearchParams, useRouter } from 'expo-router';

// Local imports
import { useMatches } from '../../../../hooks/useMatches';
import { useBetSubmission } from '../../../../hooks/useBetSubmission';
import { MatchResult } from '../../../../src/types/match';
import { MockTimeService } from '../../../../src/services/timeService';
import { TimeServiceProvider, useTimeService } from '../../../../src/contexts/TimeServiceContext';
import { useAuth } from '../../../../src/contexts/AuthContext';
import { API_CONFIG } from '../../../../src/config/api';
import { SeasonMatch } from '../../../../src/types/match';
import { colors } from '../../../../src/constants/colors';
import { useTranslation } from 'react-i18next';
import { formatTime, formatDate } from '../../../../src/utils/dateUtils';

interface TempScores {
  [key: string]: {
    home?: number;
    away?: number;
  };
}



// Custom hook for fetching matches for a specific game
const useMatchesForGame = (gameId: string) => {
  const { player } = useAuth();
  const [incomingMatches, setIncomingMatches] = useState<{ [key: string]: MatchResult }>({});
  const [pastMatches, setPastMatches] = useState<{ [key: string]: MatchResult }>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const fetchMatches = async () => {
    try {
      const { getAuthenticatedHeaders } = await import('../../../../src/config/api');
      const headers = await getAuthenticatedHeaders();
      console.log('🔧 useMatches - Using authenticated headers:', {
        hasApiKey: !!headers['X-API-Key'],
        hasAuth: !!headers['Authorization']
      });
      
      const response = await fetch(`${API_CONFIG.BASE_URL}/api/game/${gameId}/matches`, {
        headers,
      });
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(`${response.status}: ${errorData.error || 'Unknown error'}`);
      }
      
      const data = await response.json();
      
      // Convert the matches to SeasonMatch objects and bets to BetImpl objects
      const processMatches = (matches: any) => {
        const processed: { [key: string]: MatchResult } = {};
        Object.entries(matches).forEach(([key, value]: [string, any]) => {
          const match = SeasonMatch.fromJSON(value.match);
          
          const bets = value.bets ? Object.entries(value.bets).reduce((acc: { [key: string]: any }, [playerName, betData]: [string, any]) => {
            acc[betData.playerId] = {
              playerId: betData.playerId,
              playerName: betData.playerName,
              predictedHomeGoals: betData.predictedHomeGoals,
              predictedAwayGoals: betData.predictedAwayGoals,
              isModifiable: (now: Date) => {
                // Simple implementation - can be enhanced later
                return !match.isFinished() && !match.isInProgress();
              }
            };
            return acc;
          }, {}) : null;
          
          // Process scores with new structure - use playerId as key
          const scores = value.scores ? Object.entries(value.scores).reduce((acc: { [key: string]: any }, [playerName, scoreData]: [string, any]) => {
            acc[scoreData.playerId] = {
              playerId: scoreData.playerId,
              playerName: scoreData.playerName,
              points: scoreData.points
            };
            return acc;
          }, {}) : null;
          
          processed[key] = {
            match,
            bets,
            scores
          };
        });
        return processed;
      };

      setIncomingMatches(processMatches(data.incomingMatches));
      setPastMatches(processMatches(data.pastMatches));
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err : new Error('Failed to fetch matches'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // Only fetch matches if we have a gameId and a player (authenticated)
    if (gameId && player) {
      console.log('🔧 useMatchesForGame - Fetching matches for authenticated player:', player.name);
      fetchMatches();
    } else if (gameId && !player) {
      console.log('🔧 useMatchesForGame - Waiting for player authentication...');
      setLoading(true);
    }
  }, [gameId, player]);

  return {
    incomingMatches,
    pastMatches,
    loading,
    error,
    refresh: fetchMatches
  };
};

function TeamInput({ 
  teamName, 
  value, 
  onChange, 
  canModify, 
  isAway = false,
  onFocus,
  onBlur
}: { 
  teamName: string; 
  value: string; 
  onChange: (value: string) => void; 
  canModify: boolean; 
  isAway?: boolean;
  onFocus?: () => void;
  onBlur?: () => void;
}) {
  return (
    <View style={[styles.teamContainer, isAway && styles.awayTeamContainer]}>
      <Text style={[styles.teamName, isAway && styles.awayTeamName]}>
        {teamName}
      </Text>
      <TextInput
        style={[styles.betInput, !canModify && styles.disabledInput]}
        value={value}
        onChangeText={onChange}
        keyboardType="numeric"
        editable={canModify}
        onFocus={onFocus}
        onBlur={onBlur}
      />
    </View>
  );
}

// Wrapper component that has access to the time service
function MatchCard({ matchResult, tempScores, expandedMatches, onBetChange, onToggleBetSection, onFocus, onBlur, onDone }: {
  matchResult: MatchResult;
  tempScores: TempScores;
  expandedMatches: { [key: string]: boolean };
  onBetChange: (matchId: string, team: 'home' | 'away', value: string) => void;
  onToggleBetSection: (matchId: string) => void;
  onFocus: () => void;
  onBlur: () => void;
  onDone?: () => void;
}) {
  const timeService = useTimeService();
  const { player } = useAuth();
  const { t } = useTranslation();
  const now = timeService.now();
  const isFuture = !matchResult.match.isFinished() && !matchResult.match.isInProgress();
  const userBet = player && matchResult.bets ? matchResult.bets[player.id] : undefined;
  const canModify = isFuture && (userBet?.isModifiable(now) !== false);

  return (
    <View 
      key={matchResult.match.id()} 
      style={[
        styles.matchCard,
        matchResult.match.isFinished() && (
          matchResult.scores && player && matchResult.scores[player.id] && matchResult.scores[player.id].points > 0
            ? styles.successfulBetMatchCard 
            : styles.finishedMatchCard
        ),
        matchResult.match.isInProgress() && styles.inProgressMatchCard
      ]}
    >
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
            />
          </>
        ) : (
          // For past/in-progress matches, keep showing all bets/scores as before
          <>
            <TeamInput
              teamName={matchResult.match.getHomeTeam()}
              value={matchResult.match.getHomeGoals().toString()}
              onChange={(value) => onBetChange(matchResult.match.id(), 'home', value)}
              canModify={canModify}
              onFocus={() => onFocus()}
              onBlur={() => onBlur()}
            />
            <Text style={styles.vsText}>{t('common.vs')}</Text>
            <TeamInput
              teamName={matchResult.match.getAwayTeam()}
              value={matchResult.match.getAwayGoals().toString()}
              onChange={(value) => onBetChange(matchResult.match.id(), 'away', value)}
              canModify={canModify}
              isAway
              onFocus={() => onFocus()}
              onBlur={() => onBlur()}
            />
          </>
        )}
      </View>
      
      {/* Done button for future matches when editing */}
      {isFuture && onDone && (
        <TouchableOpacity 
          style={styles.doneButton}
          onPress={onDone}
        >
          <Text style={styles.doneButtonText}>{t('common.done')}</Text>
        </TouchableOpacity>
      )}
      
      {/* Odds display */}
      <View style={styles.oddsContainer}>
        <View style={styles.oddsRow}>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>1</Text>
            <Text style={styles.oddsValue}>{matchResult.match.getHomeTeamOdds().toFixed(2)}</Text>
            {matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() === matchResult.match.getHomeTeam() && (
              <Text style={styles.favoriteStar}>⭐</Text>
            )}
            {matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() !== matchResult.match.getHomeTeam() && (
              <Text style={styles.outsiderMark}>×2</Text>
            )}
          </View>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>N</Text>
            <Text style={styles.oddsValue}>{matchResult.match.getDrawOdds().toFixed(2)}</Text>
            {matchResult.match.hasClearFavorite() && (
              <Text style={styles.drawMark}>×1.5</Text>
            )}
          </View>
          <View style={styles.oddsItem}>
            <Text style={styles.oddsLabel}>2</Text>
            <Text style={styles.oddsValue}>{matchResult.match.getAwayTeamOdds().toFixed(2)}</Text>
            {matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() === matchResult.match.getAwayTeam() && (
              <Text style={styles.favoriteStar}>⭐</Text>
            )}
            {matchResult.match.hasClearFavorite() && matchResult.match.getFavoriteTeam() !== matchResult.match.getAwayTeam() && (
              <Text style={styles.outsiderMark}>×2</Text>
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
              {matchResult.match.isFinished() 
                ? (matchResult.scores && player && matchResult.scores[player.id] && matchResult.scores[player.id].points > 0
                    ? `Your bet won you ${matchResult.scores[player.id].points} points`
                    : 'Players\' bets')
                : 'Players\' bets'}
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
                  {Object.entries(matchResult.scores).map(([playerId, scoreData]) => (
                    <View key={playerId} style={styles.playerScoreRow}>
                      <Text style={styles.scoreText}>
                        {scoreData.playerName}: {scoreData.points} points
                      </Text>
                      {matchResult.bets?.[playerId] && (
                        <Text style={styles.betResultText}>
                          ({matchResult.bets[playerId].predictedHomeGoals} - {matchResult.bets[playerId].predictedAwayGoals})
                        </Text>
                      )}
                    </View>
                  ))}
                </View>
              ) : matchResult.bets && (
                <View style={styles.scoresContainer}>
                  {Object.entries(matchResult.bets)
                    .map(([playerId, betData]) => (
                      <View key={playerId} style={styles.playerScoreRow}>
                        <Text style={styles.scoreText}>
                          {betData.playerName}:
                        </Text>
                        <Text style={styles.betResultText}>
                          ({betData.predictedHomeGoals} - {betData.predictedAwayGoals})
                        </Text>
                      </View>
                    ))}
                </View>
              )}
            </View>
          )}
        </>
      )}
    </View>
  );
}

function MatchesList() {
  const { id } = useLocalSearchParams();
  const gameId = id as string;
  
  console.log('📋 MatchesList - Rendering matches list for game:', gameId);
  
  const { incomingMatches, pastMatches, loading: matchesLoading, error: matchesError, refresh } = useMatchesForGame(gameId);
  const [tempScores, setTempScores] = useState<TempScores>({});
  const [expandedMatches, setExpandedMatches] = useState<{ [key: string]: boolean }>({});
  const [refreshing, setRefreshing] = useState(false);
  const [editingMatchId, setEditingMatchId] = useState<string | null>(null);
  const [currentMatchday, setCurrentMatchday] = useState<number | null>(null);
  const { submitBet, error: submitError } = useBetSubmission(gameId);
  const { player } = useAuth();
  const { t } = useTranslation();
  
  // Debug player information
  console.log('🔍 Current player:', player);

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
  }, {} as { [key: number]: MatchResult[] });

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
    }, {} as { [key: string]: MatchResult[] });

    return groupedByTime;
  };

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  // Initialize tempScores with existing bets
  useEffect(() => {
    console.log('🔍 Initializing tempScores - Player ID:', player?.id);
    console.log('🔍 Available bets:', Object.keys(incomingMatches).map(key => ({
      matchId: key,
      betKeys: Object.keys(incomingMatches[key].bets || {})
    })));
    
    const initialTempScores: TempScores = {};
    [...Object.values(incomingMatches), ...Object.values(pastMatches)].forEach((matchResult: MatchResult) => {
      // Only use the current user's bet for tempScores
      if (player && matchResult.bets && matchResult.bets[player.id]) {
        const userBet = matchResult.bets[player.id];
        console.log('🔍 Found user bet for match:', matchResult.match.id(), 'Bet:', userBet);
        initialTempScores[matchResult.match.id()] = {
          home: userBet.predictedHomeGoals,
          away: userBet.predictedAwayGoals
        };
      } else {
        console.log('🔍 No bet found for player:', player?.id, 'in match:', matchResult.match.id());
        console.log('🔍 Available bet keys:', Object.keys(matchResult.bets || {}));
      }
    });
    console.log('🔍 Final tempScores:', initialTempScores);
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
    const matchResult = matches.find((m: MatchResult) => m.match.id() === matchId);
    if (!matchResult) return;

    // Update temporary scores
    const currentTempScores = tempScores[matchId] || {};
    const newTempScores = {
      ...currentTempScores,
      [team]: value && value.trim() !== '' ? parseInt(value) : undefined
    };

    setTempScores(prev => ({
      ...prev,
      [matchId]: newTempScores
    }));

    // Submit bet if both scores are provided and are valid numbers greater than 0
    if (newTempScores.home !== undefined && newTempScores.away !== undefined) {
      const homeGoals = Number(newTempScores.home);
      const awayGoals = Number(newTempScores.away);
      
      // Only submit if both values are valid numbers and greater than or equal to 0
      if (!isNaN(homeGoals) && !isNaN(awayGoals) && homeGoals >= 0 && awayGoals >= 0) {
        console.log('🔍 Submitting bet:', { matchId, homeGoals, awayGoals, playerId: player?.id });
        await submitBet(matchId, homeGoals, awayGoals);
        console.log('🔍 Bet submitted successfully');
        // Refresh matches data to get the updated bet
        await refresh();
      }
    }
  };

  const handleFocus = (matchId: string) => {
    setEditingMatchId(matchId);
  };

  const handleBlur = () => {
    // Don't submit bet if fields are empty when blurring
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

  if (matchesLoading && !refreshing) {
    return (
      <View style={styles.container}>
        <ActivityIndicator testID="loading-indicator" size="large" color={colors.primary} />
      </View>
    );
  }

  // Filter matches to only show the one that is being edited, if any, else show all matches
  const filteredMatches = matches.filter((matchResult: MatchResult) => 
    !editingMatchId || matchResult.match.id() === editingMatchId
  );

  // Check if the currently editing match has a clear favorite
  const editingMatch = editingMatchId ? matches.find(m => m.match.id() === editingMatchId) : null;
  const shouldShowLegend = !editingMatchId || (editingMatch && editingMatch.match.hasClearFavorite());

  // Get current matchday matches grouped by time
  const currentMatchdayMatches = currentMatchday ? getMatchesByTime(currentMatchday) : {};
  const sortedTimes = Object.keys(currentMatchdayMatches).sort();

  return (
    <KeyboardAvoidingView 
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <ScrollView 
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
        {matchesError ? (
          <View style={styles.errorContainer}>
            <Text style={styles.errorText}>{t('games.failedToLoadMatches')} {matchesError.message}</Text>
            <Text style={styles.refreshHint}>{t('games.pullToRefresh')}</Text>
          </View>
        ) : sortedMatchdays.length === 0 ? (
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>{t('games.noMatchesAvailable')}</Text>
          </View>
        ) : (
          <>
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
              // Get the date from the first match to display the day
              const matchDate = matchesAtTime[0]?.match.getDate();
              const dayDisplay = matchDate ? formatDate(matchDate) : '';
              
              // Filter matches for this time slot
              const filteredMatchesAtTime = matchesAtTime.filter((matchResult: MatchResult) => 
                !editingMatchId || matchResult.match.id() === editingMatchId
              );
              
              // Only show time group if there are matches to display
              if (filteredMatchesAtTime.length === 0) {
                return null;
              }
              
              return (
                <View key={time} style={styles.timeGroup}>
                  <View style={styles.timeHeaderContainer}>
                    <Text style={styles.timeHeader}>{time}</Text>
                    <Text style={styles.dayHeader}>{dayDisplay}</Text>
                  </View>
                  {filteredMatchesAtTime.map((matchResult: MatchResult) => (
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
                  <Text style={styles.legendMark}>×1.5</Text>
                  <Text style={styles.legendText}>{t('games.drawBonus')}</Text>
                </View>
                <View style={styles.legendItem}>
                  <Text style={styles.legendMark}>×2</Text>
                  <Text style={styles.legendText}>{t('games.outsiderWinBonus')}</Text>
                </View>
              </View>
            )}
          </>
        )}
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

// Wrap the app with the TimeServiceProvider using MockTimeService
export default function GameScreen() {
  const { id } = useLocalSearchParams();
  const gameId = id as string;
  
  console.log('🏠 GameScreen - Rendering game screen with gameId:', gameId);
  
  const mockTime = new Date('2024-03-20T20:10:00');
  const timeService = React.useMemo(() => new MockTimeService(mockTime), []);
  
  return (
    <View testID="game-screen" style={styles.container}>
      <TimeServiceProvider service={timeService}>
        <MatchesList />
      </TimeServiceProvider>
    </View>
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
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    margin: 16,
    color: '#fff',
  },
  matchCard: {
    backgroundColor: '#f5f5f5',
    padding: 16,
    borderRadius: 8,
    marginHorizontal: 16,
    marginBottom: 12,
  },
  finishedMatchCard: {
    backgroundColor: '#9e9e9e',
    opacity: 0.7,
  },
  successfulBetMatchCard: {
    backgroundColor: '#2e7d32',
    opacity: 0.9,
  },
  inProgressMatchCard: {
    backgroundColor: '#f5e663', // A softer but more yellow tone
    opacity: 0.9,
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
  errorContainer: {
    padding: 20,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 200,
  },
  errorText: {
    color: 'red',
    fontSize: 16,
    textAlign: 'center',
    marginBottom: 20,
  },
  refreshHint: {
    color: '#666',
    fontSize: 14,
    textAlign: 'center',
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
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  emptyText: {
    color: '#fff',
    fontSize: 18,
    textAlign: 'center',
  },
  doneButton: {
    backgroundColor: '#4CAF50',
    paddingVertical: 8,
    paddingHorizontal: 16,
    borderRadius: 6,
    alignSelf: 'center',
    marginTop: 8,
  },
  doneButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: 'bold',
  },
  oddsContainer: {
    marginTop: 8,
    flexDirection: 'row',
    justifyContent: 'space-between',
    gap: 12,
  },
  oddsRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    width: '100%',
  },
  oddsItem: {
    flex: 1,
    alignItems: 'center',
    paddingVertical: 4,
    paddingHorizontal: 6,
    backgroundColor: '#e0e0e0',
    borderRadius: 4,
    minHeight: 28,
    justifyContent: 'center',
    marginHorizontal: 2,
    position: 'relative',
  },
  oddsLabel: {
    fontSize: 12,
    color: '#333',
    fontWeight: 'bold',
    marginBottom: 1,
  },
  oddsValue: {
    fontSize: 14,
    color: '#333',
    fontWeight: '600',
  },
  favoriteStar: {
    fontSize: 12,
    color: '#ffd700', // Gold color for favorite
    position: 'absolute',
    top: 2,
    right: 2,
  },
  outsiderMark: {
    fontSize: 12,
    color: colors.primary, // Primary color for outsider
    position: 'absolute',
    top: 2,
    right: 2,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 2,
    paddingVertical: 1,
    borderRadius: 2,
  },
  drawMark: {
    fontSize: 12,
    color: colors.primary, // Primary color for draw
    position: 'absolute',
    top: 2,
    right: 2,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    paddingHorizontal: 2,
    paddingVertical: 1,
    borderRadius: 2,
  },
  legendContainer: {
    marginTop: 16,
    marginBottom: 20,
    marginHorizontal: 16,
    padding: 12,
    backgroundColor: '#333',
    borderRadius: 8,
    borderWidth: 1,
    borderColor: '#444',
  },
  legendTitle: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 8,
    textAlign: 'center',
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 4,
  },
  legendStar: {
    fontSize: 16,
    marginRight: 8,
    width: 30,
    textAlign: 'left',
  },
  legendMark: {
    fontSize: 16,
    marginRight: 8,
    color: colors.primary,
    width: 30,
    textAlign: 'left',
  },
  legendText: {
    fontSize: 14,
    color: '#ccc',
    flex: 1,
  },
}); 