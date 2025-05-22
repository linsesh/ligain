import React, { useState } from 'react';
import { View, Text, StyleSheet, ActivityIndicator, TextInput } from 'react-native';
import { useMatches } from '../hooks/useMatches';

export default function TabOneScreen() {
  const { matches, loading, error } = useMatches();
  const [bets, setBets] = useState<{ [key: string]: { home: string; away: string } }>({});

  const handleBetChange = (matchId: string, team: 'home' | 'away', value: string) => {
    // Only allow numbers and max 2 digits
    if (/^\d{0,2}$/.test(value)) {
      setBets(prev => ({
        ...prev,
        [matchId]: {
          ...prev[matchId],
          [team]: value
        }
      }));
      // TODO: Send bet to server here
    }
  };

  if (loading) {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="large" color="#0000ff" />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.container}>
        <Text style={styles.errorText}>Error: {error.message}</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Matches</Text>
      {matches.map((match) => (
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
                value={match.isFinished() ? match.getHomeGoals().toString() : (bets[match.id()]?.home || '')}
                onChangeText={(value) => handleBetChange(match.id(), 'home', value)}
                keyboardType="numeric"
                maxLength={2}
                placeholder="0"
                editable={!match.isFinished()}
              />
            </View>
            <View style={styles.teamContainer}>
              <TextInput
                style={styles.betInput}
                value={match.isFinished() ? match.getAwayGoals().toString() : (bets[match.id()]?.away || '')}
                onChangeText={(value) => handleBetChange(match.id(), 'away', value)}
                keyboardType="numeric"
                maxLength={2}
                placeholder="0"
                editable={!match.isFinished()}
              />
              <Text style={[styles.teamName, styles.awayTeamName]}>{match.getAwayTeam()}</Text>
            </View>
          </View>
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
    backgroundColor: '#e0e0e0',
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
});
