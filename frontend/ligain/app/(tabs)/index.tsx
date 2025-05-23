import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput, Keyboard, TouchableOpacity, Alert } from 'react-native';
import { useMatches } from '../../hooks/useMatches';
import { useBets } from '../../hooks/useBets';
import { useBetSubmission } from '../../hooks/useBetSubmission';
import { Ionicons } from '@expo/vector-icons';

interface TempScores {
  [key: string]: {
    home?: number;
    away?: number;
  };
}

export default function TabOneScreen() {
  const { matches, loading: matchesLoading, error: matchesError } = useMatches();
  const { bets, loading: betsLoading, error: betsError } = useBets(); 
  const [tempScores, setTempScores] = useState<TempScores>({});
  const [expandedMatches, setExpandedMatches] = useState<{ [key: string]: boolean }>({});
  const { submitBet, error: submitError } = useBetSubmission();

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
    const match = matches.find(m => m.id() === matchId);
    if (!match) return;

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

  if (matchesLoading || betsLoading) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="large" color="#0000ff" />
      </View>
    );
  }

  if (matchesError || betsError) {
    return (
      <View style={styles.container}>
        <Text style={styles.errorText}>Error: {matchesError?.message || betsError?.message}</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Matches</Text>
      {matches
        .sort((a, b) => {
          if (a.getMatchday() !== b.getMatchday()) {
            return a.getMatchday() - b.getMatchday();
          }
          return a.getDate().getTime() - b.getDate().getTime();
        })
        .map((match) => (
        <View 
          key={match.id()} 
          style={[
            styles.matchCard,
            match.isFinished() && styles.finishedMatchCard
          ]}
        >
          <View style={styles.bettingContainer}>
            <View style={styles.teamContainer}>
              <Text style={styles.teamName}>{match.getHomeTeam()}</Text>
              <TextInput
                style={styles.betInput}
                value={match.isFinished() 
                  ? match.getHomeGoals().toString() 
                  : (tempScores[match.id()]?.home?.toString() || '')}
                onChangeText={(value) => handleBetChange(match.id(), 'home', value)}
                keyboardType="numeric"
                maxLength={2}
                placeholder="0"
                editable={!match.isFinished()}
                returnKeyType="done"
                onSubmitEditing={() => Keyboard.dismiss()}
              />
            </View>
            <View style={styles.teamContainer}>
              <TextInput
                style={styles.betInput}
                value={match.isFinished() 
                  ? match.getAwayGoals().toString() 
                  : (tempScores[match.id()]?.away?.toString() || '')}
                onChangeText={(value) => handleBetChange(match.id(), 'away', value)}
                keyboardType="numeric"
                maxLength={2}
                placeholder="0"
                editable={!match.isFinished()}
                returnKeyType="done"
                onSubmitEditing={() => Keyboard.dismiss()}
              />
              <Text style={[styles.teamName, styles.awayTeamName]}>{match.getAwayTeam()}</Text>
            </View>
          </View>
          
          {match.isFinished() && (
            <>
              <TouchableOpacity 
                style={styles.toggleButton} 
                onPress={() => toggleBetSection(match.id())}
              >
                <Text style={styles.toggleButtonText}>Your Bet</Text>
                <Ionicons 
                  name={expandedMatches[match.id()] ? "chevron-up" : "chevron-down"} 
                  size={24} 
                  color="#333" 
                />
              </TouchableOpacity>
              
              {expandedMatches[match.id()] && (
                <View style={styles.betResultContainer}>
                  <View style={styles.betResultContent}>
                    <Text style={styles.betResultText}>
                      {match.getHomeTeam()}: {bets.find(bet => bet.match.id() === match.id())?.predictedHomeGoals || '-'}
                    </Text>
                    <Text style={styles.betResultText}>
                      {match.getAwayTeam()}: {bets.find(bet => bet.match.id() === match.id())?.predictedAwayGoals || '-'}
                    </Text>
                  </View>
                </View>
              )}
            </>
          )}
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    backgroundColor: '#25292e',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 16,
    color: '#fff',
  },
  matchCard: {
    backgroundColor: '#f5f5f5',
    padding: 16,
    borderRadius: 8,
    marginBottom: 12,
  },
  finishedMatchCard: {
    backgroundColor: '#9e9e9e',
    opacity: 0.7,
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
});
