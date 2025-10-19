import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Modal,
  useWindowDimensions,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors } from '../constants/colors';
import { useTranslation } from '../hooks/useTranslation';
import { SyncOpportunity } from '../../hooks/useBetSynchronization';

interface BetSyncModalProps {
  visible: boolean;
  syncOpportunity: SyncOpportunity | null;
  onSynchronize: () => void;
  onNotNow: () => void;
  loading?: boolean;
}

export const BetSyncModal: React.FC<BetSyncModalProps> = ({
  visible,
  syncOpportunity,
  onSynchronize,
  onNotNow,
  loading = false,
}) => {
  const { width } = useWindowDimensions();
  const { t } = useTranslation();

  if (!syncOpportunity) return null;

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
            <Ionicons name="sync" size={40} color={colors.primary} style={styles.icon} />
            <Text style={[styles.title, { color: colors.text }]}>
              {t('betSync.title')}
            </Text>
          </View>

          {/* Content */}
          <View style={styles.content}>
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
          </View>

          {/* Buttons */}
          <View style={styles.buttons}>
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
});
