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

interface TempScores {
  [key: string]: {
    home?: number;
    away?: number;
  };
}

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
    <View style={styles.teamContainer}>
      {!isAway && <Text style={styles.teamName}>{teamName}</Text>}
      <TextInput
        style={[
          styles.betInput,
          !canModify && { backgroundColor: '#e0e0e0' }
        ]}
        value={value}
        onChangeText={onChange}
        keyboardType="numeric"
        maxLength={2}
        placeholder="0"
        editable={canModify}
        returnKeyType="done"
        onSubmitEditing={() => Keyboard.dismiss()}
        onFocus={onFocus}
        onBlur={onBlur}
      />
      {isAway && <Text style={[styles.teamName, styles.awayTeamName]}>{teamName}</Text>}
    </View>
  );
}

// Wrapper component that has access to the time service
function MatchCard({ matchResult, tempScores, expandedMatches, onBetChange, onToggleBetSection, onFocus, onBlur }: {
  matchResult: MatchResult;
  tempScores: TempScores;
  expandedMatches: { [key: string]: boolean };
  onBetChange: (matchId: string, team: 'home' | 'away', value: string) => void;
  onToggleBetSection: (matchId: string) => void;
  onFocus: () => void;
  onBlur: () => void;
}) {
  const timeService = useTimeService();
  const { player } = useAuth();
  const now = timeService.now();
  const canModify = !matchResult.match.isFinished() && 
                   !matchResult.match.isInProgress() && 
                   (matchResult.bets?.[player?.id || '']?.isModifiable(now) ?? true);

  return (
    <View 
      key={matchResult.match.id()} 
      style={[
        styles.matchCard,
        matchResult.match.isFinished() && (
          matchResult.scores && Object.values(matchResult.scores).some(scoreData => scoreData.points > 0)
            ? styles.successfulBetMatchCard 
            : styles.finishedMatchCard
        ),
        matchResult.match.isInProgress() && styles.inProgressMatchCard
      ]}
    >
      <View style={styles.bettingContainer}>
        <TeamInput
          teamName={matchResult.match.getHomeTeam()}
          value={matchResult.match.isFinished() 
            ? matchResult.match.getHomeGoals().toString() 
            : (tempScores[matchResult.match.id()]?.home?.toString() || '')}
          onChange={(value) => onBetChange(matchResult.match.id(), 'home', value)}
          canModify={canModify}
          onFocus={() => onFocus()}
          onBlur={() => onBlur()}
        />
        <Text style={styles.vsText}>vs</Text>
        <TeamInput
          teamName={matchResult.match.getAwayTeam()}
          value={matchResult.match.isFinished() 
            ? matchResult.match.getAwayGoals().toString() 
            : (tempScores[matchResult.match.id()]?.away?.toString() || '')}
          onChange={(value) => onBetChange(matchResult.match.id(), 'away', value)}
          canModify={canModify}
          isAway
          onFocus={() => onFocus()}
          onBlur={() => onBlur()}
        />
      </View>
      
      {(matchResult.match.isFinished() || matchResult.match.isInProgress()) && (
        <>
          <TouchableOpacity 
            style={styles.toggleButton} 
            onPress={() => onToggleBetSection(matchResult.match.id())}
          >
            <Text style={styles.toggleButtonText}>
              {matchResult.match.isFinished() 
                ? (matchResult.scores && Object.values(matchResult.scores).some(scoreData => scoreData.points > 0)
                    ? `Your bet won you ${Object.values(matchResult.scores).find(scoreData => scoreData.points > 0)?.points || 0} points`
                    : 'Other players\' bets')
                : 'Other players\' bets'}
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
  
  // Create a custom hook that accepts the game ID
  const useMatchesForGame = (gameId: string) => {
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
      fetchMatches();
    }, [gameId]);

    return {
      incomingMatches,
      pastMatches,
      loading,
      error,
      refresh: fetchMatches
    };
  };

  const { incomingMatches, pastMatches, loading: matchesLoading, error: matchesError, refresh } = useMatchesForGame(gameId);
  const [tempScores, setTempScores] = useState<TempScores>({});
  const [expandedMatches, setExpandedMatches] = useState<{ [key: string]: boolean }>({});
  const [refreshing, setRefreshing] = useState(false);
  const [editingMatchId, setEditingMatchId] = useState<string | null>(null);
  const { submitBet, error: submitError } = useBetSubmission();
  const { player } = useAuth();

  // Combine incoming and past matches
  const matches = [...Object.values(incomingMatches), ...Object.values(pastMatches)];

  const onRefresh = React.useCallback(async () => {
    setRefreshing(true);
    await refresh();
    setRefreshing(false);
  }, [refresh]);

  // Initialize tempScores with existing bets
  useEffect(() => {
    const initialTempScores: TempScores = {};
    [...Object.values(incomingMatches), ...Object.values(pastMatches)].forEach((matchResult: MatchResult) => {
      // Find the current user's bet (we'll need to get this from auth context later)
      // For now, we'll use the first bet as a placeholder
      if (matchResult.bets && Object.keys(matchResult.bets).length > 0) {
        const firstBetId = Object.keys(matchResult.bets)[0];
        const firstBet = matchResult.bets[firstBetId];
        initialTempScores[matchResult.match.id()] = {
          home: firstBet.predictedHomeGoals,
          away: firstBet.predictedAwayGoals
        };
      }
    });
    setTempScores(initialTempScores);
  }, [incomingMatches, pastMatches]);

  useEffect(() => {
    if (submitError) {
      Alert.alert(
        "Bet Not Saved",
        "We couldn't save your bet. Please check your internet connection and try again.",
        [{ text: "OK" }]
      );
    }
  }, [submitError]);

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
      [team]: value ? parseInt(value) : undefined
    };

    setTempScores(prev => ({
      ...prev,
      [matchId]: newTempScores
    }));

    // Submit bet if both scores are provided and are valid numbers
    if (newTempScores.home !== undefined && newTempScores.away !== undefined) {
      const homeGoals = Number(newTempScores.home);
      const awayGoals = Number(newTempScores.away);
      
      // Only submit if both values are valid numbers
      if (!isNaN(homeGoals) && !isNaN(awayGoals)) {
        await submitBet(matchId, homeGoals, awayGoals);
      }
    }
  };

  const handleFocus = (matchId: string) => {
    setEditingMatchId(matchId);
  };

  const handleBlur = () => {
    setEditingMatchId(null);
  };

  if (matchesLoading && !refreshing) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="large" color="#0000ff" />
      </View>
    );
  }

  // Filter matches to only show the one that is being edited, if any, else show all matches
  const filteredMatches = matches.filter((matchResult: MatchResult) => 
    !editingMatchId || matchResult.match.id() === editingMatchId
  );

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
            colors={['#ffffff']}
            tintColor="#ffffff"
            progressBackgroundColor="#25292e"
            progressViewOffset={20}
          />
        }
      >
        {matchesError ? (
          <View style={styles.errorContainer}>
            <Text style={styles.errorText}>Failed to load matches: {matchesError.message}</Text>
            <Text style={styles.refreshHint}>Pull down to refresh</Text>
          </View>
        ) : (
          filteredMatches
            .sort((a: MatchResult, b: MatchResult) => {
              if (a.match.getMatchday() !== b.match.getMatchday()) {
                return a.match.getMatchday() - b.match.getMatchday();
              }
              return a.match.getDate().getTime() - b.match.getDate().getTime();
            })
            .map((matchResult: MatchResult) => (
              <MatchCard
                key={matchResult.match.id()}
                matchResult={matchResult}
                tempScores={tempScores}
                expandedMatches={expandedMatches}
                onBetChange={handleBetChange}
                onToggleBetSection={toggleBetSection}
                onFocus={() => handleFocus(matchResult.match.id())}
                onBlur={handleBlur}
              />
            ))
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
    <TimeServiceProvider service={timeService}>
      <MatchesList />
    </TimeServiceProvider>
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
    padding: 8,
    backgroundColor: '#f8f8f8',
    borderRadius: 4,
  },
  betResultText: {
    fontSize: 14,
    color: '#666',
  },
  scoresContainer: {
    marginTop: 6,
  },
  playerScoreRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginVertical: 1,
  },
  scoreText: {
    fontSize: 14,
    color: '#666',
  },
  loadingContainer: {
    padding: 20,
    alignItems: 'center',
    justifyContent: 'center',
    minHeight: 200,
  },
}); 