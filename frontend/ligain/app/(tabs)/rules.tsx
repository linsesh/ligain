import React, { useState } from 'react';
import { View, ScrollView, useWindowDimensions } from 'react-native';
import { Text } from '../../src/components/ui/Text';
import { useTranslation } from '../../src/hooks/useTranslation';

const ACCENTS = ['#f25702', '#e9a317', '#469dff', '#2e7d32', '#f25702', '#e9a317', '#f25702'];

function RuleItem({ title, accent, children }: { title: string; accent: string; children?: React.ReactNode }) {
  return (
    <View className="mb-5">
      <View style={{ borderLeftWidth: 3, borderLeftColor: accent }} className="pl-3 mb-2">
        <Text className="text-base font-hk-semibold text-foreground">{title}</Text>
      </View>
      {children}
    </View>
  );
}

export default function RulesScreen() {
  const { t } = useTranslation();
  const { width } = useWindowDimensions();
  const [currentPage, setCurrentPage] = useState(0);

  const sections = [
    {
      titleKey: 'rules.philosophy',
      content: (accent: string) => (
        <View>
          <Text className="text-base text-foreground leading-6 mb-3">{t('rules.philosophyDescription1')}</Text>
          <Text className="text-base text-foreground leading-6">{t('rules.philosophyDescription2')}</Text>
        </View>
      ),
    },
    {
      titleKey: 'rules.basicPoints',
      content: (accent: string) => (
        <View>
          <RuleItem title={t('rules.exactScore')} accent={accent} />
          <RuleItem title={t('rules.closeScore')} accent={accent}>
            <Text className="text-base text-foreground leading-6 mb-3">{t('rules.closeScoreDescription')}</Text>
            <Text className="text-base text-foreground-secondary leading-6">{t('rules.closeScoreExample1')}</Text>
            <Text className="text-base text-foreground-secondary leading-6">{t('rules.closeScoreExample2')}</Text>
            <Text className="text-base text-foreground-secondary leading-6">{t('rules.closeScoreExample3')}</Text>
          </RuleItem>
          <RuleItem title={t('rules.goodResult')} accent={accent}>
            <Text className="text-base text-foreground-secondary leading-6">{t('rules.goodResultExample')}</Text>
          </RuleItem>
          <RuleItem title={t('rules.missedBet')} accent={accent}>
            <Text className="text-base text-foreground leading-6">{t('rules.missedBetDescription')}</Text>
          </RuleItem>
        </View>
      ),
    },
    {
      titleKey: 'rules.multipliers',
      content: (accent: string) => (
        <View>
          <Text className="text-base text-foreground leading-6 mb-3">{t('rules.multipliersDescription1')}</Text>
          <Text className="text-base text-foreground leading-6 mb-5">{t('rules.multipliersDescription2')}</Text>
          <RuleItem title={t('rules.oddsDifferenceHigh')} accent={accent}>
            <Text className="text-base text-foreground leading-6 mb-2">{t('rules.favoriteWin')}</Text>
            <Text className="text-base text-foreground leading-6 mb-2">{t('rules.draw')}</Text>
            <Text className="text-base text-foreground leading-6 mb-3">{t('rules.underdogWin')}</Text>
            <Text className="text-base text-foreground-secondary leading-6">{t('rules.oddsDifferenceHighExample')}</Text>
          </RuleItem>
          <RuleItem title={t('rules.oddsDifferenceLow')} accent={accent}>
            <Text className="text-base text-foreground leading-6 mb-3">{t('rules.noMultiplier')}</Text>
            <Text className="text-base text-foreground-secondary leading-6">{t('rules.oddsDifferenceLowExample')}</Text>
          </RuleItem>
        </View>
      ),
    },
    {
      titleKey: 'rules.riskBonus',
      content: (accent: string) => (
        <View>
          <RuleItem title={t('rules.riskBonus50')} accent={accent}>
            <Text className="text-base text-foreground-secondary leading-6">{t('rules.riskBonus50Example')}</Text>
          </RuleItem>
          <RuleItem title={t('rules.riskBonus25')} accent={accent}>
            <Text className="text-base text-foreground-secondary leading-6">{t('rules.riskBonus25Example')}</Text>
          </RuleItem>
          <Text className="text-base text-foreground leading-6">{t('rules.riskBonusNonCumulative')}</Text>
        </View>
      ),
    },
    {
      titleKey: 'rules.oddsUpdates',
      content: (accent: string) => (
        <Text className="text-base text-foreground leading-6">{t('rules.oddsUpdatesDescription')}</Text>
      ),
    },
    {
      titleKey: 'rules.noBet',
      content: (accent: string) => (
        <Text className="text-base text-foreground leading-6">{t('rules.noBetDescription')}</Text>
      ),
    },
    {
      titleKey: 'rules.completeExample',
      content: (accent: string) => (
        <Text className="text-base text-foreground-secondary leading-6">{t('rules.completeExampleText')}</Text>
      ),
    },
  ];

  return (
    <View className="flex-1" style={{ backgroundColor: 'transparent' }}>
      <ScrollView
        horizontal
        pagingEnabled
        showsHorizontalScrollIndicator={false}
        onScroll={(e) => setCurrentPage(Math.round(e.nativeEvent.contentOffset.x / width))}
        scrollEventThrottle={32}
        className="flex-1"
      >
        {sections.map((section, i) => {
          const accent = ACCENTS[i];
          return (
            <View key={i} style={{ width }} className="flex-1 px-4 pt-2 pb-4">
              <View className="flex-1 rounded-2xl bg-background overflow-hidden">
                {/* Coloured header */}
                <View style={{ backgroundColor: accent }} className="px-5 py-6">
                  <Text className="text-white text-xs font-hk-semibold tracking-widest uppercase opacity-80 mb-1">
                    {i + 1} / {sections.length}
                  </Text>
                  <Text className="text-white text-2xl font-hk-bold leading-8">
                    {t(section.titleKey)}
                  </Text>
                </View>

                {/* Content */}
                <ScrollView
                  className="flex-1"
                  contentContainerClassName="p-5"
                  showsVerticalScrollIndicator={false}
                >
                  {section.content(accent)}
                </ScrollView>
              </View>
            </View>
          );
        })}
      </ScrollView>

      {/* Page dots */}
      <View className="flex-row justify-center items-center gap-2 py-4">
        {sections.map((_, i) => (
          <View
            key={i}
            style={{
              backgroundColor: i === currentPage ? ACCENTS[i] : '#ccd2d7',
              width: i === currentPage ? 16 : 8,
              height: 8,
            }}
            className="rounded-full"
          />
        ))}
      </View>
    </View>
  );
}
