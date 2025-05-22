import React from 'react';
import { View, Text, StyleSheet, ActivityIndicator } from 'react-native';
import { useMatches } from '../hooks/useMatches';

export default function TabOneScreen() {
  const { matches, loading, error } = useMatches();

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
        <View key={match.id()} style={styles.matchCard}>
          <Text style={styles.matchTitle}>
            {match.getHomeTeam()} vs {match.getAwayTeam()}
          </Text>
          {match.isFinished() ? (
            <Text style={styles.score}>
              {match.getHomeGoals()} - {match.getAwayGoals()}
            </Text>
          ) : (
            <Text style={styles.date}>
              {new Date(match.getDate()).toLocaleDateString()}
            </Text>
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
    backgroundColor: '#fff',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  matchCard: {
    backgroundColor: '#f5f5f5',
    padding: 16,
    borderRadius: 8,
    marginBottom: 12,
  },
  matchTitle: {
    fontSize: 18,
    fontWeight: '600',
  },
  score: {
    fontSize: 16,
    color: '#666',
    marginTop: 8,
  },
  date: {
    fontSize: 14,
    color: '#666',
    marginTop: 8,
  },
  errorText: {
    color: 'red',
    fontSize: 16,
  },
});
