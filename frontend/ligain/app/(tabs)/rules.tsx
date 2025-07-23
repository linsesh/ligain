import React from 'react';
import { View, Text, ScrollView, StyleSheet } from 'react-native';
import { colors } from '../../src/constants/colors';
import { useTranslation } from '../../src/hooks/useTranslation';

export default function RulesScreen() {
  const { t } = useTranslation();
  
  return (
    <ScrollView style={styles.container}>
      <View style={styles.content}>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>{t('rules.philosophy')}</Text>
            <Text style={styles.ruleDescription}>{t('rules.philosophyDescription1')}</Text>
            <Text style={styles.ruleDescription}>{t('rules.philosophyDescription2')}</Text>
            <Text style={styles.ruleDescription}>{t('rules.philosophyDescription3')}</Text>
        </View>
        
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>{t('rules.basicPoints')}</Text>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>{t('rules.exactScore')}</Text>
            <Text style={styles.example}>{t('rules.exactScoreExample')}</Text>
          </View>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>{t('rules.closeScore')}</Text>
            <Text style={styles.ruleDescription}>
              {t('rules.closeScoreDescription')}
            </Text>
            <Text style={styles.example}>{t('rules.closeScoreExample1')}</Text>
            <Text style={styles.example}>{t('rules.closeScoreExample2')}</Text>
            <Text style={styles.example}>{t('rules.closeScoreExample3')}</Text>
          </View>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>{t('rules.goodResult')}</Text>
            <Text style={styles.example}>{t('rules.goodResultExample')}</Text>
          </View>

          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>{t('rules.missedBet')}</Text>
            <Text style={styles.ruleDescription}>{t('rules.missedBetDescription')}</Text>
          </View>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>{t('rules.multipliers')}</Text>
          
          <Text style={styles.ruleDescription}>{t('rules.multipliersDescription1')}</Text>
          <Text style={styles.ruleDescription}>{t('rules.multipliersDescription2')}</Text>

          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>{t('rules.oddsDifferenceHigh')}</Text>
            <Text style={styles.ruleDescription}>{t('rules.favoriteWin')}</Text>
            <Text style={styles.ruleDescription}>{t('rules.draw')}</Text>
            <Text style={styles.ruleDescription}>{t('rules.underdogWin')}</Text>
            <Text style={styles.example}>{t('rules.oddsDifferenceHighExample')}</Text>
          </View>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>{t('rules.oddsDifferenceLow')}</Text>
            <Text style={styles.ruleDescription}>{t('rules.noMultiplier')}</Text>
            <Text style={styles.example}>{t('rules.oddsDifferenceLowExample')}</Text>
          </View>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>{t('rules.riskBonus')}</Text>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>{t('rules.riskBonus50')}</Text>
            <Text style={styles.example}>{t('rules.riskBonus50Example')}</Text>
          </View>
          
          <View style={styles.ruleItem}>
            <Text style={styles.ruleTitle}>{t('rules.riskBonus25')}</Text>
            <Text style={styles.example}>{t('rules.riskBonus25Example')}</Text>
          </View>

          <Text style={styles.ruleDescription}>{t('rules.riskBonusNonCumulative')}</Text>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>{t('rules.oddsUpdates')}</Text>
          <Text style={styles.ruleDescription}>{t('rules.oddsUpdatesDescription')}</Text>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>{t('rules.completeExample')}</Text>
          <Text style={styles.example}>
            {t('rules.completeExampleText')}
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