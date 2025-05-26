import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert, ScrollView, RefreshControl } from 'react-native';
import { useMatches } from '../../hooks/useMatches';
import { useBetSubmission } from '../../hooks/useBetSubmission';
import { Ionicons } from '@expo/vector-icons';
import { MatchResult } from '../../src/types/match';

interface TempScores {
  [key: string]: {
    home?: number;
    away?: number;
  };
}

export default function TabOneScreen() {
  const { incomingMatches, pastMatches, loading: matchesLoading, error: matchesError, refresh } = useMatches();
  const [tempScores, setTempScores] = useState<TempScores>({});
  const [expandedMatches, setExpandedMatches] = useState<{ [key: string]: boolean }>({});
  const [refreshing, setRefreshing] = useState(false);
  const { submitBet, error: submitError } = useBetSubmission();

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
      if (matchResult.bets?.['Player1']) {
        initialTempScores[matchResult.match.id()] = {
          home: matchResult.bets['Player1'].predictedHomeGoals,
          away: matchResult.bets['Player1'].predictedAwayGoals
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

  if (matchesLoading && !refreshing) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="large" color="#0000ff" />
      </View>
    );
  }

  if (matchesError) {
    return (
      <View style={styles.container}>
        <Text style={styles.errorText}>Error: {matchesError?.message}</Text>
      </View>
    );
  }

  return (
    <ScrollView 
      style={styles.container}
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
      <Text style={styles.title}>Matches</Text>
      {matches
        .sort((a: MatchResult, b: MatchResult) => {
          if (a.match.getMatchday() !== b.match.getMatchday()) {
            return a.match.getMatchday() - b.match.getMatchday();
          }
          return a.match.getDate().getTime() - b.match.getDate().getTime();
        })
        .map((matchResult: MatchResult) => (
        <View 
          key={matchResult.match.id()} 
          style={[
            styles.matchCard,
            matchResult.match.isFinished() && (
              matchResult.scores && matchResult.scores['Player1'] > 0 
                ? styles.successfulBetMatchCard 
                : styles.finishedMatchCard
            )
          ]}
        >
          <View style={styles.bettingContainer}>
            <View style={styles.teamContainer}>
              <Text style={styles.teamName}>{matchResult.match.getHomeTeam()}</Text>
              <TextInput
                style={styles.betInput}
                value={matchResult.match.isFinished() 
                  ? matchResult.match.getHomeGoals().toString() 
                  : (tempScores[matchResult.match.id()]?.home?.toString() || '')}
                onChangeText={(value) => handleBetChange(matchResult.match.id(), 'home', value)}
                keyboardType="numeric"
                maxLength={2}
                placeholder="0"
                editable={!matchResult.match.isFinished()}
                returnKeyType="done"
                onSubmitEditing={() => Keyboard.dismiss()}
              />
            </View>
            <Text style={styles.vsText}>vs</Text>
            <View style={styles.teamContainer}>
              <TextInput
                style={styles.betInput}
                value={matchResult.match.isFinished() 
                  ? matchResult.match.getAwayGoals().toString() 
                  : (tempScores[matchResult.match.id()]?.away?.toString() || '')}
                onChangeText={(value) => handleBetChange(matchResult.match.id(), 'away', value)}
                keyboardType="numeric"
                maxLength={2}
                placeholder="0"
                editable={!matchResult.match.isFinished()}
                returnKeyType="done"
                onSubmitEditing={() => Keyboard.dismiss()}
              />
              <Text style={[styles.teamName, styles.awayTeamName]}>{matchResult.match.getAwayTeam()}</Text>
            </View>
          </View>
          
          {matchResult.match.isFinished() && (
            <>
              <TouchableOpacity 
                style={styles.toggleButton} 
                onPress={() => toggleBetSection(matchResult.match.id())}
              >
                <Text style={styles.toggleButtonText}>
                  {matchResult.scores?.['Player1'] !== undefined 
                    ? `Your bet won you ${matchResult.scores['Player1']} points`
                    : 'Your bet'}
                </Text>
                <Ionicons 
                  name={expandedMatches[matchResult.match.id()] ? "chevron-up" : "chevron-down"} 
                  size={24} 
                  color="#333" 
                />
              </TouchableOpacity>
              
              {expandedMatches[matchResult.match.id()] && (
                <View style={styles.betResultContainer}>
                  {matchResult.bets?.['Player1'] ? (
                    <View style={styles.betResultContent}>
                      <Text style={styles.betResultText}>
                        {matchResult.match.getHomeTeam()}: {matchResult.bets['Player1'].predictedHomeGoals}
                      </Text>
                      <Text style={styles.betResultText}>
                        {matchResult.match.getAwayTeam()}: {matchResult.bets['Player1'].predictedAwayGoals}
                      </Text>
                    </View>
                  ) : (
                    <Text style={styles.betResultText}>No bet placed yet</Text>
                  )}
                  {matchResult.scores && (
                    <View style={styles.scoresContainer}>
                      <Text style={styles.scoresTitle}>All players' points:</Text>
                      {Object.entries(matchResult.scores).map(([player, score]) => (
                        <View key={player} style={styles.playerScoreRow}>
                          <Text style={styles.scoreText}>
                            {player}: {score} points
                          </Text>
                          {matchResult.bets?.[player] && (
                            <Text style={styles.betResultText}>
                              ({matchResult.bets[player].predictedHomeGoals} - {matchResult.bets[player].predictedAwayGoals})
                            </Text>
                          )}
                        </View>
                      ))}
                    </View>
                  )}
                </View>
              )}
            </>
          )}
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#25292e',
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
  errorText: {
    color: 'red',
    fontSize: 16,
  },
  toggleButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 8,
    paddingHorizontal: 16,
    backgroundColor: '#f0f0f0',
    borderRadius: 4,
    marginTop: 8,
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
    marginTop: 4,
  },
  betResultContent: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  betResultText: {
    fontSize: 14,
    color: '#666',
  },
  scoresContainer: {
    marginTop: 12,
    paddingTop: 12,
    borderTopWidth: 1,
    borderTopColor: '#ddd',
  },
  scoresTitle: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#333',
    marginBottom: 4,
  },
  scoreText: {
    fontSize: 14,
    color: '#666',
    marginVertical: 2,
  },
  playerScoreRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginVertical: 2,
  },
});
