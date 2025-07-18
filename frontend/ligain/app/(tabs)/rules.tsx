import React from 'react';
import { View, Text, ScrollView, StyleSheet } from 'react-native';
import { colors } from '../../src/constants/colors';

export default function RulesScreen() {
  return (
    <ScrollView style={styles.container}>
      <View style={styles.content}>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Philosophy behind the rules</Text>
            <Text style={styles.ruleDescription}>The idea is simple, being as close as possible to the real match result.</Text>
            <Text style={styles.ruleDescription}>We believe that when a result is 3-0, having predicted 2-1 is not as good as having predicted 2-0.</Text>
            <Text style={styles.ruleDescription}>We also want to reward players who took risks and bet on the underdog, or were alone in their prediction.</Text>
        </View>
        
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Basic points attribution</Text>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>• Exact score : 500 points</Text>
            <Text style={styles.example}>Example : Prediction 2-1, Match 2-1 → 500 points</Text>
          </View>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>• Close score : 400 points</Text>
            <Text style={styles.ruleDescription}>
              A score is considered close if there is the same goal difference as in the prediction AND the difference in the total goals of both teams between the prediction and the match is less than or equal to 2.
            </Text>
            <Text style={styles.example}>Example : Prediction 2-1, Match 3-2 → 400 points (same goal difference of 1, total differs by 2)</Text>
            <Text style={styles.example}>Example : Prediction 1-0, Match 2-1 → 400 points (same goal difference of 1, total differs by 2)</Text>
            <Text style={styles.example}>Example : Prediction 2-1, Match 4-3 → 300 points (same goal difference of 1, total differs by 4)</Text>
          </View>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>• Good result (neither close nor exact) : 300 points</Text>
            <Text style={styles.example}>Example : Prediction 4-1, Match 1-0 → 300 points (good result but not close)</Text>
          </View>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Multipliers based on odds</Text>
          
          <Text style={styles.ruleDescription}>The idea here is to determine if there is a team which is clearly considered as the favorite, which makes the other the underdog.</Text>
          <Text style={styles.ruleDescription}>The bonus is active only if you have predicted a draw or a win for the underdog.</Text>

          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>Odds difference {'>'} 1.5:</Text>
            <Text style={styles.ruleDescription}>• Favorite win : points × 1 (no multiplier)</Text>
            <Text style={styles.ruleDescription}>• Draw : points × 1.5</Text>
            <Text style={styles.ruleDescription}>• Underdog win : points × 2</Text>
            <Text style={styles.example}>Example : Team A at 1.3 vs Team B at 3.0 Multiplier active because the difference is 1.7</Text>
          </View>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>Odds difference {'<'} 1.5 :</Text>
            <Text style={styles.ruleDescription}>• No multiplier</Text>
            <Text style={styles.example}>Example : Team A at 2.0 vs Team B at 2.5 → No multiplier</Text>
          </View>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Risk bonus</Text>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>• 50% or less of the players have the same good result : +10%</Text>
            <Text style={styles.example}>Example : 2/4 players predict 2-1 and are correct → +10% on their points</Text>
          </View>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>• 25% or less of the players have the good result : +25%</Text>
            <Text style={styles.example}>Example : Only 1/4 players predict 2-1 and are correct → +25% on their points</Text>
          </View>

          <Text style={styles.ruleDescription}>The risks bonuses are non-cumulative.</Text>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Complete example</Text>
          <Text style={styles.example}>
            Competition between 4 players{'\n'}
            Team A at 1.3 vs Team B at 3.0{'\n'}
            Prediction : 1-1{'\n'}
            The 3 other players predict 2-1{'\n'}
            Match : 2-2{'\n'}
            Close score = 400 points {'\n'} 
            × 1.5 (draw) = 600 points{'\n'}
            Alone in the draw prediction : +25% = 750 points{'\n'}
          </Text>
        </View>
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#25292e',
  },
  content: {
    padding: 20,
  },
  title: {
    fontSize: 28,
    fontWeight: 'bold',
    color: colors.primary,
    marginBottom: 30,
    textAlign: 'center',
  },
  section: {
    marginBottom: 30,
    backgroundColor: '#2f353a',
    borderRadius: 12,
    padding: 16,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 16,
  },
  ruleItem: {
    marginBottom: 20,
  },
  ruleTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: colors.primary,
    marginBottom: 8,
  },
  ruleDescription: {
    fontSize: 14,
    color: '#e0e0e0',
    marginBottom: 8,
    lineHeight: 20,
  },
  example: {
    fontSize: 14,
    color: '#b0b0b0',
    fontStyle: 'italic',
    marginTop: 4,
    lineHeight: 18,
  },
}); 