import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Modal,
  useWindowDimensions,
  ScrollView,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../constants/colors';
import { useTranslation } from '../hooks/useTranslation';
import { SyncOpportunity } from '../../hooks/useBetSynchronization';

export interface SyncResult {
  successful: Array<{
    matchId: string;
    homeTeam: string;
    awayTeam: string;
    matchday: number;
    predictedHomeGoals: number;
    predictedAwayGoals: number;
  }>;
  failed: Array<{
    match: {
      matchId: string;
      homeTeam: string;
      awayTeam: string;
      matchday: number;
      predictedHomeGoals: number;
      predictedAwayGoals: number;
    };
    error: any;
  }>;
}

interface BetSyncModalProps {
  visible: boolean;
  syncOpportunity: SyncOpportunity | null;
  onSynchronize: () => void;
  onNotNow: () => void;
  onRetryFailed?: () => void;
  loading?: boolean;
  syncResult?: SyncResult | null;
  mode?: 'initial' | 'success' | 'partialSuccess' | 'failure';
}

export const BetSyncModal: React.FC<BetSyncModalProps> = ({
  visible,
  syncOpportunity,
  onSynchronize,
  onNotNow,
  onRetryFailed,
  loading = false,
  syncResult,
  mode = 'initial',
}) => {
  const { width } = useWindowDimensions();
  const { t } = useTranslation();

  if (!syncOpportunity) return null;

  const renderInitialContent = () => (
    <>
      <Text style={[styles.message, { color: colors.text }]}>
        {t('betSync.message', { 
          count: syncOpportunity.matchesToSync.length,
          gameName: syncOpportunity.sourceGameName 
        })}
      </Text>
      
      <Text style={[styles.subtitle, { color: colors.textSecondary }]}>
        {t('betSync.onlyUnbetted')}
      </Text>

      {/* Match preview */}
      <View style={styles.matchesPreview}>
        <Text style={[styles.matchesTitle, { color: colors.text }]}>
          {t('betSync.matchesToSync', { count: syncOpportunity.matchesToSync.length })}
        </Text>
        {(() => {
          // Group matches by matchday
          const matchesByMatchday = syncOpportunity.matchesToSync.reduce((acc, match) => {
            if (!acc[match.matchday]) {
              acc[match.matchday] = 0;
            }
            acc[match.matchday]++;
            return acc;
          }, {} as Record<number, number>);

          return Object.entries(matchesByMatchday)
            .sort(([a], [b]) => parseInt(a) - parseInt(b))
            .map(([matchday, count]) => (
              <Text key={matchday} style={[styles.matchItem, { color: colors.textSecondary }]}>
                {t('betSync.matchday', { matchday: parseInt(matchday), count })}
              </Text>
            ));
        })()}
      </View>
    </>
  );

  const renderSuccessContent = () => (
    <Text style={[styles.message, { color: colors.text }]}>
      {t('betSync.success.message', { count: syncResult?.successful.length || 0 })}
    </Text>
  );

  const renderPartialSuccessContent = () => (
    <ScrollView style={styles.scrollContent} showsVerticalScrollIndicator={false}>
      <Text style={[styles.message, { color: colors.text }]}>
        {t('betSync.partialSuccess.message', { 
          successful: syncResult?.successful.length || 0,
          total: (syncResult?.successful.length || 0) + (syncResult?.failed.length || 0)
        })}
      </Text>

      {syncResult && syncResult.successful.length > 0 && (
        <View style={styles.matchesPreview}>
          <Text style={[styles.matchesTitle, { color: colors.text }]}>
            {t('betSync.partialSuccess.successfulMatches')}
          </Text>
          {syncResult.successful.map((match, index) => (
            <Text key={index} style={[styles.matchItem, { color: colors.success || '#4CAF50' }]}>
              {match.homeTeam} vs {match.awayTeam} ({match.predictedHomeGoals}-{match.predictedAwayGoals})
            </Text>
          ))}
        </View>
      )}

      {syncResult && syncResult.failed.length > 0 && (
        <View style={[styles.matchesPreview, { backgroundColor: '#ffebee' }]}>
          <Text style={[styles.matchesTitle, { color: colors.text }]}>
            {t('betSync.partialSuccess.failedMatches')}
          </Text>
          {syncResult.failed.map((failedMatch, index) => (
            <Text key={index} style={[styles.matchItem, { color: colors.error || '#f44336' }]}>
              {failedMatch.match.homeTeam} vs {failedMatch.match.awayTeam} ({failedMatch.match.predictedHomeGoals}-{failedMatch.match.predictedAwayGoals})
            </Text>
          ))}
        </View>
      )}
    </ScrollView>
  );

  const renderFailureContent = () => (
    <Text style={[styles.message, { color: colors.text }]}>
      {t('betSync.failure.message')}
    </Text>
  );

  const renderContent = () => {
    switch (mode) {
      case 'success':
        return renderSuccessContent();
      case 'partialSuccess':
        return renderPartialSuccessContent();
      case 'failure':
        return renderFailureContent();
      default:
        return renderInitialContent();
    }
  };

  const renderButtons = () => {
    switch (mode) {
      case 'success':
        return (
          <TouchableOpacity
            style={[styles.button, styles.singleButton, { backgroundColor: colors.primary }]}
            onPress={onNotNow}
          >
            <Text style={styles.synchronizeButtonText}>
              {t('betSync.success.close')}
            </Text>
          </TouchableOpacity>
        );
      
      case 'partialSuccess':
        return (
          <>
            <TouchableOpacity
              style={[styles.button, styles.notNowButton]}
              onPress={onNotNow}
            >
              <Text style={[styles.notNowButtonText, { color: colors.text }]}>
                {t('betSync.partialSuccess.close')}
              </Text>
            </TouchableOpacity>
            
            {onRetryFailed && syncResult && syncResult.failed.length > 0 && (
              <TouchableOpacity
                style={[styles.button, styles.synchronizeButton, { backgroundColor: colors.primary }]}
                onPress={onRetryFailed}
              >
                <Text style={styles.synchronizeButtonText}>
                  {t('betSync.partialSuccess.retryFailed')}
                </Text>
              </TouchableOpacity>
            )}
          </>
        );
      
      case 'failure':
        return (
          <TouchableOpacity
            style={[styles.button, styles.singleButton, { backgroundColor: colors.primary }]}
            onPress={onNotNow}
          >
            <Text style={styles.synchronizeButtonText}>
              {t('betSync.failure.close')}
            </Text>
          </TouchableOpacity>
        );
      
      default:
        return (
          <>
            <TouchableOpacity
              style={[styles.button, styles.notNowButton]}
              onPress={onNotNow}
              disabled={loading}
            >
              <Text style={[styles.notNowButtonText, { color: colors.text }]}>
                {t('betSync.notNow')}
              </Text>
            </TouchableOpacity>
            
            <TouchableOpacity
              style={[styles.button, styles.synchronizeButton, { backgroundColor: colors.primary }]}
              onPress={onSynchronize}
              disabled={loading}
            >
              {loading ? (
                <Text style={styles.synchronizeButtonText}>
                  {t('common.loading')}
                </Text>
              ) : (
                <Text style={styles.synchronizeButtonText}>
                  {t('betSync.synchronize')}
                </Text>
              )}
            </TouchableOpacity>
          </>
        );
    }
  };

  const getTitle = () => {
    switch (mode) {
      case 'success':
        return t('betSync.success.title');
      case 'partialSuccess':
        return t('betSync.partialSuccess.title');
      case 'failure':
        return t('betSync.failure.title');
      default:
        return t('betSync.title');
    }
  };

  const getIcon = () => {
    switch (mode) {
      case 'success':
        return 'checkmark-circle';
      case 'partialSuccess':
        return 'warning';
      case 'failure':
        return 'close-circle';
      default:
        return 'sync';
    }
  };

  const getIconColor = () => {
    switch (mode) {
      case 'success':
        return colors.success || '#4CAF50';
      case 'partialSuccess':
        return colors.warning || '#FF9800';
      case 'failure':
        return colors.error || '#f44336';
      default:
        return colors.primary;
    }
  };

  return (
    <Modal
      visible={visible}
      animationType="slide"
      transparent={true}
      onRequestClose={onNotNow}
    >
      <View style={styles.modalOverlay}>
        <View style={[styles.modalContent, { width: width - 40 }]}>
          {/* Header */}
          <View style={styles.header}>
            <Ionicons name={getIcon()} size={40} color={getIconColor()} style={styles.icon} />
            <Text style={[styles.title, { color: colors.text }]}>
              {getTitle()}
            </Text>
          </View>

          {/* Content */}
          <View style={styles.content}>
            {renderContent()}
          </View>

          {/* Buttons */}
          <View style={styles.buttons}>
            {renderButtons()}
          </View>
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  modalContent: {
    backgroundColor: colors.card,
    borderRadius: 16,
    overflow: 'hidden',
    maxHeight: '80%',
  },
  header: {
    alignItems: 'center',
    padding: 24,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  icon: {
    marginBottom: 12,
  },
  title: {
    fontSize: 20,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  content: {
    padding: 24,
  },
  message: {
    fontSize: 16,
    lineHeight: 24,
    textAlign: 'center',
    marginBottom: 16,
  },
  subtitle: {
    fontSize: 14,
    lineHeight: 20,
    textAlign: 'center',
    marginBottom: 20,
    fontStyle: 'italic',
  },
  matchesPreview: {
    backgroundColor: colors.background,
    borderRadius: 8,
    padding: 16,
  },
  matchesTitle: {
    fontSize: 14,
    fontWeight: '600',
    marginBottom: 8,
  },
  matchItem: {
    fontSize: 13,
    lineHeight: 18,
    marginBottom: 4,
  },
  buttons: {
    flexDirection: 'row',
    padding: 24,
    borderTopWidth: 1,
    borderTopColor: colors.border,
    gap: 12,
  },
  button: {
    flex: 1,
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
  notNowButton: {
    backgroundColor: colors.background,
    borderWidth: 1,
    borderColor: colors.border,
  },
  notNowButtonText: {
    fontSize: 14,
    fontWeight: '600',
    textAlign: 'center',
  },
  synchronizeButton: {
    // backgroundColor set via style prop
  },
  synchronizeButtonText: {
    color: '#FFFFFF',
    fontSize: 14,
    fontWeight: '600',
    textAlign: 'center',
  },
  scrollContent: {
    maxHeight: 200,
  },
  singleButton: {
    flex: 1,
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
});
